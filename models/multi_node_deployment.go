package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/dborm"
	"gorm.io/gorm"
)

// MultiNodeDeployment 记录应用的多节点部署信息
type MultiNodeDeployment struct {
	ID            uuid.UUID `gorm:"type:char(36);primary_key"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	ApplicationID uuid.UUID      `gorm:"type:char(36);not null;index"`
	ReleaseID     uuid.UUID      `gorm:"type:char(36);not null;index"`
	Strategy      string         `gorm:"size:50;not null;default:'parallel'"` // 部署策略: parallel/sequential/canary
	Status        string         `gorm:"size:50;not null;default:'pending'"`  // 整体状态: pending/running/success/failed/partial
	TotalNodes    int            `gorm:"not null;default:0"`                  // 总节点数
	SuccessNodes  int            `gorm:"not null;default:0"`                  // 成功节点数
	FailedNodes   int            `gorm:"not null;default:0"`                  // 失败节点数
	StartedAt     time.Time
	FinishedAt    *time.Time
	LogText       string `gorm:"type:text"` // 整体部署日志

	// 关联关系
	Application     *Application     `gorm:"foreignKey:ApplicationID"`
	Release         *Release         `gorm:"foreignKey:ReleaseID"`
	NodeDeployments []NodeDeployment `gorm:"foreignKey:MultiNodeDeploymentID"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (m *MultiNodeDeployment) BeforeCreate(tx *gorm.DB) (err error) {
	m.ID = uuid.New()
	return
}

// NodeDeployment 记录单个节点的部署状态
type NodeDeployment struct {
	ID                    uuid.UUID `gorm:"type:char(36);primary_key"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             gorm.DeletedAt `gorm:"index"`
	MultiNodeDeploymentID uuid.UUID      `gorm:"type:char(36);not null;index"`
	SSHHostID             uuid.UUID      `gorm:"type:char(36);not null;index"`
	Status                string         `gorm:"size:50;not null;default:'pending'"` // pending/running/success/failed
	LogText               string         `gorm:"type:text"`                          // 部署日志
	StartedAt             time.Time
	FinishedAt            *time.Time
	ErrorMessage          string `gorm:"type:text"` // 错误信息

	// 运行时信息
	ContainerID    string `gorm:"size:255"`                  // 容器ID
	SystemPort     int    `gorm:"default:0"`                 // 分配的系统端口
	HealthStatus   string `gorm:"size:50;default:'unknown'"` // 健康状态: healthy/unhealthy/unknown
	QuadletFile    string `gorm:"type:text"`                 // 生成的Quadlet配置
	EnvFile        string `gorm:"type:text"`                 // 环境变量文件内容

	// 关联关系
	SSHHost             *SSHHost             `gorm:"foreignKey:SSHHostID"`
	MultiNodeDeployment *MultiNodeDeployment `gorm:"foreignKey:MultiNodeDeploymentID"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (n *NodeDeployment) BeforeCreate(tx *gorm.DB) (err error) {
	n.ID = uuid.New()
	return
}

// TableName specifies the table name for the MultiNodeDeployment model
func (MultiNodeDeployment) TableName() string {
	return "multi_node_deployments"
}

// TableName specifies the table name for the NodeDeployment model
func (NodeDeployment) TableName() string {
	return "node_deployments"
}

// CreateMultiNodeDeployment creates a new multi-node deployment record
func CreateMultiNodeDeployment(applicationID, releaseID uuid.UUID, strategy string, hostIDs []uuid.UUID) (*MultiNodeDeployment, error) {
	multiDeploy := &MultiNodeDeployment{
		ApplicationID: applicationID,
		ReleaseID:     releaseID,
		Strategy:      strategy,
		Status:        "pending",
		TotalNodes:    len(hostIDs),
		StartedAt:     time.Now(),
	}

	tx := dborm.Db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建多节点部署记录
	if err := tx.Create(multiDeploy).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 为每个主机创建节点部署记录
	for _, hostID := range hostIDs {
		nodeDeploy := &NodeDeployment{
			MultiNodeDeploymentID: multiDeploy.ID,
			SSHHostID:             hostID,
			Status:                "pending",
			StartedAt:             time.Now(),
		}
		if err := tx.Create(nodeDeploy).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return multiDeploy, nil
}

// GetMultiNodeDeploymentByID retrieves a multi-node deployment by ID
func GetMultiNodeDeploymentByID(id uuid.UUID) (*MultiNodeDeployment, error) {
	var multiDeploy MultiNodeDeployment
	if err := dborm.Db.Preload("Application").Preload("Release").Preload("NodeDeployments.SSHHost").Where("id = ?", id).First(&multiDeploy).Error; err != nil {
		return nil, err
	}
	return &multiDeploy, nil
}

// GetNodeDeploymentByID retrieves a node deployment by ID
func GetNodeDeploymentByID(id uuid.UUID) (*NodeDeployment, error) {
	var nodeDeploy NodeDeployment
	if err := dborm.Db.Preload("SSHHost").Preload("MultiNodeDeployment").Where("id = ?", id).First(&nodeDeploy).Error; err != nil {
		return nil, err
	}
	return &nodeDeploy, nil
}

// ListMultiNodeDeploymentsByApp retrieves multi-node deployments by application ID
func ListMultiNodeDeploymentsByApp(appID uuid.UUID) ([]*MultiNodeDeployment, error) {
	var deployments []*MultiNodeDeployment
	if err := dborm.Db.Preload("Release").Preload("NodeDeployments.SSHHost").Where("application_id = ?", appID).Order("created_at DESC").Find(&deployments).Error; err != nil {
		return nil, err
	}
	return deployments, nil
}

// UpdateMultiNodeDeploymentStatus updates the status of multi-node deployment
func UpdateMultiNodeDeploymentStatus(id uuid.UUID, status string, logText string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if logText != "" {
		updates["log_text"] = logText
	}
	if status == "success" || status == "failed" || status == "partial" {
		now := time.Now()
		updates["finished_at"] = &now
	}
	return dborm.Db.Model(&MultiNodeDeployment{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateNodeDeploymentStatus updates the status of a node deployment
func UpdateNodeDeploymentStatus(id uuid.UUID, status, logText, errorMessage string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if logText != "" {
		updates["log_text"] = logText
	}
	if errorMessage != "" {
		updates["error_message"] = errorMessage
	}
	if status == "success" || status == "failed" {
		now := time.Now()
		updates["finished_at"] = &now
	}
	return dborm.Db.Model(&NodeDeployment{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateNodeDeploymentRuntimeInfo updates runtime information for a node deployment
func UpdateNodeDeploymentRuntimeInfo(id uuid.UUID, containerID string, systemPort int, healthStatus string) error {
	updates := map[string]interface{}{}
	if containerID != "" {
		updates["container_id"] = containerID
	}
	if systemPort > 0 {
		updates["system_port"] = systemPort
	}
	if healthStatus != "" {
		updates["health_status"] = healthStatus
	}
	return dborm.Db.Model(&NodeDeployment{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateNodeDeploymentFiles updates the generated files for a node deployment
func UpdateNodeDeploymentFiles(id uuid.UUID, quadletFile, envFile string) error {
	updates := map[string]interface{}{}
	if quadletFile != "" {
		updates["quadlet_file"] = quadletFile
	}
	if envFile != "" {
		updates["env_file"] = envFile
	}
	return dborm.Db.Model(&NodeDeployment{}).Where("id = ?", id).Updates(updates).Error
}

// GetPendingNodeDeployments retrieves all pending node deployments
func GetPendingNodeDeployments() ([]*NodeDeployment, error) {
	var nodeDeployments []*NodeDeployment
	if err := dborm.Db.Preload("SSHHost").Preload("MultiNodeDeployment.Application").Preload("MultiNodeDeployment.Release").Where("status = ?", "pending").Find(&nodeDeployments).Error; err != nil {
		return nil, err
	}
	return nodeDeployments, nil
}

// GetNodeDeploymentsByMultiNodeID retrieves node deployments by multi-node deployment ID
func GetNodeDeploymentsByMultiNodeID(multiNodeID uuid.UUID) ([]*NodeDeployment, error) {
	var nodeDeployments []*NodeDeployment
	if err := dborm.Db.Preload("SSHHost").Where("multi_node_deployment_id = ?", multiNodeID).Order("created_at ASC").Find(&nodeDeployments).Error; err != nil {
		return nil, err
	}
	return nodeDeployments, nil
}

// UpdateMultiNodeDeploymentProgress updates the progress counters
func UpdateMultiNodeDeploymentProgress(id uuid.UUID) error {
	var successCount, failedCount int64

	// 统计成功和失败的节点数
	dborm.Db.Model(&NodeDeployment{}).Where("multi_node_deployment_id = ? AND status = ?", id, "success").Count(&successCount)
	dborm.Db.Model(&NodeDeployment{}).Where("multi_node_deployment_id = ? AND status = ?", id, "failed").Count(&failedCount)

	return dborm.Db.Model(&MultiNodeDeployment{}).Where("id = ?", id).Updates(map[string]interface{}{
		"success_nodes": successCount,
		"failed_nodes":  failedCount,
	}).Error
}