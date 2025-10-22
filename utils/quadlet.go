package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// QuadletInfo contains parsed information from a Quadlet file
type QuadletInfo struct {
	Image         string            `json:"image"`
	Volumes       []string          `json:"volumes"`        // Host paths that need to exist
	EnvFile       string            `json:"env_file"`       // Environment file path
	PublishPort   string            `json:"publish_port"`   // PublishPort directive (e.g., "9092:9090")
	Environment   map[string]string `json:"environment"`    // Environment variables (supports values and no-value vars)
	Labels        map[string]string `json:"labels"`         // Pod labels (Label= key for .pod units)
	ExitPolicy    string            `json:"exit_policy"`    // Pod exit policy (ExitPolicy= key for .pod units)
	Policy        string            `json:"policy"`         // Image pull policy (Policy= key for .image units)
	InterfaceName string            `json:"interface_name"` // Network interface name (InterfaceName= key for .network units)
}

// ParseQuadletFile parses a Quadlet file content and extracts relevant information
func ParseQuadletFile(content string) (*QuadletInfo, error) {
	info := &QuadletInfo{
		Volumes:     make([]string, 0),
		Environment: make(map[string]string),
		Labels:      make(map[string]string),
	}

	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse Image directive
		if strings.HasPrefix(line, "Image=") {
			info.Image = strings.TrimPrefix(line, "Image=")
			continue
		}

		// Parse Volume directive - extract host path
		if strings.HasPrefix(line, "Volume=") {
			volumeSpec := strings.TrimPrefix(line, "Volume=")
			hostPath := extractHostPathFromVolume(volumeSpec)
			if hostPath != "" {
				info.Volumes = append(info.Volumes, hostPath)
			}
			continue
		}

		// Parse EnvironmentFile directive
		if strings.HasPrefix(line, "EnvironmentFile=") {
			envPath := strings.TrimPrefix(line, "EnvironmentFile=")
			info.EnvFile = envPath
			continue
		}

		// Parse PublishPort directive
		if strings.HasPrefix(line, "PublishPort=") {
			publishPort := strings.TrimPrefix(line, "PublishPort=")
			info.PublishPort = publishPort
			continue
		}

		// Parse Environment directive (supports variables without values)
		if strings.HasPrefix(line, "Environment=") {
			envSpec := strings.TrimPrefix(line, "Environment=")
			if strings.Contains(envSpec, "=") {
				// Variable with value: VAR=value
				parts := strings.SplitN(envSpec, "=", 2)
				info.Environment[parts[0]] = parts[1]
			} else {
				// Variable without value: VAR (retrieves from host)
				info.Environment[envSpec] = ""
			}
			continue
		}

		// Parse Label directive (for .pod units)
		if strings.HasPrefix(line, "Label=") {
			labelSpec := strings.TrimPrefix(line, "Label=")
			if strings.Contains(labelSpec, "=") {
				parts := strings.SplitN(labelSpec, "=", 2)
				info.Labels[parts[0]] = parts[1]
			}
			continue
		}

		// Parse ExitPolicy directive (for .pod units)
		if strings.HasPrefix(line, "ExitPolicy=") {
			info.ExitPolicy = strings.TrimPrefix(line, "ExitPolicy=")
			continue
		}

		// Parse Policy directive (for .image units)
		if strings.HasPrefix(line, "Policy=") {
			info.Policy = strings.TrimPrefix(line, "Policy=")
			continue
		}

		// Parse InterfaceName directive (for .network units)
		if strings.HasPrefix(line, "InterfaceName=") {
			info.InterfaceName = strings.TrimPrefix(line, "InterfaceName=")
			continue
		}
	}

	return info, nil
}

// extractHostPathFromVolume extracts the host path from a volume specification
// Examples:
//
//	"/var/linkding/data:/etc/linkding/data" -> "/var/linkding/data"
//	"/host/path:/container/path:ro" -> "/host/path"
//	"volume_name:/container/path" -> "" (named volume, no host path to create)
func extractHostPathFromVolume(volumeSpec string) string {
	parts := strings.Split(volumeSpec, ":")
	if len(parts) < 2 {
		return ""
	}

	hostPath := parts[0]

	// Check if this is an absolute path (starts with /)
	// Named volumes don't start with / and don't need directory creation
	if !strings.HasPrefix(hostPath, "/") {
		return ""
	}

	return hostPath
}

