package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/dborm"
	"gorm.io/gorm"
)

// SelfHostedDatabaseType 自托管数据库类型枚举
type SelfHostedDatabaseType string

const (
	PostgreSQL SelfHostedDatabaseType = "postgresql"
	// 为未来扩展预留空间
	MySQL   SelfHostedDatabaseType = "mysql"
	MongoDB SelfHostedDatabaseType = "mongodb"
	Redis   SelfHostedDatabaseType = "redis"
)

// DatabaseStatus 数据库状态枚举
type DatabaseStatus string

const (
	DatabaseStatusPending DatabaseStatus = "pending"
	DatabaseStatusRunning DatabaseStatus = "running"
	DatabaseStatusStopped DatabaseStatus = "stopped"
	DatabaseStatusFailed  DatabaseStatus = "failed"
)

// Database 数据库实例模型
type SelfHostedDatabase struct {
	ID        uuid.UUID `gorm:"type:char(36);primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt    gorm.DeletedAt         `gorm:"index"`
	Name         string                 `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Type         SelfHostedDatabaseType `gorm:"size:50;not null" json:"type"`
	Version      string                 `gorm:"size:50;not null;default:'latest'" json:"version"`
	CustomImage  string                 `gorm:"size:500" json:"custom_image"` // 自定义镜像来源，如 docker.io/postgres:16-alpine
	Status       DatabaseStatus         `gorm:"size:50;not null;default:'pending'" json:"status"`
	Port         int                    `gorm:"not null" json:"port"`                       // 外部访问端口
	InternalPort int                    `gorm:"not null;default:5432" json:"internal_port"` // 容器内部端口
	Username     string                 `gorm:"size:100;not null" json:"username"`
	Password     string                 `gorm:"size:255;not null" json:"password"`      // 建议后续加密存储
	DatabaseName string                 `gorm:"size:100;not null" json:"database_name"` // 默认数据库名
	DataPath     string                 `gorm:"size:500;not null" json:"data_path"`     // 数据持久化路径
	ConfigPath   string                 `gorm:"size:500" json:"config_path"`            // 配置文件路径
	IsRemote     bool                   `gorm:"default:false" json:"is_remote"`         // 是否为远程数据库
	SSHHostID    *uuid.UUID             `gorm:"type:char(36);index" json:"ssh_host_id"`               // 远程主机ID（可选）
	ExtraConfig  JSONB                  `gorm:"type:jsonb" json:"extra_config"`         // 额外配置参数
	LastCheckAt  *time.Time             `json:"last_check_at"`                          // 最后健康检查时间
}

// BeforeCreate will set a UUID rather than numeric ID.
func (db *SelfHostedDatabase) BeforeCreate(tx *gorm.DB) (err error) {
	db.ID = uuid.New()
	return
}

// TableName specifies the table name for the SelfHostedDatabase model
func (SelfHostedDatabase) TableName() string {
	return "databases"
}

// CreateDatabase creates a new database record
func CreateDatabase(name string, dbType SelfHostedDatabaseType, version, customImage, username, password, databaseName, dataPath string, port, internalPort int, isRemote bool, sshHostID *uuid.UUID, extraConfig JSONB) (*SelfHostedDatabase, error) {
	database := &SelfHostedDatabase{
		Name:         name,
		Type:         dbType,
		Version:      version,
		CustomImage:  customImage,
		Status:       DatabaseStatusPending,
		Port:         port,
		InternalPort: internalPort,
		Username:     username,
		Password:     password,
		DatabaseName: databaseName,
		DataPath:     dataPath,
		IsRemote:     isRemote,
		SSHHostID:    sshHostID,
		ExtraConfig:  extraConfig,
	}

	if err := dborm.Db.Create(database).Error; err != nil {
		return nil, err
	}

	return database, nil
}

// GetDatabaseByID retrieves a database by ID
func GetDatabaseByID(id uuid.UUID) (*SelfHostedDatabase, error) {
	var database SelfHostedDatabase
	if err := dborm.Db.First(&database, id).Error; err != nil {
		return nil, err
	}
	return &database, nil
}

// GetDatabaseByName retrieves a database by name
func GetDatabaseByName(name string) (*SelfHostedDatabase, error) {
	var database SelfHostedDatabase
	if err := dborm.Db.Where("name = ?", name).First(&database).Error; err != nil {
		return nil, err
	}
	return &database, nil
}

// GetAllDatabases retrieves all databases
func GetAllDatabases() ([]SelfHostedDatabase, error) {
	var databases []SelfHostedDatabase
	if err := dborm.Db.Find(&databases).Error; err != nil {
		return nil, err
	}
	return databases, nil
}

// UpdateDatabaseStatus updates database status
func UpdateDatabaseStatus(id uuid.UUID, status DatabaseStatus) error {
	now := time.Now()
	return dborm.Db.Model(&SelfHostedDatabase{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        status,
		"last_check_at": &now,
	}).Error
}

// UpdateDatabasePassword updates database password
func UpdateDatabasePassword(id uuid.UUID, newPassword string) error {
	return dborm.Db.Model(&SelfHostedDatabase{}).Where("id = ?", id).Update("password", newPassword).Error
}

// DeleteDatabase deletes a database record
func DeleteDatabase(id uuid.UUID) error {
	return dborm.Db.Delete(&SelfHostedDatabase{}, id).Error
}

// GetDatabasesByType retrieves databases by type
func GetDatabasesByType(dbType SelfHostedDatabaseType) ([]SelfHostedDatabase, error) {
	var databases []SelfHostedDatabase
	if err := dborm.Db.Where("type = ?", dbType).Find(&databases).Error; err != nil {
		return nil, err
	}
	return databases, nil
}

// GetRunningDatabases retrieves all running databases
func GetRunningDatabases() ([]SelfHostedDatabase, error) {
	var databases []SelfHostedDatabase
	if err := dborm.Db.Where("status = ?", DatabaseStatusRunning).Find(&databases).Error; err != nil {
		return nil, err
	}
	return databases, nil
}
