package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/dborm"
	"gorm.io/gorm"
)

// JSONB a custom type for handling JSONB data in GORM.
// It allows storing flexible key-value data.
type JSONB struct {
	Data interface{}
}

// Value implements the driver.Valuer interface, allowing our custom JSONB type to be written to the database.
func (j JSONB) Value() (driver.Value, error) {
	if j.Data == nil {
		return nil, nil
	}
	return json.Marshal(j.Data)
}

// Scan implements the sql.Scanner interface, allowing our custom JSONB type to be read from the database.
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		j.Data = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	if len(bytes) == 0 {
		j.Data = nil
		return nil
	}
	var data interface{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return err
	}
	j.Data = data
	return nil
}

// MarshalJSON implements json.Marshaler
func (j JSONB) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.Data)
}

// UnmarshalJSON implements json.Unmarshaler
func (j *JSONB) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &j.Data)
}

// Application 代表一个实际运行的环境实例 (e.g., my-app-prod, my-app-staging).
// 这是系统的核心模型，存储了应用的"意图状态"。
type Application struct {
	ID          uuid.UUID `gorm:"type:char(36);primary_key"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	ProjectID   uuid.UUID      `gorm:"type:char(36);not null"`
	Name        string         `gorm:"size:255;not null;unique"`
	Description string         `gorm:"size:255"` // 用于生成 .container 文件中的描述信息

	RepoURL   *string `gorm:"size:500"`                     // 仓库URL，必须是完整的可访问URL，如 https://github.com/user/repo 或 https://your-gitea.com/user/repo，可为空以支持本地推送
	Branch    *string `gorm:"size:255;default:'main'"`      // 可选的分支名称，用于GitHub部署，默认main
	BuildDir  *string `gorm:"size:255;default:'/'"`         // 可选的构建目录，默认根目录
	BuildType *string `gorm:"size:50;default:'dockerfile'"` // 可选的构建类型，如 dockerfile, railpack, nixpacks 等，默认dockerfile

	ActiveReleaseID *uuid.UUID `gorm:"type:char(36);index"` // 指向当前线上运行的版本, 使用指针以允许为空
	TargetPort      int        `gorm:"not null"`              // 容器内部监听的端口
	Status          string     `gorm:"size:50;not null;default:'stopped'"`
	ProviderAuthID  *uuid.UUID `gorm:"type:char(36);index"` // 可选关联第三方平台授权，支持本地CLI推送不关联场景 这个是关联到仓库的。通过这个授权+RepoURL可以访问代码仓库

	// 灵活的运行时配置 (Runtime Configuration)
	Volumes          JSONB   `gorm:"type:jsonb"` // 存储多个卷挂载, e.g., [{"host_path": "/var/data", "container_path": "/data"}]
	ExecCommand      *string `gorm:"size:255"`   // 可选的容器启动命令 (override image's default command)
	AutoUpdatePolicy *string `gorm:"size:50"`    // 可选的自动更新策略 (e.g., "registry")"

	// 关联关系 (GORM Associations)
	ActiveRelease        *Release              `gorm:"foreignKey:ActiveReleaseID"`
	Releases             []Release             `gorm:"foreignKey:ApplicationID"`
	EnvironmentVariables []EnvironmentVariable `gorm:"foreignKey:ApplicationID"`
	Routings             []Routing             `gorm:"foreignKey:ApplicationID"`
	Deployments          []Deployment          `gorm:"foreignKey:ApplicationID"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (app *Application) BeforeCreate(tx *gorm.DB) (err error) {
	app.ID = uuid.New()
	return
}

// TableName specifies the table name for the Application model
func (Application) TableName() string {
	return "applications"
}

// CreateApplication creates a new application record
func CreateApplication(projectID uuid.UUID, name, description string, repoURL *string, targetPort int, volumes JSONB, execCommand, autoUpdatePolicy, branch, buildDir, buildType *string, providerAuthID *uuid.UUID) (*Application, error) {
	// Check if application name already exists globally
	var existingApp Application
	if err := dborm.Db.Where("name = ?", name).First(&existingApp).Error; err == nil {
		return nil, errors.New("应用名称已存在，请选择其他名称")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Validate repoURL: if provided, must be a full URL
	if repoURL != nil && *repoURL != "" && !strings.HasPrefix(*repoURL, "http://") && !strings.HasPrefix(*repoURL, "https://") {
		return nil, errors.New("仓库URL必须是完整的可访问URL，例如 https://github.com/user/repo 或 https://your-gitea.com/user/repo")
	}

	application := &Application{
		ProjectID:        projectID,
		Name:             name,
		Description:      description,
		RepoURL:          repoURL,
		TargetPort:       targetPort,
		Status:           "stopped",
		Volumes:          volumes,
		ExecCommand:      execCommand,
		AutoUpdatePolicy: autoUpdatePolicy,
		Branch:           branch,
		BuildDir:         buildDir,
		BuildType:        buildType,
		ProviderAuthID:   providerAuthID,
	}

	if err := dborm.Db.Create(application).Error; err != nil {
		return nil, err
	}

	return application, nil
}

// GetApplicationByID retrieves an application by its ID
func GetApplicationByID(id uuid.UUID) (*Application, error) {
	var application Application
	if err := dborm.Db.Where("id = ?", id).First(&application).Error; err != nil {
		return nil, err
	}
	return &application, nil
}

// ListApplications retrieves all applications
func ListApplications() ([]*Application, error) {
	var applications []*Application
	if err := dborm.Db.Find(&applications).Error; err != nil {
		return nil, err
	}
	return applications, nil
}

// UpdateApplication updates an existing application
func UpdateApplication(id uuid.UUID, name, description string, repoURL *string, targetPort int, status string, volumes JSONB, execCommand, autoUpdatePolicy, branch, buildDir, buildType *string, providerAuthID *uuid.UUID) (*Application, error) {
	application, err := GetApplicationByID(id)
	if err != nil {
		return nil, err
	}

	application.Name = name
	application.Description = description
	application.RepoURL = repoURL
	application.TargetPort = targetPort
	application.Status = status
	application.Volumes = volumes
	application.ExecCommand = execCommand
	application.AutoUpdatePolicy = autoUpdatePolicy
	application.Branch = branch
	application.BuildDir = buildDir
	application.BuildType = buildType
	application.ProviderAuthID = providerAuthID
	if err := dborm.Db.Save(application).Error; err != nil {
		return nil, err
	}

	return application, nil
}

// UpdateApplicationFromFrontend updates an existing application (frontend version)
func UpdateApplicationFromFrontend(id uuid.UUID, description string, repoURL *string, targetPort int, status string, volumes JSONB, execCommand, autoUpdatePolicy, branch, buildDir, buildType *string, providerAuthID *uuid.UUID) (*Application, error) {
	application, err := GetApplicationByID(id)
	if err != nil {
		return nil, err
	}

	// Validate repoURL: if provided, must be a full URL
	if repoURL != nil && *repoURL != "" && !strings.HasPrefix(*repoURL, "http://") && !strings.HasPrefix(*repoURL, "https://") {
		return nil, errors.New("仓库URL必须是完整的可访问URL，例如 https://github.com/user/repo 或 https://your-gitea.com/user/repo")
	}

	application.Description = description
	application.RepoURL = repoURL
	application.TargetPort = targetPort
	application.Volumes = volumes
	application.ExecCommand = execCommand
	application.AutoUpdatePolicy = autoUpdatePolicy
	application.Branch = branch
	application.BuildDir = buildDir
	application.BuildType = buildType
	application.ProviderAuthID = providerAuthID
	if err := dborm.Db.Save(application).Error; err != nil {
		return nil, err
	}

	return application, nil
}

// DeleteApplication deletes an application by its ID
func DeleteApplication(id uuid.UUID) error {
	return dborm.Db.Where("id = ?", id).Delete(&Application{}).Error
}

// GetApplicationByName retrieves an application by its name
func GetApplicationByName(name string) (*Application, error) {
	var application Application
	if err := dborm.Db.Where("name = ?", name).First(&application).Error; err != nil {
		return nil, err
	}
	return &application, nil
}

// ListApplicationsByProjectName retrieves applications by project name
func ListApplicationsByProjectName(projectName string) ([]*Application, error) {
	var applications []*Application
	if err := dborm.Db.Joins("JOIN projects ON applications.project_id = projects.id").
		Where("projects.name = ?", projectName).
		Find(&applications).Error; err != nil {
		return nil, err
	}
	return applications, nil
}

// GetApplicationByProjectNameAndAppName retrieves an application by project name and app name
func GetApplicationByProjectNameAndAppName(projectName, appName string) (*Application, error) {
	var application Application
	if err := dborm.Db.Joins("JOIN projects ON applications.project_id = projects.id").
		Where("projects.name = ? AND applications.name = ?", projectName, appName).
		First(&application).Error; err != nil {
		return nil, err
	}
	return &application, nil
}
