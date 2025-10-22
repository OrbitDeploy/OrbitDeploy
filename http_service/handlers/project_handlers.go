package handlers

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/opentdp/go-helper/dborm"
	"github.com/youfun/OrbitDeploy/models"
	"github.com/youfun/OrbitDeploy/services"
)

// isGitHubRepository checks if the given repository URL is a GitHub repository
func isGitHubRepository(gitRepository string) bool {
	return strings.Contains(gitRepository, "github.com")
}

// getDefaultGitHubTokenIDForUser retrieves the default GitHub token ID for a user
func getDefaultGitHubTokenIDForUser(userID uint) (uuid.UUID, error) {
	var token models.GitHubToken
	err := dborm.Db.Where("user_id = ? AND is_active = ?", userID, true).
		Order("created_at desc"). // 假设最新的为默认
		First(&token).Error
	if err != nil {
		return uuid.Nil, err
	}
	return token.ID, nil
}

// Original CreateProjectHandler - 更新为新的项目模型 (原则五)
func CreateProjectHandlerOriginal(c echo.Context) error {
	type CreateProjectRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	var req CreateProjectRequest
	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid request body")
	}

	project, err := models.CreateProject(req.Name, req.Description)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to create project")
	}

	response := ProjectResponse{
		Uid:         EncodeFriendlyID(PrefixProject, project.ID),
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}

	return SendCreated(c, response)
}

// NewCreateProjectHandler creates a new project handler with dependency injection
// 这实现了原则一：依赖外置，不内建
func NewCreateProjectHandler(projectManager *services.Manager) echo.HandlerFunc {
	return func(c echo.Context) error {
		type CreateProjectRequest struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}

		var req CreateProjectRequest
		if err := c.Bind(&req); err != nil {
			return SendError(c, http.StatusBadRequest, "Invalid request body")
		}

		// 首先通过项目管理器执行 Setup 流程 (集成 project_service.go 的 Setup 功能)暂时注释掉
		// projectSetup, err := projectManager.Setup(req.Name)
		// if err != nil {
		// 	return SendError(c, http.StatusInternalServerError, "Failed to setup project environment: "+err.Error())
		// }

		// 创建项目记录，并填充 Setup 的结果
		project, err := models.CreateProjectWithSetup(req.Name, req.Description, "", "")
		if err != nil {
			// 如果数据库创建失败，清理已创建的系统环境
			_ = projectManager.Teardown(req.Name)
			return SendError(c, http.StatusInternalServerError, "Failed to create project: "+err.Error())
		}

		response := ProjectResponse{
			Uid:         EncodeFriendlyID(PrefixProject, project.ID),
			Name:        project.Name,
			Description: project.Description,
			CreatedAt:   project.CreatedAt,
			UpdatedAt:   project.UpdatedAt,
		}

		return SendCreated(c, response)
	}
}

// CreateProjectHandler - 为了向后兼容，重定向到原有逻辑
func CreateProjectHandler(c echo.Context) error {
	return CreateProjectHandlerOriginal(c)
}

func ListProjectsHandler(c echo.Context) error {
	projects, err := models.ListProjects()
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to list projects")
	}

	var responses []ProjectResponse
	for _, p := range projects {
		responses = append(responses, ProjectResponse{
			Uid:         EncodeFriendlyID(PrefixProject, p.ID),
			Name:        p.Name,
			Description: p.Description,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		})
	}

	return SendSuccess(c, responses)
}

func GetProjectHandler(c echo.Context) error {
	projectUID := c.Param("projectId")
	projectID, err := DecodeFriendlyID(PrefixProject, projectUID)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid project ID format")
	}

	project, err := models.GetProjectByID(projectID)
	if err != nil {
		return SendError(c, http.StatusNotFound, "Project not found")
	}

	response := ProjectResponse{
		Uid:         EncodeFriendlyID(PrefixProject, project.ID),
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}

	return SendSuccess(c, response)
}

