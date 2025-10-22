package handlers

import (
	"net/http"

	"github.com/OrbitDeploy/OrbitDeploy/models"
	"github.com/labstack/echo/v4"
)

// Environment Variable Handlers

// CreateEnvironmentVariableHandler creates a new environment variable
func CreateEnvironmentVariableHandler(c echo.Context) error {
	appIDStr := c.Param("appId")
	appID, err := DecodeFriendlyID(PrefixApplication, appIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid application ID format")
	}

	var req CreateEnvironmentVariableRequest
	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid request body")
	}

	// Validate required fields
	if req.Key == "" {
		return SendError(c, http.StatusBadRequest, "Key is required")
	}

	envVar, err := models.CreateEnvironmentVariable(appID, req.Key, req.Value, req.IsEncrypted)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to create environment variable")
	}

	return SendCreated(c, toEnvironmentVariableResponse(envVar))
}

// ListEnvironmentVariablesHandler lists all environment variables for an application
func ListEnvironmentVariablesHandler(c echo.Context) error {
	appIDStr := c.Param("appId")
	appID, err := DecodeFriendlyID(PrefixApplication, appIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid application ID format")
	}

	envVars, err := models.ListEnvironmentVariablesByApplicationID(appID)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to list environment variables")
	}

	responses := make([]EnvironmentVariableResponse, len(envVars))
	for i, envVar := range envVars {
		responses[i] = *toEnvironmentVariableResponse(envVar)
	}

	return SendSuccess(c, responses)
}

// UpdateEnvironmentVariableHandler updates an environment variable
func UpdateEnvironmentVariableHandler(c echo.Context) error {
	envVarIDStr := c.Param("envVarId")
	envVarID, err := DecodeFriendlyID(PrefixEnvVar, envVarIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid environment variable ID format")
	}

	var req UpdateEnvironmentVariableRequest
	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid request body")
	}

	// Validate required fields
	if req.Key == "" {
		return SendError(c, http.StatusBadRequest, "Key is required")
	}

	envVar, err := models.UpdateEnvironmentVariable(envVarID, req.Key, req.Value, req.IsEncrypted)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to update environment variable")
	}

	return SendSuccess(c, toEnvironmentVariableResponse(envVar))
}

// DeleteEnvironmentVariableHandler deletes an environment variable
func DeleteEnvironmentVariableHandler(c echo.Context) error {
	envVarIDStr := c.Param("envVarId")
	envVarID, err := DecodeFriendlyID(PrefixEnvVar, envVarIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid environment variable ID format")
	}

	err = models.DeleteEnvironmentVariable(envVarID)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to delete environment variable")
	}

	return SendSuccess(c, map[string]string{"message": "Environment variable deleted"})
}

// Helper functions

func toEnvironmentVariableResponse(envVar *models.EnvironmentVariable) *EnvironmentVariableResponse {
	// Get decrypted value for API response
	value, err := envVar.GetDecryptedValue()
	if err != nil {
		// If decryption fails, return empty value but log the error
		value = ""
	}

	return &EnvironmentVariableResponse{
		Uid:            EncodeFriendlyID(PrefixEnvVar, envVar.ID),
		ApplicationUid: EncodeFriendlyID(PrefixApplication, envVar.ApplicationID),
		Key:            envVar.Key,
		Value:          value,
		IsEncrypted:    envVar.IsEncrypted,
		CreatedAt:      envVar.CreatedAt,
		UpdatedAt:      envVar.UpdatedAt,
	}
}
