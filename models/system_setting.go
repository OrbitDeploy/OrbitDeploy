package models

import (
	"log"

	"github.com/opentdp/go-helper/dborm"
	"gorm.io/gorm"
)

// SystemSetting stores key-value pairs for system-wide settings.
type SystemSetting struct {
	gorm.Model
	Key   string `gorm:"uniqueIndex;not null"`
	Value string `gorm:"type:text"`
}

// GetSystemSetting retrieves a system setting by its key.
func GetSystemSetting(key string) (string, error) {
	var setting SystemSetting
	result := dborm.Db.Where("key = ?", key).First(&setting)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return "", nil // Return empty string if not found, not an error
		}
		return "", result.Error
	}
	return setting.Value, nil
}

// SetSystemSetting creates or updates a system setting.
func SetSystemSetting(key string, value string) error {
	log.Printf("SetSystemSetting called for key: %s, value: %s", key, value)
	var setting SystemSetting
	result := dborm.Db.Where("key = ?", key).First(&setting)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			log.Printf("Key %s not found, creating new setting.", key)
			// Create new setting
			setting = SystemSetting{Key: key, Value: value}
			err := dborm.Db.Create(&setting).Error
			if err != nil {
				log.Printf("Error creating setting for key %s: %v", key, err)
			}
			return err
		}
		log.Printf("Error finding setting for key %s: %v", key, result.Error)
		return result.Error
	}

	log.Printf("Key %s found, updating existing setting. Old value: %s, New value: %s", key, setting.Value, value)
	// Update existing setting
	setting.Value = value
	err := dborm.Db.Save(&setting).Error
	if err != nil {
		log.Printf("Error updating setting for key %s: %v", key, err)
	}
	return err
}
