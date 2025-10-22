package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/dborm"
	"gorm.io/gorm"
)

// Routing 将公网的域名和端口映射到一个应用。
type Routing struct {
	ID            uuid.UUID `gorm:"type:char(36);primary_key"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	ApplicationID uuid.UUID      `gorm:"type:char(36);not null;index"`
	DomainName    string         `gorm:"size:255;not null;uniqueIndex"`
	HostPort      int            `gorm:"not null;uniqueIndex"` // 主机上暴露的端口，必须唯一
	IsActive      bool           `gorm:"not null;default:true"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (r *Routing) BeforeCreate(tx *gorm.DB) (err error) {
	r.ID = uuid.New()
	return
}

// TableName specifies the table name for the Routing model
func (Routing) TableName() string {
	return "routings"
}

// CreateRouting creates a new routing record
func CreateRouting(applicationID uuid.UUID, domainName string, hostPort int, isActive bool) (*Routing, error) {
	routing := &Routing{
		ApplicationID: applicationID,
		DomainName:    domainName,
		HostPort:      hostPort,
		IsActive:      isActive,
		CreatedAt:     time.Now(),
	}

	if err := dborm.Db.Create(routing).Error; err != nil {
		return nil, err
	}

	return routing, nil
}

// GetRoutingByID retrieves a routing by its ID
func GetRoutingByID(id uuid.UUID) (*Routing, error) {
	var routing Routing
	if err := dborm.Db.Where("id = ?", id).First(&routing).Error; err != nil {
		return nil, err
	}
	return &routing, nil
}

// ListRoutings retrieves all routings
func ListRoutings() ([]*Routing, error) {
	var routings []*Routing
	if err := dborm.Db.Find(&routings).Error; err != nil {
		return nil, err
	}
	return routings, nil
}

// ListRoutingsByAppID retrieves all routings for a specific application
func ListRoutingsByAppID(applicationID uuid.UUID) ([]*Routing, error) {
	var routings []*Routing
	if err := dborm.Db.Where("application_id = ?", applicationID).Order("created_at desc").Find(&routings).Error; err != nil {
		return nil, err
	}
	return routings, nil
}

// UpdateRouting updates an existing routing
func UpdateRouting(id uuid.UUID, domainName string, hostPort int, isActive bool) (*Routing, error) {
	routing, err := GetRoutingByID(id)
	if err != nil {
		return nil, err
	}

	routing.DomainName = domainName
	routing.HostPort = hostPort
	routing.IsActive = isActive
	if err := dborm.Db.Save(routing).Error; err != nil {
		return nil, err
	}

	return routing, nil
}

// DeleteRouting deletes a routing by its ID
func DeleteRouting(id uuid.UUID) error {
	return dborm.Db.Where("id = ?", id).Delete(&Routing{}).Error
}

func GetActiveRoutingsByApplicationID(applicationID uuid.UUID) ([]*Routing, error) {
	var routings []*Routing
	if err := dborm.Db.Where("application_id = ? AND is_active = ?", applicationID, true).Find(&routings).Error; err != nil {
		return nil, err
	}
	return routings, nil
}