// GetHostDirectoriesFromQuadlet returns all host directories that need to be created
func GetHostDirectoriesFromQuadlet(content string) ([]string, error) {
	info, err := ParseQuadletFile(content)
	if err != nil {
		return nil, err
	}

	directories := make([]string, 0)

	// Add volume directories
	directories = append(directories, info.Volumes...)

	// Add environment file directory if specified
	if info.EnvFile != "" && strings.HasPrefix(info.EnvFile, "/") {
		// Extract directory from file path
		envDir := extractDirFromPath(info.EnvFile)
		if envDir != "" {
			directories = append(directories, envDir)
		}
	}

	return directories, nil
}

// extractDirFromPath extracts the directory path from a file path
func extractDirFromPath(filePath string) string {
	lastSlash := strings.LastIndex(filePath, "/")
	if lastSlash <= 0 {
		return ""
	}
	return filePath[:lastSlash]
}

// ExtractImageFromQuadlet extracts the container image from Quadlet content
func ExtractImageFromQuadlet(content string) (string, error) {
	info, err := ParseQuadletFile(content)
	if err != nil {
		return "", err
	}
	return info.Image, nil
}

// ValidateQuadletContent performs basic validation on Quadlet file content
func ValidateQuadletContent(content string) error {
	// Check for required sections
	hasUnit := strings.Contains(content, "[Unit]")
	hasContainer := strings.Contains(content, "[Container]")
	hasInstall := strings.Contains(content, "[Install]")

	if !hasUnit || !hasContainer || !hasInstall {
		return fmt.Errorf("quadlet file must contain [Unit], [Container], and [Install] sections")
	}

	return nil
}

// ExtractHostPortFromQuadlet extracts the host port from PublishPort directive in Quadlet content
// PublishPort format: "host_port:container_port" (e.g., "9092:9090")
// Returns the host port (the first number before the colon)
func ExtractHostPortFromQuadlet(content string) (int, error) {
	info, err := ParseQuadletFile(content)
	if err != nil {
		return 0, err
	}

	if info.PublishPort == "" {
		return 0, fmt.Errorf("no PublishPort directive found in Quadlet file")
	}

	// Parse the PublishPort format: "host_port:container_port"
	parts := strings.Split(info.PublishPort, ":")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid PublishPort format: %s", info.PublishPort)
	}

	hostPort := strings.TrimSpace(parts[0])
	port, err := strconv.Atoi(hostPort)
	if err != nil {
		return 0, fmt.Errorf("invalid host port in PublishPort: %s", hostPort)
	}

	return port, nil
}

// ListQuadlets uses 'podman quadlet list' to get installed Quadlets
func ListQuadlets() ([]string, error) {
	output, err := executeQuadletCommand("list")
	if err != nil {
		return nil, fmt.Errorf("failed to list quadlets: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	quadlets := make([]string, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			// Extract container name from quadlet filename
			if strings.HasSuffix(line, ".container") {
				containerName := strings.TrimSuffix(line, ".container")
				quadlets = append(quadlets, containerName)
			}
		}
	}

	return quadlets, nil
}

// InstallQuadlet uses 'podman quadlet install' to install a Quadlet
func InstallQuadlet(name string, content string) error {
	// Write quadlet content to a temporary file
	tmpFile, err := writeQuadletToTempFile(name, content)
	if err != nil {
		return fmt.Errorf("failed to create temporary quadlet file: %w", err)
	}
	defer os.Remove(tmpFile)

	// Use podman quadlet install
	_, err = executeQuadletCommand("install", tmpFile)
	if err != nil {
		return fmt.Errorf("failed to install quadlet %s: %w", name, err)
	}

	return nil
}

// RemoveQuadlet uses 'podman quadlet rm' to remove a Quadlet
func RemoveQuadlet(name string) error {
	_, err := executeQuadletCommand("rm", name)
	if err != nil {
		return fmt.Errorf("failed to remove quadlet %s: %w", name, err)
	}
	return nil
}

