// 它通过创建专属的、无特权的 Linux 用户，并启用 linger来实现服务的持久化运行。
//
// 核心理念:
// 1. 安全隔离: 每个项目运行在独立的 Linux 用户下，遵循最小权限原则。
// 2. 自动化: 提供简单的 Go API 来处理复杂的系统命令 (useradd, loginctl)。
// 3. 持久化: 通过启用 linger，确保用户服务在系统重启后也能像系统服务一样运行。
//
// 注意：执行此包中的操作需要 root 或 sudo 权限。
package services

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// UserSuffix 是附加到项目名称后以形成用户名的后缀。
	UserSuffix = "-user"
)

// Project 描述了一个已配置好的项目环境。
// 这个结构体是 Setup() 方法成功后的返回值，包含了后续所有操作所需的所有关键信息。
type Project struct {
	Name     string // 项目的原始名称, e.g., "my-web-app"
	Username string // 为项目创建的专属 Linux 用户名, e.g., "my-web-app-user"

	// HomeDir 是此项目的根目录，也是实现隔离的核心。
	// 后续所有与项目相关的文件都应存放在此目录下，例如：
	// - Quadlet 文件: HomeDir/.config/containers/systemd/
	// - 环境变量文件: HomeDir/.env
	// - 持久化 Volume 数据: HomeDir/data/
	HomeDir string
}

// Manager 是管理项目环境的主要结构体。
type Manager struct {
	basePath string // 所有项目 home 目录的基础路径, e.g., "/var/lib/my-tool/projects"
}

// NewManager 创建一个新的 Manager 实例。
// 它会确保 basePath 存在。
func NewManager(basePath string) (*Manager, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("无法创建基础目录 '%s': %w", basePath, err)
	}
	return &Manager{basePath: basePath}, nil
}

// createProjectDir ensures the project's home directory exists with proper permissions.
func (m *Manager) createProjectDir(homeDir string) error {
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		return fmt.Errorf("failed to create project directory %s: %w", homeDir, err)
	}
	// Ensure permissions are set
	if err := os.Chmod(homeDir, 0755); err != nil {
		return fmt.Errorf("failed to set permissions on project directory %s: %w", homeDir, err)
	}
	return nil
}

// findCommandPath searches for the command in common system paths.
func findCommandPath(cmd string) (string, error) {
	paths := []string{"/usr/sbin", "/usr/bin", "/bin", "/usr/local/sbin", "/usr/local/bin"}
	for _, path := range paths {
		fullPath := filepath.Join(path, cmd)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}
	return "", fmt.Errorf("command %s not found in common paths", cmd)
}

// Setup 为一个新项目配置完整的运行环境。
// 返回的 *Project 对象是后续部署步骤的“钥匙”。
func (m *Manager) Setup(projectName string) (*Project, error) {
	if strings.Trim(projectName, " ") == "" {
		return nil, fmt.Errorf("项目名称不能为空")
	}

	proj := &Project{
		Name:     projectName,
		Username: strings.ReplaceAll(projectName, " ", "-") + UserSuffix, // Sanitize spaces in username for Linux compatibility
		HomeDir:  filepath.Join(m.basePath, projectName),
	}

	// Ensure the project directory exists before proceeding
	if err := m.createProjectDir(proj.HomeDir); err != nil {
		return nil, fmt.Errorf("failed to create project directory for '%s': %w", proj.Name, err)
	}

	if userExists(proj.Username) {
		fmt.Printf("用户 '%s' 已存在，跳过创建步骤。\n", proj.Username)
		if err := m.runCommand("loginctl", "enable-linger", proj.Username); err != nil {
			return nil, fmt.Errorf("为已存在的用户 '%s' 启用 linger 失败: %w", proj.Username, err)
		}
		return proj, nil
	}

	// Find full path for useradd
	useraddPath, err := findCommandPath("useradd")
	if err != nil {
		return nil, fmt.Errorf("useradd not found: %w", err)
	}

	fmt.Printf("正在为项目 '%s' 创建用户 '%s'...\n", proj.Name, proj.Username)
	err = m.runCommand(useraddPath,
		"--system",
		"--home-dir", proj.HomeDir,
		"--create-home",
		"--shell", "/bin/false",
		proj.Username,
	)
	if err != nil {
		return nil, fmt.Errorf("创建用户 '%s' 失败: %w", proj.Username, err)
	}

	fmt.Printf("正在为用户 '%s' 启用 linger...\n", proj.Username)
	if err := m.runCommand("loginctl", "enable-linger", proj.Username); err != nil {
		_ = m.Teardown(projectName)
		return nil, fmt.Errorf("为用户 '%s' 启用 linger 失败: %w", proj.Username, err)
	}

	fmt.Printf("项目 '%s' 的环境已成功设置。\n", proj.Name)
	return proj, nil
}

// Teardown 彻底清理一个项目的运行环境。
func (m *Manager) Teardown(projectName string) error {
	username := strings.ReplaceAll(projectName, " ", "-") + UserSuffix // Sanitize spaces in username for consistency
	fmt.Printf("正在清理项目 '%s' 的环境 (用户: %s)...\n", projectName, username)

	if !userExists(username) {
		fmt.Printf("用户 '%s' 不存在，无需清理。\n", username)
		return nil
	}

	if err := m.runCommand("loginctl", "disable-linger", username); err != nil {
		fmt.Printf("警告: 禁用用户 '%s' 的 linger 失败: %v\n", username, err)
	}

	if err := m.runCommand("userdel", "--remove", username); err != nil {
		return fmt.Errorf("删除用户 '%s' 失败: %w", username, err)
	}

	fmt.Printf("项目 '%s' 的环境已成功清理。\n", projectName)
	return nil
}

func (m *Manager) runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("执行命令 '%s %s' 失败: %w, 错误输出: %s", name, strings.Join(args, " "), err, stderr.String())
	}
	return nil
}

func userExists(username string) bool {
	return exec.Command("id", username).Run() == nil
}
