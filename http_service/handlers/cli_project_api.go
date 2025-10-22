package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/opentdp/go-helper/logman"
	"github.com/youfun/OrbitDeploy/models"
	"github.com/youfun/OrbitDeploy/services"
)

// validateApplicationTokenPermission validates if the current token has permission for the application
func validateApplicationTokenPermission(c echo.Context, applicationID uuid.UUID) error {
	authType := c.Get("auth_type")
	if authType == "app_token" {
		// Using application token, check if it matches the application
		if app, ok := c.Get("application").(*models.Application); ok {
			if app.ID != applicationID {
				return echo.NewHTTPError(http.StatusForbidden, "Token does not have permission for this application")
			}
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, "Application context not found")
		}
	}
	// JWT tokens have full access, no additional validation needed
	return nil
}

// UploadProjectImage handles image upload for projects
// Endpoint: POST /api/projects/{project_id}/images
func UploadProjectImage(c echo.Context) error {
	projectIDStr := c.Param("project_id")
	if projectIDStr == "" {
		return SendError(c, http.StatusBadRequest, "project_id is required")
	}

	_, err := DecodeFriendlyID(PrefixProject, projectIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid project_id format")
	}

	// Handle multipart form data - expect image file
	file, err := c.FormFile("image")
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Image file is required")
	}

	// Validate file extension (support .tar and .tar.gz for Docker images)
	if !strings.HasSuffix(file.Filename, ".tar") && !strings.HasSuffix(file.Filename, ".tar.gz") {
		return SendError(c, http.StatusBadRequest, "Only .tar and .tar.gz files are supported")
	}

	// Get optional metadata from form
	description := c.FormValue("description")
	if description == "" {
		description = fmt.Sprintf("CLI upload %s", time.Now().Format("2006-01-02 15:04:05"))
	}

	// Generate unique tag based on timestamp
	tag := fmt.Sprintf("cli-upload-%s", time.Now().Format("20060102150405"))
	tempImageName := fmt.Sprintf("cli-temp:%s", tag)

	// Create temporary directory for upload
	tempDir := "/tmp/cli-image-uploads"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to create temporary directory")
	}

	// Save uploaded file to temporary location
	tempFilePath := filepath.Join(tempDir, fmt.Sprintf("%s_%s", tag, file.Filename))
	src, err := file.Open()
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to open uploaded file")
	}
	defer src.Close()

	dst, err := os.Create(tempFilePath)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to create temporary file")
	}
	defer dst.Close()
	defer os.Remove(tempFilePath) // Clean up temp file

	// Copy uploaded file content
	if _, err := io.Copy(dst, src); err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to save uploaded file")
	}

	// Load the image using podman
	cmd := exec.Command("podman", "load", "-i", tempFilePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return SendError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to load image: %s", string(output)))
	}

	// Parse loaded image name from podman output
	loadedImageName := parseLoadedImageName(string(output))
	if loadedImageName == "" {
		loadedImageName = tempImageName
	}

	// Tag with our temporary name if different
	if loadedImageName != tempImageName {
		tagCmd := exec.Command("podman", "tag", loadedImageName, tempImageName)
		if _, tagErr := tagCmd.CombinedOutput(); tagErr != nil {
			// Use the originally loaded name if tagging fails
			tempImageName = loadedImageName
		}
	}

	response := map[string]interface{}{
		"image_id":    tempImageName,
		"size":        file.Size,
		"tag":         tag,
		"description": description,
		"uploaded_at": time.Now().Format(time.RFC3339),
	}

	return SendSuccess(c, response)
}

