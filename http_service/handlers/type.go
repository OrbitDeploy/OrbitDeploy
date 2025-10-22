package handlers

import (
	"time"

	"github.com/OrbitDeploy/OrbitDeploy/models"
	"github.com/google/uuid"
)

type ProjectResponse struct {
	Uid         string    `json:"uid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// type JSONB map[string]interface{}

type ApplicationResponse struct {
	Uid         string `json:"uid"`
	ProjectUid  string `json:"projectUid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TargetPort  int    `json:"targetPort"`
	Status      string `json:"status"`
}

type ApplicationDetailResponse struct {
	Uid               string       `json:"uid"`
	ProjectUid        string       `json:"projectUid"`
	Name              string       `json:"name"`
	Description       string       `json:"description"`
	RepoURL           *string      `json:"repoUrl,omitempty"`
	ActiveReleaseUid  *string      `json:"activeReleaseUid,omitempty"`
	TargetPort        int          `json:"targetPort"`
	Status            string       `json:"status"`
	Volumes           models.JSONB `json:"volumes,omitempty"`
	ExecCommand       *string      `json:"execCommand,omitempty"`
	AutoUpdatePolicy  *string      `json:"autoUpdatePolicy,omitempty"`
	Branch            *string      `json:"branch,omitempty"`
	BuildDir          *string      `json:"buildDir,omitempty"`
	BuildType         *string      `json:"buildType,omitempty"`
	CreatedAt         time.Time    `json:"createdAt"`
	UpdatedAt         time.Time    `json:"updatedAt"`
	ActiveReleaseInfo *ReleaseInfo `json:"activeReleaseInfo,omitempty"`
}
type RunningDeploymentResponse struct {
	DeploymentResponse
	Domains  []string `json:"domains"`
	HostPort int      `json:"hostPort"`
}

type ReleaseInfo struct {
	Uid       string  `json:"uid"`
	ImageName *string `json:"imageName,omitempty"`
}
type UpdateApplicationRequest struct {
	Description      string      `json:"description"`
	RepoURL          *string     `json:"repoUrl,omitempty"`
	TargetPort       int         `json:"targetPort"`
	Status           string      `json:"status"`
	Volumes          interface{} `json:"volumes"`
	ExecCommand      *string     `json:"execCommand"`
	AutoUpdatePolicy *string     `json:"autoUpdatePolicy"`
	Branch           *string     `json:"branch"`
	BuildDir         *string     `json:"buildDir,omitempty"`
	BuildType        *string     `json:"buildType,omitempty"`
	ProviderAuthUid  *string     `json:"providerAuthUid,omitempty"`
}
type CreateReleaseRequest struct {
	ImageName       string                 `json:"imageName"`
	BuildSourceInfo map[string]interface{} `json:"buildSourceInfo"`
	Status          string                 `json:"status"`
}

type ReleaseResponse struct {
	Uid             string                 `json:"uid"`
	ApplicationUid  string                 `json:"applicationUid"`
	ImageName       string                 `json:"imageName"`
	BuildSourceInfo map[string]interface{} `json:"buildSourceInfo"`
	Status          string                 `json:"status"`
	CreatedAt       time.Time              `json:"createdAt"`
	UpdatedAt       time.Time              `json:"updatedAt"`
}

type RoutingRequest struct {
	DomainName string `json:"domainName"`
	HostPort   int    `json:"hostPort"`
	IsActive   bool   `json:"isActive"`
}

