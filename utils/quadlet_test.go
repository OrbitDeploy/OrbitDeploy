package utils

import (
	"strings"
	"testing"
)

func TestParseQuadletFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectImage string
		expectVols  int
		expectEnv   string
		expectEnvironmentVars int
		expectLabels int
	}{
		{
			name: "linkding example",
			content: `[Unit]
Description=Linkding Bookmark Manager
After=network-online.target

[Container]
Image=ghcr.io/sissbruecker/linkding:latest
PublishPort=9092:9090
Volume=/var/linkding/data:/etc/linkding/data
EnvironmentFile=/var/linkding/config/linkding.env

[Install]
WantedBy=default.target`,
			expectImage: "ghcr.io/sissbruecker/linkding:latest",
			expectVols:  1,
			expectEnv:   "/var/linkding/config/linkding.env",
			expectEnvironmentVars: 0,
			expectLabels: 0,
		},
		{
			name: "multiple volumes",
			content: `[Unit]
Description=Test Container

[Container]
Image=nginx:latest
Volume=/host/path1:/container/path1
Volume=/host/path2:/container/path2:ro
Volume=named_volume:/container/path3

[Install]
WantedBy=default.target`,
			expectImage: "nginx:latest",
			expectVols:  2, // named_volume should be excluded
			expectEnv:   "",
			expectEnvironmentVars: 0,
			expectLabels: 0,
		},
		{
			name: "new features test",
			content: `[Unit]
Description=Test Container with new features

[Container]
Image=nginx:latest
Environment=MY_VAR=value
Environment=HOST_VAR
Label=version=1.0
Label=environment=prod
ExitPolicy=always
Policy=always

[Network]
InterfaceName=eth0

[Install]
WantedBy=default.target`,
			expectImage: "nginx:latest",
			expectVols:  0,
			expectEnv:   "",
			expectEnvironmentVars: 2, // MY_VAR and HOST_VAR
			expectLabels: 2, // version and environment
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ParseQuadletFile(tt.content)
			if err != nil {
				t.Fatalf("ParseQuadletFile failed: %v", err)
			}

			if info.Image != tt.expectImage {
				t.Errorf("Expected image %s, got %s", tt.expectImage, info.Image)
			}

			if len(info.Volumes) != tt.expectVols {
				t.Errorf("Expected %d volumes, got %d: %v", tt.expectVols, len(info.Volumes), info.Volumes)
			}

			if info.EnvFile != tt.expectEnv {
				t.Errorf("Expected env file %s, got %s", tt.expectEnv, info.EnvFile)
			}

			if len(info.Environment) != tt.expectEnvironmentVars {
				t.Errorf("Expected %d environment variables, got %d: %v", tt.expectEnvironmentVars, len(info.Environment), info.Environment)
			}

			if len(info.Labels) != tt.expectLabels {
				t.Errorf("Expected %d labels, got %d: %v", tt.expectLabels, len(info.Labels), info.Labels)
			}

			// Test specific values for the new features test
			if tt.name == "new features test" {
				if info.Environment["MY_VAR"] != "value" {
					t.Errorf("Expected MY_VAR=value, got %s", info.Environment["MY_VAR"])
				}
				if val, exists := info.Environment["HOST_VAR"]; !exists || val != "" {
					t.Errorf("Expected HOST_VAR with empty value (from host), got %s (exists: %v)", val, exists)
				}
				if info.Labels["version"] != "1.0" {
					t.Errorf("Expected version=1.0, got %s", info.Labels["version"])
				}
				if info.ExitPolicy != "always" {
					t.Errorf("Expected ExitPolicy=always, got %s", info.ExitPolicy)
				}
				if info.Policy != "always" {
					t.Errorf("Expected Policy=always, got %s", info.Policy)
				}
				if info.InterfaceName != "eth0" {
					t.Errorf("Expected InterfaceName=eth0, got %s", info.InterfaceName)
				}
			}
		})
	}
}

