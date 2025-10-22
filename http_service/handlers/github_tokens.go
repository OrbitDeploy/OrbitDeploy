package handlers

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/opentdp/go-helper/logman"
	"github.com/youfun/OrbitDeploy/models"
)

// GitHubTokenRequest represents the request body for creating/updating GitHub tokens
type GitHubTokenRequest struct {
	Name        string     `json:"name"`
	Token       string     `json:"token"`
	Permissions string     `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// GitHubTokenResponse represents the response format for GitHub token operations
type GitHubTokenResponse struct {
	Uid         string     `json:"uid"`
	Name        string     `json:"name"`
	Permissions string     `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TokenValidationResponse represents the response for token validation
type TokenValidationResponse struct {
	Valid       bool     `json:"valid"`
	Permissions []string `json:"permissions"`
	Username    string   `json:"username"`
	RateLimit   struct {
		Remaining int `json:"remaining"`
		Total     int `json:"total"`
		ResetAt   int `json:"reset_at"`
	} `json:"rate_limit"`
}

// CreateGitHubToken handles POST /api/github-tokens
func CreateGitHubToken(c echo.Context) error {
	userID, err := getUserUUIDFromContext(c)
	if err != nil {
		return err
	}

	var req GitHubTokenRequest
	if err := c.Bind(&req); err != nil {
		// 打印原始请求体和错误信息
		body := ""
		if bodyBytes, readErr := io.ReadAll(c.Request().Body); readErr == nil {
			body = string(bodyBytes)
		}
		logman.Error("Bind GitHubTokenRequest failed", "error", err, "body", body)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON payload")
	}

	// Validate required fields
	if strings.TrimSpace(req.Name) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Token name is required")
	}
	if strings.TrimSpace(req.Token) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Token is required")
	}

	// Validate GitHub token format (basic check)
	if !strings.HasPrefix(req.Token, "ghp_") && !strings.HasPrefix(req.Token, "github_pat_") {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid GitHub token format")
	}

	// Test token validity before storing
	isValid, err := validateGitHubTokenAPI(req.Token)
	if err != nil {
		logman.Error("Failed to validate GitHub token", "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to validate token with GitHub API")
	}
	if !isValid {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid GitHub token")
	}

	// Create token record
	token, err := models.CreateGitHubToken(userID, req.Name, req.Token, req.Permissions, req.ExpiresAt)
	if err != nil {
		logman.Error("Failed to create GitHub token", "user_id", userID, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create token")
	}

	response := convertToTokenResponse(token)
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    response,
		"message": "GitHub token created successfully",
	})
}

// ListGitHubTokens handles GET /api/github-tokens
func ListGitHubTokens(c echo.Context) error {
	userID, err := getUserUUIDFromContext(c)
	if err != nil {
		return err
	}

	tokens, err := models.GetGitHubTokensByUserID(userID)
	if err != nil {
		logman.Error("Failed to list GitHub tokens", "user_id", userID, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list tokens")
	}

	responses := make([]GitHubTokenResponse, len(tokens))
	for i, token := range tokens {
		responses[i] = convertToTokenResponse(&token)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    responses,
	})
}

// UpdateGitHubToken handles PUT /api/github-tokens/:id
func UpdateGitHubToken(c echo.Context) error {
	userID, err := getUserUUIDFromContext(c)
	if err != nil {
		return err
	}

	tokenIDStr := c.Param("id")
	tokenID, err := DecodeFriendlyID(PrefixGitHubToken, tokenIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid token ID format")
	}

	var req GitHubTokenRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON payload")
	}

	// Validate required fields
	if strings.TrimSpace(req.Name) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Token name is required")
	}

	// Update token (excluding the token itself)
	err = models.UpdateGitHubToken(tokenID, userID, req.Name, req.Permissions, req.ExpiresAt)
	if err != nil {
		logman.Error("Failed to update GitHub token", "token_id", tokenID, "user_id", userID, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update token")
	}

	// Get updated token
	token, _, err := models.GetGitHubTokenByID(tokenID, userID, false)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Token not found")
	}

	response := convertToTokenResponse(token)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    response,
		"message": "GitHub token updated successfully",
	})
}

