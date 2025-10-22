package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/OrbitDeploy/OrbitDeploy/models"
)

func TestGenerateProjectEnvPath(t *testing.T) {
	service := NewDeploymentEnvironmentService()

	tests := []struct {
		name           string
		projectHomeDir string
		containerName  string
		expected       string
	}{
		{
			name:           "normal_project_home",
			projectHomeDir: "/home/project-user",
			containerName:  "my-app",
			expected:       "/home/project-user/.config/env/my-app.env",
		},
		{
			name:           "empty_home_dir_fallback",
			projectHomeDir: "",
			containerName:  "my-app",
			expected:       filepath.Join(getHomeDir(), ".config", "env", "my-app.env"),
		},
		{
			name:           "special_characters_in_app_name",
			projectHomeDir: "/home/test-user",
			containerName:  "my-web-app-v2",
			expected:       "/home/test-user/.config/env/my-web-app-v2.env",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GenerateProjectEnvPath(tt.projectHomeDir, tt.containerName)
			if result != tt.expected {
				t.Errorf("GenerateProjectEnvPath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateProjectDirectoryPermissions(t *testing.T) {
	// 创建临时目录进行测试
	tempDir, err := os.MkdirTemp("", "orbitdeploy_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	orchestrator := &DeploymentOrchestrator{}

	project := &models.Project{
		HomeDir:  tempDir,
		Username: "test-user",
	}

	// 测试目录权限验证
	err = orchestrator.validateProjectDirectoryPermissions(project)
	if err != nil {
		t.Errorf("validateProjectDirectoryPermissions() failed: %v", err)
	}

	// 验证目录是否被创建
	expectedDirs := []string{
		filepath.Join(tempDir, ".config"),
		filepath.Join(tempDir, ".config", "containers"),
		filepath.Join(tempDir, ".config", "containers", "systemd"),
		filepath.Join(tempDir, ".config", "env"),
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Expected directory %s was not created", dir)
		}
	}
}

func TestBackwardCompatibility(t *testing.T) {
	service := NewDeploymentEnvironmentService()

	// 测试旧函数仍然工作
	result := service.GenerateDefaultEnvPath("test-project", "test-app")
	expected := "%h/.config/orbit/test-project/env/test-app.env"

	if result != expected {
		t.Errorf("GenerateDefaultEnvPath() = %v, want %v", result, expected)
	}
}

// 辅助函数
func getHomeDir() string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "/root"
	}
	return homeDir
}
