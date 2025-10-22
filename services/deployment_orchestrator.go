package services

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/OrbitDeploy/OrbitDeploy/models"
	"github.com/OrbitDeploy/OrbitDeploy/utils"
	"github.com/google/uuid"
	"github.com/opentdp/go-helper/command"
	"github.com/opentdp/go-helper/dborm"
	"github.com/opentdp/go-helper/logman"
)

// SSELogSender SSE日志发送函数的接口，避免循环依赖
type SSELogSender func(deploymentID uuid.UUID, message string)

// DeploymentOrchestrator 部署编排服务，负责协调整个部署流程
type DeploymentOrchestrator struct {
	buildService  *BuildService
	envService    *DeploymentEnvironmentService
	podmanService *PodmanService
	sseLogSender  SSELogSender // SSE日志发送函数
}

// NewDeploymentOrchestrator 创建新的部署编排服务实例
func NewDeploymentOrchestrator(buildService *BuildService, envService *DeploymentEnvironmentService, podmanService *PodmanService) *DeploymentOrchestrator {
	return &DeploymentOrchestrator{
		buildService:  buildService,
		envService:    envService,
		podmanService: podmanService,
		sseLogSender:  nil, // 将在后续设置
	}
}

// SetSSELogSender 设置SSE日志发送函数（避免循环依赖）
func (do *DeploymentOrchestrator) SetSSELogSender(sender SSELogSender) {
	do.sseLogSender = sender
}

// sendDeploymentLog 发送部署日志到数据库和SSE客户端
func (do *DeploymentOrchestrator) sendDeploymentLog(deploymentID uuid.UUID, message string) {
	// 发送到SSE客户端（实时）
	if do.sseLogSender != nil {
		do.sseLogSender(deploymentID, message)
	}

	// 保存到数据库
	do.AppendLogToDB(deploymentID, message, "INFO", "SYSTEM")

	// 记录到日志
	logman.Info("部署日志", "deployment_id", deploymentID, "message", message)
}

// AppendLogToDB 将日志追加到数据库
func (do *DeploymentOrchestrator) AppendLogToDB(deploymentID uuid.UUID, message string, level string, source string) error {
	logEntry := &models.DeploymentLog{
		DeploymentID: deploymentID,
		Timestamp:    time.Now(),
		Message:      message,
		Level:        level,
		Source:       source,
	}
	return models.CreateDeploymentLog(logEntry)
}

// updateDeploymentLogInDB 更新数据库中的部署日志（已废弃，保留用于向后兼容）
// 推荐使用 AppendLogToDB 替代
func (do *DeploymentOrchestrator) updateDeploymentLogInDB(deploymentID uuid.UUID, additionalLog string) error {
	// 使用新的 AppendLogToDB 方法
	return do.AppendLogToDB(deploymentID, additionalLog, "INFO", "SYSTEM")
}

// CreateDeploymentRequest 创建部署请求结构
type CreateDeploymentRequest struct {
	ReleaseID *uuid.UUID `json:"releaseId"` // 如果为nil，表示需要重新构建
}

// CreateDeployment 创建部署并启动异步部署流程
func (do *DeploymentOrchestrator) CreateDeployment(appID uuid.UUID, req CreateDeploymentRequest) (*models.Deployment, error) {
	logman.Info("开始创建部署", "app_id", appID)

	// 1. 获取应用信息
	application, err := models.GetApplicationByID(appID)
	if err != nil {
		logman.Error("获取应用信息失败", "app_id", appID, "error", err)
		return nil, fmt.Errorf("获取应用信息失败: %w", err)
	}
	logman.Info("获取应用信息成功", "app_name", application.Name)

	// 2. 确定要部署的 Release
	var releaseID uuid.UUID
	var needsBuild bool

	if req.ReleaseID != nil {
		// 使用指定的 Release
		releaseID = *req.ReleaseID
		needsBuild = false
		logman.Info("使用指定的发布版本", "app_id", appID, "release_id", releaseID)
	} else {
		// 需要重新构建版本 - 只创建 Release 记录，不执行构建
		logman.Info("创建新的构建版本记录", "app_id", appID)
		release, err := do.createNewReleaseRecord(application)
		if err != nil {
			logman.Error("创建新版本记录失败", "app_id", appID, "error", err)
			return nil, fmt.Errorf("创建新版本记录失败: %w", err)
		}
		releaseID = release.ID
		needsBuild = true
		logman.Info("新版本记录创建成功，准备异步构建", "release_id", releaseID)
	}

	// 3. 创建 Deployment 记录
	initialLogText := "开始部署流程...\n"
	if needsBuild {
		initialLogText = "开始构建 Release...\n"
	}

	version := time.Now().Format("060102150405")
	serviceName := application.Name + "-" + version + ".service"

	// Update the Release with the generated version
	if err := models.UpdateReleaseVersion(releaseID, version); err != nil {
		logman.Error("更新Release版本失败", "release_id", releaseID, "error", err)
		return nil, fmt.Errorf("更新Release版本失败: %w", err)
	}

	deployment, err := models.CreateDeployment(
		appID,
		releaseID,
		"in_progress",
		initialLogText,
		serviceName,
		time.Now(),
		nil,
	)
	if err != nil {
		logman.Error("创建部署记录失败", "app_id", appID, "error", err)
		return nil, fmt.Errorf("创建部署记录失败: %w", err)
	}
	logman.Info("部署记录创建成功", "deployment_id", deployment.ID)

	// 4. 启动异步构建+部署流程
	if needsBuild {
		logman.Info("启动异步构建+部署流程", "deployment_id", deployment.ID)
		go do.startBuildAndDeploymentAsync(deployment.ID)
	} else {
		logman.Info("启动异步部署流程", "deployment_id", deployment.ID)
		go do.startDeploymentAsync(deployment.ID)
	}

	return deployment, nil
}