// DeleteGitHubToken handles DELETE /api/github-tokens/:id
func DeleteGitHubToken(c echo.Context) error {
	userID, err := getUserUUIDFromContext(c)
	if err != nil {
		return err
	}

	tokenIDStr := c.Param("id")
	tokenID, err := DecodeFriendlyID(PrefixGitHubToken, tokenIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid token ID format")
	}

	// Check if token exists and belongs to user
	_, _, err = models.GetGitHubTokenByID(tokenID, userID, false)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Token not found")
	}

	// Delete token
	err = models.DeleteGitHubToken(tokenID, userID)
	if err != nil {
		logman.Error("Failed to delete GitHub token", "token_id", tokenID, "user_id", userID, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete token")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "GitHub token deleted successfully",
	})
}

// TestGitHubToken handles POST /api/github-tokens/:id/test
func TestGitHubToken(c echo.Context) error {
	userID, err := getUserUUIDFromContext(c)
	if err != nil {
		return err
	}

	tokenIDStr := c.Param("id")
	tokenID, err := DecodeFriendlyID(PrefixGitHubToken, tokenIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid token ID format")
	}

	// Get token with decryption
	_, decryptedToken, err := models.GetGitHubTokenByID(tokenID, userID, true)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Token not found")
	}

	// Test token with GitHub API
	isValid, err := validateGitHubTokenAPI(decryptedToken)
	if err != nil {
		logman.Error("Failed to test GitHub token", "token_id", tokenID, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to test token")
	}

	var responseData TokenValidationResponse
	responseData.Valid = isValid

	if isValid {
		// Get additional token info if valid
		tokenInfo, err := getGitHubTokenInfo(decryptedToken)
		if err == nil {
			responseData.Username = tokenInfo.Username
			responseData.Permissions = tokenInfo.Permissions
			responseData.RateLimit = tokenInfo.RateLimit
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    responseData,
	})
}

// Helper function to get user ID from context
func getUserUUIDFromContext(c echo.Context) (uuid.UUID, error) {
	// Prefer JWT middleware key "userID" (friendly id)
	var uidStr string
	if v := c.Get("userID"); v != nil {
		if s, ok := v.(string); ok {
			uidStr = s
		}
	}
	if uidStr == "" {
		// Fallback key used elsewhere
		if v := c.Get("user_id"); v != nil {
			if s, ok := v.(string); ok {
				uidStr = s
			}
		}
	}
	if uidStr == "" {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}
	// Decode friendly user id to UUID
	id, err := DecodeFriendlyID(PrefixUser, uidStr)
	if err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID")
	}
	return id, nil
}

// Helper function to convert model to response
func convertToTokenResponse(token *models.GitHubToken) GitHubTokenResponse {
	return GitHubTokenResponse{
		Uid:         EncodeFriendlyID(PrefixGitHubToken, token.ID),
		Name:        token.Name,
		Permissions: token.Permissions,
		ExpiresAt:   token.ExpiresAt,
		LastUsedAt:  token.LastUsedAt,
		IsActive:    token.IsActive,
		CreatedAt:   token.CreatedAt,
		UpdatedAt:   token.UpdatedAt,
	}
}

// validateGitHubTokenAPI validates a GitHub token by calling the GitHub API
func validateGitHubTokenAPI(token string) (bool, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// getGitHubTokenInfo gets additional information about a GitHub token
func getGitHubTokenInfo(token string) (*struct {
	Username    string
	Permissions []string
	RateLimit   struct {
		Remaining int `json:"remaining"`
		Total     int `json:"total"`
		ResetAt   int `json:"reset_at"`
	}
}, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid token or API error")
	}

	// Parse rate limit headers
	rateLimit := struct {
		Remaining int `json:"remaining"`
		Total     int `json:"total"`
		ResetAt   int `json:"reset_at"`
	}{}
	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		if r, err := strconv.Atoi(remaining); err == nil {
			rateLimit.Remaining = r
		}
	}
	if total := resp.Header.Get("X-RateLimit-Limit"); total != "" {
		if t, err := strconv.Atoi(total); err == nil {
			rateLimit.Total = t
		}
	}
	if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
		if r, err := strconv.Atoi(reset); err == nil {
			rateLimit.ResetAt = r
		}
	}

	// For permissions, we'd need to check scopes, but GitHub API doesn't directly expose them in /user
	// This is a simplified version; in production, you might need to use a different endpoint or parse scopes from a test request
	permissions := []string{"repo"} // Placeholder; adjust based on actual token scopes if available

	return &struct {
		Username    string
		Permissions []string
		RateLimit   struct {
			Remaining int `json:"remaining"`
			Total     int `json:"total"`
			ResetAt   int `json:"reset_at"`
		}
	}{
		Username:    "user", // In a real implementation, parse from response body
		Permissions: permissions,
		RateLimit:   rateLimit,
	}, nil
}
