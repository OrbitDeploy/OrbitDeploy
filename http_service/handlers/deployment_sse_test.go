package handlers

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/OrbitDeploy/OrbitDeploy/models"
	"github.com/labstack/echo/v4"
	"github.com/opentdp/go-helper/dborm"
	"github.com/stretchr/testify/assert"
)

// TestDeploymentLogsSSEEnhanced_ValidDeploymentId tests the SSE handler with valid deployment ID
func TestDeploymentLogsSSEEnhanced_ValidDeploymentId(t *testing.T) {
	// Setup test database
	tmpDir := t.TempDir()
	config := &dborm.Config{
		Type:   "sqlite",
		DbName: tmpDir + "/test.db",
	}
	if dborm.Connect(config) == nil {
		t.Fatal("failed to connect to test database")
	}
	defer dborm.Destroy()

	// Auto-migrate models
	err := dborm.Db.AutoMigrate(&models.Deployment{}, &models.Application{}, &models.Release{})
	assert.NoError(t, err)

	// Create test deployment in database
	deployment := &models.Deployment{
		ApplicationID: 1,
		ReleaseID:     1,
		Status:        "running",
		LogText:       "Test log line 1\nTest log line 2",
		ServiceName:   "test-service",
	}
	err = dborm.Db.Create(deployment).Error
	assert.NoError(t, err)

	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Set deployment ID parameter
	c.SetParamNames("deploymentId")
	c.SetParamValues(strconv.Itoa(int(deployment.ID)))

	// Call handler
	err = DeploymentLogsSSEEnhanced(c)
	
	// The handler should not return an error for valid deployment ID
	assert.NoError(t, err)
}

// TestDeploymentLogsSSEEnhanced_InvalidDeploymentId tests the SSE handler with invalid deployment ID
func TestDeploymentLogsSSEEnhanced_InvalidDeploymentId(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Set invalid deployment ID parameter
	c.SetParamNames("deploymentId")
	c.SetParamValues("invalid")

	// Call handler
	err := DeploymentLogsSSEEnhanced(c)
	
	// Should return error for invalid deployment ID format
	assert.Error(t, err)
	if httpErr, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		assert.Contains(t, httpErr.Message, "Invalid deploymentId format")
	}
}

// TestDeploymentLogsSSEEnhanced_MissingDeploymentId tests the SSE handler with missing deployment ID
func TestDeploymentLogsSSEEnhanced_MissingDeploymentId(t *testing.T) {
	// Setup Echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// No deployment ID parameter set

	// Call handler
	err := DeploymentLogsSSEEnhanced(c)
	
	// Should return error for missing deployment ID
	assert.Error(t, err)
	if httpErr, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		assert.Contains(t, httpErr.Message, "deploymentId is required")
	}
}