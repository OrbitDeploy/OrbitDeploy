package handlers

import (
	"log"
	"net/http"

	"github.com/OrbitDeploy/OrbitDeploy/models"
	"github.com/labstack/echo/v4"
)

// ChangePasswordRequest represents the password change request payload
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangePassword handles user password change
func ChangePassword(c echo.Context) error {

	// Get user ID from JWT middleware context as friendly ID string
	var userIDStr string
	if v := c.Get("userID"); v != nil {
		if s, ok := v.(string); ok {
			userIDStr = s
		}
	}
	if userIDStr == "" {
		if v := c.Get("user_id"); v != nil { // fallback key
			if s, ok := v.(string); ok {
				userIDStr = s
			}
		}
	}
	if userIDStr == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	// Decode friendly ID to UUID
	userID, err := DecodeFriendlyID(PrefixUser, userIDStr)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}
	// Bind the request body to the struct
	var req ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Failed to decode password change request: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON body")
	}

	// Validate input
	if req.CurrentPassword == "" || req.NewPassword == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Current password and new password are required")
	}

	if len(req.NewPassword) < 6 {
		return echo.NewHTTPError(http.StatusBadRequest, "New password must be at least 6 characters long")
	}

	// Get user from database
	user, err := models.GetUserByID(userID)
	if err != nil {
		log.Printf("User not found: %v", err)
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	// Check current password
	if !user.CheckPassword(req.CurrentPassword) {
		log.Printf("Invalid current password for user: %s", user.Username)
		return echo.NewHTTPError(http.StatusUnauthorized, "Current password is incorrect")
	}

	// Update password
	if err := user.HashPassword(req.NewPassword); err != nil {
		log.Printf("Failed to hash new password: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update password")
	}

	// Save user with new password
	if err := models.UpdateUserPassword(user); err != nil {
		log.Printf("Failed to update user password: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update password")
	}

	log.Printf("Password changed successfully for user: %s", user.Username)

	// Return a success response
	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"message": "Password changed successfully",
	})
}
