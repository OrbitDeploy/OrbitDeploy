package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/OrbitDeploy/OrbitDeploy/models"
	"github.com/OrbitDeploy/fastcaddy"
	"github.com/google/uuid"
	"github.com/opentdp/go-helper/logman"
	"gorm.io/gorm"
)

// ApplicationService 处理应用相关的业务逻辑
type ApplicationService struct {
	db            *gorm.DB
	podmanService *PodmanService
}

// NewApplicationService 创建新的应用服务实例
// 依赖从外部传入，不在内部创建 (原则一：依赖外置，不内建)
func NewApplicationService(db *gorm.DB, podmanService *PodmanService) *ApplicationService {
	return &ApplicationService{
		db:            db,
		podmanService: podmanService,
	}
}

// DeleteApplicationWithCleanup 删除应用及其所有相关资源
func (as *ApplicationService) DeleteApplicationWithCleanup(appID uuid.UUID, appName string) error {
	// 开始事务
	tx := as.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("开始事务失败: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取应用信息
	app, err := as.getApplicationByID(appID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("获取应用失败: %w", err)
	}

	// 验证应用名称
	if app.Name != appName {
		tx.Rollback()
		return fmt.Errorf("应用名称不匹配")
	}

	logman.Info("开始删除应用", "app_id", appID, "app_name", appName)

	// 1. 停止并删除相关的容器和服务
	if err := as.stopAndRemoveContainers(app.Name); err != nil {
		logman.Warn("停止容器失败", "app_name", app.Name, "error", err)
		// 不中断流程，继续删除其他资源
	}

	// 2. 获取所有相关的 Release，用于清理镜像
	var releases []models.Release
	if err := tx.Where("application_id = ?", appID).Find(&releases).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("获取应用的 Release 失败: %w", err)
	}

	// 获取所有相关的 Routing，用于清理域名
	var routings []models.Routing
	if err := tx.Where("application_id = ? AND is_active = ?", appID, true).Find(&routings).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("获取应用的 Routing 失败: %w", err)
	}

	// 3. 删除相关的数据库记录（按依赖顺序）

	// 删除 Deployments
	if err := tx.Where("application_id = ?", appID).Delete(&models.Deployment{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除 Deployment 记录失败: %w", err)
	}
	logman.Info("已删除 Deployment 记录", "app_id", appID)

	// 删除 Routings
	if err := tx.Where("application_id = ?", appID).Delete(&models.Routing{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除 Routing 记录失败: %w", err)
	}
	logman.Info("已删除 Routing 记录", "app_id", appID)

	// 删除 Environment Variables
	if err := tx.Where("application_id = ?", appID).Delete(&models.EnvironmentVariable{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除 EnvironmentVariable 记录失败: %w", err)
	}
	logman.Info("已删除 EnvironmentVariable 记录", "app_id", appID)

	// 删除 Releases
	if err := tx.Where("application_id = ?", appID).Delete(&models.Release{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除 Release 记录失败: %w", err)
	}
	logman.Info("已删除 Release 记录", "app_id", appID)

	// 删除 Application
	if err := tx.Where("id = ?", appID).Delete(&models.Application{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除 Application 记录失败: %w", err)
	}
	logman.Info("已删除 Application 记录", "app_id", appID)

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	// 4. 清理Docker镜像（在事务外进行，避免影响数据库操作）
	for _, release := range releases {
		if err := as.removeDockerImage(release.ImageName); err != nil {
			logman.Warn("删除Docker镜像失败", "image", release.ImageName, "error", err)
			// 不中断流程，继续删除其他镜像
		}
	}

	// 清理域名路由（在事务外进行） 暂时不清理
	// as.cleanupDomainRoutes(routings)

	// 5. 清理本地文件
	if err := as.cleanupLocalFiles(app.Name); err != nil {
		logman.Warn("清理本地文件失败", "app_name", app.Name, "error", err)
		// 不中断流程
	}

	logman.Info("应用删除完成", "app_id", appID, "app_name", appName)
	return nil
}

// stopAndRemoveContainers 停止并删除容器和相关服务
func (as *ApplicationService) stopAndRemoveContainers(ServiceName string) error {
	logman.Info("开始停止和删除容器", "service", ServiceName)

	// --- 1. 停止并禁用 systemd 服务 ---
	logman.Info("正在停止服务...", "service", ServiceName)
	if err := as.runSystemctlCommand("stop", ServiceName, true); err != nil {
		logman.Warn("停止 systemd 服务失败，将继续尝试清理", "service", ServiceName, "error", err)
	} else {
		logman.Info("服务已停止", "service", ServiceName)
	}

	logman.Info("正在禁用服务自启...", "service", ServiceName)
	if err := as.runSystemctlCommand("disable", ServiceName, true); err != nil {
		logman.Warn("禁用 systemd 服务失败", "service", ServiceName, "error", err)
	} else {
		logman.Info("服务已禁用", "service", ServiceName)
	}

	// --- 2. 删除 systemd 服务文件 ---
	homeDir, _ := os.UserHomeDir()
	serviceFilePath := filepath.Join(homeDir, ".config", "systemd", "user", ServiceName+".service")
	quadletFilePath := filepath.Join("/usr/share", "containers", "systemd", ServiceName+".container")
	_ = as.removeFileIfExists(quadletFilePath) // 同时尝试清理 quadlet 文件
	if err := as.removeFileIfExists(serviceFilePath); err != nil {
		logman.Warn("删除 service 文件失败", "file", serviceFilePath, "error", err)
	} else {
		logman.Info("已删除 service 文件", "file", serviceFilePath)
	}

	// --- 3. 重新加载 systemd 配置 ---
	logman.Info("正在重新加载 systemd 配置...")
	if err := as.runSystemctlCommand("daemon-reload", "", true); err != nil {
		logman.Warn("重新加载 systemd 配置失败", "error", err)
	}

	// --- 4. 停止并删除 Podman 容器 ---
	stopContainerCmd := exec.Command("podman", "stop", ServiceName)
	if output, err := stopContainerCmd.CombinedOutput(); err != nil {
		logman.Warn("停止容器失败", "service", ServiceName, "output", string(output))
	} else {
		logman.Info("已停止容器", "service", ServiceName)
	}

	removeContainerCmd := exec.Command("podman", "rm", ServiceName)
	if output, err := removeContainerCmd.CombinedOutput(); err != nil {
		logman.Warn("删除容器失败", "service", ServiceName, "output", string(output))
	} else {
		logman.Info("已删除容器", "service", ServiceName)
	}

	logman.Info("容器和相关服务清理完成", "service", ServiceName)
	return nil
}

// runSystemctlCommand 是一个辅助函数，用于执行 systemctl 命令
func (as *ApplicationService) runSystemctlCommand(command, serviceUnitName string, isUser bool) error {
	args := []string{}
	if isUser {
		args = append(args, "--user")
	}
	args = append(args, command)
	if serviceUnitName != "" {
		args = append(args, serviceUnitName)
	}

	cmd := exec.Command("systemctl", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("执行 '%s' 失败: %s, output: %s", cmd.String(), err, string(output))
	}
	return nil
}

// removeDockerImage 删除Docker镜像
func (as *ApplicationService) removeDockerImage(imageName string) error {
	if imageName == "" {
		return nil
	}

	// 检查镜像是否存在
	if !as.podmanService.CheckImageExists(imageName) {
		logman.Info("镜像不存在，跳过删除", "image", imageName)
		return nil
	}

	removeCmd := exec.Command("podman", "rmi", imageName)
	if output, err := removeCmd.CombinedOutput(); err != nil {
		logman.Warn("删除镜像失败", "image", imageName, "output", string(output))
		return fmt.Errorf("删除镜像失败: %s", string(output))
	}

	logman.Info("已删除镜像", "image", imageName)
	return nil
}

// cleanupLocalFiles 清理本地文件
func (as *ApplicationService) cleanupLocalFiles(appName string) error {
	// 定义可能需要清理的文件路径
	filesToClean := []string{
		// systemd用户服务文件
		filepath.Join(os.Getenv("HOME"), ".config", "systemd", "user", appName+".service"),
		// quadlet文件
		filepath.Join("/usr/share", "containers", "systemd", appName+".container"),

		// 可能的日志文件
		filepath.Join("/tmp", appName+".log"),
		filepath.Join("/var/log", appName+".log"),
	}

	for _, filePath := range filesToClean {
		if err := as.removeFileIfExists(filePath); err != nil {
			logman.Warn("删除文件失败", "file", filePath, "error", err)
		}
	}

	// 清理可能的配置目录
	configDirs := []string{
		filepath.Join(os.Getenv("HOME"), ".config", "orbitdeploy", appName),
		filepath.Join("/tmp", "orbitdeploy", appName),
	}

	for _, dirPath := range configDirs {
		if err := as.removeDirIfExists(dirPath); err != nil {
			logman.Warn("删除目录失败", "dir", dirPath, "error", err)
		}
	}

	logman.Info("本地文件清理完成", "app", appName)
	return nil
}

// removeFileIfExists 删除文件（如果存在）
func (as *ApplicationService) removeFileIfExists(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // 文件不存在，无需删除
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("删除文件失败 %s: %w", filePath, err)
	}

	logman.Info("已删除文件", "file", filePath)
	return nil
}

// removeDirIfExists 删除目录（如果存在）
func (as *ApplicationService) removeDirIfExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil // 目录不存在，无需删除
	}

	if err := os.RemoveAll(dirPath); err != nil {
		return fmt.Errorf("删除目录失败 %s: %w", dirPath, err)
	}

	logman.Info("已删除目录", "dir", dirPath)
	return nil
}

// cleanupDomainRoutes 清理域名路由
func (as *ApplicationService) cleanupDomainRoutes(routings []models.Routing) {
	fc := fastcaddy.New()
	for _, routing := range routings {
		if err := fc.DeleteRoute(routing.DomainName); err != nil {
			logman.Warn("删除域名路由失败", "domain", routing.DomainName, "error", err)
			// 不中断流程，继续删除其他路由
		} else {
			logman.Info("已删除域名路由", "domain", routing.DomainName)
		}
	}
}

// getApplicationByID 内部方法，使用注入的db获取应用信息
func (as *ApplicationService) getApplicationByID(id uuid.UUID) (*models.Application, error) {
	var application models.Application
	if err := as.db.Where("id = ?", id).First(&application).Error; err != nil {
		return nil, err
	}
	return &application, nil
}

// ValidateApplicationDeletion 验证应用是否可以被删除
func (as *ApplicationService) ValidateApplicationDeletion(appID uuid.UUID, appName string) error {
	app, err := as.getApplicationByID(appID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("应用不存在")
		}
		return fmt.Errorf("获取应用失败: %w", err)
	}

	if app.Name != appName {
		return fmt.Errorf("应用名称不匹配")
	}

	return nil
}

