package services

import (
	"github.com/google/uuid"
	"github.com/youfun/OrbitDeploy/models"
	"gorm.io/gorm"
)

// DatabaseService handles business logic for self-hosted databases.
type DatabaseService struct {
	db *gorm.DB
}

// NewDatabaseService creates a new instance of DatabaseService.
func NewDatabaseService(db *gorm.DB) *DatabaseService {
	return &DatabaseService{db: db}
}

// CreateDatabase creates a new self-hosted database record.
func (s *DatabaseService) CreateDatabase(req *models.SelfHostedDatabase) (*models.SelfHostedDatabase, error) {
	db := &models.SelfHostedDatabase{
		Name:         req.Name,
		Type:         req.Type,
		Version:      req.Version,
		CustomImage:  req.CustomImage,
		Status:       models.DatabaseStatusPending,
		Port:         req.Port,
		InternalPort: req.InternalPort,
		Username:     req.Username,
		Password:     req.Password, // Note: Password should be encrypted in a real app
		DatabaseName: req.DatabaseName,
		DataPath:     req.DataPath,
		ConfigPath:   req.ConfigPath,
		IsRemote:     req.IsRemote,
		SSHHostID:    req.SSHHostID,
		ExtraConfig:  req.ExtraConfig,
	}

	if err := s.db.Create(db).Error; err != nil {
		return nil, err
	}
	return db, nil
}

// GetDatabaseByID retrieves a database by its UUID.
func (s *DatabaseService) GetDatabaseByID(id uuid.UUID) (*models.SelfHostedDatabase, error) {
	var db models.SelfHostedDatabase
	if err := s.db.First(&db, id).Error; err != nil {
		return nil, err
	}
	return &db, nil
}

// GetAllDatabases retrieves all database records.
func (s *DatabaseService) GetAllDatabases() ([]models.SelfHostedDatabase, error) {
	var dbs []models.SelfHostedDatabase
	if err := s.db.Find(&dbs).Error; err != nil {
		return nil, err
	}
	return dbs, nil
}

// DeleteDatabase deletes a database record by its UUID.
func (s *DatabaseService) DeleteDatabase(id uuid.UUID) error {
	return s.db.Delete(&models.SelfHostedDatabase{}, id).Error
}

// UpdateDatabase updates a database record.
func (s *DatabaseService) UpdateDatabase(db *models.SelfHostedDatabase) error {
	return s.db.Save(db).Error
}
