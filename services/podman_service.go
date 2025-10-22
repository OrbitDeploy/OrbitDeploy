package services

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/opentdp/go-helper/logman"
)

// PodmanService 处理 Podman 相关操作
type PodmanService struct{}

// NewPodmanService 创建新的 Podman 服务实例
func NewPodmanService() *PodmanService {
	return &PodmanService{}
}

// LoadPodmanImage 加载 Podman 镜像
func (ps *PodmanService) LoadPodmanImage(imageName string) error {
	logman.Info("加载 Podman 镜像", "image_name", imageName)

	// 检查镜像是否已存在
	checkCmd := exec.Command("podman", "image", "exists", imageName)
	if err := checkCmd.Run(); err == nil {
		logman.Info("镜像已存在", "image", imageName)
		return nil
	}

	// 拉取镜像
	pullCmd := exec.Command("podman", "pull", imageName)
	if output, err := pullCmd.CombinedOutput(); err != nil {
		logman.Error("拉取镜像失败", "image", imageName, "output", string(output), "error", err)
		return fmt.Errorf("拉取镜像失败: %s", string(output))
	}

	logman.Info("镜像加载成功", "image", imageName)
	return nil
}

// CheckImageExists 检查镜像是否存在
func (ps *PodmanService) CheckImageExists(imageName string) bool {
	checkCmd := exec.Command("podman", "image", "exists", imageName)
	return checkCmd.Run() == nil
}

// PullImage 拉取镜像
func (ps *PodmanService) PullImage(imageName string) error {
	logman.Info("拉取镜像", "image", imageName)

	pullCmd := exec.Command("podman", "pull", imageName)
	if output, err := pullCmd.CombinedOutput(); err != nil {
		logman.Error("拉取镜像失败", "image", imageName, "output", string(output), "error", err)
		return fmt.Errorf("拉取镜像失败: %s", string(output))
	}

	logman.Info("镜像拉取成功", "image", imageName)
	return nil
}

// RunHealthCheck 运行指定容器的健康检查
func (ps *PodmanService) RunHealthCheck(containerName string) error {
	logman.Info("运行健康检查", "container", containerName)

	checkCmd := exec.Command("podman", "healthcheck", "run", containerName)
	if output, err := checkCmd.CombinedOutput(); err != nil {
		logman.Error("健康检查失败", "container", containerName, "output", string(output), "error", err)
		return fmt.Errorf("健康检查失败: %s", string(output))
	}

	logman.Info("健康检查成功", "container", containerName)
	return nil
}

// GetImageUser 检查镜像的运行 UID 和 GID
func (ps *PodmanService) GetImageUser(imageName string) (uid int, gid int, err error) {
	logman.Info("检查镜像用户", "image", imageName)

	inspectCmd := exec.Command("podman", "inspect", "--format", "{{.Config.User}}", imageName)
	output, err := inspectCmd.Output()
	if err != nil {
		logman.Error("检查镜像用户失败", "image", imageName, "error", err)
		return 0, 0, fmt.Errorf("检查镜像用户失败: %v", err)
	}

	userGroup := strings.TrimSpace(string(output))
	parts := strings.Split(userGroup, ":")

	if len(parts) >= 1 {
		uid, err = strconv.Atoi(parts[0])
		if err != nil {
			logman.Error("解析 UID 失败", "image", imageName, "value", parts[0], "error", err)
			return 0, 0, fmt.Errorf("解析 UID 失败: %v", err)
		}
	}

	if len(parts) >= 2 {
		gid, err = strconv.Atoi(parts[1])
		if err != nil {
			logman.Error("解析 GID 失败", "image", imageName, "value", parts[1], "error", err)
			return 0, 0, fmt.Errorf("解析 GID 失败: %v", err)
		}
	} else {
		gid = 0 // 默认 GID 为 0
	}

	logman.Info("镜像用户检查成功", "image", imageName, "uid", uid, "gid", gid)
	return uid, gid, nil
}

// CheckContainerRunning 检查容器是否正在运行，支持 Quadlet 命名规则
func (ps *PodmanService) CheckContainerRunningWithQuadlet(containerName string) bool {
	// logman.Info("检查容器运行状态", "container", containerName)

	// 处理 Quadlet 命名规则：如果不以 "systemd-" 开头，添加前缀
	// 先去掉 .service 后缀（如果存在）
	// containerName = strings.TrimSuffix(containerName, ".service")
	// if !strings.HasPrefix(containerName, "systemd-") {
	// 	containerName = "systemd-" + containerName
	// 	fmt.Println("Adjusted container name for Quadlet:", containerName)
	// }
	containerName = adjustContainerName(containerName)

	psCmd := exec.Command("podman", "ps", "--filter", "name="+containerName, "--format", "{{.Status}}")
	output, err := psCmd.Output()
	if err != nil {
		// logman.Error("检查容器状态失败", "container", containerName, "error", err)
		return false
	}

	status := strings.TrimSpace(string(output))
	lowerStatus := strings.ToLower(status)
	isRunning := strings.HasPrefix(lowerStatus, "up") || strings.Contains(lowerStatus, "running")

	logman.Info("容器状态检查完成", "container", containerName, "status", status, "running", isRunning)
	return isRunning
}
