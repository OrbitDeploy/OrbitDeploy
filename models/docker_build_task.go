package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/dborm"
	"gorm.io/gorm"
)

// DockerBuildTaskStatus 定义Docker构建任务状态的类型
type DockerBuildTaskStatus string

const (
	DockerBuildStatusPending   DockerBuildTaskStatus = "pending"
	DockerBuildStatusRunning   DockerBuildTaskStatus = "running"
	DockerBuildStatusCompleted DockerBuildTaskStatus = "completed"
	DockerBuildStatusFailed    DockerBuildTaskStatus = "failed"
	DockerBuildStatusPaused    DockerBuildTaskStatus = "paused"
)

// DockerBuildPayload 定义Docker构建任务的负载
type DockerBuildPayload struct {
	AppID       uuid.UUID         `json:"app_id"`
	ReleaseID   uuid.UUID         `json:"release_id"`
	LogText     string            `json:"log_text"`
	Dockerfile  string            `json:"dockerfile"`
	ContextPath string            `json:"context_path"`
	BuildArgs   map[string]string `json:"build_args"`
}

// DockerBuildTask Docker构建任务模型
type DockerBuildTask struct {
	ID         uuid.UUID             `gorm:"type:char(36);primary_key"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt        `gorm:"index"`
	UUID       string                `gorm:"uniqueIndex;not null"`
	Payload    string                `gorm:"type:text;not null"` // JSON encoded DockerBuildPayload
	Status     DockerBuildTaskStatus `gorm:"size:50;not null;default:'pending'"`
	Log        string                `gorm:"type:text"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (t *DockerBuildTask) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	return
}

// TableName specifies the table name for the DockerBuildTask model
func (DockerBuildTask) TableName() string {
	return "docker_build_tasks"
}

// CreateDockerBuildTask creates a new docker build task record
func CreateDockerBuildTask(uuid string, payload string, status DockerBuildTaskStatus) (*DockerBuildTask, error) {
	task := &DockerBuildTask{
		UUID:    uuid,
		Payload: payload,
		Status:  status,
	}

	if err := dborm.Db.Create(task).Error; err != nil {
		return nil, err
	}

	return task, nil
}

// GetDockerBuildTaskByUUID retrieves a docker build task by its UUID
func GetDockerBuildTaskByUUID(uuid string) (*DockerBuildTask, error) {
	var task DockerBuildTask
	if err := dborm.Db.Where("uuid = ?", uuid).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateDockerBuildTaskStatus updates the status and log of a docker build task
func UpdateDockerBuildTaskStatus(uuid string, status DockerBuildTaskStatus, log string) error {
	return dborm.Db.Model(&DockerBuildTask{}).Where("uuid = ?", uuid).Updates(map[string]interface{}{
		"status":     status,
		"log":        log,
		"updated_at": time.Now(),
	}).Error
}

// ListDockerBuildTasksByStatus retrieves docker build tasks by status
func ListDockerBuildTasksByStatus(status DockerBuildTaskStatus) ([]*DockerBuildTask, error) {
	var tasks []*DockerBuildTask
	if err := dborm.Db.Where("status = ?", status).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// DequeueDockerBuildTask dequeues a pending task and marks it as running
func DequeueDockerBuildTask() (*DockerBuildTask, error) {
	var task DockerBuildTask
	err := dborm.Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("status = ?", DockerBuildStatusPending).Order("created_at ASC").First(&task).Error; err != nil {
			return err
		}
		return tx.Model(&task).Update("status", DockerBuildStatusRunning).Error
	})
	if err != nil {
		return nil, err
	}
	return &task, nil
}
