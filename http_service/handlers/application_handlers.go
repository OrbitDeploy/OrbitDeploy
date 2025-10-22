package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/youfun/OrbitDeploy/models"
)

// Application Handlers

func CreateApplicationHandler(c echo.Context) error {
	projectUID := c.Param("projectId")
	projectID, err := DecodeFriendlyID(PrefixProject, projectUID)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid project UID")
	}

	var req CreateApplicationRequest
	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid request body")
	}

	var providerAuthID *uuid.UUID
	if req.ProviderAuthUid != nil && *req.ProviderAuthUid != "" {
		id, err := DecodeFriendlyID(PrefixProviderAuth, *req.ProviderAuthUid)
		if err != nil {
			return SendError(c, http.StatusBadRequest, "Invalid providerAuthUid")
		}
		providerAuthID = &id
	}

	application, err := models.CreateApplication(projectID, req.Name, req.Description, req.RepoURL, req.TargetPort, models.JSONB{Data: req.Volumes}, req.ExecCommand, req.AutoUpdatePolicy, req.Branch, req.BuildDir, req.BuildType, providerAuthID)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to create application")
	}

	var activeReleaseUID *string
	if application.ActiveReleaseID != nil {
		uid := EncodeFriendlyID(PrefixRelease, *application.ActiveReleaseID)
		activeReleaseUID = &uid
	}

	response := ApplicationDetailResponse{
		Uid:              EncodeFriendlyID(PrefixApplication, application.ID),
		ProjectUid:       EncodeFriendlyID(PrefixProject, application.ProjectID),
		Name:             application.Name,
		Description:      application.Description,
		RepoURL:          application.RepoURL,
		ActiveReleaseUid: activeReleaseUID,
		TargetPort:       application.TargetPort,
		Status:           application.Status,
		Volumes:          application.Volumes,
		ExecCommand:      application.ExecCommand,
		AutoUpdatePolicy: application.AutoUpdatePolicy,
		Branch:           application.Branch,
		BuildDir:         application.BuildDir,
		BuildType:        application.BuildType,
		CreatedAt:        application.CreatedAt,
		UpdatedAt:        application.UpdatedAt,
	}

	return SendCreated(c, response)
}

func GetApplicationHandler(c echo.Context) error {
	appUID := c.Param("appId")
	appID, err := DecodeFriendlyID(PrefixApplication, appUID)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid application UID")
	}

	application, err := models.GetApplicationByID(appID)
	if err != nil {
		return SendError(c, http.StatusNotFound, "Application not found")
	}

	var activeReleaseUID *string
	if application.ActiveReleaseID != nil {
		uid := EncodeFriendlyID(PrefixRelease, *application.ActiveReleaseID)
		activeReleaseUID = &uid
	}

	response := ApplicationDetailResponse{
		Uid:              EncodeFriendlyID(PrefixApplication, application.ID),
		ProjectUid:       EncodeFriendlyID(PrefixProject, application.ProjectID),
		Name:             application.Name,
		Description:      application.Description,
		RepoURL:          application.RepoURL,
		ActiveReleaseUid: activeReleaseUID,
		TargetPort:       application.TargetPort,
		Status:           application.Status,
		Volumes:          application.Volumes,
		ExecCommand:      application.ExecCommand,
		AutoUpdatePolicy: application.AutoUpdatePolicy,
		Branch:           application.Branch,
		BuildDir:         application.BuildDir,
		BuildType:        application.BuildType,
		CreatedAt:        application.CreatedAt,
		UpdatedAt:        application.UpdatedAt,
	}

	return SendSuccess(c, response)
}

// GetApplicationByNameHandler handles GET /api/apps/by-name/:name
func GetApplicationByNameHandler(c echo.Context) error {
	appName := c.Param("name")
	if appName == "" {
		return SendError(c, http.StatusBadRequest, "Application name is required")
	}

	application, err := models.GetApplicationByName(appName)
	if err != nil {
		return SendError(c, http.StatusNotFound, "Application not found")
	}

	var activeReleaseUID *string
	if application.ActiveReleaseID != nil {
		uid := EncodeFriendlyID(PrefixRelease, *application.ActiveReleaseID)
		activeReleaseUID = &uid
	}

	response := ApplicationDetailResponse{
		Uid:              EncodeFriendlyID(PrefixApplication, application.ID),
		ProjectUid:       EncodeFriendlyID(PrefixProject, application.ProjectID),
		Name:             application.Name,
		Description:      application.Description,
		RepoURL:          application.RepoURL,
		ActiveReleaseUid: activeReleaseUID,
		TargetPort:       application.TargetPort,
		Status:           application.Status,
		Volumes:          application.Volumes,
		ExecCommand:      application.ExecCommand,
		AutoUpdatePolicy: application.AutoUpdatePolicy,
		Branch:           application.Branch,
		BuildDir:         application.BuildDir,
		BuildType:        application.BuildType,
		CreatedAt:        application.CreatedAt,
		UpdatedAt:        application.UpdatedAt,
	}

	return SendSuccess(c, response)
}

