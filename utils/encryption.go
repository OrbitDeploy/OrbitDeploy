package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
)

// getEncryptionKey derives a key from environment variable or uses a default for development
func getEncryptionKey() []byte {
	// Try to get key from environment
	if keyStr := os.Getenv("ORBIT_ENCRYPTION_KEY"); keyStr != "" {
		hash := sha256.Sum256([]byte(keyStr))
		return hash[:]
	}
	
	// Default development key (should be replaced in production)
	defaultKey := "orbit-deploy-default-encryption-key-change-in-production"
	hash := sha256.Sum256([]byte(defaultKey))
	return hash[:]
}

// EncryptValue encrypts a string value using AES-256-GCM
func EncryptValue(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	key := getEncryptionKey()
	
	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	
	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptValue decrypts a string value using AES-256-GCM
func DecryptValue(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	key := getEncryptionKey()
	
	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Check minimum size
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext_bytes := data[:nonceSize], data[nonceSize:]
	
	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext_bytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}