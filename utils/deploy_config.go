package utils

import (
	"fmt"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// DeployConfigInfo contains parsed information from go-web-deploy.toml
type DeployConfigInfo struct {
	App           string                    `toml:"app"`
	PrimaryRegion string                    `toml:"primary_region"`
	KillSignal    string                    `toml:"kill_signal"`
	Build         BuildConfig               `toml:"build"`
	Env           map[string]string         `toml:"env"`
	Mounts        []MountConfig             `toml:"mounts"`
	HTTPService   HTTPServiceConfig         `toml:"http_service"`
	VM            []VMConfig                `toml:"vm"`
	HealthCheck   *HealthCheckConfig        `toml:"health_check,omitempty"`
	Domains       []DomainConfig            `toml:"domains,omitempty"`
	Deploy        *DeployStrategyConfig     `toml:"deploy,omitempty"`
	Logging       *LoggingConfig            `toml:"logging,omitempty"`
	Security      *SecurityConfig           `toml:"security,omitempty"`
	Network       *NetworkConfig            `toml:"network,omitempty"`
	Resources     *ResourceConfig           `toml:"resources,omitempty"`
}

// BuildConfig represents build configuration
type BuildConfig struct {
	LocalImage   string `toml:"local_image"`
	RemoteImage  string `toml:"remote_image"`
	TagStrategy  string `toml:"tag_strategy"`
}

// MountConfig represents mount configuration  
type MountConfig struct {
	Source                   string `toml:"source"`
	Destination              string `toml:"destination"`
	AutoExtendSizeThreshold  int    `toml:"auto_extend_size_threshold"`
	AutoExtendSizeIncrement  string `toml:"auto_extend_size_increment"`
	AutoExtendSizeLimit      string `toml:"auto_extend_size_limit"`
}

// HTTPServiceConfig represents HTTP service configuration
type HTTPServiceConfig struct {
	InternalPort       int                `toml:"internal_port"`
	ForceHTTPS         bool               `toml:"force_https"`
	AutoStopMachines   string             `toml:"auto_stop_machines"`
	AutoStartMachines  bool               `toml:"auto_start_machines"`
	MinMachinesRunning int                `toml:"min_machines_running"`
	Processes          []string           `toml:"processes"`
	Concurrency        ConcurrencyConfig  `toml:"concurrency"`
}

// ConcurrencyConfig represents concurrency configuration
type ConcurrencyConfig struct {
	Type      string `toml:"type"`
	HardLimit int    `toml:"hard_limit"`
	SoftLimit int    `toml:"soft_limit"`
}

// VMConfig represents VM configuration
type VMConfig struct {
	Size   string `toml:"size"`
	CPUs   *int   `toml:"cpus,omitempty"`
	Memory string `toml:"memory,omitempty"`
}

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	Path        string `toml:"path"`
	Interval    string `toml:"interval"`
	Timeout     string `toml:"timeout"`
	Retries     int    `toml:"retries"`
	GracePeriod string `toml:"grace_period"`
}

// DomainConfig represents domain configuration
type DomainConfig struct {
	Name    string `toml:"name"`
	Primary bool   `toml:"primary"`
}

// DeployStrategyConfig represents deployment strategy configuration
type DeployStrategyConfig struct {
	Strategy       string `toml:"strategy"`
	MaxUnavailable int    `toml:"max_unavailable"`
	MaxSurge       int    `toml:"max_surge"`
	Timeout        string `toml:"timeout"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Driver  string                 `toml:"driver"`
	Level   string                 `toml:"level"`
	Options map[string]string      `toml:"options"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	DisableSecurityLabel bool     `toml:"disable_security_label"`
	AddCapabilities      []string `toml:"add_capabilities"`
	DropCapabilities     []string `toml:"drop_capabilities"`
	RunAsNonRoot         bool     `toml:"run_as_non_root"`
	UserID               *int     `toml:"user_id,omitempty"`
	GroupID              *int     `toml:"group_id,omitempty"`
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	Mode           string   `toml:"mode"`
	CustomNetworks []string `toml:"custom_networks,omitempty"`
	DNS            []string `toml:"dns,omitempty"`
}

