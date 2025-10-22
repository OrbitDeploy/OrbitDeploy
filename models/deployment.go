package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/dborm"
	"gorm.io/gorm"
)

// Deployment 记录将一个 Release 应用到 Application 的过程，是纯粹的日志。
// 用于追踪部署历史和排查问题。
type Deployment struct {
	ID            uuid.UUID `gorm:"type:char(36);primary_key"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	ApplicationID uuid.UUID      `gorm:"type:char(36);not null;index"`
	ReleaseID     uuid.UUID      `gorm:"type:char(36);not null;index"`
	Status        string         `gorm:"size:50;not null;default:'pending'"`
	LogText       string         `gorm:"type:text"`
	StartedAt     time.Time
	FinishedAt    *time.Time
	ServiceName   string `gorm:"size:50;not null;default:''"`
	Snapshot      string `gorm:"type:text"`    // JSON snapshot of environment variables at deployment time
	SystemPort    *int   `gorm:"default:null"` // 系统分配的端口，可选字段（仅在部署成功时分配）

	Release     Release     `gorm:"foreignKey:ReleaseID"`
	Application Application `gorm:"foreignKey:ApplicationID"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (d *Deployment) BeforeCreate(tx *gorm.DB) (err error) {
	d.ID = uuid.New()
	return
}

// TableName specifies the table name for the Deployment model
func (Deployment) TableName() string {
	return "deployments"
}

// CreateDeployment creates a new deployment record
func CreateDeployment(applicationID, releaseID uuid.UUID, status, logText, serviceName string, startedAt time.Time, finishedAt *time.Time) (*Deployment, error) {
	// Create snapshot of environment variables
	snapshot, err := CreateSnapshotForDeployment(applicationID)
	if err != nil {
		// Log the error but don't fail the deployment
		snapshot = ""
	}

	deployment := &Deployment{
		ApplicationID: applicationID,
		ReleaseID:     releaseID,
		Status:        status,
		LogText:       logText,
		ServiceName:   serviceName, // Set the service name
		StartedAt:     startedAt,
		FinishedAt:    finishedAt,
		Snapshot:      snapshot, // Store the environment variable snapshot
	}

	if err := dborm.Db.Create(deployment).Error; err != nil {
		return nil, err
	}

	return deployment, nil
}

// GetDeploymentByID retrieves a deployment by its ID
func GetDeploymentByID(id uuid.UUID) (*Deployment, error) {
	var deployment Deployment
	if err := dborm.Db.Where("id = ?", id).First(&deployment).Error; err != nil {
		return nil, err
	}
	return &deployment, nil
}

// ListDeployments retrieves all deployments
func ListDeployments() ([]*Deployment, error) {
	var deployments []*Deployment
	if err := dborm.Db.Find(&deployments).Error; err != nil {
		return nil, err
	}
	return deployments, nil
}

// ListDeploymentsByAppID retrieves deployments by application ID with Release preloaded
func ListDeploymentsByAppID(appID uuid.UUID) ([]*Deployment, error) {
	var deployments []*Deployment
	if err := dborm.Db.
		Preload("Release").
		Where("application_id = ?", appID).
		Order("created_at DESC").
		Find(&deployments).Error; err != nil {
		return nil, err
	}
	return deployments, nil
}

// UpdateDeployment updates an existing deployment
func UpdateDeployment(id uuid.UUID, status, logText string, finishedAt *time.Time) (*Deployment, error) {
	deployment, err := GetDeploymentByID(id)
	if err != nil {
		return nil, err
	}

	deployment.Status = status
	deployment.LogText = logText
	deployment.FinishedAt = finishedAt
	if err := dborm.Db.Save(deployment).Error; err != nil {
		return nil, err
	}

	return deployment, nil
}

// DeleteDeployment deletes a deployment by its ID
func DeleteDeployment(id uuid.UUID) error {
	return dborm.Db.Where("id = ?", id).Delete(&Deployment{}).Error
}

// GetRunningDeploymentsByAppID retrieves running deployments for an application
func GetRunningDeploymentsByAppID(appID uuid.UUID) ([]*Deployment, error) {
	var deployments []*Deployment
	if err := dborm.Db.
		Preload("Release").
		Where("application_id = ? AND status = ?", appID, "running").
		Order("created_at DESC").
		Find(&deployments).Error; err != nil {
		return nil, err
	}
	return deployments, nil
}

func GetRunningDeploymentsByName(appName string) ([]*Deployment, error) {
	var deployments []*Deployment
	if err := dborm.Db.
		Preload("Release").
		Where("name = ? AND status = ?", appName, "running").
		Order("created_at DESC").
		Find(&deployments).Error; err != nil {
		return nil, err
	}
	return deployments, nil
}
func GetAllRunningDeployments() ([]*Deployment, error) {
	var deployments []*Deployment
	if err := dborm.Db.
		Preload("Release").
		Preload("Application").
		Where("status = ?", "success").
		Order("created_at DESC").
		Find(&deployments).Error; err != nil {
		return nil, err
	}
	return deployments, nil
}

// ListAllDeployments retrieves all deployments with preloaded Application and Release associations
func ListAllDeployments() ([]*Deployment, error) {
	var deployments []*Deployment
	if err := dborm.Db.Preload("Application").Preload("Release").Find(&deployments).Error; err != nil {
		return nil, err
	}
	return deployments, nil
}
func IsPortInUse(port int) (bool, error) {
	var count int64
	err := dborm.Db.Model(&Deployment{}).Where("system_port = ?", port).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
func UpdateDeploymentSystemPort(deploymentID uuid.UUID, systemPort int) error {
	return dborm.Db.Model(&Deployment{}).Where("id = ?", deploymentID).Update("system_port", systemPort).Error
}
