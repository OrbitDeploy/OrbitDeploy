package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/dborm"
	"gorm.io/gorm"
)

// ProviderAuth 存储第三方代码仓库平台的授权信息
type ProviderAuth struct {
	ID             uuid.UUID `gorm:"type:char(36);primary_key"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	Platform       string         `gorm:"size:50;not null;index"` // 平台名称（小写），如 'github', 'gitlab', 'bitbucket', 'gitea'

	// OAuth2 客户端凭据（用于gitlab、gitea等）
	ClientID     string `gorm:"size:255"` // OAuth2客户端ID
	ClientSecret string `gorm:"size:255"` // OAuth2客户端密钥（加密存储）
	RedirectURI  string `gorm:"size:500"` // OAuth回调URI

	// Bitbucket App Password专用字段
	Username    string `gorm:"size:255"` // Bitbucket专用用户名
	AppPassword string `gorm:"size:255"` // Bitbucket专用应用密码（加密存储）

	// GitHub Apps专用字段
	AppID          string `gorm:"size:255"`  // GitHub App唯一ID
	Slug           string `gorm:"size:255"`  // GitHub App Slug（用于安装URL）
	PrivateKey     string `gorm:"type:text"` // PEM格式私钥，加密存储，用于JWT生成
	WebhookSecret  string `gorm:"size:255"`  // Webhook验证密钥（可选）
	InstallationID uint   `gorm:"default:0"` // GitHub安装ID（安装后生成）

	// 通用字段
	Scopes   string `gorm:"size:500"`              // 授权范围，JSON格式存储
	IsActive bool   `gorm:"not null;default:true"` // 启用/禁用状态

}

// BeforeCreate will set a UUID rather than numeric ID.
func (pa *ProviderAuth) BeforeCreate(tx *gorm.DB) (err error) {
	pa.ID = uuid.New()
	return
}

// TableName specifies the table name for the ProviderAuth model
func (ProviderAuth) TableName() string {
	return "provider_auths"
}

// 临时加密密钥（生产环境应从配置或环境变量读取）
var providerAuthEncryptionKey = []byte("provider-auth-encryption-key-32b") // 32字节密钥

// BeforeSave GORM钩子，自动加密敏感字段
func (pa *ProviderAuth) BeforeSave(tx *gorm.DB) error {
	if pa.ClientSecret != "" && !isProviderAuthEncrypted(pa.ClientSecret) {
		encrypted, err := encryptProviderAuth(pa.ClientSecret)
		if err != nil {
			return err
		}
		pa.ClientSecret = encrypted
	}

	if pa.AppPassword != "" && !isProviderAuthEncrypted(pa.AppPassword) {
		encrypted, err := encryptProviderAuth(pa.AppPassword)
		if err != nil {
			return err
		}
		pa.AppPassword = encrypted
	}

	if pa.PrivateKey != "" && !isProviderAuthEncrypted(pa.PrivateKey) {
		encrypted, err := encryptProviderAuth(pa.PrivateKey)
		if err != nil {
			return err
		}
		pa.PrivateKey = encrypted
	}

	return nil
}

// AfterFind GORM钩子，自动解密敏感字段
func (pa *ProviderAuth) AfterFind(tx *gorm.DB) error {
	if pa.ClientSecret != "" && isProviderAuthEncrypted(pa.ClientSecret) {
		decrypted, err := decryptProviderAuth(pa.ClientSecret)
		if err != nil {
			return err
		}
		pa.ClientSecret = decrypted
	}

	if pa.AppPassword != "" && isProviderAuthEncrypted(pa.AppPassword) {
		decrypted, err := decryptProviderAuth(pa.AppPassword)
		if err != nil {
			return err
		}
		pa.AppPassword = decrypted
	}

	if pa.PrivateKey != "" && isProviderAuthEncrypted(pa.PrivateKey) {
		decrypted, err := decryptProviderAuth(pa.PrivateKey)
		if err != nil {
			return err
		}
		pa.PrivateKey = decrypted
	}

	return nil
}

// 加密函数
func encryptProviderAuth(plaintext string) (string, error) {
	block, err := aes.NewCipher(providerAuthEncryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// 解密函数
func decryptProviderAuth(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(providerAuthEncryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(data) < gcm.NonceSize() {
		return "", errors.New("malformed ciphertext")
	}

	nonce, ciphertextBytes := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// 检查字符串是否已加密（简单检查是否为Base64格式）
func isProviderAuthEncrypted(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil && len(s) > 20 // 加密后的字符串通常较长
}

// CreateProviderAuth 创建新的第三方平台授权记录
func CreateProviderAuth(platform string, clientID, clientSecret, redirectURI, username, appPassword, appID, slug, privateKey, webhookSecret string, installationID uint, scopes string) (*ProviderAuth, error) {
	// 平台名称统一小写
	platform = strings.ToLower(platform)

	// 检查同一平台同一应用是否已存在授权 (removed user check for self-deployment)
	var existing ProviderAuth
	query := dborm.Db.Where("platform = ?", platform)

	if err := query.First(&existing).Error; err == nil {
		return nil, errors.New("该平台在此应用的授权已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	providerAuth := &ProviderAuth{
		// UserID:         userID, - removed
		Platform:       platform,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		RedirectURI:    redirectURI,
		Username:       username,
		AppPassword:    appPassword,
		AppID:          appID,
		Slug:           slug,
		PrivateKey:     privateKey,
		WebhookSecret:  webhookSecret,
		InstallationID: installationID,
		Scopes:         scopes,
		IsActive:       true,
	}

	if err := dborm.Db.Create(providerAuth).Error; err != nil {
		return nil, err
	}

	return providerAuth, nil
}

// GetProviderAuthByID 根据ID获取授权记录
func GetProviderAuthByID(id uuid.UUID) (*ProviderAuth, error) {
	var providerAuth ProviderAuth
	if err := dborm.Db.Where("id = ?", id).First(&providerAuth).Error; err != nil {
		return nil, err
	}
	return &providerAuth, nil
}

// ListProviderAuths 获取所有授权记录
func ListProviderAuths() ([]*ProviderAuth, error) {
	var providerAuths []*ProviderAuth
	if err := dborm.Db.Find(&providerAuths).Error; err != nil {
		return nil, err
	}
	return providerAuths, nil
}

// ListProviderAuthsByPlatform 获取特定平台的授权记录
func ListProviderAuthsByPlatform(platform string) ([]*ProviderAuth, error) {
	var providerAuths []*ProviderAuth
	platform = strings.ToLower(platform)
	if err := dborm.Db.Where("platform = ?", platform).Find(&providerAuths).Error; err != nil {
		return nil, err
	}
	return providerAuths, nil
}

// UpdateProviderAuth 更新授权记录
func UpdateProviderAuth(id uuid.UUID, clientID, clientSecret, redirectURI, username, appPassword, appID, slug, privateKey, webhookSecret string, installationID uint, scopes string, isActive bool) (*ProviderAuth, error) {
	providerAuth, err := GetProviderAuthByID(id)
	if err != nil {
		return nil, err
	}

	// Log before update
	fmt.Printf("Before update: ProviderAuth ID=%d, InstallationID=%d\n", providerAuth.ID, providerAuth.InstallationID)

	providerAuth.ClientID = clientID
	providerAuth.ClientSecret = clientSecret
	providerAuth.RedirectURI = redirectURI
	providerAuth.Username = username
	providerAuth.AppPassword = appPassword
	providerAuth.AppID = appID
	providerAuth.Slug = slug
	providerAuth.PrivateKey = privateKey
	providerAuth.WebhookSecret = webhookSecret
	providerAuth.InstallationID = installationID
	providerAuth.Scopes = scopes
	providerAuth.IsActive = isActive

	if err := dborm.Db.Save(providerAuth).Error; err != nil {
		return nil, err
	}

	// Log after update
	fmt.Printf("After update: ProviderAuth ID=%d, InstallationID=%d\n", providerAuth.ID, providerAuth.InstallationID)

	return providerAuth, nil
}

// DeleteProviderAuth 删除授权记录
func DeleteProviderAuth(id uuid.UUID) error {
	return dborm.Db.Where("id = ?", id).Delete(&ProviderAuth{}).Error
}

// ActivateProviderAuth 激活授权记录
func ActivateProviderAuth(id uuid.UUID) error {
	return dborm.Db.Model(&ProviderAuth{}).Where("id = ?", id).Update("is_active", true).Error
}

// DeactivateProviderAuth 停用授权记录
func DeactivateProviderAuth(id uuid.UUID) error {
	return dborm.Db.Model(&ProviderAuth{}).Where("id = ?", id).Update("is_active", false).Error
}