func TestExtractHostPathFromVolume(t *testing.T) {
	tests := []struct {
		volumeSpec string
		expected   string
	}{
		{"/var/linkding/data:/etc/linkding/data", "/var/linkding/data"},
		{"/host/path:/container/path:ro", "/host/path"},
		{"named_volume:/container/path", ""}, // named volume
		{"/absolute/path:/container", "/absolute/path"},
		{"relative/path:/container", ""}, // relative path
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.volumeSpec, func(t *testing.T) {
			result := extractHostPathFromVolume(tt.volumeSpec)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetHostDirectoriesFromQuadlet(t *testing.T) {
	content := `[Unit]
Description=Test Container

[Container]
Image=nginx:latest
Volume=/var/app/data:/data
Volume=/var/app/logs:/logs
EnvironmentFile=/etc/app/config.env

[Install]
WantedBy=default.target`

	dirs, err := GetHostDirectoriesFromQuadlet(content)
	if err != nil {
		t.Fatalf("GetHostDirectoriesFromQuadlet failed: %v", err)
	}

	expected := []string{"/var/app/data", "/var/app/logs", "/etc/app"}
	if len(dirs) != len(expected) {
		t.Fatalf("Expected %d directories, got %d: %v", len(expected), len(dirs), dirs)
	}

	for i, dir := range dirs {
		if dir != expected[i] {
			t.Errorf("Expected directory %s, got %s", expected[i], dir)
		}
	}
}

func TestExtractHostPortFromQuadlet(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectPort  int
		expectError bool
	}{
		{
			name: "valid PublishPort",
			content: `[Unit]
Description=Test Container

[Container]
Image=nginx:latest
PublishPort=9092:9090

[Install]
WantedBy=default.target`,
			expectPort:  9092,
			expectError: false,
		},
		{
			name: "different port format",
			content: `[Unit]
Description=Test Container

[Container]
Image=nginx:latest
PublishPort=8080:80

[Install]
WantedBy=default.target`,
			expectPort:  8080,
			expectError: false,
		},
		{
			name: "no PublishPort",
			content: `[Unit]
Description=Test Container

[Container]
Image=nginx:latest

[Install]
WantedBy=default.target`,
			expectPort:  0,
			expectError: true,
		},
		{
			name: "invalid PublishPort format",
			content: `[Unit]
Description=Test Container

[Container]
Image=nginx:latest
PublishPort=8080

[Install]
WantedBy=default.target`,
			expectPort:  0,
			expectError: true,
		},
		{
			name: "invalid port number",
			content: `[Unit]
Description=Test Container

[Container]
Image=nginx:latest
PublishPort=abc:80

[Install]
WantedBy=default.target`,
			expectPort:  0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port, err := ExtractHostPortFromQuadlet(tt.content)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if port != tt.expectPort {
					t.Errorf("Expected port %d, got %d", tt.expectPort, port)
				}
			}
		})
	}
}

func TestGenerateProjectQuadlet(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		description string
		imageName   string
		publishPort int
		expectError bool
		checkFunc   func(t *testing.T, content string)
	}{
		{
			name:        "basic project generation",
			projectName: "conflux",
			description: "Conflux Node",
			imageName:   "",
			publishPort: 8545,
			expectError: false,
			checkFunc: func(t *testing.T, content string) {
				// Check that %h is preserved for systemd (not interpreted by Go)
				if !strings.Contains(content, "%h/.config/conflux.env") {
					t.Errorf("Expected %%h/.config/conflux.env in EnvironmentFile, got content: %s", content)
				}
				
				// Check that the problematic %!h pattern does not exist
				if strings.Contains(content, "%!h(") {
					t.Errorf("Found problematic %%!h pattern in content: %s", content)
				}
				
				// Check basic structure
				if !strings.Contains(content, "[Unit]") {
					t.Errorf("Missing [Unit] section")
				}
				if !strings.Contains(content, "[Container]") {
					t.Errorf("Missing [Container] section")
				}
				if !strings.Contains(content, "[Install]") {
					t.Errorf("Missing [Install] section")
				}
				
				// Check that description appears
				if !strings.Contains(content, "Conflux Node") {
					t.Errorf("Description not found in content")
				}
				
				// Check default image format
				if !strings.Contains(content, "localhost/conflux:latest") {
					t.Errorf("Expected default image format, got content: %s", content)
				}
			},
		},
		{
			name:        "custom image name",
			projectName: "myapp",
			description: "My Application",
			imageName:   "custom/image:v1.0",
			publishPort: 3000,
			expectError: false,
			checkFunc: func(t *testing.T, content string) {
				if !strings.Contains(content, "custom/image:v1.0") {
					t.Errorf("Custom image name not found")
				}
				if !strings.Contains(content, "%h/.config/myapp.env") {
					t.Errorf("Expected %%h/.config/myapp.env in EnvironmentFile")
				}
			},
		},
		{
			name:        "empty project name",
			projectName: "",
			description: "Test",
			imageName:   "",
			publishPort: 8080,
			expectError: true,
		},
		{
			name:        "invalid port - too low",
			projectName: "test",
			description: "Test",
			imageName:   "",
			publishPort: 0,
			expectError: true,
		},
		{
			name:        "invalid port - too high", 
			projectName: "test",
			description: "Test",
			imageName:   "",
			publishPort: 70000,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := GenerateProjectQuadlet(tt.projectName, tt.description, tt.imageName, tt.publishPort)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tt.checkFunc != nil {
					tt.checkFunc(t, content)
				}
			}
		})
	}
}