// GetRunningDeploymentsByAppID 获取应用实际运行中的部署（通过 Podman 检查）
func (as *ApplicationService) GetRunningDeploymentsByAppID(appID uuid.UUID) ([]*models.Deployment, error) {
	// 获取所有部署（预加载 Release）
	var deployments []*models.Deployment
	if err := as.db.Where("application_id = ?", appID).Preload("Release").Find(&deployments).Error; err != nil {
		return nil, err
	}

	// 过滤实际运行的部署
	var runningDeployments []*models.Deployment
	for _, deployment := range deployments {
		if as.podmanService.CheckContainerRunningWithQuadlet(deployment.ServiceName) {
			runningDeployments = append(runningDeployments, deployment)
		}
	}

	return runningDeployments, nil
}

// cleanupApplicationData 仅清理应用自身相关的数据目录
func (as *ApplicationService) cleanupApplicationData(appName string) error {
	configDirs := []string{
		filepath.Join(os.Getenv("HOME"), ".config", "orbitdeploy", appName),
		filepath.Join("/tmp", "orbitdeploy", appName),
	}
	for _, dirPath := range configDirs {
		if err := as.removeDirIfExists(dirPath); err != nil {
			logman.Warn("删除目录失败", "dir", dirPath, "error", err)
		}
	}
	logman.Info("应用数据清理完成", "app", appName)
	return nil
}
