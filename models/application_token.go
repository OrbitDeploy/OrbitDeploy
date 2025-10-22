package models

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/dborm"
	"github.com/youfun/OrbitDeploy/utils"
	"gorm.io/gorm"
)

// ApplicationToken represents an access token for a specific application
type ApplicationToken struct {
	ID            uuid.UUID `gorm:"type:char(36);primary_key"`
	ApplicationID uuid.UUID `gorm:"type:char(36);not null;index"`
	Name          string    `json:"name" gorm:"not null;size:100"`           // 用户自定义的令牌名称
	TokenHash     string    `json:"-" gorm:"not null;size:500"`              // 加密后的token
	ExpiresAt     *time.Time `json:"expires_at"`                              // 令牌过期时间（可选）
	LastUsedAt    *time.Time `json:"last_used_at"`                            // 最后使用时间
	IsActive      bool      `json:"is_active" gorm:"default:true;not null"`  // 是否激活
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// 关联关系
	Application Application `json:"application,omitempty" gorm:"foreignKey:ApplicationID"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (t *ApplicationToken) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	return
}

// TableName specifies the table name for the ApplicationToken model
func (ApplicationToken) TableName() string {
	return "application_tokens"
}

// encryptAppToken 加密应用令牌
// 使用统一的加密工具，通过环境变量 ORBIT_ENCRYPTION_KEY 配置加密密钥
func encryptAppToken(plaintext string) (string, error) {
	return utils.EncryptValue(plaintext)
}

// decryptAppToken 解密应用令牌
// 使用统一的加密工具，通过环境变量 ORBIT_ENCRYPTION_KEY 配置加密密钥
func decryptAppToken(ciphertext string) (string, error) {
	return utils.DecryptValue(ciphertext)
}

// GenerateRandomToken generates a random token string
func GenerateRandomToken() (string, error) {
	bytes := make([]byte, 32) // 256-bit token
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// CreateApplicationToken creates a new application token record
func CreateApplicationToken(applicationID uuid.UUID, name string, expiresAt *time.Time) (*ApplicationToken, string, error) {
	if name == "" {
		return nil, "", fmt.Errorf("name is required")
	}

	// 生成随机令牌
	token, err := GenerateRandomToken()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	// 加密令牌
	encryptedToken, err := encryptAppToken(token)
	if err != nil {
		return nil, "", fmt.Errorf("failed to encrypt token: %w", err)
	}

	appToken := &ApplicationToken{
		ApplicationID: applicationID,
		Name:          name,
		TokenHash:     encryptedToken,
		ExpiresAt:     expiresAt,
		IsActive:      true,
	}

	if err := dborm.Db.Create(appToken).Error; err != nil {
		return nil, "", err
	}

	return appToken, token, nil
}

// GetApplicationTokensByAppID retrieves all application tokens for an app (without decrypted tokens)
func GetApplicationTokensByAppID(applicationID uuid.UUID) ([]ApplicationToken, error) {
	var tokens []ApplicationToken
	if err := dborm.Db.Where("application_id = ? AND is_active = ?", applicationID, true).
		Order("created_at DESC").Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

// GetApplicationTokenByID retrieves an application token by its ID
func GetApplicationTokenByID(tokenID uuid.UUID, applicationID uuid.UUID, decrypt bool) (*ApplicationToken, string, error) {
	var token ApplicationToken
	if err := dborm.Db.Where("id = ? AND application_id = ? AND is_active = ?", 
		tokenID, applicationID, true).First(&token).Error; err != nil {
		return nil, "", err
	}

	var decryptedToken string
	if decrypt {
		var err error
		decryptedToken, err = decryptAppToken(token.TokenHash)
		if err != nil {
			return nil, "", fmt.Errorf("failed to decrypt token: %w", err)
		}
	}

	return &token, decryptedToken, nil
}

// ValidateApplicationToken validates a token and returns the associated application and token info
func ValidateApplicationToken(tokenString string) (*Application, *ApplicationToken, error) {
	// 获取所有激活的token
	var tokens []ApplicationToken
	if err := dborm.Db.Where("is_active = ?", true).
		Preload("Application").Find(&tokens).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to query tokens: %w", err)
	}

	// 尝试匹配token
	for _, token := range tokens {
		// 检查是否过期
		if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
			continue
		}

		// 解密并比较token
		decryptedToken, err := decryptAppToken(token.TokenHash)
		if err != nil {
			continue // 解密失败，跳过
		}

		if decryptedToken == tokenString {
			// 找到匹配的token，更新最后使用时间
			if err := UpdateAppTokenLastUsed(token.ID); err != nil {
				// 记录警告但不影响主要功能
				fmt.Printf("Warning: failed to update token last used time: %v\n", err)
			}
			return &token.Application, &token, nil
		}
	}

	return nil, nil, fmt.Errorf("invalid token")
}

// UpdateApplicationToken updates an application token (excluding the token itself)
func UpdateApplicationToken(tokenID uuid.UUID, applicationID uuid.UUID, name string, expiresAt *time.Time) error {
	updates := map[string]interface{}{
		"name":       name,
		"expires_at": expiresAt,
		"updated_at": time.Now(),
	}

	return dborm.Db.Model(&ApplicationToken{}).
		Where("id = ? AND application_id = ?", tokenID, applicationID).
		Updates(updates).Error
}

// DeleteApplicationToken soft deletes an application token
func DeleteApplicationToken(tokenID uuid.UUID, applicationID uuid.UUID) error {
	return dborm.Db.Model(&ApplicationToken{}).
		Where("id = ? AND application_id = ?", tokenID, applicationID).
		Update("is_active", false).Error
}

// UpdateAppTokenLastUsed updates the last used timestamp for a token
func UpdateAppTokenLastUsed(tokenID uuid.UUID) error {
	now := time.Now()
	return dborm.Db.Model(&ApplicationToken{}).
		Where("id = ?", tokenID).
		Update("last_used_at", &now).Error
}