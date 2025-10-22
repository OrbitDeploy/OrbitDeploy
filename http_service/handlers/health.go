package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HealthCheck responds with the health status of the application
func HealthCheck(c echo.Context) error {
	{
		log.Println("Health check requested")

		response := map[string]interface{}{
			"status":  "ok",
			"message": "Application is running",
		}

		return c.JSON(http.StatusOK, response)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}
}
