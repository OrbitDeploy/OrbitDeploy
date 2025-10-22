package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/dborm"
	"gorm.io/gorm"
)

// GitHubToken represents a GitHub access token for private repository access
type GitHubToken struct {
	ID          uuid.UUID `gorm:"type:char(36);primary_key"`
	UserID      uuid.UUID `gorm:"type:char(36);not null;index"`
	Name        string    `json:"name" gorm:"not null;size:100"`           // 用户自定义的令牌名称
	TokenHash   string    `json:"-" gorm:"not null;size:500"`              // 加密后的token
	Permissions string    `json:"permissions" gorm:"type:text"`            // JSON格式的权限范围
	ExpiresAt   *time.Time `json:"expires_at"`                             // 令牌过期时间（可选）
	LastUsedAt  *time.Time `json:"last_used_at"`                           // 最后使用时间
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (t *GitHubToken) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	return
}

func (GitHubToken) TableName() string {
	return "github_tokens"
}

// ProjectCredential represents the association between projects and their authentication credentials
type ProjectCredential struct {
	ID             uuid.UUID `gorm:"type:char(36);primary_key"`
	ProjectID      uuid.UUID `gorm:"type:char(36);not null;index;uniqueIndex:uniq_project_credential"`
	CredentialType string    `json:"credential_type" gorm:"not null;size:20;uniqueIndex:uniq_project_credential"` // 'github_token', 'ssh_key'
	CredentialID   uuid.UUID `gorm:"type:char(36);not null;uniqueIndex:uniq_project_credential"`           // 关联到github_tokens或ssh_keys表的ID
	IsDefault      bool      `json:"is_default" gorm:"default:false"`         // 是否为该项目的默认凭证
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (c *ProjectCredential) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return
}

func (ProjectCredential) TableName() string {
	return "project_credentials"
}

// 加密密钥 - 在生产环境中应该从环境变量或配置文件中读取
const encryptionKey = "web-deploy-github-token-key-32b!" // 32字节密钥

