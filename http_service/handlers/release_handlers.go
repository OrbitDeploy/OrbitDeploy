package handlers

import (
	"net/http"

	"github.com/OrbitDeploy/OrbitDeploy/models"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Release Handlers

func CreateReleaseAndBuildHandler(c echo.Context) error {
	appIDStr := c.Param("appId")
	appID, err := DecodeFriendlyID(PrefixApplication, appIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid application ID format")
	}

	var req CreateReleaseRequest
	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid request body")
	}

	release, err := models.CreateRelease(appID, req.ImageName, models.JSONB{Data: req.BuildSourceInfo}, req.Status)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to create release")
	}

	return SendCreated(c, toReleaseResponse(release))
}

func ListReleasesHandler(c echo.Context) error {
	appIDStr := c.Param("appId")
	appID, err := DecodeFriendlyID(PrefixApplication, appIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid application ID format")
	}

	releases, err := models.ListReleasesByAppID(appID)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to list releases")
	}

	// Convert to API models
	var responses []*ReleaseResponse
	for _, r := range releases {
		responses = append(responses, toReleaseResponse(r))
	}

	return SendSuccess(c, responses)
}

func GetReleaseHandler(c echo.Context) error {
	releaseIDStr := c.Param("releaseId")
	releaseID, err := DecodeFriendlyID(PrefixRelease, releaseIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid release ID format")
	}

	release, err := models.GetReleaseByID(releaseID)
	if err != nil {
		return SendError(c, http.StatusNotFound, "Release not found")
	}

	return SendSuccess(c, toReleaseResponse(release))
}

// GetLastReleaseHandler gets the latest release for an application
func GetLatestReleaseHandler(c echo.Context) error {
	appIDStr := c.Param("appId")
	appID, err := DecodeFriendlyID(PrefixApplication, appIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid application ID format")
	}

	var LatestRelease *models.Release
	LatestRelease, err = models.GetLatestRelease(appID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return SendError(c, http.StatusNotFound, "No releases found for this application")
		}
	}

	return SendSuccess(c, toReleaseResponse(LatestRelease))
}

func toReleaseResponse(r *models.Release) *ReleaseResponse {
	var buildSourceInfo map[string]interface{}
	if r.BuildSourceInfo.Data != nil {
		if data, ok := r.BuildSourceInfo.Data.(map[string]interface{}); ok {
			buildSourceInfo = data
		}
	}

	return &ReleaseResponse{
		Uid:             EncodeFriendlyID(PrefixRelease, r.ID),
		ApplicationUid:  EncodeFriendlyID(PrefixApplication, r.ApplicationID),
		ImageName:       r.ImageName,
		BuildSourceInfo: buildSourceInfo,
		Status:          r.Status,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}
