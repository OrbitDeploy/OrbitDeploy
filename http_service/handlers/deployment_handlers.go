package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/OrbitDeploy/OrbitDeploy/models"
	"github.com/OrbitDeploy/OrbitDeploy/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/opentdp/go-helper/dborm"
	"github.com/opentdp/go-helper/logman"
)

// NewCreateDeploymentHandler 是一个工厂函数，它接收依赖，返回真正的 Handler
// 这实现了原则三：服务单例，生命周期与程序相同
func NewCreateDeploymentHandler(deploymentOrchestrator *services.DeploymentOrchestrator) echo.HandlerFunc {
	// 返回的这个函数是一个闭包，它可以访问外部的 deploymentOrchestrator 变量
	return func(c echo.Context) error {
		return createDeploymentHandlerImpl(c, deploymentOrchestrator)
	}
}

// createDeploymentHandlerImpl 是 CreateDeploymentHandler 的内部实现
func createDeploymentHandlerImpl(c echo.Context, deploymentOrchestrator *services.DeploymentOrchestrator) error {
	logman.Info("开始处理创建部署请求")

	appUID := c.Param("appId")
	appID, err := DecodeFriendlyID(PrefixApplication, appUID)
	if err != nil {
		logman.Error("解析应用UID失败", "app_uid", appUID, "error", err)
		return SendError(c, http.StatusBadRequest, "无效的应用UID")
	}
	logman.Info("应用UID解析成功", "app_id", appID)

	// 使用 services 包中的 CreateDeploymentRequest 结构体
	var req services.CreateDeploymentRequest
	if err := c.Bind(&req); err != nil {
		logman.Error("绑定请求体失败", "error", err)
		return SendError(c, http.StatusBadRequest, "无效的请求体")
	}
	logman.Info("请求体绑定成功")

	// 直接使用从外部注入的、早已创建好的 deploymentOrchestrator 实例
	// 不再需要在这里 New()
	logman.Info("调用部署服务创建部署", "app_id", appID)
	deployment, err := deploymentOrchestrator.CreateDeployment(appID, req)
	if err != nil {
		logman.Error("创建部署失败", "app_id", appID, "error", err)
		return SendError(c, http.StatusInternalServerError, "创建部署失败: "+err.Error())
	}
	logman.Info("部署创建成功", "deployment_id", deployment.ID)

	// 返回创建的部署信息
	return SendCreated(c, deployment)
}

// CreateDeploymentHandler 保留向后兼容性，但标记为已弃用
// Deprecated: 使用 NewCreateDeploymentHandler 工厂函数代替
func CreateDeploymentHandler(c echo.Context) error {
	logman.Warn("CreateDeploymentHandler 已弃用，请使用依赖注入方式")
	// 创建部署编排服务实例（向后兼容）
	buildService := services.NewBuildService()
	envService := services.NewDeploymentEnvironmentService()
	podmanService := services.NewPodmanService()
	do := services.NewDeploymentOrchestrator(buildService, envService, podmanService)
	return createDeploymentHandlerImpl(c, do)
}

func ListDeploymentsByAppHandler(c echo.Context) error {
	appUID := c.Param("appId")
	appID, err := DecodeFriendlyID(PrefixApplication, appUID)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid application UID")
	}

	deployments, err := models.ListDeploymentsByAppID(appID)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to list deployments")
	}

	result := make([]DeploymentResponse, 0, len(deployments))
	for _, d := range deployments {
		resp := DeploymentResponse{
			Uid:            EncodeFriendlyID(PrefixDeployment, d.ID),
			ApplicationUid: EncodeFriendlyID(PrefixApplication, d.ApplicationID),
			ReleaseUid:     EncodeFriendlyID(PrefixRelease, d.ReleaseID),
			Status:         d.Status,
			LogText:        d.LogText,
			StartedAt:      d.StartedAt,
			FinishedAt:     d.FinishedAt,
			CreatedAt:      d.CreatedAt,
			UpdatedAt:      d.UpdatedAt,
			SystemPort:     d.SystemPort,
		}

		if d.Release.ID != uuid.Nil {
			if d.Release.Version != "" {
				resp.Version = &d.Release.Version
			}
			if d.Release.ImageName != "" {
				resp.ImageName = &d.Release.ImageName
			}

			resp.ReleaseStatus = d.Release.Status
		}
		if resp.SystemPort == nil {
			resp.SystemPort = new(int)
			*resp.SystemPort = 0
		}

		result = append(result, resp)
	}

	fmt.Println(result)
	// 3. Let the framework handle JSON marshaling automatically.
	return SendSuccess(c, result)
}