// ResourceConfig represents resource configuration
type ResourceConfig struct {
	CPULimit      string `toml:"cpu_limit"`
	MemoryLimit   string `toml:"memory_limit"`
	CPURequest    string `toml:"cpu_request"`
	MemoryRequest string `toml:"memory_request"`
}

// ParseDeployConfig parses a go-web-deploy.toml configuration file
func ParseDeployConfig(content string) (*DeployConfigInfo, error) {
	var config DeployConfigInfo
	
	if err := toml.Unmarshal([]byte(content), &config); err != nil {
		return nil, fmt.Errorf("failed to parse deploy config: %w", err)
	}
	
	// Validate required fields
	if config.App == "" {
		return nil, fmt.Errorf("app name is required")
	}
	
	// Set defaults
	if config.PrimaryRegion == "" {
		config.PrimaryRegion = "default"
	}
	if config.KillSignal == "" {
		config.KillSignal = "SIGTERM"
	}
	if config.Build.TagStrategy == "" {
		config.Build.TagStrategy = "latest"
	}
	
	return &config, nil
}

// GenerateQuadletFromDeployConfig converts a DeployConfigInfo to Quadlet format
func GenerateQuadletFromDeployConfig(config *DeployConfigInfo) (string, error) {
	var quadlet strings.Builder
	
	// [Unit] section
	quadlet.WriteString("[Unit]\n")
	quadlet.WriteString(fmt.Sprintf("Description=%s container\n", config.App))
	quadlet.WriteString("After=network.target\n")
	quadlet.WriteString("\n")
	
	// [Container] section
	quadlet.WriteString("[Container]\n")
	
	// Image configuration
	if config.Build.LocalImage != "" {
		quadlet.WriteString(fmt.Sprintf("Image=%s\n", config.Build.LocalImage))
	} else if config.Build.RemoteImage != "" {
		quadlet.WriteString(fmt.Sprintf("Image=%s\n", config.Build.RemoteImage))
	} else {
		return "", fmt.Errorf("either local_image or remote_image must be specified")
	}
	
	// Container name
	quadlet.WriteString(fmt.Sprintf("ContainerName=%s\n", config.App))
	
	// Environment variables
	for key, value := range config.Env {
		quadlet.WriteString(fmt.Sprintf("Environment=%s=%s\n", key, value))
	}
	
	// Port mapping
	if config.HTTPService.InternalPort > 0 {
		// Default to same port for host:container mapping
		quadlet.WriteString(fmt.Sprintf("PublishPort=%d:%d\n", 
			config.HTTPService.InternalPort, config.HTTPService.InternalPort))
	}
	
	// Volume mounts
	for _, mount := range config.Mounts {
		quadlet.WriteString(fmt.Sprintf("Volume=%s:%s\n", mount.Source, mount.Destination))
	}
	
	// Health check
	if config.HealthCheck != nil && config.HealthCheck.Path != "" {
		quadlet.WriteString(fmt.Sprintf("HealthCmd=curl --fail http://localhost:%d%s\n", 
			config.HTTPService.InternalPort, config.HealthCheck.Path))
		if config.HealthCheck.Interval != "" {
			quadlet.WriteString(fmt.Sprintf("HealthInterval=%s\n", config.HealthCheck.Interval))
		}
		if config.HealthCheck.Timeout != "" {
			quadlet.WriteString(fmt.Sprintf("HealthTimeout=%s\n", config.HealthCheck.Timeout))
		}
		if config.HealthCheck.Retries > 0 {
			quadlet.WriteString(fmt.Sprintf("HealthRetries=%d\n", config.HealthCheck.Retries))
		}
	}
	
	// Security configuration
	if config.Security != nil {
		if config.Security.RunAsNonRoot {
			if config.Security.UserID != nil {
				quadlet.WriteString(fmt.Sprintf("User=%d\n", *config.Security.UserID))
			}
			if config.Security.GroupID != nil {
				quadlet.WriteString(fmt.Sprintf("Group=%d\n", *config.Security.GroupID))
			}
		}
		
		for _, cap := range config.Security.AddCapabilities {
			quadlet.WriteString(fmt.Sprintf("AddCapability=%s\n", cap))
		}
		
		for _, cap := range config.Security.DropCapabilities {
			quadlet.WriteString(fmt.Sprintf("DropCapability=%s\n", cap))
		}
		
		if config.Security.DisableSecurityLabel {
			quadlet.WriteString("SecurityLabelDisable=true\n")
		}
	}
	
	// Network configuration
	if config.Network != nil {
		if config.Network.Mode != "" && config.Network.Mode != "bridge" {
			quadlet.WriteString(fmt.Sprintf("Network=%s\n", config.Network.Mode))
		}
		
		for _, dns := range config.Network.DNS {
			quadlet.WriteString(fmt.Sprintf("DNS=%s\n", dns))
		}
	}
	
	// Logging configuration
	if config.Logging != nil && config.Logging.Driver != "" {
		quadlet.WriteString(fmt.Sprintf("LogDriver=%s\n", config.Logging.Driver))
		
		for key, value := range config.Logging.Options {
			quadlet.WriteString(fmt.Sprintf("LogOpt=%s=%s\n", key, value))
		}
	}
	
	// Auto restart
	quadlet.WriteString("Restart=always\n")
	
	// Kill signal
	if config.KillSignal != "" && config.KillSignal != "SIGTERM" {
		quadlet.WriteString(fmt.Sprintf("KillSignal=%s\n", config.KillSignal))
	}
	
	quadlet.WriteString("\n")
	
	// [Install] section
	quadlet.WriteString("[Install]\n")
	quadlet.WriteString("WantedBy=default.target\n")
	
	return quadlet.String(), nil
}

