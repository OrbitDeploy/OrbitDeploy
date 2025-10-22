package http_service

import (
	"embed"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/youfun/OrbitDeploy/http_service/handlers"
	"github.com/youfun/OrbitDeploy/models"
	"github.com/youfun/OrbitDeploy/services"
)

// RegisterRoutes sets up all HTTP routes for the application using Echo
func RegisterRoutes() {
	// This function is kept for compatibility but redirects to Echo implementation
	log.Println("Warning: RegisterRoutes is deprecated, use NewEchoServer() instead")
}

// SetInstallationScripts sets the installation scripts for environment handlers
func SetInstallationScripts(podmanScript, caddyScript func() string) {
	handlers.SetInstallScripts(podmanScript, caddyScript)
}

// SetDockerBuildQueueService sets the docker build queue service for handlers
func SetDockerBuildQueueService(svc *services.DockerBuildQueueService) {
	handlers.SetDockerBuildQueueService(svc)
}

// NewEchoServer creates and configures a new Echo server with all routes
func NewEchoServer(assets embed.FS) *echo.Echo {
	// This function is kept for backward compatibility
	// In production, use NewEchoServerWithDependencies instead
	log.Println("Warning: NewEchoServer is deprecated for dependency injection, consider using NewEchoServerWithDependencies")
	return createEchoServerWithRoutes(assets, nil, nil, nil, nil, nil)
}

// NewEchoServerWithDependencies creates and configures a new Echo server with dependency injection
// 这实现了原则二：统一组装，集中管理
func NewEchoServerWithDependencies(assets embed.FS, appService *services.ApplicationService, deploymentOrchestrator *services.DeploymentOrchestrator, projectManager *services.Manager, databaseService *services.DatabaseService, databaseOrchestrator *services.DatabaseOrchestrator) *echo.Echo {
	return createEchoServerWithRoutes(assets, appService, deploymentOrchestrator, projectManager, databaseService, databaseOrchestrator)
}