// CreateProjectDeployment handles deployment creation for projects
// Endpoint: POST /api/projects/{project_id}/deployments
func CreateProjectDeployment(c echo.Context) error {
	projectIDStr := c.Param("project_id")
	if projectIDStr == "" {
		return SendError(c, http.StatusBadRequest, "project_id is required")
	}

	// Parse project ID
	projectID, err := DecodeFriendlyID(PrefixProject, projectIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid project_id format")
	}

	// Parse request body
	var req struct {
		ImageID string                 `json:"image_id"`
		AppName string                 `json:"app_name"`
		Source  string                 `json:"source"`
		Config  map[string]interface{} `json:"config"`
	}

	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid request body")
	}

	if req.AppName == "" {
		return SendError(c, http.StatusBadRequest, "app_name is required")
	}

	if req.Source == "" {
		req.Source = "cli"
	}

	// Check if application exists, or suggest creation
	app, err := models.GetApplicationByName(req.AppName)
	if err != nil {
		return SendError(c, http.StatusNotFound, fmt.Sprintf("Application '%s' not found. Please create it first using the web interface.", req.AppName))
	}

	// Verify application belongs to the specified project
	if app.ProjectID != projectID {
		return SendError(c, http.StatusBadRequest, "Application does not belong to the specified project")
	}

	// Create services needed for deployment orchestrator
	buildService := services.NewBuildService()
	envService := services.NewDeploymentEnvironmentService()
	podmanService := services.NewPodmanService()
	deploymentOrchestrator := services.NewDeploymentOrchestrator(buildService, envService, podmanService)

	// Create deployment request (for new build, set ReleaseID to nil)
	deployReq := services.CreateDeploymentRequest{
		ReleaseID: nil, // This will trigger a new build from latest
	}

	logman.Info("Creating deployment for project application", "project_id", projectID, "app_name", req.AppName, "app_id", app.ID)

	// Create and start deployment
	deployment, err := deploymentOrchestrator.CreateDeployment(app.ID, deployReq)
	if err != nil {
		logman.Error("Failed to create deployment", "project_id", projectID, "app_name", req.AppName, "error", err)
		return SendError(c, http.StatusInternalServerError, "Failed to create deployment: "+err.Error())
	}

	logman.Info("Deployment created successfully", "project_id", projectID, "app_name", req.AppName, "deployment_id", deployment.ID)

	response := map[string]interface{}{
		"deployment_id": EncodeFriendlyID(PrefixDeployment, deployment.ID),
		"status":        deployment.Status,
		"message":       "Deployment task has been created.",
		"project_id":    projectIDStr,
		"app_name":      req.AppName,
		"service_name":  deployment.ServiceName,
		"created_at":    deployment.CreatedAt.Format(time.RFC3339),
	}

	// Return 202 Accepted as specified in documentation
	c.Response().WriteHeader(http.StatusAccepted)
	return SendSuccess(c, response)
}

// GetDeploymentResult handles getting deployment result
// Endpoint: GET /api/deployments/{deployment_id}
func GetDeploymentResult(c echo.Context) error {
	deploymentIDParam := c.Param("deployment_id")

	// Parse deployment ID using helper function
	deploymentID, err := DecodeFriendlyID(PrefixDeployment, deploymentIDParam)
	if err != nil {
		return SendError(c, http.StatusBadRequest, err.Error())
	}

	// Get deployment with related data
	deployment, err := models.GetDeploymentByID(deploymentID)
	if err != nil {
		logman.Error("Failed to get deployment", "deployment_id", deploymentID, "error", err)
		return SendError(c, http.StatusNotFound, "Deployment not found")
	}

	// Get application info
	app, err := models.GetApplicationByID(deployment.ApplicationID)
	if err != nil {
		logman.Error("Failed to get application", "application_id", deployment.ApplicationID, "error", err)
		return SendError(c, http.StatusNotFound, "Application not found")
	}

	// Validate application token permission if using app token
	if err := validateApplicationTokenPermission(c, app.ID); err != nil {
		return err
	}

	// Get release info
	release, err := models.GetReleaseByID(deployment.ReleaseID)
	if err != nil {
		logman.Error("Failed to get release", "release_id", deployment.ReleaseID, "error", err)
		return SendError(c, http.StatusInternalServerError, "Failed to get release info")
	}

	// Get routing info to determine URLs
	routings, err := models.GetActiveRoutingsByApplicationID(deployment.ApplicationID)
	if err != nil {
		logman.Warn("Failed to get routings", "app_id", deployment.ApplicationID, "error", err)
		routings = []*models.Routing{} // Continue with empty routings
	}

	// Build URLs from routings
	var urls []string
	for _, routing := range routings {
		if routing.IsActive {
			// Default to HTTP since there's no UseHTTPS field in the model
			protocol := "http"
			url := fmt.Sprintf("%s://%s", protocol, routing.DomainName)
			urls = append(urls, url)
		}
	}

	// If no custom domains, provide system port info
	if len(urls) == 0 && deployment.SystemPort != nil {
		urls = append(urls, fmt.Sprintf("http://localhost:%d", *deployment.SystemPort))
	}

	// Map deployment status to CLI expected format
	status := "PENDING"
	switch deployment.Status {
	case "success":
		status = "SUCCESS"
	case "failed":
		status = "FAILED"
	case "in_progress":
		status = "RUNNING"
	default:
		status = "PENDING"
	}

	var errorMessage *string
	if deployment.Status == "failed" {
		// Extract error from log text if available
		lines := strings.Split(deployment.LogText, "\n")
		for _, line := range lines {
			if strings.Contains(line, "失败") || strings.Contains(line, "failed") || strings.Contains(line, "error") {
				errorMessage = &line
				break
			}
		}
		if errorMessage == nil {
			defaultError := "Deployment failed. Check logs for details."
			errorMessage = &defaultError
		}
	}

	response := map[string]interface{}{
		"deployment_id": deploymentIDParam,
		"status":        status,
		"created_at":    deployment.CreatedAt.Format(time.RFC3339),
		"started_at":    deployment.StartedAt.Format(time.RFC3339),
		"app_name":      app.Name,
		"release_id":    EncodeFriendlyID(PrefixRelease, release.ID),
		"version":       release.Version,
		"image_name":    release.ImageName,
		"urls":          urls,
		"error_message": errorMessage,
	}

	if deployment.FinishedAt != nil {
		response["finished_at"] = deployment.FinishedAt.Format(time.RFC3339)
	}

	return SendSuccess(c, response)
}

