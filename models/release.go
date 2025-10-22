package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/dborm"
	"gorm.io/gorm"
)

// Release 代表一次成功的构建产物，是一个不可变的 Docker 镜像版本。
type Release struct {
	ID              uuid.UUID `gorm:"type:char(36);primary_key"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
	ApplicationID   uuid.UUID      `gorm:"type:char(36);not null;index"`
	Version         string         `gorm:"size:50;not null;default:''"`
	ImageName       string         `gorm:"size:255;not null"`                  // 最终的镜像名称和标签
	BuildSourceInfo JSONB          `gorm:"type:jsonb"`                         // 构建源信息, e.g., {"commit_sha": "...", "branch": "main"}
	Status          string         `gorm:"size:50;not null;default:'pending'"` // 构建状态 (pending, building, success, failed)
	// SystemPort      *int   `gorm:"default:null"`                       // 系统分配的端口，可选字段
}

// BeforeCreate will set a UUID rather than numeric ID.
func (r *Release) BeforeCreate(tx *gorm.DB) (err error) {
	r.ID = uuid.New()
	return
}

// TableName specifies the table name for the Release model
func (Release) TableName() string {
	return "releases"
}

// CreateRelease creates a new release record
func CreateRelease(applicationID uuid.UUID, imageName string, buildSourceInfo JSONB, status string) (*Release, error) {
	release := &Release{
		ApplicationID:   applicationID,
		ImageName:       imageName,
		BuildSourceInfo: buildSourceInfo,
		Status:          status,
	}

	if err := dborm.Db.Create(release).Error; err != nil {
		return nil, err
	}

	return release, nil
}

// CreateReleaseWithVersion creates a new release record with version
func CreateReleaseWithVersion(applicationID uuid.UUID, version, imageName string, buildSourceInfo JSONB, status string) (*Release, error) {
	release := &Release{
		ApplicationID:   applicationID,
		Version:         version,
		ImageName:       imageName,
		BuildSourceInfo: buildSourceInfo,
		Status:          status,
	}

	if err := dborm.Db.Create(release).Error; err != nil {
		return nil, err
	}

	return release, nil
}

// GetReleaseByID retrieves a release by its ID
func GetReleaseByID(id uuid.UUID) (*Release, error) {
	var release Release
	if err := dborm.Db.Where("id = ?", id).First(&release).Error; err != nil {
		return nil, err
	}
	return &release, nil
}

func GetLatestRelease(appID uuid.UUID) (*Release, error) {
	var release Release
	if err := dborm.Db.Where("application_id = ?", appID).Order("created_at DESC").First(&release).Error; err != nil {
		return nil, err
	}
	return &release, nil
}

// ListReleases retrieves all releases
func ListReleases() ([]*Release, error) {
	var releases []*Release
	if err := dborm.Db.Find(&releases).Error; err != nil {
		return nil, err
	}
	return releases, nil
}

// ListReleasesByAppID retrieves all releases for a specific application
func ListReleasesByAppID(appID uuid.UUID) ([]*Release, error) {
	var releases []*Release
	if err := dborm.Db.Where("application_id = ?", appID).Order("created_at desc").Find(&releases).Error; err != nil {
		return nil, err
	}
	return releases, nil
}

// UpdateRelease updates an existing release
func UpdateRelease(id uuid.UUID, imageName string, buildSourceInfo JSONB, status string) (*Release, error) {
	release, err := GetReleaseByID(id)
	if err != nil {
		return nil, err
	}

	release.ImageName = imageName
	release.BuildSourceInfo = buildSourceInfo
	release.Status = status
	if err := dborm.Db.Save(release).Error; err != nil {
		return nil, err
	}

	return release, nil
}

// UpdateReleaseVersion updates the version of an existing release
func UpdateReleaseVersion(id uuid.UUID, version string) error {
	return dborm.Db.Model(&Release{}).Where("id = ?", id).Update("version", version).Error
}

// DeleteRelease deletes a release by its ID
func DeleteRelease(id uuid.UUID) error {
	return dborm.Db.Where("id = ?", id).Delete(&Release{}).Error
}

// IsPortInUse checks if the given port is already assigned to any Release
// func IsPortInUse(port int) (bool, error) {
// 	var count int64
// 	err := dborm.Db.Model(&Release{}).Where("system_port = ?", port).Count(&count).Error
// 	if err != nil {
// 		return false, err
// 	}
// 	return count > 0, nil
// }
// func UpdateReleaseSystemPort(releaseID uuid.UUID, systemPort int) error {
// 	return dborm.Db.Model(&Release{}).Where("id = ?", releaseID).Update("system_port", systemPort).Error
// }
