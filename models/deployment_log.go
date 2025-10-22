package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/dborm"
	"gorm.io/gorm"
)

// DeploymentLog 部署日志条目，存储结构化的部署日志信息
type DeploymentLog struct {
	ID           uint      `gorm:"primaryKey"`
	DeploymentID uuid.UUID `gorm:"type:char(36);not null;index"`
	Timestamp    time.Time `gorm:"type:datetime(6);not null;index"`
	Level        string    `gorm:"size:20;default:'INFO'"`
	Source       string    `gorm:"size:50;default:'SYSTEM'"`
	Message      string    `gorm:"type:text;not null"`
	CreatedAt    time.Time
}

// TableName specifies the table name for the DeploymentLog model
func (DeploymentLog) TableName() string {
	return "deployment_logs"
}

// CreateDeploymentLog creates a new deployment log entry
func CreateDeploymentLog(logEntry *DeploymentLog) error {
	if logEntry.Timestamp.IsZero() {
		logEntry.Timestamp = time.Now()
	}
	if logEntry.Level == "" {
		logEntry.Level = "INFO"
	}
	if logEntry.Source == "" {
		logEntry.Source = "SYSTEM"
	}
	
	return dborm.Db.Create(logEntry).Error
}

// GetDeploymentLogs retrieves deployment logs with pagination
// limit: 返回日志条数限制
// beforeTimestamp: 获取此时间戳之前的日志（用于向上滚动加载更早的日志）
func GetDeploymentLogs(deploymentID uuid.UUID, limit int, beforeTimestamp *time.Time) ([]*DeploymentLog, error) {
	var logs []*DeploymentLog
	
	query := dborm.Db.Where("deployment_id = ?", deploymentID)
	
	// 如果提供了 beforeTimestamp，只获取该时间之前的日志
	if beforeTimestamp != nil {
		query = query.Where("timestamp < ?", *beforeTimestamp)
	}
	
	// 按时间倒序排列，获取最新的 limit 条
	// 注意：这里返回的是时间倒序，前端需要反转顺序以正序显示
	if err := query.Order("timestamp DESC").Limit(limit).Find(&logs).Error; err != nil {
		return nil, err
	}
	
	// 反转数组，使其按时间正序排列（从旧到新）
	for i := 0; i < len(logs)/2; i++ {
		j := len(logs) - 1 - i
		logs[i], logs[j] = logs[j], logs[i]
	}
	
	return logs, nil
}

// GetLatestDeploymentLogs retrieves the latest deployment logs
func GetLatestDeploymentLogs(deploymentID uuid.UUID, limit int) ([]*DeploymentLog, error) {
	return GetDeploymentLogs(deploymentID, limit, nil)
}

// DeleteDeploymentLogsByDeploymentID deletes all logs for a specific deployment
func DeleteDeploymentLogsByDeploymentID(deploymentID uuid.UUID) error {
	return dborm.Db.Where("deployment_id = ?", deploymentID).Delete(&DeploymentLog{}).Error
}

// BeforeCreate hook to set timestamp if not set
func (d *DeploymentLog) BeforeCreate(tx *gorm.DB) error {
	if d.Timestamp.IsZero() {
		d.Timestamp = time.Now()
	}
	return nil
}