// SetProjectVariables handles setting/updating project environment variables
// Endpoint: POST /api/projects/{project_id}/variables
func SetProjectVariables(c echo.Context) error {
	projectIDStr := c.Param("project_id")
	if projectIDStr == "" {
		return SendError(c, http.StatusBadRequest, "project_id is required")
	}

	_, err := DecodeFriendlyID(PrefixProject, projectIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid project_id format")
	}

	// Parse request body for environment variables
	var req struct {
		Variables []struct {
			Key         string `json:"key"`
			Value       string `json:"value"`
			IsEncrypted bool   `json:"isEncrypted"`
		} `json:"variables"`
	}

	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid request body")
	}

	// TODO: Implement actual variable setting logic
	// This would need to:
	// 1. Find the application for this project
	// 2. Get or create a configuration
	// 3. Update environment variables using configuration service
	// For now, return a response matching the API documentation

	response := map[string]interface{}{
		"message": "Environment variables updated successfully. Please redeploy to apply changes.",
	}

	return SendSuccess(c, response)
}

// GetProjectVariables handles getting project environment variables
// Endpoint: GET /api/projects/{project_id}/variables
func GetProjectVariables(c echo.Context) error {
	projectID := c.Param("project_id")
	if projectID == "" {
		return SendError(c, http.StatusBadRequest, "project_id is required")
	}

	// TODO: Implement actual variable retrieval logic
	// This would need to:
	// 1. Find the application for this project
	// 2. Get the active configuration
	// 3. List environment variables using configuration service
	// For now, return a mock response matching the API documentation

	variables := []map[string]interface{}{
		{
			"key":    "DATABASE_URL",
			"value":  "postgresql://user:p...ost:port/db",
			"secret": false,
		},
		{
			"key":        "API_KEY",
			"value":      nil,
			"secret":     true,
			"updated_at": "2025-09-01T10:00:00Z",
		},
	}

	response := map[string]interface{}{
		"variables": variables,
	}

	return SendSuccess(c, response)
}

// ProjectConfigSession represents a temporary configuration session
type ProjectConfigSession struct {
	ID          string                 `json:"id"`
	ExpiresAt   time.Time              `json:"expires_at"`
	Config      map[string]interface{} `json:"config,omitempty"`
	IsSubmitted bool                   `json:"is_submitted"`
	CreatedAt   time.Time              `json:"created_at"`
}

// In-memory storage for config sessions (in production, use Redis or database)
var configSessions = make(map[string]*ProjectConfigSession)
var configSessionsMutex sync.RWMutex

// InitiateConfigRequest represents the request to create a config session
type InitiateConfigRequest struct {
	TomlData string `json:"toml_data,omitempty"`
}

// InitiateConfigResponse represents the response with config session info
type InitiateConfigResponse struct {
	SessionID        string `json:"session_id"`
	ConfigurationURI string `json:"configuration_uri"`
	ExpiresIn        int    `json:"expires_in"`
}

