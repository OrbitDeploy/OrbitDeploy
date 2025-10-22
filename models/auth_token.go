package models


import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/dborm"
	"gorm.io/gorm"
)

// AuthToken represents a refresh token in the database
type AuthToken struct {
	ID                  uuid.UUID `gorm:"type:char(36);primary_key"`
	RefreshTokenHash    string    `json:"-" gorm:"unique;not null"` // SHA-256 hash of refresh token
	ClientDescription   string    `json:"client_description" gorm:"size:100"` // e.g., "Web Browser" or "CLI"
	ExpiresAt          time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (t *AuthToken) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	return
}

// TableName specifies the table name for the AuthToken model
func (AuthToken) TableName() string {
	return "auth_tokens"
}

// CreateAuthToken creates a new auth token record
func CreateAuthToken(tokenHash, clientDescription string, expiresAt time.Time) (*AuthToken, error) {
	authToken := &AuthToken{
		RefreshTokenHash:  tokenHash,
		ClientDescription: clientDescription,
		ExpiresAt:         expiresAt,
	}

	if err := dborm.Db.Create(authToken).Error; err != nil {
		return nil, err
	}

	return authToken, nil
}

// GetAuthTokenByHash retrieves an auth token by its hash
func GetAuthTokenByHash(tokenHash string) (*AuthToken, error) {
	var authToken AuthToken
	if err := dborm.Db.Where("refresh_token_hash = ? AND expires_at > ?", tokenHash, time.Now()).First(&authToken).Error; err != nil {
		return nil, err
	}
	return &authToken, nil
}

// DeleteAuthToken removes an auth token by its hash
func DeleteAuthToken(tokenHash string) error {
	return dborm.Db.Where("refresh_token_hash = ?", tokenHash).Delete(&AuthToken{}).Error
}

// DeleteExpiredAuthTokens removes all expired auth tokens
func DeleteExpiredAuthTokens() error {
	return dborm.Db.Where("expires_at <= ?", time.Now()).Delete(&AuthToken{}).Error
}

// DeleteAllAuthTokens removes all auth tokens (for logout all functionality)
func DeleteAllAuthTokens() error {
	return dborm.Db.Where("1 = 1").Delete(&AuthToken{}).Error
}

// IsRefreshTokenValid checks if a refresh token hash exists and is not expired
func IsRefreshTokenValid(tokenHash string) (bool, error) {
	var count int64
	err := dborm.Db.Model(&AuthToken{}).Where("refresh_token_hash = ? AND expires_at > ?", tokenHash, time.Now()).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CLIDeviceCode represents a device code for CLI authentication
type CLIDeviceCode struct {
	ID           uuid.UUID `gorm:"type:char(36);primary_key"`
	DeviceCode   string    `json:"device_code" gorm:"unique;not null;size:100"`
	UserCode     string    `json:"user_code" gorm:"unique;not null;size:20"`
	IsAuthorized bool      `json:"is_authorized" gorm:"default:false"`
	UserID       *uuid.UUID `gorm:"type:char(36);index"` // nullable until authorized
	ExpiresAt    time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (c *CLIDeviceCode) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return
}

// TableName specifies the table name for the CLIDeviceCode model
func (CLIDeviceCode) TableName() string {
	return "cli_device_codes"
}

// GenerateDeviceCodes creates a new device code and user code pair
func GenerateDeviceCodes() (*CLIDeviceCode, error) {
	// Generate device code (longer, machine-readable)
	deviceCodeBytes := make([]byte, 32)
	if _, err := rand.Read(deviceCodeBytes); err != nil {
		return nil, fmt.Errorf("failed to generate device code: %w", err)
	}
	deviceCode := base64.URLEncoding.EncodeToString(deviceCodeBytes)

	// Generate user code (shorter, human-readable)
	userCodeBytes := make([]byte, 6)
	if _, err := rand.Read(userCodeBytes); err != nil {
		return nil, fmt.Errorf("failed to generate user code: %w", err)
	}
	// Convert to uppercase alphanumeric for easier reading
	userCode := base64.RawStdEncoding.EncodeToString(userCodeBytes)[:8]

	cliCode := &CLIDeviceCode{
		DeviceCode: deviceCode,
		UserCode:   userCode,
		ExpiresAt:  time.Now().Add(10 * time.Minute), // 10 minutes expiry
	}

	if err := dborm.Db.Create(cliCode).Error; err != nil {
		return nil, err
	}

	return cliCode, nil
}

// GetDeviceCodeByDeviceCode retrieves a device code by device code
func GetDeviceCodeByDeviceCode(deviceCode string) (*CLIDeviceCode, error) {
	var cliCode CLIDeviceCode
	if err := dborm.Db.Where("device_code = ? AND expires_at > ?", deviceCode, time.Now()).First(&cliCode).Error; err != nil {
		return nil, err
	}
	return &cliCode, nil
}

// GetDeviceCodeByUserCode retrieves a device code by user code
func GetDeviceCodeByUserCode(userCode string) (*CLIDeviceCode, error) {
	var cliCode CLIDeviceCode
	if err := dborm.Db.Where("user_code = ? AND expires_at > ?", userCode, time.Now()).First(&cliCode).Error; err != nil {
		return nil, err
	}
	return &cliCode, nil
}

// AuthorizeDeviceCode authorizes a device code for a specific user
func AuthorizeDeviceCode(userCode string, userID uuid.UUID) error {
	return dborm.Db.Model(&CLIDeviceCode{}).Where("user_code = ? AND expires_at > ?", userCode, time.Now()).Updates(map[string]interface{}{
		"is_authorized": true,
		"user_id":       userID,
	}).Error
}

// DeleteExpiredDeviceCodes removes all expired device codes
func DeleteExpiredDeviceCodes() error {
	return dborm.Db.Where("expires_at <= ?", time.Now()).Delete(&CLIDeviceCode{}).Error
}