// createEchoServerWithRoutes creates the actual Echo server with optional dependency injection
func createEchoServerWithRoutes(assets embed.FS, appService *services.ApplicationService, deploymentOrchestrator *services.DeploymentOrchestrator, projectManager *services.Manager, databaseService *services.DatabaseService, databaseOrchestrator *services.DatabaseOrchestrator) *echo.Echo {

	e := echo.New()

	// Global middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// 使用 StaticWithConfig 中间件来服务嵌入的文件
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		HTML5: true,
		Root:  "frontend/dist",
		// 使用传入的参数
		Filesystem: http.FS(assets),
	}))
	// Custom global error handler
	e.HTTPErrorHandler = customHTTPErrorHandler

	// Public routes (no authentication required)
	api := e.Group("/api")
	api.POST("/setup", handlers.Setup)
	// Add public GET route for GitHub App callback (handles redirect from GitHub)
	api.GET("/providers/github/app-callback", handlers.HandleGitHubAppCallback)
	// Add public POST route for GitHub webhooks (GitHub sends unauthenticated requests)
	api.POST("/providers/github/webhook", handlers.HandleGitHubWebhook)

	// Setup routes
	setup := api.Group("/setup")
	setup.GET("/check", handlers.CheckSetup)

	// Auth routes
	auth := api.Group("/auth")
	auth.POST("/login", handlers.Login)
	auth.POST("/logout", handlers.Logout)
	auth.POST("/refresh_token", handlers.RefreshToken) // New refresh endpoint
	auth.GET("/status", handlers.CheckAuthStatus)

	// 2FA routes
	api.POST("/2fa/login", handlers.Login2FA) // Public for the second step of login
	protected2FA := api.Group("/2fa", echoAuthMiddleware)
	protected2FA.POST("/setup", handlers.Setup2FA)
	protected2FA.POST("/verify", handlers.Verify2FA)
	protected2FA.POST("/disable", handlers.Disable2FA)

	// CLI Device Code Authentication routes
	cli := api.Group("/cli")
	cli.GET("/authorize/status/:session_id", handlers.PollDeviceToken) // Fixed: should be GET with param

	// CLI Device Context Authentication routes (NEW FLOW)
	cliDeviceAuth := cli.Group("/device-auth")
	cliDeviceAuth.POST("/sessions", handlers.InitiateDeviceContextAuth)
	cliDeviceAuth.GET("/sessions/:session_id", handlers.GetDeviceContextSession)
	cliDeviceAuth.GET("/token/:session_id", handlers.PollDeviceContextToken)

	// CLI Project Configuration routes (public, no auth required)
	cli.POST("/configure/initiate", handlers.InitiateProjectConfig)
	cli.POST("/configure/submit", handlers.SubmitProjectConfig)
	cli.GET("/configure/status/:sessionId", handlers.GetProjectConfigStatus)

	// CLI Application Management routes (support both JWT and application token authentication)
	cli.POST("/apps/by-name/:appName/releases", handlers.UploadApplicationImage, echoAppTokenOrAuthMiddleware)
	cli.POST("/apps/by-name/:appName/deployments", handlers.CreateApplicationDeployment, echoAppTokenOrAuthMiddleware)
	cli.GET("/apps/by-name/:appName/config/export", handlers.ExportApplicationConfig, echoAppTokenOrAuthMiddleware)
	cli.GET("/deployments/:deployment_id", handlers.GetDeploymentResult, echoAppTokenOrAuthMiddleware)
	cli.GET("/deployments/:deployment_id/logs", handlers.DeploymentLogsSSEEnhanced, echoAppTokenOrAuthMiddleware)

	// Protected CLI routes (require authentication) - AuthorizeDeviceCode route removed

	// Protected CLI Device Context routes (NEW FLOW)
	protectedCliDeviceAuth := cliDeviceAuth.Group("", echoAuthMiddleware)
	protectedCliDeviceAuth.POST("/confirm", handlers.ConfirmDeviceContextAuth)

	// Protected routes group with authentication middleware
	protected := api.Group("", echoAuthMiddleware)

	protected.GET("/system/monitor", handlers.SystemMonitorHandler)

	protected.GET("/system/running-deployments", handlers.RunningDeploymentsHandler)

	// Upload progress WebSocket
	protected.GET("/docker-images/upload/progress", handlers.UploadProgressHandler)

	// Project and Application routes - 使用依赖注入的项目管理器
	if projectManager != nil {
		protected.POST("/projects", handlers.NewCreateProjectHandler(projectManager))
	} else {
		// Fallback for backward compatibility (when no dependency injection)
		protected.POST("/projects", func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusServiceUnavailable, "Project manager service not available - dependency injection required")
		})
	}

	protected.GET("/projects", handlers.ListProjectsHandler)
	protected.GET("/projects/:projectId", handlers.GetProjectHandler)
	protected.GET("/projects/by-name/:name", handlers.GetProjectByNameHandler)
	protected.POST("/projects/:projectId/apps", handlers.CreateApplicationHandler)
	protected.GET("/projects/:projectId/apps", handlers.ListApplicationsByProjectHandler)
	protected.GET("/projects/by-name/:name/apps", handlers.ListApplicationsByProjectNameHandler)
	protected.GET("/projects/by-name/:projectName/apps/by-name/:appName", handlers.GetApplicationByProjectNameAndAppNameHandler)
	protected.GET("/apps/:appId", handlers.GetApplicationHandler)
	protected.GET("/apps/by-name/:name", handlers.GetApplicationByNameHandler)
	protected.PUT("/apps/:appId", handlers.UpdateApplicationHandler)

	// Application token management routes
	protected.POST("/apps/:appId/tokens", handlers.CreateApplicationToken)
	protected.GET("/apps/:appId/tokens", handlers.ListApplicationTokens)
	protected.PUT("/apps/:appId/tokens/:tokenId", handlers.UpdateApplicationToken)
	protected.DELETE("/apps/:appId/tokens/:tokenId", handlers.DeleteApplicationToken)

	// ProviderAuth (第三方平台授权) management routes
	protected.POST("/provider-auths", handlers.CreateProviderAuthHandler)
	protected.GET("/provider-auths", handlers.ListProviderAuthsHandler)
	protected.GET("/provider-auths/:uid", handlers.GetProviderAuthHandler)
	protected.PUT("/provider-auths/:uid", handlers.UpdateProviderAuthHandler)
	protected.DELETE("/provider-auths/:uid", handlers.DeleteProviderAuthHandler)
	protected.POST("/provider-auths/:uid/activate", handlers.ActivateProviderAuthHandler)
	protected.POST("/provider-auths/:uid/deactivate", handlers.DeactivateProviderAuthHandler)

	// New route for fetching repositories for a provider auth
	protected.GET("/provider-auths/:uid/repositories", handlers.ListRepositoriesHandler)
	protected.GET("/provider-auths/:uid/repositories/branches", handlers.ListBranchesHandler)

	protected.GET("/providers/github/app-manifest", handlers.GenerateGitHubAppManifest)
	// Remove the protected webhook route, as it's now public
	// protected.POST("/providers/github/webhook", handlers.HandleGitHubWebhook)
	protected.POST("/provider-auths/:uid/github-install", handlers.InstallGitHubApp)

	// 使用工厂函数创建 Handler，并注入对应的服务实例
	// 这实现了原则三：服务单例，生命周期与程序相同
	if appService != nil {
		protected.DELETE("/apps/:appId", handlers.NewDeleteApplicationHandler(appService))
	} else {
		// Fallback for backward compatibility (when no dependency injection)
		protected.DELETE("/apps/:appId", func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusServiceUnavailable, "Service not available - dependency injection required")
		})
	}

	// Release routes
	protected.POST("/apps/:appId/releases", handlers.CreateReleaseAndBuildHandler)
	protected.GET("/apps/:appId/releases", handlers.ListReleasesHandler)
	protected.GET("/apps/:appId/releases/latest", handlers.GetLatestReleaseHandler)
	protected.GET("/releases/:releaseId", handlers.GetReleaseHandler)

	// Deployment routes
	if deploymentOrchestrator != nil {
		// 设置SSE日志发送函数
		deploymentOrchestrator.SetSSELogSender(handlers.SendDeploymentLogSSE)
		protected.POST("/apps/:appId/deployments", handlers.NewCreateDeploymentHandler(deploymentOrchestrator))
	} else {
		// Fallback for backward compatibility (when no dependency injection)
		protected.POST("/apps/:appId/deployments", func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusServiceUnavailable, "Service not available - dependency injection required")
		})
	}
	protected.GET("/apps/:appId/deployments", handlers.ListDeploymentsByAppHandler)
	protected.GET("/apps/:appId/deployments/running", handlers.ListRunningDeploymentsByAppHandler)
	protected.GET("/deployments/:deploymentId", handlers.GetDeploymentHandler)
	protected.GET("/deployments/:deploymentId/logs", handlers.DeploymentLogsSSEEnhanced)
	protected.GET("/deployments/:deploymentId/logs-data", handlers.GetDeploymentLogsHandler)

	// Environment Variable routes (simplified - directly associated with applications)
	protected.POST("/apps/:appId/environment-variables", handlers.CreateEnvironmentVariableHandler)
	protected.GET("/apps/:appId/environment-variables", handlers.ListEnvironmentVariablesHandler)
	protected.PUT("/environment-variables/:envVarId", handlers.UpdateEnvironmentVariableHandler)
	protected.DELETE("/environment-variables/:envVarId", handlers.DeleteEnvironmentVariableHandler)

	// Routing routes
	protected.POST("/apps/:appId/routings", handlers.CreateRoutingHandler)
	protected.GET("/apps/:appId/routings", handlers.ListRoutingsByAppHandler)
	protected.PUT("/routings/:routingId", handlers.UpdateRoutingHandler)
	protected.DELETE("/routings/:routingId", handlers.DeleteRoutingHandler)

	// Application operations routes
	protected.GET("/apps/:appId/status", handlers.GetAppRuntimeStatusHandler)
	protected.GET("/apps/:appId/logs", handlers.GetApplicationLogsHandler)
	protected.POST("/apps/:appId/actions/restart", handlers.RestartAppHandler)
	protected.POST("/apps/:appId/actions/override-deploy", handlers.OverrideDeployHandler)
	protected.GET("/projects/:projectId/branches", handlers.GetGitHubBranchesHandler)

	// GitHub token management routes
	protected.POST("/github-tokens", handlers.CreateGitHubToken)
	protected.GET("/github-tokens", handlers.ListGitHubTokens)
	protected.DELETE("/github-tokens/:uid", handlers.DeleteGitHubToken)
	protected.POST("/github-tokens/:uid/test", handlers.TestGitHubToken)

	// SSH Host management routes
	protected.GET("/ssh-hosts", handlers.ListSSHHosts)
	protected.POST("/ssh-hosts", handlers.CreateSSHHost)
	protected.GET("/ssh-hosts/:uid", handlers.GetSSHHost)
	protected.PUT("/ssh-hosts/:uid", handlers.UpdateSSHHost)
	protected.DELETE("/ssh-hosts/:uid", handlers.DeleteSSHHost)
	protected.POST("/ssh-hosts/:uid/test", handlers.TestSSHConnection)

	// Self-Hosted Database routes
	if databaseService != nil {
		protected.GET("/databases", handlers.NewListDatabasesHandler(databaseService))
		protected.POST("/databases", handlers.NewCreateDatabaseHandler(databaseService))
		protected.GET("/databases/:id", handlers.NewGetDatabaseHandler(databaseService))
		protected.PATCH("/databases/:id", handlers.NewUpdateDatabaseHandler(databaseService))
		protected.DELETE("/databases/:id", handlers.NewDeleteDatabaseHandler(databaseService))
		protected.GET("/databases/:id/connection-info", handlers.NewGetDatabaseConnectionInfoHandler(databaseService))
		if databaseOrchestrator != nil {
			protected.POST("/databases/:id/deploy", handlers.NewDeployDatabaseHandler(databaseOrchestrator))
			protected.POST("/databases/:id/start", handlers.NewStartDatabaseHandler(databaseOrchestrator))
			protected.POST("/databases/:id/stop", handlers.NewStopDatabaseHandler(databaseOrchestrator))
			protected.POST("/databases/:id/restart", handlers.NewRestartDatabaseHandler(databaseOrchestrator))
		} else {
			protected.POST("/databases/:id/deploy", func(c echo.Context) error {
				return echo.NewHTTPError(http.StatusServiceUnavailable, "Database orchestrator not available")
			})
			protected.POST("/databases/:id/start", func(c echo.Context) error {
				return echo.NewHTTPError(http.StatusServiceUnavailable, "Database orchestrator not available")
			})
			protected.POST("/databases/:id/stop", func(c echo.Context) error {
				return echo.NewHTTPError(http.StatusServiceUnavailable, "Database orchestrator not available")
			})
			protected.POST("/databases/:id/restart", func(c echo.Context) error {
				return echo.NewHTTPError(http.StatusServiceUnavailable, "Database orchestrator not available")
			})
		}
	} else {
		// Fallback for backward compatibility
		protected.GET("/databases", func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusServiceUnavailable, "Database service not available")
		})
		protected.POST("/databases", func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusServiceUnavailable, "Database service not available")
		})
		protected.GET("/databases/:id", func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusServiceUnavailable, "Database service not available")
		})
		protected.PATCH("/databases/:id", func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusServiceUnavailable, "Database service not available")
		})
		protected.DELETE("/databases/:id", func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusServiceUnavailable, "Database service not available")
		})
	}

	// SSH WebSocket connection route
	protected.GET("/ssh/connect", handlers.ConnectSSHHandler)

	// Multi-Node Deployment routes
	// protected.POST("/apps/:appId/multi-deployments", handlers.CreateMultiNodeDeployment)
	// protected.GET("/apps/:appId/multi-deployments", handlers.ListMultiNodeDeploymentsByApp)
	// protected.GET("/multi-deployments/:uid", handlers.GetMultiNodeDeployment)
	// protected.GET("/multi-deployments/:uid/nodes", handlers.GetNodeDeploymentsByMultiNode)
	// protected.PUT("/multi-deployments/:uid/status", handlers.UpdateMultiNodeDeploymentStatus)

	// // Node Deployment routes
	// protected.GET("/node-deployments/:uid", handlers.GetNodeDeployment)
	// protected.POST("/node-deployments/:uid/retry", handlers.RetryNodeDeployment)
	// protected.GET("/node-deployments/pending", handlers.GetPendingNodeDeployments)

	// Environment check and installation routes (public - needed during initial setup)
	api.GET("/environment/check", handlers.CheckDeploymentEnvironment)
	api.POST("/environment/install-podman", handlers.InstallPodman)
	api.POST("/environment/install-caddy", handlers.InstallCaddy)
	
	// Container environment check route (public - alias for /environment/check)
	api.POST("/containers/check-env", handlers.CheckDeploymentEnvironment)

	// System settings routes
	protected.GET("/system/settings/:key", handlers.GetSystemSettingHandler)
	protected.PUT("/system/settings/:key", handlers.UpdateSystemSettingHandler)

	return e
}

