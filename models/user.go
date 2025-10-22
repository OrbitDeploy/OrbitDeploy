package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/dborm"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User model for authentication
type User struct {
	ID               uuid.UUID `gorm:"type:char(36);primary_key"`
	Username         string    `json:"username" gorm:"unique;not null"`
	Password         string    `json:"-" gorm:"not null"` // Don't include in JSON
	TwoFactorSecret  string    `json:"-" gorm:"size:255"`
	TwoFactorEnabled bool      `json:"two_factor_enabled" gorm:"default:false"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New()
	return
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}

// HashPassword hashes the password using bcrypt
func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword verifies the password against the hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// CreateUser creates a new user with hashed password
func CreateUser(username, password string) (*User, error) {
	user := &User{
		Username: username,
	}

	if err := user.HashPassword(password); err != nil {
		return nil, err
	}

	if err := dborm.Db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(username string) (*User, error) {
	var user User
	if err := dborm.Db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserCount returns the number of users in the database
func GetUserCount() (int64, error) {
	var count int64
	if err := dborm.Db.Model(&User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetFirstUser returns the first user (for checking if setup is needed)
func GetFirstUser() (*User, error) {
	var user User
	if err := dborm.Db.First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(id uuid.UUID) (*User, error) {
	var user User
	if err := dborm.Db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUserPassword updates a user's password
func UpdateUserPassword(user *User) error {
	return dborm.Db.Save(user).Error
}

// Disable2FAForUser disables 2FA for a user and deletes their recovery codes in a transaction.
func Disable2FAForUser(userID uuid.UUID) error {
	return dborm.Db.Transaction(func(tx *gorm.DB) error {
		// 1. Update the user record
		userUpdate := map[string]interface{}{
			"two_factor_enabled": false,
			"two_factor_secret":  "",
		}
		if err := tx.Model(&User{}).Where("id = ?", userID).Updates(userUpdate).Error; err != nil {
			return err
		}

		// 2. Delete all recovery codes for the user
		if err := tx.Where("user_id = ?", userID).Delete(&TwoFactorRecoveryCode{}).Error; err != nil {
			return err
		}

		return nil
	})
}

// Enable2FAForUser enables 2FA for a user and stores recovery codes in a transaction.
func Enable2FAForUser(userID uuid.UUID, secret string, hashedCodes []string) error {
	return dborm.Db.Transaction(func(tx *gorm.DB) error {
		// 1. Update the user record
		userUpdate := map[string]interface{}{
			"two_factor_enabled": true,
			"two_factor_secret":  secret,
		}
		if err := tx.Model(&User{}).Where("id = ?", userID).Updates(userUpdate).Error; err != nil {
			return err
		}

		// 2. Create recovery codes
		if len(hashedCodes) > 0 {
			recoveryCodes := make([]TwoFactorRecoveryCode, len(hashedCodes))
			for i, code := range hashedCodes {
				recoveryCodes[i] = TwoFactorRecoveryCode{UserID: userID, Code: code}
			}
			if err := tx.Create(&recoveryCodes).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
