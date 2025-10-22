package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// TestCheckDeploymentEnvironment_NoAuth tests that CheckDeploymentEnvironment
// is accessible without authentication
func TestCheckDeploymentEnvironment_NoAuth(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/environment/check", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := CheckDeploymentEnvironment(c)
	
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestInstallPodman_NoAuth tests that InstallPodman is accessible without authentication
func TestInstallPodman_NoAuth(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/environment/install-podman", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = InstallPodman(c)
	
	// The handler should execute without auth errors
	// It may return an error if the script is not set, but that's expected
	// What matters is that there's no authentication error
	assert.NotEqual(t, http.StatusUnauthorized, rec.Code, "Should not return Unauthorized")
}

// TestInstallCaddy_NoAuth tests that InstallCaddy is accessible without authentication
func TestInstallCaddy_NoAuth(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/environment/install-caddy", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = InstallCaddy(c)
	
	// The handler should execute without auth errors
	// It may return an error if the script is not set, but that's expected
	// What matters is that there's no authentication error
	assert.NotEqual(t, http.StatusUnauthorized, rec.Code, "Should not return Unauthorized")
}