// PrintQuadlet uses 'podman quadlet print' to get Quadlet content
func PrintQuadlet(name string) (string, error) {
	output, err := executeQuadletCommand("print", name)
	if err != nil {
		return "", fmt.Errorf("failed to print quadlet %s: %w", name, err)
	}
	return output, nil
}

// executeQuadletCommand executes a podman quadlet command
func executeQuadletCommand(action string, args ...string) (string, error) {
	cmdArgs := []string{"quadlet", action}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command("podman", cmdArgs...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// writeQuadletToTempFile writes quadlet content to a temporary file
func writeQuadletToTempFile(name, content string) (string, error) {
	tmpFile := filepath.Join("/tmp", fmt.Sprintf("%s.container", name))
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		return "", err
	}
	return tmpFile, nil
}

// ImageUserInfo contains user information extracted from a container image
type ImageUserInfo struct {
	UID     string `json:"uid"`
	GID     string `json:"gid"`
	UserStr string `json:"user_str"` // Original user string from image
}

// InspectImageUser uses podman image inspect to get user information from an image
func InspectImageUser(imageName string) (*ImageUserInfo, error) {
	if imageName == "" {
		return nil, fmt.Errorf("image name cannot be empty")
	}

	// Execute podman image inspect command
	cmd := exec.Command("podman", "image", "inspect", "--format", "{{.Config.User}}", imageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to inspect image %s: %w (output: %s)", imageName, err, string(output))
	}

	userStr := strings.TrimSpace(string(output))
	info := &ImageUserInfo{
		UserStr: userStr,
	}

	// Parse the user string
	if userStr == "" || userStr == "<no value>" {
		// No user specified, defaults to root
		info.UID = "0"
		info.GID = "0"
	} else if strings.Contains(userStr, ":") {
		// Format: UID:GID
		parts := strings.Split(userStr, ":")
		if len(parts) >= 2 {
			info.UID = parts[0]
			info.GID = parts[1]
		} else {
			return nil, fmt.Errorf("invalid UID:GID format: %s", userStr)
		}
	} else if isNumeric(userStr) {
		// Format: UID only
		info.UID = userStr
		// Need to get GID by running a temporary container
		gid, err := getGIDFromContainer(imageName, userStr)
		if err != nil {
			// Fallback to same as UID
			info.GID = userStr
		} else {
			info.GID = gid
		}
	} else {
		// Format: username
		// Need to get UID/GID by running a temporary container
		uid, gid, err := getUIDGIDFromContainer(imageName, userStr)
		if err != nil {
			return nil, fmt.Errorf("failed to get UID/GID for user %s: %w", userStr, err)
		}
		info.UID = uid
		info.GID = gid
	}

	return info, nil
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// getGIDFromContainer gets GID for a given UID by running a temporary container
func getGIDFromContainer(imageName, uid string) (string, error) {
	cmd := exec.Command("podman", "run", "--rm", imageName, "id", "-g", uid)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get GID for UID %s: %w", uid, err)
	}
	return strings.TrimSpace(string(output)), nil
}

// getUIDGIDFromContainer gets UID and GID for a given username by running a temporary container
func getUIDGIDFromContainer(imageName, username string) (string, string, error) {
	// Get UID
	uidCmd := exec.Command("podman", "run", "--rm", imageName, "id", "-u", username)
	uidOutput, err := uidCmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to get UID for user %s: %w", username, err)
	}
	uid := strings.TrimSpace(string(uidOutput))

	// Get GID
	gidCmd := exec.Command("podman", "run", "--rm", imageName, "id", "-g", username)
	gidOutput, err := gidCmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to get GID for user %s: %w", username, err)
	}
	gid := strings.TrimSpace(string(gidOutput))

	return uid, gid, nil
}