// createNewReleaseRecord 创建新的发布版本记录（不执行构建）
func (do *DeploymentOrchestrator) createNewReleaseRecord(application *models.Application) (*models.Release, error) {
	// 1. 获取项目信息
	_, err := models.GetProjectByID(application.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("获取项目信息失败: %w", err)
	}

	// 2. 创建 Release 记录（初始状态为 building）
	buildSourceInfo := models.JSONB{
		Data: map[string]interface{}{
			"branch": func() string {
				if application.Branch != nil && *application.Branch != "" {
					return *application.Branch
				}
				return "master" // 默认分支
			}(),
			"commit_sha": "latest", // 这里可以后续改进为获取实际的 commit SHA
		},
	}

	release, err := models.CreateRelease(
		application.ID,
		"wait-for-build-"+fmt.Sprintf("%d", time.Now().Unix()), // 镜像名称将在构建完成后更新
		buildSourceInfo,
		"building",
	)
	if err != nil {
		return nil, fmt.Errorf("创建 Release 记录失败: %w", err)
	}

	logman.Info("Release 记录创建成功", "release_id", release.ID, "status", "building")
	return release, nil
}

// buildRelease 执行指定 Release 的构建过程
func (do *DeploymentOrchestrator) buildRelease(release *models.Release, application *models.Application) error {
	// 1. 执行构建
	buildReq := BuildFromApplicationRequest{
		ApplicationID: application.ID,
		Dockerfile:    "Dockerfile",
		ContextPath:   ".",
		BuildArgs:     make(map[string]string),
	}

	imageName, err := do.buildService.BuildImageFromApplication(buildReq)
	if err != nil {
		// 更新 Release 状态为失败
		models.UpdateRelease(release.ID, "", release.BuildSourceInfo, "failed")
		return fmt.Errorf("构建镜像失败: %w", err)
	}

	// 2. 更新 Release 状态为成功
	_, err = models.UpdateRelease(release.ID, imageName, release.BuildSourceInfo, "success")
	if err != nil {
		return fmt.Errorf("更新 Release 状态失败: %w", err)
	}

	logman.Info("构建完成", "release_id", release.ID, "image_name", imageName)
	return nil
}

