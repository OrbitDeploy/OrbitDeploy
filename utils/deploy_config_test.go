package utils

import (
	"testing"
)

func TestParseDeployConfig(t *testing.T) {
	tomlContent := `
app = 'test-app'
primary_region = 'us-east'
kill_signal = 'SIGTERM'

[build]
  remote_image = 'nginx:latest'
  tag_strategy = 'latest'

[env]
  PORT = '8080'
  APP_ENV = 'production'

[[mounts]]
  source = '/data'
  destination = '/app/data'
  auto_extend_size_threshold = 80
  auto_extend_size_increment = '1GB'
  auto_extend_size_limit = '10GB'

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = 'stop'
  auto_start_machines = true
  min_machines_running = 0
  processes = ['app']

  [http_service.concurrency]
    type = 'connections'
    hard_limit = 1000
    soft_limit = 1000

[[vm]]
  size = 'shared-cpu-1x'
`

	config, err := ParseDeployConfig(tomlContent)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if config.App != "test-app" {
		t.Errorf("Expected app name 'test-app', got '%s'", config.App)
	}

	if config.Build.RemoteImage != "nginx:latest" {
		t.Errorf("Expected remote image 'nginx:latest', got '%s'", config.Build.RemoteImage)
	}

	if config.HTTPService.InternalPort != 8080 {
		t.Errorf("Expected internal port 8080, got %d", config.HTTPService.InternalPort)
	}

	if len(config.Mounts) != 1 {
		t.Errorf("Expected 1 mount, got %d", len(config.Mounts))
	}

	if config.Env["PORT"] != "8080" {
		t.Errorf("Expected PORT environment variable '8080', got '%s'", config.Env["PORT"])
	}
}

func TestValidateDeployConfig(t *testing.T) {
	// Test valid config
	validConfig := &DeployConfigInfo{
		App: "test-app",
		Build: BuildConfig{
			RemoteImage: "nginx:latest",
		},
		HTTPService: HTTPServiceConfig{
			InternalPort: 8080,
		},
		Mounts: []MountConfig{
			{
				Source:      "/data",
				Destination: "/app/data",
			},
		},
		Domains: []DomainConfig{
			{
				Name:    "example.com",
				Primary: true,
			},
		},
	}

	if err := ValidateDeployConfig(validConfig); err != nil {
		t.Errorf("Expected valid config to pass validation, got error: %v", err)
	}

	// Test invalid config - missing app name
	invalidConfig := &DeployConfigInfo{
		Build: BuildConfig{
			RemoteImage: "nginx:latest",
		},
		HTTPService: HTTPServiceConfig{
			InternalPort: 8080,
		},
	}

	if err := ValidateDeployConfig(invalidConfig); err == nil {
		t.Error("Expected invalid config to fail validation")
	}

	// Test invalid config - missing image
	invalidConfig2 := &DeployConfigInfo{
		App: "test-app",
		HTTPService: HTTPServiceConfig{
			InternalPort: 8080,
		},
	}

	if err := ValidateDeployConfig(invalidConfig2); err == nil {
		t.Error("Expected config without image to fail validation")
	}

	// Test invalid config - invalid port
	invalidConfig3 := &DeployConfigInfo{
		App: "test-app",
		Build: BuildConfig{
			RemoteImage: "nginx:latest",
		},
		HTTPService: HTTPServiceConfig{
			InternalPort: 0,
		},
	}

	if err := ValidateDeployConfig(invalidConfig3); err == nil {
		t.Error("Expected config with invalid port to fail validation")
	}
}

func TestGenerateQuadletFromDeployConfig(t *testing.T) {
	config := &DeployConfigInfo{
		App: "test-app",
		Build: BuildConfig{
			RemoteImage: "nginx:latest",
		},
		Env: map[string]string{
			"PORT":    "8080",
			"APP_ENV": "production",
		},
		HTTPService: HTTPServiceConfig{
			InternalPort: 8080,
		},
		Mounts: []MountConfig{
			{
				Source:      "/data",
				Destination: "/app/data",
			},
		},
		KillSignal: "SIGTERM",
	}

	quadlet, err := GenerateQuadletFromDeployConfig(config)
	if err != nil {
		t.Fatalf("Failed to generate Quadlet: %v", err)
	}

	// Check that the generated Quadlet contains expected content
	expectedStrings := []string{
		"[Unit]",
		"[Container]", 
		"[Install]",
		"Image=nginx:latest",
		"ContainerName=test-app",
		"Environment=PORT=8080",
		"Environment=APP_ENV=production",
		"PublishPort=8080:8080",
		"Volume=/data:/app/data",
		"WantedBy=default.target",
	}

	for _, expected := range expectedStrings {
		if !contains(quadlet, expected) {
			t.Errorf("Expected Quadlet to contain '%s', but it was missing. Generated Quadlet:\n%s", expected, quadlet)
		}
	}
}

func contains(text, substr string) bool {
	return len(text) >= len(substr) && (text == substr || 
		(len(text) > len(substr) && (text[:len(substr)] == substr || 
		text[len(text)-len(substr):] == substr || 
		containsSubstring(text, substr))))
}

func containsSubstring(text, substr string) bool {
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}