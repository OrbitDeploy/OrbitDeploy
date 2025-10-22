package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/youfun/OrbitDeploy/models"
	"github.com/youfun/OrbitDeploy/services"
)

// GetSystemSettingHandler handles retrieving a system setting.
func GetSystemSettingHandler(c echo.Context) error {
	key := c.Param("key")
	value, err := models.GetSystemSetting(key)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to retrieve system setting")
	}
	return SendSuccess(c, map[string]string{"key": key, "value": value})
}

// UpdateSystemSettingHandler handles updating a system setting.
func UpdateSystemSettingHandler(c echo.Context) error {
	key := c.Param("key")
	var payload struct {
		Value string `json:"value"`
	}

	if err := c.Bind(&payload); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid request payload")
	}

	if key == "domain" {
		domainService := services.NewSystemDomainService()
		if err := domainService.UpdateSystemDomain(payload.Value); err != nil {
			return SendError(c, http.StatusInternalServerError, "Failed to update system domain")
		}
	}

	if err := models.SetSystemSetting(key, payload.Value); err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to update system setting")
	}

	return SendSuccess(c, map[string]string{"key": key, "value": payload.Value})
}