// startBuildAndDeploymentAsync 异步执行构建+部署流程
func (do *DeploymentOrchestrator) startBuildAndDeploymentAsync(deploymentID uuid.UUID) {
	// 获取部署记录
	deployment, err := models.GetDeploymentByID(deploymentID)
	if err != nil {
		logman.Error("获取部署记录失败", "deployment_id", deploymentID, "error", err)
		return
	}

	// 获取应用和 Release 信息
	application, err := models.GetApplicationByID(deployment.ApplicationID)
	if err != nil {
		logman.Error("获取应用信息失败", "deployment_id", deploymentID, "error", err)
		do.updateDeploymentFailed(deployment, "获取应用信息失败: "+err.Error())
		return
	}

	release, err := models.GetReleaseByID(deployment.ReleaseID)
	if err != nil {
		logman.Error("获取发布版本失败", "deployment_id", deploymentID, "error", err)
		do.updateDeploymentFailed(deployment, "获取发布版本失败: "+err.Error())
		return
	}

	// 检查 Release 状态并处理构建
	switch release.Status {
	case "building":
		logman.Info("检测到 Release 需要构建", "deployment_id", deploymentID, "release_id", release.ID)

		// 发送实时日志：开始构建镜像
		buildStartMsg := "开始构建镜像..."
		do.sendDeploymentLog(deploymentID, buildStartMsg)
		do.updateDeploymentLogInDB(deploymentID, buildStartMsg)

		// 执行构建
		if err := do.buildRelease(release, application); err != nil {
			logman.Error("构建失败", "deployment_id", deploymentID, "error", err)
			do.updateDeploymentFailed(deployment, "构建失败: "+err.Error())
			return
		}

		// 重新获取更新后的 Release
		release, err = models.GetReleaseByID(deployment.ReleaseID)
		if err != nil {
			logman.Error("重新获取发布版本失败", "deployment_id", deploymentID, "error", err)
			do.updateDeploymentFailed(deployment, "重新获取发布版本失败: "+err.Error())
			return
		}

		// 发送实时日志：构建完成，开始部署
		deployStartMsg := "构建完成，开始部署..."
		do.sendDeploymentLog(deploymentID, deployStartMsg)
		do.updateDeploymentLogInDB(deploymentID, deployStartMsg)
	case "success":
		logman.Info("Release 已存在且成功，直接部署", "deployment_id", deploymentID, "release_id", release.ID)

		// 发送实时日志：Release 已就绪，开始部署
		readyMsg := "Release 已就绪，开始部署..."
		do.sendDeploymentLog(deploymentID, readyMsg)
		do.updateDeploymentLogInDB(deploymentID, readyMsg)
	default:
		logman.Error("Release 状态无效", "deployment_id", deploymentID, "release_id", release.ID, "status", release.Status)
		do.updateDeploymentFailed(deployment, "Release 状态无效: "+release.Status)
		return
	}

	// 重新获取更新后的Release以确保包含最新的版本信息
	release, err = models.GetReleaseByID(deployment.ReleaseID)
	if err != nil {
		logman.Error("重新获取发布版本失败", "deployment_id", deploymentID, "error", err)
		do.updateDeploymentFailed(deployment, "重新获取发布版本失败: "+err.Error())
		return
	}

	// 执行部署流程
	if err := do.executeDeployment(deployment, application, release); err != nil {
		logman.Error("部署执行失败", "deployment_id", deploymentID, "error", err)
		do.updateDeploymentFailed(deployment, "部署执行失败: "+err.Error())
		return
	}

	// 部署成功
	now := time.Now()
	successMsg := "部署成功，应用已启动。"
	do.sendDeploymentLog(deploymentID, successMsg)

	// 重新获取部署记录以确保有最新的日志
	deployment, err = models.GetDeploymentByID(deploymentID)
	if err != nil {
		logman.Error("重新获取部署记录失败", "deployment_id", deploymentID, "error", err)
		return
	}

	_, err = models.UpdateDeployment(
		deployment.ID,
		"success",
		deployment.LogText+successMsg+"\n",
		&now,
	)
	if err != nil {
		logman.Error("更新部署状态失败", "deployment_id", deploymentID, "error", err)
	}

	logman.Info("构建+部署完成", "deployment_id", deploymentID)
}