// encryptToken 加密GitHub令牌
func encryptToken(plaintext string) (string, error) {
	key := []byte(encryptionKey)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptToken 解密GitHub令牌
func decryptToken(ciphertext string) (string, error) {
	key := []byte(encryptionKey)
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(data) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// CreateGitHubToken creates a new GitHub token record with encrypted storage
func CreateGitHubToken(userID uuid.UUID, name, token, permissions string, expiresAt *time.Time) (*GitHubToken, error) {
	if name == "" || token == "" {
		return nil, fmt.Errorf("name and token are required")
	}

	// 加密令牌
	encryptedToken, err := encryptToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt token: %w", err)
	}

	githubToken := &GitHubToken{
		UserID:      userID,
		Name:        name,
		TokenHash:   encryptedToken,
		Permissions: permissions,
		ExpiresAt:   expiresAt,
		IsActive:    true,
	}

	if err := dborm.Db.Create(githubToken).Error; err != nil {
		return nil, err
	}

	return githubToken, nil
}

// GetGitHubTokensByUserID retrieves all GitHub tokens for a user (without decrypted tokens)
func GetGitHubTokensByUserID(userID uuid.UUID) ([]GitHubToken, error) {
	var tokens []GitHubToken
	err := dborm.Db.Where("user_id = ? AND is_active = ?", userID, true).Order("created_at desc").Find(&tokens).Error
	return tokens, err
}

// GetGitHubTokenByID retrieves a GitHub token by ID with decryption capability
func GetGitHubTokenByID(tokenID uuid.UUID, userID uuid.UUID, decrypt bool) (*GitHubToken, string, error) {
	var token GitHubToken
	err := dborm.Db.Where("id = ? AND user_id = ? AND is_active = ?", tokenID, userID, true).First(&token).Error
	if err != nil {
		return nil, "", err
	}

	var decryptedToken string
	if decrypt {
		decryptedToken, err = decryptToken(token.TokenHash)
		if err != nil {
			return &token, "", fmt.Errorf("failed to decrypt token: %w", err)
		}
	}

	return &token, decryptedToken, nil
}

// UpdateGitHubToken updates a GitHub token (excluding the token itself)
func UpdateGitHubToken(tokenID uuid.UUID, userID uuid.UUID, name, permissions string, expiresAt *time.Time) error {
	updates := map[string]interface{}{
		"name":        name,
		"permissions": permissions,
		"expires_at":  expiresAt,
		"updated_at":  time.Now(),
	}

	return dborm.Db.Model(&GitHubToken{}).Where("id = ? AND user_id = ?", tokenID, userID).Updates(updates).Error
}

// DeleteGitHubToken soft deletes a GitHub token
func DeleteGitHubToken(tokenID uuid.UUID, userID uuid.UUID) error {
	return dborm.Db.Model(&GitHubToken{}).Where("id = ? AND user_id = ?", tokenID, userID).Update("is_active", false).Error
}

// UpdateTokenLastUsed updates the last used timestamp for a token
func UpdateTokenLastUsed(tokenID uuid.UUID) error {
	now := time.Now()
	return dborm.Db.Model(&GitHubToken{}).Where("id = ?", tokenID).Update("last_used_at", &now).Error
}

// CreateProjectCredential creates a new project credential association
func CreateProjectCredential(projectID uuid.UUID, credentialType string, credentialID uuid.UUID, isDefault bool) (*ProjectCredential, error) {
	// 如果设置为默认凭证，先清除该项目的其他默认凭证
	if isDefault {
		err := dborm.Db.Model(&ProjectCredential{}).Where("project_id = ? AND credential_type = ?", projectID, credentialType).Update("is_default", false).Error
		if err != nil {
			return nil, err
		}
	}

	credential := &ProjectCredential{
		ProjectID:      projectID,
		CredentialType: credentialType,
		CredentialID:   credentialID,
		IsDefault:      isDefault,
	}

	if err := dborm.Db.Create(credential).Error; err != nil {
		return nil, err
	}

	return credential, nil
}

// GetProjectCredentials retrieves all credentials for a project
func GetProjectCredentials(projectID uuid.UUID) ([]ProjectCredential, error) {
	var credentials []ProjectCredential
	err := dborm.Db.Where("project_id = ?", projectID).Order("is_default desc, created_at desc").Find(&credentials).Error
	return credentials, err
}

// GetDefaultCredentialForProject retrieves the default credential for a project
func GetDefaultCredentialForProject(projectID uuid.UUID, credentialType string) (*ProjectCredential, error) {
	var credential ProjectCredential
	err := dborm.Db.Where("project_id = ? AND credential_type = ? AND is_default = ?", projectID, credentialType, true).First(&credential).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

// DeleteProjectCredential removes a project credential association
func DeleteProjectCredential(projectID uuid.UUID, credentialID uuid.UUID, credentialType string) error {
	return dborm.Db.Where("project_id = ? AND credential_id = ? AND credential_type = ?", projectID, credentialID, credentialType).Delete(&ProjectCredential{}).Error
}

// GetGitHubTokenForProject retrieves the GitHub token associated with a project
func GetGitHubTokenForProject(projectID uuid.UUID) (*GitHubToken, string, error) {
	// 首先尝试获取默认的GitHub令牌凭证
	credential, err := GetDefaultCredentialForProject(projectID, "github_token")
	if err != nil {
		return nil, "", fmt.Errorf("no GitHub token configured for project %d", projectID)
	}

	// 获取项目信息以验证用户权限
	_, err = GetProjectByID(projectID)
	if err != nil {
		return nil, "", err
	}

	// 这里假设所有项目都属于用户ID为1的用户，实际应用中需要根据项目的所有者来确定
	// TODO: 扩展Project模型添加UserID字段
	var userID uuid.UUID // 临时硬编码，实际应该从项目或会话中获取

	// 获取并解密令牌
	token, decryptedToken, err := GetGitHubTokenByID(credential.CredentialID, userID, true)
	if err != nil {
		return nil, "", fmt.Errorf("failed to retrieve GitHub token: %w", err)
	}

	// 更新最后使用时间
	if err := UpdateTokenLastUsed(token.ID); err != nil {
		// 记录警告但不影响主要功能
		fmt.Printf("Warning: failed to update token last used time: %v\n", err)
	}

	return token, decryptedToken, nil
}