// customHTTPErrorHandler is the global error handler for Echo
func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := "Internal server error"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if msg, ok := he.Message.(string); ok {
			message = msg
		}
	}

	// Log the error
	log.Printf("HTTP Error: %d - %s - %v", code, c.Request().URL.Path, err)

	// Send standardized JSON error response
	if !c.Response().Committed {
		if c.Request().Header.Get("Content-Type") == "application/json" ||
			c.Request().Header.Get("Accept") == "application/json" {
			c.JSON(code, map[string]interface{}{
				"success": false,
				"message": message,
			})
		} else {
			c.String(code, message)
		}
	}
}

// echoAuthMiddleware converts the existing auth middleware to Echo format using JWT
func echoAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get access token from Authorization header
		authHeader := c.Request().Header.Get("Authorization")

		var token string

		// Try Authorization header first: "Bearer <token>" (case-insensitive for Bearer)
		if authHeader != "" {
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && strings.EqualFold(tokenParts[0], "Bearer") {
				token = tokenParts[1]
			}
		}

		// Fallback to query parameter for SSE/WebSocket and other cases
		if token == "" {
			token = c.QueryParam("access_token")
		}

		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}

		// Verify access token
		jwtService := services.GetJWTService()
		claims, err := jwtService.VerifyAccessToken(token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}

		// Add user info to context
		c.Set("userID", claims.UserID)
		c.Set("user_id", claims.UserID) // Also set with underscore for consistency
		c.Set("username", claims.Username)

		return next(c)
	}
}

// echoAppTokenOrAuthMiddleware supports both JWT and application token authentication
func echoAppTokenOrAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get token from Authorization header
		authHeader := c.Request().Header.Get("Authorization")

		var token string

		// Try Authorization header first: "Bearer <token>" (case-insensitive for Bearer)
		if authHeader != "" {
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && strings.EqualFold(tokenParts[0], "Bearer") {
				token = tokenParts[1]
			}
		}

		// Fallback to query parameter for SSE/WebSocket and other cases
		if token == "" {
			token = c.QueryParam("access_token")
		}

		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}

		// Try JWT token first
		jwtService := services.GetJWTService()
		claims, err := jwtService.VerifyAccessToken(token)
		if err == nil {
			// JWT token is valid
			c.Set("userID", claims.UserID)
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("auth_type", "jwt")
			return next(c)
		}

		// Try application token
		app, appToken, err := models.ValidateApplicationToken(token)
		if err == nil {
			// Application token is valid
			c.Set("applicationID", app.ID)
			c.Set("application", app)
			c.Set("appToken", appToken)
			c.Set("auth_type", "app_token")
			return next(c)
		}

		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}
}