// GetProjectByNameHandler handles GET /api/projects/by-name/:name
func GetProjectByNameHandler(c echo.Context) error {
	projectName := c.Param("name")
	if projectName == "" {
		return SendError(c, http.StatusBadRequest, "Project name is required")
	}

	project, err := models.GetProjectByName(projectName)
	if err != nil {
		return SendError(c, http.StatusNotFound, "Project not found")
	}

	response := ProjectResponse{
		Uid:         EncodeFriendlyID(PrefixProject, project.ID),
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}

	return SendSuccess(c, response)
}

func ListApplicationsByProjectHandler(c echo.Context) error {
	projectIDStr := c.Param("projectId")
	projectID, err := DecodeFriendlyID(PrefixProject, projectIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid project ID format")
	}

	var applications []models.Application
	if err := dborm.Db.
		Preload("ActiveRelease").
		Where("project_id = ?", projectID).
		Order("updated_at DESC").
		Find(&applications).Error; err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to list applications")
	}

	// 使用新的 ApplicationDetailResponse 来构建返回结果
	result := make([]ApplicationDetailResponse, 0, len(applications))
	for _, app := range applications {
		dto := ApplicationDetailResponse{
			Uid:              EncodeFriendlyID(PrefixApplication, app.ID),
			ProjectUid:       EncodeFriendlyID(PrefixProject, app.ProjectID),
			Name:             app.Name,
			Description:      app.Description,
			RepoURL:          app.RepoURL,
			TargetPort:       app.TargetPort,
			Status:           app.Status,
			Volumes:          app.Volumes,
			ExecCommand:      app.ExecCommand,
			AutoUpdatePolicy: app.AutoUpdatePolicy,
			Branch:           app.Branch,
			BuildDir:         app.BuildDir,  // Added mapping
			BuildType:        app.BuildType, // Added mapping
			CreatedAt:        app.CreatedAt,
			UpdatedAt:        app.UpdatedAt,
		}

		if app.ActiveReleaseID != nil {
			uid := EncodeFriendlyID(PrefixRelease, *app.ActiveReleaseID)
			dto.ActiveReleaseUid = &uid
		}

		if app.ActiveRelease != nil && app.ActiveRelease.ID != uuid.Nil {
			dto.ActiveReleaseInfo = &ReleaseInfo{
				Uid:       EncodeFriendlyID(PrefixRelease, app.ActiveRelease.ID),
				ImageName: &app.ActiveRelease.ImageName,
			}
		}
		result = append(result, dto)
	}

	return SendSuccess(c, result)
}

// ListApplicationsByProjectNameHandler handles GET /api/projects/by-name/:name/apps
func ListApplicationsByProjectNameHandler(c echo.Context) error {
	projectName := c.Param("name")
	if projectName == "" {
		return SendError(c, http.StatusBadRequest, "Project name is required")
	}

	applications, err := models.ListApplicationsByProjectName(projectName)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to list applications")
	}

	// 使用 ApplicationDetailResponse 来构建返回结果
	result := make([]ApplicationDetailResponse, 0, len(applications))
	for _, app := range applications {
		dto := ApplicationDetailResponse{
			Uid:              EncodeFriendlyID(PrefixApplication, app.ID),
			ProjectUid:       EncodeFriendlyID(PrefixProject, app.ProjectID),
			Name:             app.Name,
			Description:      app.Description,
			RepoURL:          app.RepoURL,
			TargetPort:       app.TargetPort,
			Status:           app.Status,
			Volumes:          app.Volumes,
			ExecCommand:      app.ExecCommand,
			AutoUpdatePolicy: app.AutoUpdatePolicy,
			Branch:           app.Branch,
			BuildDir:         app.BuildDir,  // Added mapping
			BuildType:        app.BuildType, // Added mapping
			CreatedAt:        app.CreatedAt,
			UpdatedAt:        app.UpdatedAt,
		}

		if app.ActiveReleaseID != nil {
			uid := EncodeFriendlyID(PrefixRelease, *app.ActiveReleaseID)
			dto.ActiveReleaseUid = &uid
		}

		// 如果需要预加载 ActiveRelease，可以在这里添加类似逻辑
		// 但为简化，直接使用模型数据
		result = append(result, dto)
	}

	return SendSuccess(c, result)
}
