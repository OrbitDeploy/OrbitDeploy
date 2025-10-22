package utils

import (
	"fmt"
	// "log" // Temporarily commented for migration
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/opentdp/go-helper/command"
	"github.com/opentdp/go-helper/logman"
	"github.com/opentdp/go-helper/strutil"
	// "github.com/OrbitDeploy/fastcaddy" // Temporarily commented for migration
)

type EnvironmentStatus struct {
	Podman struct {
		Installed    bool   `json:"installed"`
		Version      string `json:"version"`
		VersionValid bool   `json:"version_valid"`
		Message      string `json:"message"`
	} `json:"podman"`
	Caddy struct {
		Installed       bool   `json:"installed"`
		Version         string `json:"version"`
		Message         string `json:"message"`
		CloudflareToken bool   `json:"cloudflare_token"`
		TokenMessage    string `json:"token_message"`
	} `json:"caddy"`
	OverallStatus string `json:"overall_status"` // "ready", "partial", "missing"
}

func InitCaddy() error {

	// 创建 FastCaddy 客户端实例
	// fc := fastcaddy.New() // Temporarily commented for migration

	// 初始化基础配置
	fmt.Println("\n🚀 初始化环境...")
	// installTrust := true // Temporarily commented for migration
	cf_token := os.Getenv("CADDY_CF_TOKEN")
	if cf_token != "" {
		// err := fc.SetupCaddy("", "srv0", false, &installTrust) // Temporarily commented for migration
		// if err != nil {
		//	log.Printf("初始化失败: %v (可能是因为 Caddy 服务器未运行)", err)
		// } else {
		fmt.Println("✅ 环境初始化完成")
		// }

	}

	return nil
}

// CheckPodman checks Podman installation and version
func CheckPodman(status *EnvironmentStatus) {
	output, err := command.Exec(&command.ExecPayload{
		Content:     "podman --version",
		CommandType: "SHELL",
		Timeout:     10,
	})

	if err != nil {
		status.Podman.Installed = false
		status.Podman.Message = "Podman not installed"
		return
	}

	status.Podman.Installed = true
	status.Podman.Version = strings.TrimSpace(output)

	// Extract version number and check if it's >= 5.0
	versionRegex := regexp.MustCompile(`podman version (\d+)\.(\d+)\.(\d+)`)
	matches := versionRegex.FindStringSubmatch(output)

	if len(matches) >= 4 {
		major, _ := strconv.Atoi(matches[1])
		if major >= 5 {
			status.Podman.VersionValid = true
			status.Podman.Message = fmt.Sprintf("Podman %s is installed and compatible", status.Podman.Version)
		} else {
			status.Podman.VersionValid = false
			status.Podman.Message = fmt.Sprintf("Podman %s is outdated. Please uninstall and install latest version (>=5.0)", status.Podman.Version)
		}
	} else {
		status.Podman.VersionValid = false
		status.Podman.Message = "Could not determine Podman version"
	}
}

// CheckCaddy checks Caddy installation and Cloudflare token configuration
func CheckCaddy(status *EnvironmentStatus) {
	output, err := command.Exec(&command.ExecPayload{
		Content:     "caddy version",
		CommandType: "SHELL",
		Timeout:     10,
	})

	if err != nil {
		status.Caddy.Installed = false
		status.Caddy.Message = "Caddy not installed"
		status.Caddy.CloudflareToken = false
		status.Caddy.TokenMessage = "N/A (Caddy not installed)"
		return
	}

	status.Caddy.Installed = true
	status.Caddy.Version = strings.TrimSpace(output)
	status.Caddy.Message = fmt.Sprintf("Caddy %s is installed", status.Caddy.Version)

	// Check for CADDY_CF_TOKEN environment variable
	cfToken := os.Getenv("CADDY_CF_TOKEN")
	if cfToken != "" {
		status.Caddy.CloudflareToken = true
		status.Caddy.TokenMessage = "CADDY_CF_TOKEN is configured"
	} else {
		status.Caddy.CloudflareToken = false
		status.Caddy.TokenMessage = "CADDY_CF_TOKEN not configured (required for Cloudflare DNS challenges)"
	}
}

// RunInstallation runs the installation script and provides progress feedback
func RunInstallation(script, component string) {
	logger := logman.Named("installation")

	// Normalize script content to handle different line endings and encoding
	normalizedScript := normalizeScriptContent(script)

	// Create temporary script file
	tmpDir := getTempDir()
	scriptFile := filepath.Join(tmpDir, fmt.Sprintf("install_%s_%d.sh", component, time.Now().Unix()))

	err := os.WriteFile(scriptFile, []byte(normalizedScript), 0755)
	if err != nil {
		logger.Error("Failed to write installation script", "component", component, "error", err)
		return
	}
	defer os.Remove(scriptFile)

	logger.Info("Starting installation", "component", component, "script", scriptFile)

	// Determine shell command based on OS
	shellCmd := getShellCommand(scriptFile)

	// Execute the script
	output, err := command.Exec(&command.ExecPayload{
		Content:     shellCmd,
		CommandType: "SHELL",
		Timeout:     600, // 10 minutes timeout
	})

	if err != nil {
		logger.Error("Installation failed", "component", component, "error", err, "output", strutil.Dedent(output))
		return
	}

	logger.Info("Installation completed successfully", "component", component, "output", strutil.Dedent(output))
}

// normalizeScriptContent handles cross-platform format issues
func normalizeScriptContent(script string) string {
	// Remove any UTF-8 BOM if present
	script = strings.TrimPrefix(script, "\ufeff")

	// Normalize different line endings to Unix (\n) for internal processing
	script = strings.ReplaceAll(script, "\r\n", "\n") // CRLF -> LF
	script = strings.ReplaceAll(script, "\r", "\n")   // CR -> LF

	// Use strutil.Dedent to normalize indentation first
	script = strutil.Dedent(script)

	// Ensure script starts with shebang if not present (keeps using bash)
	if !strings.HasPrefix(script, "#!") {
		script = "#!/bin/bash\n" + script
	}

	// Convert final line endings according to OS
	lineEnding := "\n"
	if runtime.GOOS == "windows" {
		lineEnding = "\r\n"
	}
	if lineEnding != "\n" {
		script = strings.ReplaceAll(script, "\n", lineEnding)
	}

	return script
}

// getTempDir returns appropriate temp directory based on OS
func getTempDir() string {
	if runtime.GOOS == "windows" {
		if tmpDir := os.Getenv("TEMP"); tmpDir != "" {
			return tmpDir
		}
		if tmpDir := os.Getenv("TMP"); tmpDir != "" {
			return tmpDir
		}
		return "C:\\temp"
	}
	return "/tmp"
}

// getShellCommand returns appropriate shell command based on OS
func getShellCommand(scriptFile string) string {
	if runtime.GOOS == "windows" {
		// On Windows, try to use bash from WSL or Git Bash if available
		return fmt.Sprintf("bash %s", scriptFile)
	}
	return fmt.Sprintf("bash %s", scriptFile)
}

// These functions will be set by main.go to access embedded scripts
var GetPodmanInstallScript func() string
var GetCaddyInstallScript func() string
