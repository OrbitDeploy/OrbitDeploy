package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/youfun/OrbitDeploy/models"
	"github.com/youfun/OrbitDeploy/services"
)

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// SetupRequest represents the initial setup request payload
type SetupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RefreshTokenRequest represents the refresh token request payload for CLI
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	User    *User  `json:"user,omitempty"`
}

type User struct {
	Uid      string `json:"uid"`
	Username string `json:"username"`
}

//  -compatible authentication handlers with JWT

// Login handles user authentication with JWT tokens
func Login(c echo.Context) error {
	var req LoginRequest

	// Use  's data binding instead of manual JSON parsing
	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid JSON format")
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		return SendError(c, http.StatusBadRequest, "Username and password are required")
	}

	// Get user from database
	user, err := models.GetUserByUsername(req.Username)
	if err != nil {
		return SendError(c, http.StatusUnauthorized, "Invalid username or password")
	}

	// Check password
	if !user.CheckPassword(req.Password) {
		return SendError(c, http.StatusUnauthorized, "Invalid username or password")
	}

	jwtService := services.GetJWTService()

	// If 2FA is enabled, issue a temporary token and require OTP verification.
	if user.TwoFactorEnabled {
		tempToken, err := jwtService.Generate2FAToken(user.ID, user.Username)
		if err != nil {
			return SendError(c, http.StatusInternalServerError, "Failed to generate 2FA token")
		}
		return SendSuccess(c, map[string]interface{}{
			"two_factor_required": true,
			"temp_2fa_token":      tempToken,
		})
	}

	// Generate JWT tokens
	tokens, err := jwtService.GenerateTokens(user.ID, user.Username)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to generate tokens")
	}

	// Store refresh token hash in database
	refreshTokenHash := jwtService.HashRefreshToken(tokens.RefreshToken)
	_, err = models.CreateAuthToken(refreshTokenHash, "Web Browser", time.Now().Add(720*time.Hour)) // 30 days
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to store refresh token")
	}

	// Set refresh token as HttpOnly cookie
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Expires:  time.Now().Add(720 * time.Hour), // 30 days
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)

	// Send success response with access token
	return SendSuccess(c, map[string]interface{}{
		"access_token":        tokens.AccessToken,
		"two_factor_required": false, // Explicitly state that 2FA is not required
	})
}

// Logout handles user logout by invalidating refresh tokens
func Logout(c echo.Context) error {
	// Get refresh token from cookie
	cookie, err := c.Cookie("refresh_token")
	if err == nil {
		// Delete refresh token from database
		jwtService := services.GetJWTService()
		refreshTokenHash := jwtService.HashRefreshToken(cookie.Value)
		models.DeleteAuthToken(refreshTokenHash)
	}

	// Clear refresh token cookie
	clearCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
		Path:     "/",
	}
	c.SetCookie(clearCookie)

	return SendSuccess(c, map[string]interface{}{
		"message": "Logout successful",
	})
}

// RefreshToken handles token refresh for both web and CLI clients
func RefreshToken(c echo.Context) error {
	var refreshToken string

	// Try to get refresh token from cookie (Web) first
	if cookie, err := c.Cookie("refresh_token"); err == nil {
		refreshToken = cookie.Value
	} else {
		// Try to get from custom header (CLI)
		if headerToken := c.Request().Header.Get("X-Refresh-Token"); headerToken != "" {
			refreshToken = headerToken
		} else {
			// Try to get from request body (CLI)
			var req RefreshTokenRequest
			if err := c.Bind(&req); err == nil && req.RefreshToken != "" {
				refreshToken = req.RefreshToken
			}
		}
	}

	if refreshToken == "" {
		return SendError(c, http.StatusUnauthorized, "No refresh token provided")
	}

	// Verify refresh token
	jwtService := services.GetJWTService()
	claims, err := jwtService.VerifyRefreshToken(refreshToken)
	if err != nil {
		return SendError(c, http.StatusUnauthorized, "Invalid refresh token")
	}

	// Check if refresh token exists in database
	refreshTokenHash := jwtService.HashRefreshToken(refreshToken)
	valid, err := models.IsRefreshTokenValid(refreshTokenHash)
	if err != nil || !valid {
		return SendError(c, http.StatusUnauthorized, "Invalid refresh token")
	}

	// Generate new access token
	accessToken, err := jwtService.GenerateAccessToken(claims.UserID, claims.Username)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to generate access token")
	}

	// Send success response with new access token
	return SendSuccess(c, map[string]interface{}{
		"access_token": accessToken,
	})
}

// Setup handles initial admin user creation with
func Setup(c echo.Context) error {
	// Check if setup has already been done
	count, err := models.GetUserCount()
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to check user count")
	}

	if count > 0 {
		return SendError(c, http.StatusBadRequest, "Setup has already been completed")
	}

	var req SetupRequest

	// Use  's data binding
	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid JSON format")
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		return SendError(c, http.StatusBadRequest, "Username and password are required")
	}

	// Create admin user
	user, err := models.CreateUser(req.Username, req.Password)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to create user")
	}

	userData := &User{
		Uid:      EncodeFriendlyID(PrefixUser, user.ID),
		Username: user.Username,
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"message": "Setup completed successfully",
			"user":    userData,
		},
	})
}