// startDeploymentAsync 异步执行部署流程
func (do *DeploymentOrchestrator) startDeploymentAsync(deploymentID uuid.UUID) {
	// 获取部署记录
	deployment, err := models.GetDeploymentByID(deploymentID)
	if err != nil {
		logman.Error("获取部署记录失败", "deployment_id", deploymentID, "error", err)
		return
	}

	// 获取应用和 Release 信息
	application, err := models.GetApplicationByID(deployment.ApplicationID)
	if err != nil {
		logman.Error("获取应用信息失败", "deployment_id", deploymentID, "error", err)
		do.updateDeploymentFailed(deployment, "获取应用信息失败: "+err.Error())
		return
	}

	release, err := models.GetReleaseByID(deployment.ReleaseID)
	if err != nil {
		logman.Error("获取发布版本失败", "deployment_id", deploymentID, "error", err)
		do.updateDeploymentFailed(deployment, "获取发布版本失败: "+err.Error())
		return
	}

	// 检查 Release 状态
	switch release.Status {
	case "building":
		logman.Info("检测到 Release 正在构建，先执行构建", "deployment_id", deploymentID, "release_id", release.ID)

		// 发送实时日志：Release 正在构建，等待构建完成
		waitingMsg := "Release 正在构建，等待构建完成..."
		do.sendDeploymentLog(deploymentID, waitingMsg)
		do.updateDeploymentLogInDB(deploymentID, waitingMsg)

		// 执行构建
		if err := do.buildRelease(release, application); err != nil {
			logman.Error("构建失败", "deployment_id", deploymentID, "error", err)
			do.updateDeploymentFailed(deployment, "构建失败: "+err.Error())
			return
		}

		// 重新获取更新后的 Release
		release, err = models.GetReleaseByID(deployment.ReleaseID)
		if err != nil {
			logman.Error("重新获取发布版本失败", "deployment_id", deploymentID, "error", err)
			do.updateDeploymentFailed(deployment, "重新获取发布版本失败: "+err.Error())
			return
		}

		// 发送实时日志：构建完成，开始部署
		deployStartMsg := "构建完成，开始部署..."
		do.sendDeploymentLog(deploymentID, deployStartMsg)
		do.updateDeploymentLogInDB(deploymentID, deployStartMsg)
	case "success":
		// No additional action needed for success case in this function
	default:
		logman.Error("Release 状态无效", "deployment_id", deploymentID, "release_id", release.ID, "status", release.Status)
		do.updateDeploymentFailed(deployment, "Release 状态无效: "+release.Status)
		return
	}

	// 重新获取更新后的Release以确保包含最新的版本信息
	release, err = models.GetReleaseByID(deployment.ReleaseID)
	if err != nil {
		logman.Error("重新获取发布版本失败", "deployment_id", deploymentID, "error", err)
		do.updateDeploymentFailed(deployment, "重新获取发布版本失败: "+err.Error())
		return
	}

	// 执行部署流程
	if err := do.executeDeployment(deployment, application, release); err != nil {
		logman.Error("部署执行失败", "deployment_id", deploymentID, "error", err)
		do.updateDeploymentFailed(deployment, "部署执行失败: "+err.Error())
		return
	}

	// 部署成功
	now := time.Now()
	successMsg := "部署成功，应用已启动。"
	do.sendDeploymentLog(deploymentID, successMsg)

	// 重新获取部署记录以确保有最新的日志
	deployment, err = models.GetDeploymentByID(deploymentID)
	if err != nil {
		logman.Error("重新获取部署记录失败", "deployment_id", deploymentID, "error", err)
		return
	}

	_, err = models.UpdateDeployment(
		deployment.ID,
		"success",
		deployment.LogText+successMsg+"\n",
		&now,
	)
	if err != nil {
		logman.Error("更新部署状态失败", "deployment_id", deploymentID, "error", err)
	}

	logman.Info("部署完成", "deployment_id", deploymentID)
}

// executeDeployment 执行具体的部署操作
func (do *DeploymentOrchestrator) executeDeployment(deployment *models.Deployment, application *models.Application, release *models.Release) error {
	logman.Info("开始执行部署", "deployment_id", deployment.ID, "app_name", application.Name)

	// 1. 生成运行时文件
	project, err := do.generateRuntimeFiles(deployment, application, release)
	if err != nil {
		return fmt.Errorf("生成运行时文件失败: %w, deployment_id: %s", err, deployment.ID)
	}

	// 2. 执行系统级部署
	if err := do.deployToSystem(application, deployment.ServiceName, project); err != nil {
		return fmt.Errorf("系统部署失败: %w, deployment_id: %s", err, deployment.ID)
	}

	// 3. 更新应用的当前发布版本（原子化切换）
	if err := do.updateActiveRelease(application, release); err != nil {
		return fmt.Errorf("更新活跃发布版本失败: %w, deployment_id: %s", err, deployment.ID)
	}

	return nil
}

