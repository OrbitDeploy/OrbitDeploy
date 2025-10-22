package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/dborm"
	"gorm.io/gorm"
)

// SSHHost 存储远程VPS的连接信息
type SSHHost struct {
	ID          uuid.UUID `gorm:"type:char(36);primary_key"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	Name        string         `gorm:"size:255;not null;uniqueIndex"` // 主机标识名称
	Addr        string         `gorm:"size:255;not null"`             // IP地址或域名
	Port        int            `gorm:"not null;default:22"`           // SSH端口
	User        string         `gorm:"size:100;not null"`             // SSH用户名
	Password    string         `gorm:"size:255"`                      // SSH密码（加密存储）
	PrivateKey  string         `gorm:"type:text"`                     // SSH私钥（加密存储）
	Description string         `gorm:"size:500"`                      // 主机描述
	Status      string         `gorm:"size:50;default:'unknown'"`     // 连接状态: online/offline/error
	Region      string         `gorm:"size:100"`                      // 地理区域
	Tags        JSONB          `gorm:"type:jsonb"`                    // 标签信息，如 {"env": "prod", "zone": "cn-north"}

	// 性能参数
	CPUCores int `gorm:"default:0"` // CPU核心数
	MemoryGB int `gorm:"default:0"` // 内存GB
	DiskGB   int `gorm:"default:0"` // 磁盘GB

	// 管理字段
	IsActive    bool       `gorm:"not null;default:true"` // 是否启用
	LastCheckAt *time.Time                                // 最后检查时间
}

// BeforeCreate will set a UUID rather than numeric ID.
func (h *SSHHost) BeforeCreate(tx *gorm.DB) (err error) {
	h.ID = uuid.New()
	return
}

// TableName specifies the table name for the SSHHost model
func (SSHHost) TableName() string {
	return "ssh_hosts"
}

// CreateSSHHost creates a new SSH host record
func CreateSSHHost(name, addr, user, password, privateKey, description string, port int) (*SSHHost, error) {
	if port == 0 {
		port = 22
	}

	host := &SSHHost{
		Name:        name,
		Addr:        addr,
		Port:        port,
		User:        user,
		Password:    password,
		PrivateKey:  privateKey,
		Description: description,
		Status:      "unknown",
		IsActive:    true,
	}

	if err := dborm.Db.Create(host).Error; err != nil {
		return nil, err
	}

	return host, nil
}

// GetSSHHostByID retrieves an SSH host by its ID
func GetSSHHostByID(id uuid.UUID) (*SSHHost, error) {
	var host SSHHost
	if err := dborm.Db.Where("id = ?", id).First(&host).Error; err != nil {
		return nil, err
	}
	return &host, nil
}

// GetAllSSHHosts retrieves all SSH hosts
func GetAllSSHHosts() ([]*SSHHost, error) {
	var hosts []*SSHHost
	if err := dborm.Db.Find(&hosts).Error; err != nil {
		return nil, err
	}
	return hosts, nil
}

// UpdateSSHHost updates an existing SSH host
func UpdateSSHHost(id uuid.UUID, name, addr, user, password, privateKey, description string, port int) (*SSHHost, error) {
	host, err := GetSSHHostByID(id)
	if err != nil {
		return nil, err
	}

	if port == 0 {
		port = 22
	}

	host.Name = name
	host.Addr = addr
	host.Port = port
	host.User = user
	host.Password = password
	host.PrivateKey = privateKey
	host.Description = description

	if err := dborm.Db.Save(host).Error; err != nil {
		return nil, err
	}

	return host, nil
}

// DeleteSSHHost deletes an SSH host by its ID
func DeleteSSHHost(id uuid.UUID) error {
	return dborm.Db.Where("id = ?", id).Delete(&SSHHost{}).Error
}

// UpdateSSHHostStatus updates the connection status and last check time
func UpdateSSHHostStatus(id uuid.UUID, status string) error {
	now := time.Now()
	return dborm.Db.Model(&SSHHost{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       status,
		"last_check_at": &now,
	}).Error
}

// UpdateSSHHostResources updates the resource information
func UpdateSSHHostResources(id uuid.UUID, cpuCores, memoryGB, diskGB int) error {
	return dborm.Db.Model(&SSHHost{}).Where("id = ?", id).Updates(map[string]interface{}{
		"cpu_cores":  cpuCores,
		"memory_gb":  memoryGB,
		"disk_gb":    diskGB,
	}).Error
}

// GetActiveSSHHosts retrieves all active SSH hosts
func GetActiveSSHHosts() ([]*SSHHost, error) {
	var hosts []*SSHHost
	if err := dborm.Db.Where("is_active = ?", true).Find(&hosts).Error; err != nil {
		return nil, err
	}
	return hosts, nil
}

// GetSSHHostsByRegion retrieves SSH hosts by region
func GetSSHHostsByRegion(region string) ([]*SSHHost, error) {
	var hosts []*SSHHost
	if err := dborm.Db.Where("region = ? AND is_active = ?", region, true).Find(&hosts).Error; err != nil {
		return nil, err
	}
	return hosts, nil
}