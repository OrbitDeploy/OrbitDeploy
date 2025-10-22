package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/dborm"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// TwoFactorRecoveryCode model for 2FA recovery
type TwoFactorRecoveryCode struct {
	ID        uuid.UUID `gorm:"type:char(36);primary_key"`
	UserID    uuid.UUID `gorm:"type:char(36);not null;index"`
	Code      string    `gorm:"type:varchar(255);not null;uniqueIndex"` // Stores the hashed recovery code
	Used      bool      `gorm:"default:false"`
	CreatedAt time.Time
}

// BeforeCreate will set a UUID rather than numeric ID.
func (rc *TwoFactorRecoveryCode) BeforeCreate(tx *gorm.DB) (err error) {
	rc.ID = uuid.New()
	return
}

// TableName specifies the table name for the TwoFactorRecoveryCode model
func (TwoFactorRecoveryCode) TableName() string {
	return "two_factor_recovery_codes"
}

// UseRecoveryCode finds an unused recovery code for a user, validates it, and marks it as used.
// Returns true if the code was valid and successfully used.
func UseRecoveryCode(userID uuid.UUID, plainCode string) (bool, error) {
	var recoveryCodes []TwoFactorRecoveryCode
	err := dborm.Db.Where("user_id = ? AND used = ?", userID, false).Find(&recoveryCodes).Error
	if err != nil {
		return false, err
	}

	for _, rc := range recoveryCodes {
		if bcrypt.CompareHashAndPassword([]byte(rc.Code), []byte(plainCode)) == nil {
			// Found a valid, unused code. Mark it as used.
			rc.Used = true
			if err := dborm.Db.Save(&rc).Error; err != nil {
				return false, err
			}
			return true, nil
		}
	}

	// No valid, unused code found
	return false, nil
}