// generateRuntimeFiles 生成运行时配置文件
func (do *DeploymentOrchestrator) generateRuntimeFiles(deployment *models.Deployment, application *models.Application, release *models.Release) (*models.Project, error) {
	logman.Info("生成运行时配置文件", "app_name", application.Name)

	// 1. 生成环境变量内容
	envContent, err := models.GenerateEnvFileContent(application.ID)
	if err != nil {
		logman.Warn("生成环境变量内容失败", "error", err, "app_id", application.ID)
		envContent = "" // Continue with empty env content
	}

	fmt.Println("生成环境变量内容成功", envContent)

	routings, err := models.GetActiveRoutingsByApplicationID(application.ID)
	if err != nil {
		return nil, fmt.Errorf("查询路由信息失败: %w, deployment_id: %s", err, deployment.ID)
	}

	// 2. 获取项目信息并生成环境文件路径
	project, err := models.GetProjectByID(application.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("获取项目信息失败: %w, deployment_id: %s", err, deployment.ID)
	}
	envFilePath := do.envService.GenerateProjectEnvPath(project.HomeDir, application.Name)
	fmt.Println("环境文件路径", envFilePath)
	// 3. 生成 Quadlet 文件内容
	quadletContent, err := do.generateQuadletContent(application, release, routings, envFilePath)
	if err != nil {
		return nil, fmt.Errorf("生成 Quadlet 内容失败: %w, deployment_id: %s", err, deployment.ID)
	}
	fmt.Println("生成 Quadlet 内容成功", quadletContent)

	// 4. Use the already generated environment content

	// 5. 写入文件到系统
	if err := do.writeRuntimeFiles(application.Name, release.Version, quadletContent, envContent, envFilePath, project); err != nil {
		return nil, fmt.Errorf("写入运行时文件失败: %w, deployment_id: %s", err, deployment.ID)
	}

	return project, nil
}

// generateQuadletContent 生成 Quadlet 配置内容
func (do *DeploymentOrchestrator) generateQuadletContent(application *models.Application, release *models.Release, routings []*models.Routing, envFilePath string) (string, error) {
	// 使用现有的 quadlet 生成逻辑，参考 services/quadlet_service.go
	data := QuadletData{
		Description:      application.Description,
		ImageName:        release.ImageName,
		EnvFilePath:      envFilePath,
		PublishPorts:     []string{},
		Volumes:          []string{},
		ExecCommand:      "",
		AutoUpdatePolicy: "",
	}

	// 每次都生成系统端口（不管是否有路由配置）

	systemPort, err := do.generateRandomPortSystemPort()
	if err != nil {
		fmt.Println("系统端口冲突检查失败，部署将继续，但可能无法启动，请留意后续情况", err)
	}
	publishPort := fmt.Sprintf("%d:%d", systemPort, application.TargetPort)
	data.PublishPorts = append(data.PublishPorts, publishPort)

	// 更新 Release 的 SystemPort 字段
	if err := models.UpdateDeploymentSystemPort(release.ID, systemPort); err != nil {
		logman.Warn("更新 Release SystemPort 失败", "release_id", release.ID, "error", err)
	}

	// 设置卷挂载
	if application.Volumes.Data != nil {
		if volumes, ok := application.Volumes.Data.([]interface{}); ok {
			for _, v := range volumes {
				if vol, ok := v.(map[string]interface{}); ok {
					hostPath, hok := vol["host_path"].(string)
					containerPath, cok := vol["container_path"].(string)
					if hok && cok {
						data.Volumes = append(data.Volumes, hostPath+":"+containerPath)
					}
				}
			}
		}
	}

	// 设置执行命令
	if application.ExecCommand != nil {
		data.ExecCommand = *application.ExecCommand
	}

	// 设置自动更新策略
	if application.AutoUpdatePolicy != nil {
		data.AutoUpdatePolicy = *application.AutoUpdatePolicy
	}

	// 生成 Quadlet 内容
	return do.renderQuadletTemplate(data)
}

// generateRandomPortSystemPort 生成随机系统端口（10000-25000之间）
func (do *DeploymentOrchestrator) generateRandomPortSystemPort() (int, error) {
	minPort := 10001
	maxPort := 65535
	maxRetries := 100

	for i := 0; i < maxRetries; i++ {
		port := rand.Intn(maxPort-minPort+1) + minPort
		inUse, err := models.IsPortInUse(port)
		if err != nil {
			return port, fmt.Errorf("failed to check port usage: %w", err)
		}
		if !inUse {
			return port, nil
		}
	}

	return 0, fmt.Errorf("unable to find an unused port after %d attempts", maxRetries)
}

// generateEnvironmentFileContent 生成环境文件内容
func (do *DeploymentOrchestrator) generateEnvironmentFileContent(applicationID uuid.UUID) string {
	envContent, err := models.GenerateEnvFileContent(applicationID)
	if err != nil {
		logman.Error("生成环境变量内容失败", "error", err, "application_id", applicationID)
		return ""
	}
	return envContent
}

