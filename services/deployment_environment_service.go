package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/OrbitDeploy/OrbitDeploy/utils"
	"github.com/opentdp/go-helper/logman"
)

// DeploymentEnvironmentService 处理部署环境文件操作
type DeploymentEnvironmentService struct {
	dirManager *utils.DirectoryManager
}

// NewDeploymentEnvironmentService 创建新的环境服务实例
func NewDeploymentEnvironmentService() *DeploymentEnvironmentService {
	return &DeploymentEnvironmentService{
		dirManager: utils.NewDirectoryManager(),
	}
}

// CreateEnvironmentFileForDeployment 为部署创建环境文件
func (des *DeploymentEnvironmentService) CreateEnvironmentFileForDeployment(appName, envContent, envFilePath string) error {
	// 1. 解析环境文件路径
	realEnvPath, err := des.resolveSystemdPath(envFilePath, appName)
	if err != nil {
		logman.Error("解析环境变量文件路径失败", "path", envFilePath, "error", err)
		return err
	}
	logman.Info("路径解析成功", "systemd_path", envFilePath, "real_path", realEnvPath)

	// 2. 创建目录
	envDir := filepath.Dir(realEnvPath)
	logman.Info("准备创建环境变量文件目录", "dir", envDir)
	if err := os.MkdirAll(envDir, 0755); err != nil {
		logman.Error("创建环境变量文件目录失败", "dir", envDir, "error", err)
		return fmt.Errorf("创建环境变量文件目录失败 %s: %w", envDir, err)
	}

	// 3. 写入文件
	logman.Info("准备写入环境变量文件", "file", realEnvPath)
	if err := os.WriteFile(realEnvPath, []byte(envContent), 0600); err != nil {
		logman.Error("写入环境变量文件失败", "file", realEnvPath, "error", err)
		return fmt.Errorf("写入环境变量文件失败: %w", err)
	}

	logman.Info("环境变量文件创建成功", "file", realEnvPath)
	return nil
}

// resolveSystemdPath 解析 systemd 路径规范（如 %h）
func (des *DeploymentEnvironmentService) resolveSystemdPath(path, appName string) (string, error) {
	expandPath := func(path string) (string, error) {
		// 展开 systemd 路径规范
		if strings.Contains(path, "%h") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("获取用户家目录失败: %w", err)
			}
			path = strings.ReplaceAll(path, "%h", homeDir)
		}

		// 展开波浪号
		if strings.HasPrefix(path, "~/") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("获取用户家目录失败: %w", err)
			}
			path = filepath.Join(homeDir, path[2:])
		}

		// 展开环境变量
		return os.ExpandEnv(path), nil
	}

	return expandPath(path)
}

// AddEnvironmentFileToQuadlet 向 Quadlet 内容添加环境文件指令
func (des *DeploymentEnvironmentService) AddEnvironmentFileToQuadlet(quadletContent, envFilePath string) string {
	lines := strings.Split(quadletContent, "\n")
	var result []string
	containerSectionFound := false
	environmentFileAdded := false

	for _, line := range lines {
		result = append(result, line)

		// 查找 [Container] 段
		if strings.TrimSpace(line) == "[Container]" {
			containerSectionFound = true
		} else if containerSectionFound && !environmentFileAdded {
			// 检查我们是否仍在 Container 段中或已移动到另一个段
			trimmedLine := strings.TrimSpace(line)
			if strings.HasPrefix(trimmedLine, "[") && trimmedLine != "[Container]" {
				// 我们已移动到另一个段，在此段之前添加 EnvironmentFile
				result = result[:len(result)-1] // 移除当前行
				result = append(result, fmt.Sprintf("EnvironmentFile=%s", envFilePath))
				result = append(result, line) // 重新添加当前行
				environmentFileAdded = true
			}
		}
	}

	// 如果到文件末尾还没有添加 EnvironmentFile，则添加到 Container 段的末尾
	if containerSectionFound && !environmentFileAdded {
		result = append(result, fmt.Sprintf("EnvironmentFile=%s", envFilePath))
	}

	return strings.Join(result, "\n")
}

// ParseEnvironmentFilePath 从 Quadlet 配置中提取环境文件路径，或返回默认路径
func (des *DeploymentEnvironmentService) ParseEnvironmentFilePath(quadletFile, containerName string) string {
	lines := strings.Split(quadletFile, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "EnvironmentFile=") {
			return strings.TrimPrefix(line, "EnvironmentFile=")
		}
	}

	// 如果未找到，返回默认路径（注意：这里假设 projectName 为空或需要外部传入；可后续优化）
	return des.GenerateDefaultEnvPath("", containerName)
}

// GenerateDefaultEnvPath 创建默认环境文件路径，格式为：
// $HOME/.config/orbit/{projectName}/env/{containerName}.env
// 已废弃：请使用 GenerateProjectEnvPath 替代
func (des *DeploymentEnvironmentService) GenerateDefaultEnvPath(projectName, containerName string) string {
	return fmt.Sprintf("%%h/.config/orbit/%s/env/%s.env", projectName, containerName)
}

// GenerateProjectEnvPath 基于项目HomeDir创建环境文件路径，格式为：
// {projectHomeDir}/.config/env/{containerName}.env
func (des *DeploymentEnvironmentService) GenerateProjectEnvPath(projectHomeDir, containerName string) string {
	if projectHomeDir == "" {
		// 回退到系统 HOME 目录
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			homeDir = "/root"
		}
		projectHomeDir = homeDir
	}
	return filepath.Join(projectHomeDir, ".config", "env", fmt.Sprintf("%s.env", containerName))
}

// HasEnvironmentFileDirective 检查 Quadlet 文件是否已有 EnvironmentFile 指令
func (des *DeploymentEnvironmentService) HasEnvironmentFileDirective(quadletFile string) bool {
	lines := strings.Split(quadletFile, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "EnvironmentFile=") {
			return true
		}
	}
	return false
}
