package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/labstack/echo/v4"
	"github.com/youfun/OrbitDeploy/models"
	"github.com/youfun/OrbitDeploy/services"
)

// In-memory storage for device context sessions (in production, use Redis or database)
var deviceContextSessions = make(map[string]*DeviceContextSession)
var deviceContextSessionsMutex sync.RWMutex

// InitiateDeviceContextAuth creates a new device authorization session with context information
// Endpoint: POST /api/cli/device-auth/sessions
func InitiateDeviceContextAuth(c echo.Context) error {
	var req DeviceContextRequest
	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid JSON format")
	}

	// Generate session ID
	sessionBytes := make([]byte, 16)
	if _, err := rand.Read(sessionBytes); err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to generate session ID")
	}
	sessionID := base64.URLEncoding.EncodeToString(sessionBytes)
	clientIP := c.RealIP()

	// Create session
	session := &DeviceContextSession{
		SessionID:   sessionID,
		OS:          req.OS,
		DeviceName:  req.DeviceName,
		PublicIP:    clientIP,
		RequestTime: time.Now().Unix(),
		ExpiresAt:   time.Now().Add(5 * time.Minute), // 5 minutes expiry
		CreatedAt:   time.Now(),
	}

	// Store session
	deviceContextSessionsMutex.Lock()
	deviceContextSessions[sessionID] = session
	deviceContextSessionsMutex.Unlock()

	// Clean up expired sessions
	go cleanupExpiredDeviceContextSessions()

	// Build authorization URI
	scheme := "http"
	if c.Request().TLS != nil {
		scheme = "https"
	}
	host := c.Request().Host
	authURI := fmt.Sprintf("%s://%s/cli-device-auth?session_id=%s", scheme, host, sessionID)

	fmt.Println("Device Context Auth URI:", authURI) // For debugging

	// Response format matching new API design
	response := map[string]interface{}{
		"session_id": sessionID,
		"auth_url":   authURI,
		"expires_in": 300, // 5 minutes
	}

	return SendSuccess(c, response)
}

// GetDeviceContextSession returns device context information for authorization
// Endpoint: GET /api/cli/device-auth/sessions/{session_id}
func GetDeviceContextSession(c echo.Context) error {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		return SendError(c, http.StatusBadRequest, "session_id is required")
	}

	deviceContextSessionsMutex.RLock()
	session, exists := deviceContextSessions[sessionID]
	deviceContextSessionsMutex.RUnlock()

	if !exists {
		return SendError(c, http.StatusNotFound, "Session not found or expired")
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		return SendError(c, http.StatusBadRequest, "Session expired")
	}

	// Return device context for authorization page
	response := map[string]interface{}{
		"session_id":        session.SessionID,
		"os":                session.OS,
		"device_name":       session.DeviceName,
		"public_ip":         session.PublicIP,
		"request_timestamp": session.RequestTime,
	}

	return SendSuccess(c, response)
}

// ConfirmDeviceContextAuth handles device authorization confirmation
// Endpoint: POST /api/cli/device-auth/confirm
func ConfirmDeviceContextAuth(c echo.Context) error {
	// This handler requires the user to be authenticated
	userID, ok := c.Get("userID").(uuid.UUID)
	if !ok {
		return SendError(c, http.StatusUnauthorized, "User not authenticated")
	}

	var req struct {
		SessionID string `json:"session_id"`
		Approved  bool   `json:"approved"`
	}

	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid JSON format")
	}

	if req.SessionID == "" {
		return SendError(c, http.StatusBadRequest, "session_id is required")
	}

	deviceContextSessionsMutex.Lock()
	session, exists := deviceContextSessions[req.SessionID]
	if !exists {
		deviceContextSessionsMutex.Unlock()
		return SendError(c, http.StatusBadRequest, "Invalid or expired session")
	}

	if time.Now().After(session.ExpiresAt) {
		deviceContextSessionsMutex.Unlock()
		return SendError(c, http.StatusBadRequest, "Session expired")
	}

	if req.Approved {
		// Authorize the session
		session.IsAuthorized = true
		session.UserID = &userID
	} else {
		// Delete the session if denied
		delete(deviceContextSessions, req.SessionID)
		deviceContextSessionsMutex.Unlock()
		return SendSuccess(c, map[string]interface{}{
			"message": "Device authorization denied",
		})
	}

	deviceContextSessions[req.SessionID] = session
	deviceContextSessionsMutex.Unlock()

	return SendSuccess(c, map[string]interface{}{
		"message": "Device authorization confirmed",
	})
}

// PollDeviceContextToken handles CLI device context token polling
// Endpoint: GET /api/cli/device-auth/token/{session_id}
func PollDeviceContextToken(c echo.Context) error {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		return SendError(c, http.StatusBadRequest, "session_id is required")
	}

	deviceContextSessionsMutex.RLock()
	session, exists := deviceContextSessions[sessionID]
	deviceContextSessionsMutex.RUnlock()

	if !exists {
		// Session expired or not found
		response := map[string]interface{}{
			"status": "EXPIRED",
		}
		return SendSuccess(c, response)
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		response := map[string]interface{}{
			"status": "EXPIRED",
		}
		return SendSuccess(c, response)
	}

	// Check if the device has been authorized
	if !session.IsAuthorized {
		// User hasn't authorized yet
		response := map[string]interface{}{
			"status": "PENDING",
		}
		return SendSuccess(c, response)
	}

	// Get user information
	user, err := models.GetUserByID(*session.UserID)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to get user information")
	}

	// Generate JWT tokens
	jwtService := services.GetJWTService()
	tokens, err := jwtService.GenerateTokens(user.ID, user.Username)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to generate tokens")
	}

	// Store refresh token hash in database
	refreshTokenHash := jwtService.HashRefreshToken(tokens.RefreshToken)
	_, err = models.CreateAuthToken(refreshTokenHash, "CLI Client (Device Context)", time.Now().Add(720*time.Hour)) // 30 days
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to store refresh token")
	}

	// Delete the used session
	deviceContextSessionsMutex.Lock()
	delete(deviceContextSessions, sessionID)
	deviceContextSessionsMutex.Unlock()

	// User successfully authorized - response format matching API documentation
	response := map[string]interface{}{
		"status": "SUCCESS",
		"token":  tokens.AccessToken,
		"user":   user.Username, // Assuming username is email
	}

	return SendSuccess(c, response)
}

// getLocationFromIP tries to get location information from IP address
// This is a simple implementation - in production you might want to use a proper geolocation service

// cleanupExpiredDeviceContextSessions removes expired device context sessions
func cleanupExpiredDeviceContextSessions() {
	deviceContextSessionsMutex.Lock()
	defer deviceContextSessionsMutex.Unlock()

	now := time.Now()
	for id, session := range deviceContextSessions {
		if now.After(session.ExpiresAt) {
			delete(deviceContextSessions, id)
		}
	}
}

// cleanupExpiredConfigSessions removes expired configuration sessions
func cleanupExpiredConfigSessions() {
	configSessionsMutex.Lock()
	defer configSessionsMutex.Unlock()

	now := time.Now()
	for id, session := range configSessions {
		if now.After(session.ExpiresAt) || (session.IsSubmitted && now.After(session.CreatedAt.Add(time.Hour))) {
			delete(configSessions, id)
		}
	}
}