// renderQuadletTemplate 渲染 Quadlet 模板
func (do *DeploymentOrchestrator) renderQuadletTemplate(data QuadletData) (string, error) {
	template := `[Unit]
Description=%s

[Container]
Image=%s`

	content := fmt.Sprintf(template, data.Description, data.ImageName)

	// 添加执行命令
	if data.ExecCommand != "" {
		content += fmt.Sprintf("\nExec=%s", data.ExecCommand)
	}

	// 添加自动更新策略
	if data.AutoUpdatePolicy != "" {
		content += fmt.Sprintf("\nAutoUpdate=%s", data.AutoUpdatePolicy)
	}

	// 添加端口映射
	for _, port := range data.PublishPorts {
		content += fmt.Sprintf("\nPublishPort=%s", port)
	}

	// 添加卷挂载
	for _, volume := range data.Volumes {
		content += fmt.Sprintf("\nVolume=%s", volume)
	}

	// 添加环境文件
	content += fmt.Sprintf("\nEnvironmentFile=%s", data.EnvFilePath)

	// 添加 Install 段
	content += "\n\n[Install]\nWantedBy=default.target"

	return content, nil
}

// writeRuntimeFiles 写入运行时配置文件
func (do *DeploymentOrchestrator) writeRuntimeFiles(appName, version, quadletContent, envContent, envFilePath string, project *models.Project) error {
	// 1. 创建项目专属的 systemd 目录
	var systemdDir string
	if project.HomeDir != "" {
		// 使用项目的 HomeDir 创建 .config/containers/systemd 目录
		systemdDir = filepath.Join(project.HomeDir, ".config", "containers", "systemd")
	} else {
		// 回退到系统目录
		systemdDir = "/usr/share/containers/systemd"
	}

	if err := os.MkdirAll(systemdDir, 0755); err != nil {
		return fmt.Errorf("创建 systemd 目录失败: %w", err)
	}

	// 2. 写入 Quadlet 文件
	quadletFile := filepath.Join(systemdDir, fmt.Sprintf("%s-%s.container", appName, version))
	if err := os.WriteFile(quadletFile, []byte(quadletContent), 0644); err != nil {
		return fmt.Errorf("写入 Quadlet 文件失败: %w", err)
	}

	// 3. 调用envService创建环境文件（
	if err := do.envService.CreateEnvironmentFileForDeployment(appName, envContent, envFilePath); err != nil {
		return fmt.Errorf("创建环境文件失败: %w", err)
	}

	logman.Info("运行时文件写入完成", "quadlet_file", quadletFile, "env_file", envFilePath)
	return nil
}

// deployToSystem 执行系统级部署操作
func (do *DeploymentOrchestrator) deployToSystem(application *models.Application, serviceName string, project *models.Project) error {
	logman.Info("开始系统级部署", "app_name", application.Name)

	// // 1. 停止旧服务（如果存在）
	// 目前的部署策略不是停止旧服务，而是直接启动新服务，以支持回滚和零停机时间，暂时注释掉
	// oldServiceName := application.Name + ".service"
	// if err := do.stopOldService(oldServiceName); err != nil {
	// 	logman.Warn("停止旧服务失败", "service", oldServiceName, "error", err)
	// 	// 不是致命错误，继续执行
	// }

	// 2. 重新加载 systemd daemon (使用用户模式)
	if err := do.reloadUserSystemdDaemon(project); err != nil {
		return fmt.Errorf("重新加载 systemd daemon 失败: %w", err)
	}

	// Use deployment.ServiceName instead of constructing newServiceName
	// Note: deployment is passed to executeDeployment, so we need to ensure it's available here.
	// Assuming executeDeployment passes deployment to this function, or adjust accordingly.
	// newServiceName := deployment.ServiceName // Replace manual construction

	// 3. 启动新服务 (使用用户模式)
	if err := do.startUserService(serviceName, project); err != nil {
		fmt.Println(serviceName)
		return fmt.Errorf("启动服务失败: %w", err)
	}

	// 4. 检查服务状态 (使用用户模式)
	if err := do.checkUserServiceHealth(serviceName, project); err != nil {
		return fmt.Errorf("服务健康检查失败: %w", err)
	}

	logman.Info("系统级部署完成", "app_name", application.Name)
	return nil
}

// stopOldService 停止旧服务
func (do *DeploymentOrchestrator) stopOldService(serviceName string) error {
	logman.Info("停止旧服务", "service", serviceName)

	// 检查服务是否存在和运行
	checkCmd := exec.Command("systemctl", "is-active", serviceName)
	if err := checkCmd.Run(); err != nil {
		logman.Info("服务未运行或不存在", "service", serviceName)
		return nil // 服务不存在或未运行，不是错误
	}

	// 停止服务
	_, err := command.Exec(&command.ExecPayload{
		Content:     fmt.Sprintf("systemctl stop %s", serviceName),
		CommandType: "SHELL",
		Timeout:     30,
	})
	if err != nil {
		return fmt.Errorf("停止服务失败: %w", err)
	}

	logman.Info("旧服务已停止", "service", serviceName)
	return nil
}