// GetApplicationByProjectNameAndAppNameHandler handles GET /api/projects/by-name/:projectName/apps/by-name/:appName
func GetApplicationByProjectNameAndAppNameHandler(c echo.Context) error {
	projectName := c.Param("projectName")
	appName := c.Param("appName")

	if projectName == "" {
		return SendError(c, http.StatusBadRequest, "Project name is required")
	}
	if appName == "" {
		return SendError(c, http.StatusBadRequest, "Application name is required")
	}

	application, err := models.GetApplicationByProjectNameAndAppName(projectName, appName)
	if err != nil {
		return SendError(c, http.StatusNotFound, "Application not found")
	}

	var activeReleaseUID *string
	if application.ActiveReleaseID != nil {
		uid := EncodeFriendlyID(PrefixRelease, *application.ActiveReleaseID)
		activeReleaseUID = &uid
	}

	response := ApplicationDetailResponse{
		Uid:              EncodeFriendlyID(PrefixApplication, application.ID),
		ProjectUid:       EncodeFriendlyID(PrefixProject, application.ProjectID),
		Name:             application.Name,
		Description:      application.Description,
		RepoURL:          application.RepoURL,
		ActiveReleaseUid: activeReleaseUID,
		TargetPort:       application.TargetPort,
		Status:           application.Status,
		Volumes:          application.Volumes,
		ExecCommand:      application.ExecCommand,
		AutoUpdatePolicy: application.AutoUpdatePolicy,
		Branch:           application.Branch,
		BuildDir:         application.BuildDir,
		BuildType:        application.BuildType,
		CreatedAt:        application.CreatedAt,
		UpdatedAt:        application.UpdatedAt,
	}

	return SendSuccess(c, response)
}

func UpdateApplicationHandler(c echo.Context) error {
	appUID := c.Param("appId")
	appID, err := DecodeFriendlyID(PrefixApplication, appUID)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid application UID")
	}

	var req UpdateApplicationRequest
	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid request body")
	}

	var providerAuthID *uuid.UUID
	if req.ProviderAuthUid != nil && *req.ProviderAuthUid != "" {
		id, err := DecodeFriendlyID(PrefixProviderAuth, *req.ProviderAuthUid)
		if err != nil {
			return SendError(c, http.StatusBadRequest, "Invalid providerAuthUid")
		}
		providerAuthID = &id
	}

	application, err := models.UpdateApplicationFromFrontend(appID, req.Description, req.RepoURL, req.TargetPort, req.Status, models.JSONB{Data: req.Volumes}, req.ExecCommand, req.AutoUpdatePolicy, req.Branch, req.BuildDir, req.BuildType, providerAuthID)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to update application")
	}

	var activeReleaseUID *string
	if application.ActiveReleaseID != nil {
		uid := EncodeFriendlyID(PrefixRelease, *application.ActiveReleaseID)
		activeReleaseUID = &uid
	}

	response := ApplicationDetailResponse{
		Uid:              EncodeFriendlyID(PrefixApplication, application.ID),
		ProjectUid:       EncodeFriendlyID(PrefixProject, application.ProjectID),
		Name:             application.Name,
		Description:      application.Description,
		RepoURL:          application.RepoURL,
		ActiveReleaseUid: activeReleaseUID,
		TargetPort:       application.TargetPort,
		Status:           application.Status,
		Volumes:          application.Volumes,
		ExecCommand:      application.ExecCommand,
		AutoUpdatePolicy: application.AutoUpdatePolicy,
		Branch:           application.Branch,
		BuildDir:         application.BuildDir,
		BuildType:        application.BuildType,
		CreatedAt:        application.CreatedAt,
		UpdatedAt:        application.UpdatedAt,
	}

	return SendSuccess(c, response)
}
