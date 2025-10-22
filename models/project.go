package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/dborm"
	"gorm.io/gorm"
)

// Project 是最高层级的实体，代表一个代码项目或一个逻辑分组。
type Project struct {
	ID           uuid.UUID `gorm:"type:char(36);primary_key"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Name         string         `gorm:"size:255;not null;uniqueIndex"`
	Applications []Application  `gorm:"foreignKey:ProjectID"` // 一个项目可以有多个应用 ，(e.g., prod, staging)
	HomeDir      string         `gorm:"size:1024;not null;default:''"`
	Description  string         `gorm:"type:text"`                    // 项目描述
	Username     string         `gorm:"size:255;not null;default:''"` // 为项目创建的专属 Linux 用户名, e.g., "my-web-app-user"
}

// BeforeCreate will set a UUID rather than numeric ID.
func (p *Project) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = uuid.New()
	return
}

// TableName specifies the table name for the Project model
func (Project) TableName() string {
	return "projects"
}

// CreateProject creates a new project record
func CreateProject(name string, description string) (*Project, error) {
	// Check if project name already exists
	var existingProject Project
	if err := dborm.Db.Where("name = ?", name).First(&existingProject).Error; err == nil {
		return nil, errors.New("项目名称已存在，请选择其他名称")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	project := &Project{
		Name:        name,
		Description: description,
	}

	if err := dborm.Db.Create(project).Error; err != nil {
		return nil, err
	}

	return project, nil
}

// CreateProjectWithSetup creates a new project record with setup information from project manager
// 这个函数集成了 project_service.go 中 Manager 的 Setup 流程结果
func CreateProjectWithSetup(name string, description string, homeDir string, username string) (*Project, error) {
	// Check if project name already exists
	var existingProject Project
	if err := dborm.Db.Where("name = ?", name).First(&existingProject).Error; err == nil {
		return nil, errors.New("项目名称已存在，请选择其他名称")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	project := &Project{
		Name:        name,
		Description: description,
		HomeDir:     homeDir,  // 来自 Manager.Setup() 的结果
		Username:    username, // 来自 Manager.Setup() 的结果
	}

	if err := dborm.Db.Create(project).Error; err != nil {
		return nil, err
	}

	return project, nil
}

// GetProjectByID retrieves a project by its ID
func GetProjectByID(id uuid.UUID) (*Project, error) {
	var project Project
	if err := dborm.Db.Where("id = ?", id).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// ListProjects retrieves all projects
func ListProjects() ([]*Project, error) {
	var projects []*Project
	if err := dborm.Db.Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

// UpdateProject updates an existing project
func UpdateProject(id uuid.UUID, name string) (*Project, error) {
	project, err := GetProjectByID(id)
	if err != nil {
		return nil, err
	}

	project.Name = name
	if err := dborm.Db.Save(project).Error; err != nil {
		return nil, err
	}

	return project, nil
}

// DeleteProject deletes a project by its ID
func DeleteProject(id uuid.UUID) error {
	return dborm.Db.Where("id = ?", id).Delete(&Project{}).Error
}

// GetProjectByName retrieves a project by its name
func GetProjectByName(name string) (*Project, error) {
	var project Project
	if err := dborm.Db.Where("name = ?", name).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}