// startService 启动服务
func (do *DeploymentOrchestrator) startService(serviceName string) error {
	logman.Info("启动服务", "service", serviceName)

	_, err := command.Exec(&command.ExecPayload{
		Content:     fmt.Sprintf("systemctl start %s", serviceName),
		CommandType: "SHELL",
		Timeout:     60,
	})
	if err != nil {
		return fmt.Errorf("启动服务失败: %w", err)
	}

	logman.Info("服务已启动", "service", serviceName)
	return nil
}

// checkServiceHealth 检查服务健康状态
func (do *DeploymentOrchestrator) checkServiceHealth(serviceName string) error {
	logman.Info("检查服务健康状态", "service", serviceName, "timestamp", time.Now().Format(time.RFC3339))

	// 等待更长时间，确保服务稳定
	time.Sleep(6 * time.Second) // 从5秒增加到6秒

	checkCmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := checkCmd.CombinedOutput()
	if err != nil {
		logman.Error("服务健康检查失败", "service", serviceName, "error", err, "output", string(output), "timestamp", time.Now().Format(time.RFC3339))
		// 额外诊断：获取服务状态详情
		statusCmd := exec.Command("systemctl", "status", serviceName)
		statusOutput, statusErr := statusCmd.CombinedOutput()
		if statusErr != nil {
			logman.Error("获取服务状态失败", "service", serviceName, "status_error", statusErr, "status_output", string(statusOutput))
		} else {
			logman.Info("服务状态详情", "service", serviceName, "status_output", string(statusOutput))
		}
		return fmt.Errorf("服务未正常运行: %w, output: %s", err, string(output))
	}

	logman.Info("服务健康检查通过", "service", serviceName, "output", string(output), "timestamp", time.Now().Format(time.RFC3339))
	return nil
}

// updateActiveRelease 更新应用的活跃发布版本（原子化切换）
func (do *DeploymentOrchestrator) updateActiveRelease(application *models.Application, release *models.Release) error {
	logman.Info("更新活跃发布版本", "app_id", application.ID, "release_id", release.ID)

	// 使用事务确保原子性
	tx := dborm.Db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新应用的 ActiveReleaseID 和状态
	if err := tx.Model(application).Updates(map[string]interface{}{
		"ActiveReleaseID": release.ID,
		"Status":          "running",
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新应用状态失败: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	logman.Info("活跃发布版本更新完成", "app_id", application.ID, "release_id", release.ID)
	return nil
}

// updateReleaseSystemPort 更新 Release 的 SystemPort 字段

// updateDeploymentFailed 更新部署状态为失败
func (do *DeploymentOrchestrator) updateDeploymentFailed(deployment *models.Deployment, errorMsg string) {
	// 重新获取最新的部署记录以确保有最新的日志
	latestDeployment, err := models.GetDeploymentByID(deployment.ID)
	if err != nil {
		logman.Error("获取最新部署记录失败", "deployment_id", deployment.ID, "error", err)
		latestDeployment = deployment // fallback to the passed deployment
	}

	now := time.Now()
	_, err = models.UpdateDeployment(
		latestDeployment.ID,
		"failed",
		latestDeployment.LogText+"部署失败: "+errorMsg+"\n",
		&now,
	)
	if err != nil {
		logman.Error("更新部署失败状态失败", "deployment_id", latestDeployment.ID, "error", err)
	}

	logman.Info("部署状态已更新为失败", "deployment_id", latestDeployment.ID, "error_msg", errorMsg)
}

// reloadUserSystemdDaemon 重新加载用户模式的 systemd daemon
func (do *DeploymentOrchestrator) reloadUserSystemdDaemon(project *models.Project) error {
	if project.Username == "" {
		// 回退到系统模式
		return ReloadSystemdDaemon()
	}

	// 验证项目目录权限
	if err := do.validateProjectDirectoryPermissions(project); err != nil {
		return fmt.Errorf("项目目录权限验证失败: %w", err)
	}

	logman.Info("执行用户模式 systemctl daemon-reload 命令", "username", project.Username)

	// 使用 su 切换到项目用户执行命令
	cmd := fmt.Sprintf("su - %s -c 'systemctl --user daemon-reload'", project.Username)
	_, err := command.Exec(&command.ExecPayload{
		Content:     cmd,
		CommandType: "SHELL",
		Timeout:     30,
	})
	if err != nil {
		logman.Error("用户模式 systemctl daemon-reload 执行失败", "username", project.Username, "error", err)
		return fmt.Errorf("failed to reload user systemd daemon: %w", err)
	}

	logman.Info("用户模式 systemctl daemon-reload 执行成功", "username", project.Username)
	return nil
}

// validateProjectDirectoryPermissions 验证项目目录权限
func (do *DeploymentOrchestrator) validateProjectDirectoryPermissions(project *models.Project) error {
	if project.HomeDir == "" {
		return nil // 没有设置 HomeDir，跳过验证
	}

	// 验证 HomeDir 存在且可写
	if err := utils.NewDirectoryManager().ValidateDirectoryWritePermissions(project.HomeDir); err != nil {
		return fmt.Errorf("项目 HomeDir 权限验证失败: %w", err)
	}

	// 验证或创建必要的子目录
	requiredDirs := []string{
		filepath.Join(project.HomeDir, ".config"),
		filepath.Join(project.HomeDir, ".config", "containers"),
		filepath.Join(project.HomeDir, ".config", "containers", "systemd"),
		filepath.Join(project.HomeDir, ".config", "env"),
	}

	for _, dir := range requiredDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败 %s: %w", dir, err)
		}
		if err := utils.NewDirectoryManager().ValidateDirectoryWritePermissions(dir); err != nil {
			return fmt.Errorf("目录权限验证失败 %s: %w", dir, err)
		}
	}

	logman.Info("项目目录权限验证通过", "home_dir", project.HomeDir, "username", project.Username)
	return nil
}