func ListRunningDeploymentsByAppHandler(c echo.Context) error {
	identifier := c.Param("name") // Assuming route is /apps/:name, but now accepts name or ID
	if identifier == "" {
		return SendError(c, http.StatusBadRequest, "Application identifier is required")
	}

	var appID uuid.UUID
	// Prioritize name lookup
	if app, err := models.GetApplicationByName(identifier); err == nil {
		appID = app.ID
	} else {
		// Fallback to ID parsing
		parsedID, err := DecodeFriendlyID(PrefixApplication, identifier)
		if err != nil {
			logman.Error("Invalid application identifier", "identifier", identifier, "error", err)
			return SendError(c, http.StatusBadRequest, "Invalid application identifier")
		}
		appID = parsedID
	}

	deployments, err := models.GetRunningDeploymentsByAppID(appID) // Changed to correct function
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to list running deployments")
	}

	// Also get routings for the application to include domain info
	routings, err := models.GetActiveRoutingsByApplicationID(appID)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to get routings")
	}

	// Create a map of host ports to domains for efficient lookup
	portToDomains := make(map[int][]string)
	for _, routing := range routings {
		if routing.IsActive {
			portToDomains[routing.HostPort] = append(portToDomains[routing.HostPort], routing.DomainName)
		}
	}

	result := make([]RunningDeploymentResponse, 0, len(deployments))
	for _, d := range deployments {
		resp := RunningDeploymentResponse{
			DeploymentResponse: DeploymentResponse{
				Uid:            EncodeFriendlyID(PrefixDeployment, d.ID),
				ApplicationUid: EncodeFriendlyID(PrefixApplication, d.ApplicationID),
				ReleaseUid:     EncodeFriendlyID(PrefixRelease, d.ReleaseID),
				Status:         d.Status,
				LogText:        d.LogText,
				StartedAt:      d.StartedAt,
				FinishedAt:     d.FinishedAt,
				CreatedAt:      d.CreatedAt,
				UpdatedAt:      d.UpdatedAt,
			},
		}

		if d.Release.ID != uuid.Nil {
			if d.Release.Version != "" {
				resp.Version = &d.Release.Version
			}
			if d.Release.ImageName != "" {
				resp.ImageName = &d.Release.ImageName
			}

			resp.ReleaseStatus = d.Release.Status
		}
		if resp.SystemPort == nil {
			resp.SystemPort = new(int)
			*resp.SystemPort = 0
		}

		// Find host port from routing - assuming there's a port associated
		// This is a simplification; in reality you might need more complex logic
		for port, domains := range portToDomains {
			resp.HostPort = port
			resp.Domains = domains
			break // For now, take the first port found
		}

		result = append(result, resp)
	}

	return SendSuccess(c, result)
}

func GetDeploymentHandler(c echo.Context) error {
	deploymentUID := c.Param("deploymentId")
	deploymentID, err := DecodeFriendlyID(PrefixDeployment, deploymentUID)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid deployment UID")
	}

	// Get deployment with Release preloaded
	var deployment models.Deployment
	if err := dborm.Db.Preload("Release").Where("id = ?", deploymentID).First(&deployment).Error; err != nil {
		return SendError(c, http.StatusNotFound, "Deployment not found")
	}

	// 构建独立的API响应结构体，避免直接暴露数据库模型
	resp := DeploymentResponse{
		Uid:            EncodeFriendlyID(PrefixDeployment, deployment.ID),
		ApplicationUid: EncodeFriendlyID(PrefixApplication, deployment.ApplicationID),
		ReleaseUid:     EncodeFriendlyID(PrefixRelease, deployment.ReleaseID),
		Status:         deployment.Status,
		LogText:        deployment.LogText,
		StartedAt:      deployment.StartedAt,
		FinishedAt:     deployment.FinishedAt,
		CreatedAt:      deployment.CreatedAt,
		UpdatedAt:      deployment.UpdatedAt,
	}

	// Populate version and other fields from Release
	if deployment.Release.ID != uuid.Nil {
		if deployment.Release.Version != "" {
			resp.Version = &deployment.Release.Version
		}
		if deployment.Release.ImageName != "" {
			resp.ImageName = &deployment.Release.ImageName
		}

		resp.ReleaseStatus = deployment.Release.Status
	}
	if resp.SystemPort == nil {
		resp.SystemPort = new(int)
		*resp.SystemPort = 0
	}

	return SendSuccess(c, resp)
}

// GetDeploymentLogsHandler 获取部署的结构化日志（分页）
// GET /api/deployments/:deploymentId/logs-data?limit=200&before_timestamp=2024-01-01T10:00:00Z
func GetDeploymentLogsHandler(c echo.Context) error {
	deploymentUID := c.Param("deploymentId")
	deploymentID, err := DecodeFriendlyID(PrefixDeployment, deploymentUID)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid deployment UID")
	}

	// Validate application token permission if using app token
	authType := c.Get("auth_type")
	if authType == "app_token" {
		// Get deployment to check application ownership
		deployment, err := models.GetDeploymentByID(deploymentID)
		if err != nil {
			return SendError(c, http.StatusNotFound, "Deployment not found")
		}

		if err := validateApplicationTokenPermission(c, deployment.ApplicationID); err != nil {
			return err
		}
	}

	// 解析查询参数
	limit := 200 // 默认返回200条
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	var beforeTimestamp *time.Time
	if beforeStr := c.QueryParam("before_timestamp"); beforeStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, beforeStr); err == nil {
			beforeTimestamp = &parsedTime
		}
	}

	// 获取日志
	logs, err := models.GetDeploymentLogs(deploymentID, limit, beforeTimestamp)
	if err != nil {
		logman.Error("获取部署日志失败", "deployment_id", deploymentID, "error", err)
		return SendError(c, http.StatusInternalServerError, "Failed to get deployment logs")
	}

	// 构建响应
	result := make([]DeploymentLogResponse, 0, len(logs))
	for _, log := range logs {
		result = append(result, DeploymentLogResponse{
			ID:        log.ID,
			Timestamp: log.Timestamp,
			Level:     log.Level,
			Source:    log.Source,
			Message:   log.Message,
		})
	}

	return SendSuccess(c, result)
}