// ValidateDeployConfig validates a deploy configuration
func ValidateDeployConfig(config *DeployConfigInfo) error {
	if config.App == "" {
		return fmt.Errorf("app name is required")
	}
	
	// Validate app name format (should be valid for systemd service names)
	if strings.Contains(config.App, " ") || strings.Contains(config.App, "/") {
		return fmt.Errorf("app name cannot contain spaces or slashes")
	}
	
	// Validate image configuration
	if config.Build.LocalImage == "" && config.Build.RemoteImage == "" {
		return fmt.Errorf("either local_image or remote_image must be specified")
	}
	
	// Validate HTTP service configuration
	if config.HTTPService.InternalPort <= 0 || config.HTTPService.InternalPort > 65535 {
		return fmt.Errorf("internal_port must be between 1 and 65535")
	}
	
	// Validate mount paths
	for _, mount := range config.Mounts {
		if mount.Source == "" || mount.Destination == "" {
			return fmt.Errorf("mount source and destination are required")
		}
		if !strings.HasPrefix(mount.Destination, "/") {
			return fmt.Errorf("mount destination must be an absolute path")
		}
	}
	
	// Validate domain names
	for _, domain := range config.Domains {
		if domain.Name == "" {
			return fmt.Errorf("domain name cannot be empty")
		}
	}
	
	return nil
}

// ExtractImageName extracts the image name from DeployConfigInfo
func ExtractImageName(config *DeployConfigInfo) string {
	if config.Build.LocalImage != "" {
		return config.Build.LocalImage
	}
	return config.Build.RemoteImage
}

// ExtractHostPort extracts the host port from DeployConfigInfo
func ExtractHostPort(config *DeployConfigInfo) int {
	return config.HTTPService.InternalPort
}

// GetRequiredDirectories returns directories that need to be created for mounts
func GetRequiredDirectories(config *DeployConfigInfo) []string {
	var directories []string
	
	for _, mount := range config.Mounts {
		// If source looks like a directory path (starts with /), add it
		if strings.HasPrefix(mount.Source, "/") {
			directories = append(directories, mount.Source)
		}
	}
	
	return directories
}