// startUserService 启动用户模式服务
func (do *DeploymentOrchestrator) startUserService(serviceName string, project *models.Project) error {
	if project.Username == "" {
		// 回退到系统模式
		return do.startService(serviceName)
	}

	logman.Info("启动用户模式服务", "service", serviceName, "username", project.Username)

	// 使用 su 切换到项目用户执行命令
	cmd := fmt.Sprintf("su - %s -c 'systemctl --user start %s'", project.Username, serviceName)
	_, err := command.Exec(&command.ExecPayload{
		Content:     cmd,
		CommandType: "SHELL",
		Timeout:     60,
	})
	if err != nil {
		return fmt.Errorf("启动用户模式服务失败: %w", err)
	}

	logman.Info("用户模式服务已启动", "service", serviceName, "username", project.Username)
	return nil
}

// checkUserServiceHealth 检查用户模式服务健康状态
func (do *DeploymentOrchestrator) checkUserServiceHealth(serviceName string, project *models.Project) error {
	if project.Username == "" {
		// 回退到系统模式
		return do.checkServiceHealth(serviceName)
	}

	logman.Info("检查用户模式服务健康状态", "service", serviceName, "username", project.Username, "timestamp", time.Now().Format(time.RFC3339))

	// 等待更长时间，确保服务稳定
	time.Sleep(6 * time.Second)

	// 使用 su 切换到项目用户执行命令
	cmd := fmt.Sprintf("su - %s -c 'systemctl --user is-active %s'", project.Username, serviceName)
	output, err := command.Exec(&command.ExecPayload{
		Content:     cmd,
		CommandType: "SHELL",
		Timeout:     30,
	})

	if err != nil {
		logman.Error("用户模式服务健康检查失败", "service", serviceName, "username", project.Username, "error", err, "output", output, "timestamp", time.Now().Format(time.RFC3339))

		// 额外诊断：获取服务状态详情
		statusCmd := fmt.Sprintf("su - %s -c 'systemctl --user status %s'", project.Username, serviceName)
		statusOutput, statusErr := command.Exec(&command.ExecPayload{
			Content:     statusCmd,
			CommandType: "SHELL",
			Timeout:     30,
		})
		if statusErr != nil {
			logman.Error("获取用户模式服务状态失败", "service", serviceName, "username", project.Username, "status_error", statusErr, "status_output", statusOutput)
		} else {
			logman.Info("用户模式服务状态详情", "service", serviceName, "username", project.Username, "status_output", statusOutput)
		}
		return fmt.Errorf("用户模式服务未正常运行: %w, output: %s", err, output)
	}

	logman.Info("用户模式服务健康检查通过", "service", serviceName, "username", project.Username, "output", output, "timestamp", time.Now().Format(time.RFC3339))
	return nil
}