// InitiateProjectConfig creates a temporary configuration session
// Endpoint: POST /api/cli/projects/config-sessions
func InitiateProjectConfig(c echo.Context) error {
	var req InitiateConfigRequest
	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid JSON format")
	}

	// Generate session ID
	sessionBytes := make([]byte, 16)
	if _, err := rand.Read(sessionBytes); err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to generate session ID")
	}
	sessionID := base64.URLEncoding.EncodeToString(sessionBytes)

	// Create session
	session := &ProjectConfigSession{
		ID:        sessionID,
		ExpiresAt: time.Now().Add(10 * time.Minute), // 10 minutes expiry as per documentation
		Config:    make(map[string]interface{}),
		CreatedAt: time.Now(),
	}

	// Store session
	configSessionsMutex.Lock()
	configSessions[sessionID] = session
	configSessionsMutex.Unlock()

	// Clean up expired sessions
	go cleanupExpiredConfigSessions()

	// Build configuration URI
	scheme := "http"
	if c.Request().TLS != nil {
		scheme = "https"
	}
	host := c.Request().Host

	configURI := fmt.Sprintf("%s://%s/cli-configure?session=%s&expires=%d", scheme, host, sessionID, time.Now().Add(10*time.Minute).Unix())

	// Add toml parameter if provided
	if req.TomlData != "" {
		configURI += "&toml=" + url.QueryEscape(base64.StdEncoding.EncodeToString([]byte(req.TomlData)))
	}

	// Response format matching API documentation
	response := InitiateConfigResponse{
		SessionID:        sessionID,
		ConfigurationURI: configURI,
		ExpiresIn:        600, // 10 minutes as specified in documentation
	}

	return SendSuccess(c, response)
}

// SubmitConfigRequest represents the project configuration submission
type SubmitConfigRequest struct {
	SessionID string                 `json:"session_id"`
	Config    map[string]interface{} `json:"config"`
}

// SubmitProjectConfig handles project configuration submission
func SubmitProjectConfig(c echo.Context) error {
	var req SubmitConfigRequest
	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid JSON format")
	}

	if req.SessionID == "" {
		return SendError(c, http.StatusBadRequest, "session_id is required")
	}

	// Get session
	configSessionsMutex.RLock()
	session, exists := configSessions[req.SessionID]
	configSessionsMutex.RUnlock()

	if !exists {
		return SendError(c, http.StatusBadRequest, "Invalid or expired session")
	}

	if time.Now().After(session.ExpiresAt) {
		return SendError(c, http.StatusBadRequest, "Session expired")
	}

	if session.IsSubmitted {
		return SendError(c, http.StatusBadRequest, "Configuration already submitted")
	}

	// Validate config
	if req.Config["name"] == nil || req.Config["name"].(string) == "" {
		return SendError(c, http.StatusBadRequest, "Project name is required")
	}

	// Store configuration and mark as submitted
	configSessionsMutex.Lock()
	session.Config = req.Config
	session.IsSubmitted = true
	configSessions[req.SessionID] = session
	configSessionsMutex.Unlock()

	return SendSuccess(c, map[string]interface{}{
		"message": "Configuration submitted successfully",
		"config":  req.Config,
	})
}

// GetProjectConfigStatus checks the status of a configuration session
// Endpoint: GET /api/cli/projects/config-sessions/{session_id}
func GetProjectConfigStatus(c echo.Context) error {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		return SendError(c, http.StatusBadRequest, "session_id is required")
	}

	configSessionsMutex.RLock()
	session, exists := configSessions[sessionID]
	configSessionsMutex.RUnlock()

	if !exists {
		return SendError(c, http.StatusNotFound, "Session not found")
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		return SendError(c, http.StatusBadRequest, "Session expired")
	}

	// Response format matching API documentation
	if session.IsSubmitted {
		// User has successfully configured
		response := map[string]interface{}{
			"status":       "SUCCESS",
			"project_id":   session.Config["project_id"],
			"project_name": session.Config["project_name"],
			"spec":         session.Config,
			"environment":  session.Config["env"],
		}
		return SendSuccess(c, response)
	} else {
		// User hasn't completed configuration yet
		response := map[string]interface{}{
			"status": "PENDING",
		}
		return SendSuccess(c, response)
	}
}

