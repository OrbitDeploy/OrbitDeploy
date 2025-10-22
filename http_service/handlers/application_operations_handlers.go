package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/OrbitDeploy/OrbitDeploy/models"
	"github.com/OrbitDeploy/OrbitDeploy/services"
	"github.com/labstack/echo/v4"
)

// Application Operations Handlers

func GetAppRuntimeStatusHandler(c echo.Context) error {
	appIDStr := c.Param("appId")
	appID, err := DecodeFriendlyID(PrefixApplication, appIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid application ID format")
	}

	application, err := models.GetApplicationByID(appID)
	if err != nil {
		return SendError(c, http.StatusNotFound, "Application not found")
	}

	// For now, return the status from the model
	status := map[string]interface{}{
		"appId":  EncodeFriendlyID(PrefixApplication, application.ID),
		"status": application.Status,
	}

	return SendSuccess(c, status)
}

func RestartAppHandler(c echo.Context) error {
	appIDStr := c.Param("appId")
	appID, err := DecodeFriendlyID(PrefixApplication, appIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid application ID format")
	}

	// Check if application exists
	_, err = models.GetApplicationByID(appID)
	if err != nil {
		return SendError(c, http.StatusNotFound, "Application not found")
	}

	// TODO: Implement actual restart logic using services
	// For now, just return success
	return SendSuccess(c, nil)
}

func OverrideDeployHandler(c echo.Context) error {
	appIDStr := c.Param("appId")
	appID, err := DecodeFriendlyID(PrefixApplication, appIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid application ID format")
	}

	// Check if application exists
	_, err = models.GetApplicationByID(appID)
	if err != nil {
		return SendError(c, http.StatusNotFound, "Application not found")
	}

	// TODO: Implement actual override deploy logic using services
	// For now, just return success
	return SendSuccess(c, nil)
}

// GetApplicationLogsHandler returns aggregated logs for an application
func GetApplicationLogsHandler(c echo.Context) error {
	appIDStr := c.Param("appId")
	appID, err := DecodeFriendlyID(PrefixApplication, appIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid application ID format")
	}

	// Check if application exists
	application, err := models.GetApplicationByID(appID)
	if err != nil {
		return SendError(c, http.StatusNotFound, "Application not found")
	}

	// Parse query parameters for filtering
	limit := 100 // default limit
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	level := c.QueryParam("level") // info, error, warn, debug

	// Parse time filters
	var _, _ *time.Time // startTime, endTime (currently unused but ready for implementation)
	if startTimeStr := c.QueryParam("startTime"); startTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			_ = parsed // startTime = &parsed
		}
	}
	if endTimeStr := c.QueryParam("endTime"); endTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			_ = parsed // endTime = &parsed
		}
	}

	// TODO: Implement actual log aggregation logic
	// For now, return mock data structure that matches the frontend expectations
	logs := []map[string]interface{}{
		{
			"timestamp": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			"level":     "info",
			"message":   "Application started successfully",
			"source":    "container",
		},
		{
			"timestamp": time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
			"level":     "info",
			"message":   "Health check passed",
			"source":    "system",
		},
		{
			"timestamp": time.Now().Add(-15 * time.Minute).Format(time.RFC3339),
			"level":     "warn",
			"message":   "High memory usage detected",
			"source":    "system",
		},
	}

	// Apply level filter if specified
	if level != "" {
		filteredLogs := []map[string]interface{}{}
		for _, log := range logs {
			if log["level"] == level {
				filteredLogs = append(filteredLogs, log)
			}
		}
		logs = filteredLogs
	}

	// Apply pagination
	totalCount := len(logs)
	start := offset
	end := offset + limit
	if start > totalCount {
		start = totalCount
	}
	if end > totalCount {
		end = totalCount
	}
	if start < end {
		logs = logs[start:end]
	} else {
		logs = []map[string]interface{}{}
	}

	response := map[string]interface{}{
		"appId":      EncodeFriendlyID(PrefixApplication, application.ID),
		"logs":       logs,
		"totalCount": totalCount,
		"hasMore":    offset+limit < totalCount,
	}

	return SendSuccess(c, response)
}

// GetBranchesHandler fetches branches for a given repo URL (GitHub only for now)
func GetGitHubBranchesHandler(c echo.Context) error {
	repoUrl := c.QueryParam("repoUrl")
	if repoUrl == "" {
		return SendError(c, http.StatusBadRequest, "repoUrl is required")
	}

	// Only support GitHub for now
	if !strings.Contains(repoUrl, "github.com") {
		return SendSuccess(c, map[string]interface{}{"branches": []string{}})
	}

	// Extract owner/repo from URL
	parts := strings.Split(strings.TrimSuffix(repoUrl, ".git"), "/")
	if len(parts) < 2 {
		return SendError(c, http.StatusBadRequest, "Invalid GitHub repo URL")
	}
	repoPath := strings.Join(parts[len(parts)-2:], "/")

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/branches", repoPath)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to create API request")
	}

	// TODO: Add token support if needed for private repos
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "API request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SendError(c, http.StatusBadRequest, "Failed to fetch branches from GitHub")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to read API response")
	}

	var branches []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &branches); err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to parse API response")
	}

	branchNames := make([]string, len(branches))
	for i, branch := range branches {
		branchNames[i] = branch.Name
	}

	return SendSuccess(c, map[string]interface{}{"branches": branchNames})
}

// NewDeleteApplicationHandler 是一个工厂函数，它接收依赖，返回真正的 Handler
// 这实现了原则三：服务单例，生命周期与程序相同
func NewDeleteApplicationHandler(appService *services.ApplicationService) echo.HandlerFunc {
	// 返回的这个函数是一个闭包，它可以访问外部的 appService 变量
	return func(c echo.Context) error {
		appIDStr := c.Param("appId")
		appID, err := DecodeFriendlyID(PrefixApplication, appIDStr)
		if err != nil {
			return SendError(c, http.StatusBadRequest, "Invalid application ID format")
		}

		// 解析请求体
		type DeleteApplicationRequest struct {
			ApplicationName string `json:"application_name"`
		}

		var req DeleteApplicationRequest
		if err := c.Bind(&req); err != nil {
			return SendError(c, http.StatusBadRequest, "Invalid request body")
		}

		if req.ApplicationName == "" {
			return SendError(c, http.StatusBadRequest, "Application name is required for confirmation")
		}

		// 直接使用从外部注入的、早已创建好的 appService 实例
		// 不再需要在这里 New()
		if err := appService.ValidateApplicationDeletion(appID, req.ApplicationName); err != nil {
			return SendError(c, http.StatusBadRequest, err.Error())
		}

		// 执行删除操作
		if err := appService.DeleteApplicationWithCleanup(appID, req.ApplicationName); err != nil {
			return SendError(c, http.StatusInternalServerError, fmt.Sprintf("删除应用失败: %v", err))
		}

		return SendSuccess(c, map[string]interface{}{
			"message": "应用删除成功",
			"appId":   appIDStr,
		})
	}
}