// CheckSetup checks if setup has been completed
func CheckSetup(c echo.Context) error {
	count, err := models.GetUserCount()
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to check user count")
	}

	return SendSuccess(c, map[string]interface{}{
		"setup_required": count == 0,
	})
}

// CheckAuthStatus checks if user is authenticated using JWT
func CheckAuthStatus(c echo.Context) error {
	// Get access token from Authorization header or query param
	authHeader := c.Request().Header.Get("Authorization")
	var token string
	if authHeader != "" {
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) == 2 && strings.EqualFold(tokenParts[0], "Bearer") {
			token = tokenParts[1]
		}
	}
	if token == "" {
		token = c.QueryParam("access_token")
	}
	if token == "" {
		return SendError(c, http.StatusUnauthorized, "No access token provided")
	}

	// Verify access token
	jwtService := services.GetJWTService()
	claims, err := jwtService.VerifyAccessToken(token)
	if err != nil {
		return SendError(c, http.StatusUnauthorized, "Invalid access token")
	}

	// Fetch the full user model to get all details, including 2FA status
	user, err := models.GetUserByID(claims.UserID)
	if err != nil {
		return SendError(c, http.StatusNotFound, "User not found")
	}

	userData := &User{
		Uid:      EncodeFriendlyID(PrefixUser, user.ID),
		Username: user.Username,
	}

	return SendSuccess(c, map[string]interface{}{
		"message":            "Authenticated",
		"user":               userData,
		"two_factor_enabled": user.TwoFactorEnabled,
	})
}

// CLI Device Code Authentication handlers

// InitiateDeviceCodeRequest represents the device code initiation request
type InitiateDeviceCodeRequest struct {
	ClientID string `json:"client_id,omitempty"` // Optional for future use
}

// InitiateDeviceCodeResponse represents the device code initiation response
type InitiateDeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete,omitempty"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// TokenRequest represents the device code token request
type TokenRequest struct {
	DeviceCode string `json:"device_code"`
}

// TokenResponse represents the device code token response
type TokenResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	Error        string `json:"error,omitempty"`
}

// InitiateDeviceCode handles CLI device code flow initiation
// Endpoint: POST /api/cli/auth/sessions
// func InitiateDeviceCode(c echo.Context) error {
// 	// Generate device code and user code
// 	deviceCode, err := models.GenerateDeviceCodes()
// 	if err != nil {
// 		return SendError(c, http.StatusInternalServerError, "Failed to generate device codes")
// 	}

// 	// Clean up expired device codes
// 	go models.DeleteExpiredDeviceCodes()

// 	// Build verification URI (get base URL from request)
// 	scheme := "http"
// 	if c.Request().TLS != nil {
// 		scheme = "https"
// 	}
// 	host := c.Request().Host
// 	verificationURI := fmt.Sprintf("%s://%s/cli-authorize", scheme, host)

// 	// Response format matching API documentation
// 	response := map[string]interface{}{
// 		"session_id": deviceCode.DeviceCode,
// 		"auth_url":   verificationURI,
// 		"expires_in": 300, // 5 minutes as specified in documentation
// 	}

// 	return SendSuccess(c, response)
// }

// PollDeviceToken handles CLI device code token polling
// Endpoint: GET /api/cli/auth/sessions/{session_id}
func PollDeviceToken(c echo.Context) error {
	sessionId := c.Param("session_id")
	if sessionId == "" {
		return SendError(c, http.StatusBadRequest, "session_id is required")
	}

	// Get device code from database
	deviceCode, err := models.GetDeviceCodeByDeviceCode(sessionId)
	if err != nil {
		// Session expired or not found
		response := map[string]interface{}{
			"status": "EXPIRED",
		}
		return SendSuccess(c, response)
	}

	// Check if the device code has been authorized
	if !deviceCode.IsAuthorized {
		// User hasn't logged in yet
		response := map[string]interface{}{
			"status": "PENDING",
		}
		return SendSuccess(c, response)
	}

	// Get user information
	user, err := models.GetUserByID(*deviceCode.UserID)
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
	_, err = models.CreateAuthToken(refreshTokenHash, "CLI Client", time.Now().Add(720*time.Hour)) // 30 days
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to store refresh token")
	}

	// Delete the used device code
	go func() {
		models.DeleteExpiredDeviceCodes()
	}()

	// User successfully logged in - response format matching API documentation
	response := map[string]interface{}{
		"status":     "SUCCESS",
		"token":      tokens.AccessToken,
		"user_email": user.Username, // Assuming username is email
	}

	return SendSuccess(c, response)
}