// UploadApplicationImage handles image upload for applications by name
// Endpoint: POST /api/apps/by-name/:appName/releases
func UploadApplicationImage(c echo.Context) error {
	appName := c.Param("appName")
	if appName == "" {
		return SendError(c, http.StatusBadRequest, "appName is required")
	}

	// Get application by name
	app, err := models.GetApplicationByName(appName)
	if err != nil {
		return SendError(c, http.StatusNotFound, "Application not found: "+appName)
	}

	// Validate application token permission if using app token
	if err := validateApplicationTokenPermission(c, app.ID); err != nil {
		return err
	}

	// Handle multipart form data
	file, err := c.FormFile("image")
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Image file is required")
	}

	// Validate file extension
	if !strings.HasSuffix(file.Filename, ".tar") && !strings.HasSuffix(file.Filename, ".tar.gz") {
		return SendError(c, http.StatusBadRequest, "Only .tar and .tar.gz files are supported")
	}

	// Get metadata
	version := c.FormValue("version")
	if version == "" {
		version = time.Now().Format("20060102150405")
	}

	description := c.FormValue("description")
	if description == "" {
		description = "CLI upload " + version
	}

	// Save uploaded file temporarily
	src, err := file.Open()
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to open uploaded file")
	}
	defer src.Close()

	// Create temporary directory for upload
	tempDir := "/tmp/cli-app-uploads"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to create temporary directory")
	}

	// Save file to temporary location
	tempFilePath := filepath.Join(tempDir, fmt.Sprintf("%s_%s_%s", appName, version, file.Filename))
	dst, err := os.Create(tempFilePath)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to create temporary file")
	}
	defer dst.Close()
	defer os.Remove(tempFilePath) // Clean up

	if _, err := io.Copy(dst, src); err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to save uploaded file")
	}

	// Load image with podman
	tempImageName := fmt.Sprintf("%s:%s", appName, version)
	cmd := exec.Command("podman", "load", "-i", tempFilePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logman.Error("Failed to load image", "error", err, "output", string(output))
		return SendError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to load image: %s", string(output)))
	}

	// Parse loaded image name
	loadedImageName := parseLoadedImageName(string(output))
	if loadedImageName == "" {
		loadedImageName = tempImageName
	}

	// Tag with our desired name if different
	finalImageName := loadedImageName
	if loadedImageName != tempImageName {
		tagCmd := exec.Command("podman", "tag", loadedImageName, tempImageName)
		if _, tagErr := tagCmd.CombinedOutput(); tagErr != nil {
			logman.Warn("Failed to tag image", "error", tagErr)
		} else {
			finalImageName = tempImageName
		}
	}

	// Create Release record
	buildSourceInfo := models.JSONB{
		Data: map[string]interface{}{
			"type":        "cli_upload",
			"filename":    file.Filename,
			"description": description,
			"uploaded_at": time.Now().Format(time.RFC3339),
		},
	}

	release, err := models.CreateReleaseWithVersion(app.ID, version, finalImageName, buildSourceInfo, "success")
	if err != nil {
		logman.Error("Failed to create release", "app_id", app.ID, "error", err)
		return SendError(c, http.StatusInternalServerError, "Failed to create release record")
	}

	logman.Info("Release created successfully", "app_name", appName, "release_id", release.ID, "image_name", finalImageName)

	response := map[string]interface{}{
		"release_id":  fmt.Sprintf("rel-%d", release.ID),
		"version":     version,
		"description": description,
		"image_size":  file.Size,
		"image_name":  finalImageName,
		"status":      "success",
		"app_name":    appName,
		"app_id":      app.ID,
		"created_at":  release.CreatedAt.Format(time.RFC3339),
	}

	return SendSuccess(c, response)
}

