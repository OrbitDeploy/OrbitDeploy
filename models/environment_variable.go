package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/dborm"
	"github.com/youfun/OrbitDeploy/utils"
	"gorm.io/gorm"
)

// EnvironmentVariable 存储应用的单个环境变量，支持选择性加密
type EnvironmentVariable struct {
	ID            uuid.UUID `gorm:"type:char(36);primary_key"`
	ApplicationID uuid.UUID `gorm:"type:char(36);not null;index"`
	Key           string    `gorm:"not null;size:255"`
	Value         string    `gorm:"type:text"`                  // 存储加密后的值
	IsEncrypted   bool      `gorm:"not null;default:false"`    // 是否加密
	CreatedAt     time.Time
	UpdatedAt     time.Time

	// Relationships
	Application Application `gorm:"foreignKey:ApplicationID"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (ev *EnvironmentVariable) BeforeCreate(tx *gorm.DB) (err error) {
	ev.ID = uuid.New()
	return
}

// TableName specifies the table name for the EnvironmentVariable model
func (EnvironmentVariable) TableName() string {
	return "environment_variables"
}

// CreateEnvironmentVariable creates a new environment variable
func CreateEnvironmentVariable(applicationID uuid.UUID, key, value string, isEncrypted bool) (*EnvironmentVariable, error) {
	var storedValue string
	var err error

	if isEncrypted {
		storedValue, err = utils.EncryptValue(value)
		if err != nil {
			return nil, err
		}
	} else {
		storedValue = value
	}

	envVar := &EnvironmentVariable{
		ApplicationID: applicationID,
		Key:           key,
		Value:         storedValue,
		IsEncrypted:   isEncrypted,
	}

	if err := dborm.Db.Create(envVar).Error; err != nil {
		return nil, err
	}

	return envVar, nil
}

// CreateEnvironmentVariableInTx creates a new environment variable within a transaction
func CreateEnvironmentVariableInTx(tx *gorm.DB, applicationID uuid.UUID, key, value string, isEncrypted bool) (*EnvironmentVariable, error) {
	var storedValue string
	var err error

	if isEncrypted {
		storedValue, err = utils.EncryptValue(value)
		if err != nil {
			return nil, err
		}
	} else {
		storedValue = value
	}

	envVar := &EnvironmentVariable{
		ApplicationID: applicationID,
		Key:           key,
		Value:         storedValue,
		IsEncrypted:   isEncrypted,
	}

	if err := tx.Create(envVar).Error; err != nil {
		return nil, err
	}

	return envVar, nil
}

// GetEnvironmentVariableByID retrieves an environment variable by its ID
func GetEnvironmentVariableByID(id uuid.UUID) (*EnvironmentVariable, error) {
	var envVar EnvironmentVariable
	if err := dborm.Db.Where("id = ?", id).First(&envVar).Error; err != nil {
		return nil, err
	}
	return &envVar, nil
}

// ListEnvironmentVariablesByApplicationID retrieves all environment variables for an application
func ListEnvironmentVariablesByApplicationID(applicationID uuid.UUID) ([]*EnvironmentVariable, error) {
	var envVars []*EnvironmentVariable
	if err := dborm.Db.Where("application_id = ?", applicationID).Find(&envVars).Error; err != nil {
		return nil, err
	}
	return envVars, nil
}

// UpdateEnvironmentVariable updates an existing environment variable
func UpdateEnvironmentVariable(id uuid.UUID, key, value string, isEncrypted bool) (*EnvironmentVariable, error) {
	envVar, err := GetEnvironmentVariableByID(id)
	if err != nil {
		return nil, err
	}

	var storedValue string
	if isEncrypted {
		storedValue, err = utils.EncryptValue(value)
		if err != nil {
			return nil, err
		}
	} else {
		storedValue = value
	}

	envVar.Key = key
	envVar.Value = storedValue
	envVar.IsEncrypted = isEncrypted

	if err := dborm.Db.Save(envVar).Error; err != nil {
		return nil, err
	}

	return envVar, nil
}

// DeleteEnvironmentVariable deletes an environment variable by its ID
func DeleteEnvironmentVariable(id uuid.UUID) error {
	return dborm.Db.Where("id = ?", id).Delete(&EnvironmentVariable{}).Error
}

// DeleteEnvironmentVariablesByApplicationID deletes all environment variables for an application
func DeleteEnvironmentVariablesByApplicationID(applicationID uuid.UUID) error {
	return dborm.Db.Where("application_id = ?", applicationID).Delete(&EnvironmentVariable{}).Error
}

// GetDecryptedValue returns the decrypted value of an environment variable
func (env *EnvironmentVariable) GetDecryptedValue() (string, error) {
	if !env.IsEncrypted {
		return env.Value, nil
	}
	return utils.DecryptValue(env.Value)
}

// GenerateEnvFileContent generates .env file content from environment variables
func GenerateEnvFileContent(applicationID uuid.UUID) (string, error) {
	envVars, err := ListEnvironmentVariablesByApplicationID(applicationID)
	if err != nil {
		return "", err
	}

	var content string
	for _, envVar := range envVars {
		value, err := envVar.GetDecryptedValue()
		if err != nil {
			return "", err
		}
		content += envVar.Key + "=" + value + "\n"
	}

	return content, nil
}

// CreateSnapshotForDeployment creates a snapshot of environment variables for deployment
func CreateSnapshotForDeployment(applicationID uuid.UUID) (string, error) {
	envVars, err := ListEnvironmentVariablesByApplicationID(applicationID)
	if err != nil {
		return "", err
	}

	var snapshot []map[string]interface{}
	for _, envVar := range envVars {
		value, err := envVar.GetDecryptedValue()
		if err != nil {
			return "", err
		}
		snapshot = append(snapshot, map[string]interface{}{
			"key":         envVar.Key,
			"value":       value,
			"isEncrypted": envVar.IsEncrypted,
		})
	}

	// Convert to JSON string for storage
	jsonBytes, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}