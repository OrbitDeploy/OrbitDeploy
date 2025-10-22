package services

import (
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims represents the claims structure for JWT tokens
type JWTClaims struct {
	UserID   uuid.UUID `json:"sub"`
	Username string    `json:"username"`
	TokenType string    `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// JWTService interface defines JWT operations
type JWTService interface {
	GenerateTokens(userID uuid.UUID, username string) (TokenPair, error)
	GenerateAccessToken(userID uuid.UUID, username string) (string, error)
	Generate2FAToken(userID uuid.UUID, username string) (string, error)
	VerifyAccessToken(tokenString string) (*JWTClaims, error)
	VerifyRefreshToken(tokenString string) (*JWTClaims, error)
	Verify2FAToken(tokenString string) (*JWTClaims, error)
	HashRefreshToken(token string) string
}

// jwtService implements JWTService
type jwtService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
	issuer        string
	audience      string
}

// NewJWTService creates a new JWT service instance
func NewJWTService() JWTService {
	return &jwtService{
		accessSecret:  []byte(getEnvOrDefault("JWT_ACCESS_SECRET", "access-secret-key-change-in-production")),
		refreshSecret: []byte(getEnvOrDefault("JWT_REFRESH_SECRET", "refresh-secret-key-change-in-production")),
		accessTTL:     parseDurationOrDefault(getEnvOrDefault("JWT_ACCESS_TTL", "15m"), 15*time.Minute),
		refreshTTL:    parseDurationOrDefault(getEnvOrDefault("JWT_REFRESH_TTL", "720h"), 720*time.Hour), // 30 days
		issuer:        getEnvOrDefault("JWT_ISSUER", "go-webui"),
		audience:      getEnvOrDefault("JWT_AUDIENCE", "go-webui-users"),
	}
}

// GenerateTokens generates both access and refresh tokens
func (j *jwtService) GenerateTokens(userID uuid.UUID, username string) (TokenPair, error) {
	accessToken, err := j.GenerateAccessToken(userID, username)
	if err != nil {
		return TokenPair{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := j.generateRefreshToken(userID, username)
	if err != nil {
		return TokenPair{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GenerateAccessToken generates a new access token
func (j *jwtService) GenerateAccessToken(userID uuid.UUID, username string) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID:    userID,
		Username:  username,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Audience:  []string{j.audience},
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.accessSecret)
}

// generateRefreshToken generates a new refresh token
func (j *jwtService) generateRefreshToken(userID uuid.UUID, username string) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID:    userID,
		Username:  username,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Audience:  []string{j.audience},
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.refreshSecret)
}

// VerifyAccessToken verifies and parses an access token
func (j *jwtService) VerifyAccessToken(tokenString string) (*JWTClaims, error) {
	return j.verifyToken(tokenString, j.accessSecret, "access")
}

// VerifyRefreshToken verifies and parses a refresh token
func (j *jwtService) VerifyRefreshToken(tokenString string) (*JWTClaims, error) {
	return j.verifyToken(tokenString, j.refreshSecret, "refresh")
}

// Generate2FAToken generates a short-lived token for the 2FA verification step.
func (j *jwtService) Generate2FAToken(userID uuid.UUID, username string) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID:    userID,
		Username:  username,
		TokenType: "2fa_pending",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Audience:  []string{j.audience},
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)), // Short TTL for 2FA
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.accessSecret) // Use access secret for this temp token
}

// Verify2FAToken verifies and parses a 2FA pending token.
func (j *jwtService) Verify2FAToken(tokenString string) (*JWTClaims, error) {
	return j.verifyToken(tokenString, j.accessSecret, "2fa_pending")
}

// verifyToken verifies and parses a token with the given secret
func (j *jwtService) verifyToken(tokenString string, secret []byte, expectedType string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		if claims.TokenType != expectedType {
			return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.TokenType)
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// HashRefreshToken creates a SHA-256 hash of the refresh token for database storage
func (j *jwtService) HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDurationOrDefault(s string, defaultDuration time.Duration) time.Duration {
	if duration, err := time.ParseDuration(s); err == nil {
		return duration
	}
	return defaultDuration
}

// Global JWT service instance
var jwtSvc JWTService

// GetJWTService returns the global JWT service instance
func GetJWTService() JWTService {
	if jwtSvc == nil {
		jwtSvc = NewJWTService()
	}
	return jwtSvc
}