// CreateApplicationDeployment handles deployment creation for applications by name
// Endpoint: POST /api/apps/by-name/:appName/deployments
func CreateApplicationDeployment(c echo.Context) error {
	appName := c.Param("appName")
	if appName == "" {
		return SendError(c, http.StatusBadRequest, "appName is required")
	}

	// Get application by name (fallback remains)
	app, err := models.GetApplicationByName(appName)
	if err != nil {
		return SendError(c, http.StatusNotFound, "Application not found: "+appName)
	}

	// Validate application token permission if using app token
	if err := validateApplicationTokenPermission(c, app.ID); err != nil {
		return err
	}

	// Parse request body
	var req struct {
		ReleaseUid string                 `json:"release_uid"`
		Source     string                 `json:"source"`
		Metadata   map[string]interface{} `json:"metadata"`
	}

	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid request body")
	}

	// Decode the friendly Release UID from the request
	if req.ReleaseUid == "" {
		return SendError(c, http.StatusBadRequest, "release_uid is required")
	}
	releaseID, err := DecodeFriendlyID(PrefixRelease, req.ReleaseUid)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid release_uid format")
	}

	// Set default source
	if req.Source == "" {
		req.Source = "cli"
	}

	// Create services needed for deployment orchestrator
	buildService := services.NewBuildService()
	envService := services.NewDeploymentEnvironmentService()
	podmanService := services.NewPodmanService()
	deploymentOrchestrator := services.NewDeploymentOrchestrator(buildService, envService, podmanService)

	// Create deployment request with the decoded UUID
	deployReq := services.CreateDeploymentRequest{
		ReleaseID: &releaseID,
	}

	logman.Info("Creating deployment for application", "app_name", appName, "app_id", app.ID, "release_uid", req.ReleaseUid)

	// Create and start deployment
	deployment, err := deploymentOrchestrator.CreateDeployment(app.ID, deployReq)
	if err != nil {
		logman.Error("Failed to create deployment", "app_name", appName, "error", err)
		return SendError(c, http.StatusInternalServerError, "Failed to create deployment: "+err.Error())
	}

	logman.Info("Deployment created successfully", "app_name", appName, "deployment_id", deployment.ID)

	// Populate the response DTO with encoded friendly UIDs
	response := DeploymentResponse{
		Uid:            EncodeFriendlyID(PrefixDeployment, deployment.ID),
		ApplicationUid: EncodeFriendlyID(PrefixApplication, app.ID),
		ReleaseUid:     EncodeFriendlyID(PrefixRelease, deployment.ReleaseID),
		Status:         deployment.Status,
		LogText:        deployment.LogText, // Assumes LogText exists on the deployment model
		StartedAt:      deployment.StartedAt,
		FinishedAt:     deployment.FinishedAt,
		CreatedAt:      deployment.CreatedAt,
		UpdatedAt:      deployment.UpdatedAt,
		SystemPort:     deployment.SystemPort,
	}

	// Return 202 Accepted for async deployment
	return c.JSON(http.StatusAccepted, response)
}

// ExportApplicationConfig exports application configuration as TOML
// Endpoint: GET /api/apps/by-name/:appName/config/export
func ExportApplicationConfig(c echo.Context) error {
	appName := c.Param("appName")
	if appName == "" {
		return SendError(c, http.StatusBadRequest, "appName is required")
	}

	// Get application by name
	app, err := models.GetApplicationByName(appName)
	if err != nil {
		return SendError(c, http.StatusNotFound, "Application not found: "+appName)
	}

	// Validate application token permission if using app token
	if err := validateApplicationTokenPermission(c, app.ID); err != nil {
		return err
	}

	// Generate TOML configuration
	tomlConfig := fmt.Sprintf(`api_version = "webdeploy.io/v1"
kind = "DeploymentSpec"

# Application Configuration
app_name = "%s"
project = "project-%d"
environment = "production"

[container]
publish_port = %d

# Generated from Web UI configuration at %s
# Modify as needed for your deployment requirements
`,
		app.Name,
		app.ProjectID,
		app.TargetPort,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	// Set response headers for file download
	c.Response().Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-orbitdeploy.toml"`, appName))

	return c.String(http.StatusOK, tomlConfig)
}

// parseDeploymentID extracts numeric ID from CLI deployment ID format (deploy-123 -> 123)
func parseDeploymentID(deploymentIDParam string) (uint, error) {
	if deploymentIDParam == "" {
		return 0, fmt.Errorf("deployment_id is required")
	}

	// Try to parse deployment_id as a number (remove deploy- prefix if present)
	deploymentIDStr := strings.TrimPrefix(deploymentIDParam, "deploy-")

	// Convert to uint
	var deploymentID uint
	if _, err := fmt.Sscanf(deploymentIDStr, "%d", &deploymentID); err != nil {
		return 0, fmt.Errorf("invalid deployment_id format: %s", deploymentIDParam)
	}

	return deploymentID, nil
}
