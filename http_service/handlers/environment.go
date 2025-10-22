package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/youfun/OrbitDeploy/utils"
)

// InstallationProgress represents installation progress
type InstallationProgress struct {
	Stage    string `json:"stage"`
	Progress int    `json:"progress"`
	Message  string `json:"message"`
	Error    string `json:"error,omitempty"`
}

// CheckDeploymentEnvironment checks the deployment environment status
func CheckDeploymentEnvironment(c echo.Context) error {

	status := &utils.EnvironmentStatus{}

	// Check Podman
	utils.CheckPodman(status)

	// Check Caddy
	utils.CheckCaddy(status)

	// Determine overall status
	if status.Podman.Installed && status.Podman.VersionValid && status.Caddy.Installed {
		status.OverallStatus = "ready"
	} else if status.Podman.Installed || status.Caddy.Installed {
		status.OverallStatus = "partial"
	} else {
		status.OverallStatus = "missing"
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    status,
	})

}

// InstallPodman installs Podman using the embedded script
func InstallPodman(c echo.Context) error {
	if c.Request().Method != http.MethodPost {
		return echo.NewHTTPError(http.StatusMethodNotAllowed, "Method not allowed")
	}

	// Get the embedded script from the global variable
	script := utils.GetPodmanInstallScript()
	if script == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "Installation script not found")
	}

	go utils.RunInstallation(script, "podman")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Podman installation started. Check progress via WebSocket.",
	})

}

// InstallCaddy installs Caddy using the embedded script
func InstallCaddy(c echo.Context) error {

	// Get the embedded script from the global variable
	script := utils.GetCaddyInstallScript()
	if script == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "Installation script not found")
	}

	go utils.RunInstallation(script, "caddy")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Caddy installation started. Check progress via WebSocket.",
	})
}

// SetInstallScripts sets the functions to access embedded scripts
func SetInstallScripts(podmanScript, caddyScript func() string) {
	utils.GetPodmanInstallScript = podmanScript
	utils.GetCaddyInstallScript = caddyScript
}
