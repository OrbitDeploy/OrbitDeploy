package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// CreateEnvFileDirectory 创建环境文件目录并设置权限为 0755
// 这是一个通用函数，供其他函数调用，用于确保环境文件路径的目录存在
func CreateFileDirectory(envFilePath string) error {
	envDir := filepath.Dir(envFilePath)
	if err := os.MkdirAll(envDir, 0755); err != nil {
		return fmt.Errorf("failed to create env file directory %s: %w", envDir, err)
	}
	// 确保权限被设置
	if err := os.Chmod(envDir, 0755); err != nil {
		return fmt.Errorf("failed to set permissions on env file directory %s: %w", envDir, err)
	}
	return nil
}
func adjustContainerName(containerName string) string {
	containerName = strings.TrimSuffix(containerName, ".service")
	if !strings.HasPrefix(containerName, "systemd-") {
		containerName = "systemd-" + containerName
		// fmt.Println("Adjusted container name for Quadlet:", containerName)
	}
	return containerName
}
func GetProjectsBaseDir() string {
	if os.Geteuid() != 0 {
		log.Fatal("错误：此工具需要以 root 或 sudo 权限运行。")
	}

	// =================================================================
	// 步骤 1: 动态获取部署工具自己的名字
	// =================================================================
	// fmt.Println("▶ 步骤 1: 动态获取工具的可执行文件名...")

	// 获取当前程序的可执行文件完整路径
	executablePath, err := os.Executable()
	if err != nil {
		log.Fatalf("无法获取可执行文件路径: %v", err)
	}

	// 从完整路径中提取出文件名
	toolName := filepath.Base(executablePath)

	fmt.Printf("✔ 工具的可执行文件名为: %s\n\n", toolName)

	// =================================================================
	// 步骤 2: 使用动态获取的名字构建基础路径
	// =================================================================
	fmt.Println("▶ 步骤 2: 构建动态的基础数据路径...")

	// 使用 toolName 变量，而不是硬编码的字符串
	projectsBaseDir := filepath.Join("/var/lib", toolName, "projects")

	fmt.Printf("✔ 所有项目将被存放在: %s\n\n", projectsBaseDir)

	return projectsBaseDir
}