type RoutingResponse struct {
	Uid            string    `json:"uid"`
	ApplicationUid string    `json:"applicationUid"`
	DomainName     string    `json:"domainName"`
	HostPort       int       `json:"hostPort"`
	IsActive       bool      `json:"isActive"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type ConfigurationResponse struct {
	Uid            string    `json:"uid"`
	ApplicationUid string    `json:"applicationUid"`
	Version        int       `json:"version"`
	EnvVars        string    `json:"envVars"`
	IsActive       bool      `json:"isActive"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type CreateConfigurationRequest struct {
	Version  int    `json:"version"`
	EnvVars  string `json:"envVars"`
	IsActive bool   `json:"isActive"`
}
type DeploymentResponse struct {
	Uid            string     `json:"uid"`
	ApplicationUid string     `json:"applicationUid"`
	ReleaseUid     string     `json:"releaseUid"`
	Status         string     `json:"status"`
	LogText        string     `json:"logText"`
	StartedAt      time.Time  `json:"startedAt"`
	FinishedAt     *time.Time `json:"finishedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	SystemPort     *int       `json:"systemPort,omitempty"`
	// Fields from the related Release model
	Version       *string `json:"version,omitempty"`
	ImageName     *string `json:"imageName,omitempty"`
	ReleaseStatus string  `json:"releaseStatus,omitempty"`
}

type DeploymentLogResponse struct {
	ID        uint      `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Source    string    `json:"source"`
	Message   string    `json:"message"`
}
type UpdateConfigurationRequest struct {
	Version  int    `json:"version"`
	EnvVars  string `json:"envVars"`
	IsActive *bool  `json:"isActive,omitempty"` // Optional field to update active status
}

// Environment Variable API Types

type EnvironmentVariableResponse struct {
	Uid            string    `json:"uid"`
	ApplicationUid string    `json:"applicationUid"`
	Key            string    `json:"key"`
	Value          string    `json:"value"`
	IsEncrypted    bool      `json:"isEncrypted"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type CreateEnvironmentVariableRequest struct {
	Key         string `json:"key" validate:"required"`
	Value       string `json:"value"`
	IsEncrypted bool   `json:"isEncrypted"`
}

type UpdateEnvironmentVariableRequest struct {
	Key         string `json:"key" validate:"required"`
	Value       string `json:"value"`
	IsEncrypted bool   `json:"isEncrypted"`
}

type CreateConfigurationWithVariablesRequest struct {
	Version              int                                `json:"version"`
	IsActive             bool                               `json:"isActive"`
	EnvironmentVariables []CreateEnvironmentVariableRequest `json:"environmentVariables"`
}

type ConfigurationWithVariablesResponse struct {
	Uid                  string                        `json:"uid"`
	ApplicationUid       string                        `json:"applicationUid"`
	Version              int                           `json:"version"`
	IsActive             bool                          `json:"isActive"`
	CreatedAt            time.Time                     `json:"createdAt"`
	UpdatedAt            time.Time                     `json:"updatedAt"`
	EnvironmentVariables []EnvironmentVariableResponse `json:"environmentVariables"`
}

// SSH Host Management Types

type SSHHostRequest struct {
	Name        string `json:"name"`
	Addr        string `json:"addr"`
	Port        int    `json:"port"`
	User        string `json:"user"`
	Password    string `json:"password"`
	PrivateKey  string `json:"private_key"`
	Description string `json:"description"`
}

// SSHHostResponse represents SSH host data for API responses (separated from database model)
type SSHHostResponse struct {
	Uid         string `json:"uid"`
	Name        string `json:"name"`
	Addr        string `json:"addr"`
	Port        int    `json:"port"`
	User        string `json:"user"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Region      string `json:"region"`
	CPUCores    int    `json:"cpuCores"`
	MemoryGB    int    `json:"memoryGB"`
	DiskGB      int    `json:"diskGB"`
	IsActive    bool   `json:"isActive"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// Multi-Node Deployment Types

type CreateMultiNodeDeploymentRequest struct {
	ApplicationID uint   `json:"application_id"`
	ReleaseID     uint   `json:"release_id"`
	Strategy      string `json:"strategy"` // parallel/sequential/canary
	HostIDs       []uint `json:"host_ids"`
}

type MultiNodeDeploymentResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type NodeDeploymentResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
type CreateApplicationRequest struct {
	Name             string      `json:"name"`
	Description      string      `json:"description"`
	RepoURL          *string     `json:"repoUrl,omitempty"`
	TargetPort       int         `json:"targetPort"`
	Volumes          interface{} `json:"volumes"`
	ExecCommand      *string     `json:"execCommand"`
	AutoUpdatePolicy *string     `json:"autoUpdatePolicy"`
	Branch           *string     `json:"branch"`
	BuildDir         *string     `json:"buildDir,omitempty"`
	BuildType        *string     `json:"buildType,omitempty"`
	ProviderAuthUid  *string     `json:"providerAuthUid,omitempty"`
}

// ApplicationTokenResponse represents the response for an application token
type ApplicationTokenResponse struct {
	Uid            string     `json:"uid"`
	ApplicationUid string     `json:"applicationUid"`
	Name           string     `json:"name"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
	LastUsedAt     *time.Time `json:"lastUsedAt,omitempty"`
	IsActive       bool       `json:"isActive"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

// CreateApplicationTokenResponse represents the response when creating a new application token
type CreateApplicationTokenResponse struct {
	ApplicationTokenResponse
	Token string `json:"token"`
}

// AuthorizeDeviceCodeRequest represents the device authorization request
type AuthorizeDeviceCodeRequest struct {
	UserCode string `json:"user_code"`
}

// DeviceContextRequest represents the device context information for new flow
type DeviceContextRequest struct {
	OS         string `json:"os"`
	DeviceName string `json:"device_name"`
}

// DeviceContextSession represents a device authorization session with context
type DeviceContextSession struct {
	SessionID    string     `json:"session_id"`
	OS           string     `json:"os"`
	DeviceName   string     `json:"device_name"`
	PublicIP     string     `json:"public_ip"`
	RequestTime  int64      `json:"request_time"`
	IsAuthorized bool       `json:"is_authorized"`
	UserID       *uuid.UUID `json:"user_id,omitempty"`
	ExpiresAt    time.Time  `json:"expires_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

// Self-Hosted Database Types

type CreateDatabaseRequest struct {
	Name         string                        `json:"name" validate:"required"`
	Type         models.SelfHostedDatabaseType `json:"type" validate:"required"`
	Version      string                        `json:"version" validate:"required"`
	CustomImage  string                        `json:"custom_image"` // 可选：自定义镜像来源
	Port         int                           `json:"port" validate:"required"`
	InternalPort int                           `json:"internal_port"`
	Username     string                        `json:"username" validate:"required"`
	Password     string                        `json:"password" validate:"required"`
	DatabaseName string                        `json:"database_name" validate:"required"`
	DataPath     string                        `json:"data_path" validate:"required"`
	ConfigPath   string                        `json:"config_path"`
	IsRemote     bool                          `json:"is_remote"`
	SSHHostUid   *string                       `json:"ssh_host_uid,omitempty"`
	ExtraConfig  models.JSONB                  `json:"extra_config"`
}

type UpdateDatabaseRequest struct {
	Port        *int          `json:"port,omitempty"`
	Username    *string       `json:"username,omitempty"`
	Password    *string       `json:"password,omitempty"`
	DataPath    *string       `json:"data_path,omitempty"`
	ConfigPath  *string       `json:"config_path,omitempty"`
	ExtraConfig *models.JSONB `json:"extra_config,omitempty"`
}

type DatabaseResponse struct {
	Uid          string                        `json:"uid"`
	Name         string                        `json:"name"`
	Type         models.SelfHostedDatabaseType `json:"type"`
	Version      string                        `json:"version"`
	CustomImage  string                        `json:"custom_image,omitempty"`
	Status       models.DatabaseStatus         `json:"status"`
	Port         int                           `json:"port"`
	InternalPort int                           `json:"internal_port"`
	Username     string                        `json:"username"`
	DatabaseName string                        `json:"database_name"`
	DataPath     string                        `json:"data_path"`
	ConfigPath   string                        `json:"config_path"`
	IsRemote     bool                          `json:"is_remote"`
	SSHHostUid   *string                       `json:"ssh_host_uid,omitempty"`
	ExtraConfig  models.JSONB                  `json:"extra_config"`
	LastCheckAt  *time.Time                    `json:"last_check_at,omitempty"`
	CreatedAt    time.Time                     `json:"created_at"`
	UpdatedAt    time.Time                     `json:"updated_at"`
}

type DatabaseConnectionInfoResponse struct {
	Host             string `json:"host"`
	Port             int    `json:"port"`
	User             string `json:"user"`
	Password         string `json:"password,omitempty"` // omitempty to hide when not requested
	Database         string `json:"database"`
	ConnectionString string `json:"connection_string"`
}
