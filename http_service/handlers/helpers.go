package handlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/youfun/OrbitDeploy/services"
)

var dockerBuildQueueService *services.DockerBuildQueueService

// SetDockerBuildQueueService sets the global docker build queue service
func SetDockerBuildQueueService(svc *services.DockerBuildQueueService) {
	dockerBuildQueueService = svc
}

// SendSuccess sends a success response with default 200 status
func SendSuccess(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// SendError sends an error response
func SendError(c echo.Context, statusCode int, message string) error {
	return c.JSON(statusCode, map[string]interface{}{
		"success": false,
		"message": message,
	})
}

// SendCreated sends a created response with 201 status
func SendCreated(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func GetUserIDFromContext(c echo.Context) (uuid.UUID, error) {
	var userIDStr string
	if v := c.Get("userID"); v != nil {
		if s, ok := v.(string); ok {
			userIDStr = s
		}
	}
	if userIDStr == "" {
		if v := c.Get("user_id"); v != nil {
			if s, ok := v.(string); ok {
				userIDStr = s
			}
		}
	}
	if userIDStr == "" {
		return uuid.Nil, errors.New("unauthorized")
	}

	userID, err := DecodeFriendlyID(PrefixUser, userIDStr)
	if err != nil {
		return uuid.Nil, errors.New("unauthorized")
	}
	return userID, nil
}