// AddPermissionsToQuadlet adds Service section with permission-setting commands to Quadlet content
func AddPermissionsToQuadlet(quadletContent string, userInfo *ImageUserInfo, hostPaths []string) (string, error) {
	if userInfo == nil || len(hostPaths) == 0 {
		// No user info or no host paths, return original content
		return quadletContent, nil
	}

	// Check if [Service] section already exists
	if strings.Contains(quadletContent, "[Service]") {
		// [Service] section already exists, don't modify to avoid conflicts
		return quadletContent, nil
	}

	// Build the Service section
	serviceSection := "[Service]\n"

	// Add ExecStartPre commands for each host path
	for _, hostPath := range hostPaths {
		if hostPath != "" {
			// Create directory if it doesn't exist
			serviceSection += fmt.Sprintf("ExecStartPre=-/bin/mkdir -p %s\n", hostPath)
			// Set ownership
			serviceSection += fmt.Sprintf("ExecStartPre=-/bin/chown %s:%s %s\n", userInfo.UID, userInfo.GID, hostPath)
		}
	}

	serviceSection += "\n"

	// Find the position to insert the Service section
	// Insert before [Container] section if it exists
	containerPos := strings.Index(quadletContent, "[Container]")
	if containerPos != -1 {
		// Insert before [Container]
		return quadletContent[:containerPos] + serviceSection + quadletContent[containerPos:], nil
	}

	// If no [Container] section found, append at the end
	return quadletContent + "\n" + serviceSection, nil
}

// GenerateProjectQuadlet 生成项目的Quadlet文件内容
// GenerateProjectQuadlet 根据项目信息生成 Quadlet 配置文件内容
// func GenerateProjectQuadlet(projectName, description, imageName string, publishPort int) (string, error) {
// 	if projectName == "" {
// 		return "", fmt.Errorf("项目名称不能为空")
// 	}

// 	if publishPort <= 0 || publishPort > 65535 {
// 		return "", fmt.Errorf("外部端口必须在1-65535范围内")
// 	}

// 	if containerPort <= 0 || containerPort > 65535 {
// 		return "", fmt.Errorf("容器端口必须在1-65535范围内")
// 	}

// 	// 生成一个确定的系统端口，避免每次都变
// 	// 这里使用简单的映射，你可以根据需要实现更复杂的逻辑
// 	systemPort := 10000 + (publishPort % 15000)

// 	if imageName == "" {
// 		imageName = fmt.Sprintf("localhost/%s:latest", projectName)
// 	}

// 	// --- 这里是关键的修改 ---
// 	// 将 EnvironmentFile 中的 `~` 替换为 `%h`
// 	quadletTemplate := `[Unit]
// Description=%s
// After=network-online.target

// [Container]
// Image=%s
// PublishPort=%d:%d
// EnvironmentFile=%h/.config/%s/%s.env

// [Install]
// WantedBy=default.target`

// 	quadletContent := fmt.Sprintf(
// 		quadletTemplate,
// 		description, // 1st
// 		imageName,   // 2nd
// 		systemPort,  // 3rd
// 		projectName, // 5th
// 		projectName, // 6th
// 	)

// 	return quadletContent, nil
// }

// GenerateProjectQuadlet 生成项目的Quadlet文件内容
func GenerateProjectQuadlet(projectName, description, imageName string, publishPort int) (string, error) {
	if projectName == "" {
		return "", fmt.Errorf("项目名称不能为空")
	}

	if publishPort <= 0 || publishPort > 65535 {
		return "", fmt.Errorf("端口必须在1-65535范围内")
	}

	// 生成随机的系统端口（10000-25000之间）
	systemPort := 10000 + (publishPort % 15000) // 简单的伪随机生成

	// 如果镜像名称为空，使用默认的项目镜像格式
	if imageName == "" {
		imageName = fmt.Sprintf("localhost/%s:latest", projectName)
	}

	// 使用需求中指定的模板格式
	quadletTemplate := `
		[Unit]
		Description=%s
		After=network-online.target

		[Container]
		Image=%s
		PublishPort=%d:%d
		EnvironmentFile=%%h/.config/%s.env
		[Install]
		WantedBy=default.target`

	quadletContent := fmt.Sprintf(quadletTemplate, description, imageName, systemPort, publishPort, projectName)

	return quadletContent, nil
}
