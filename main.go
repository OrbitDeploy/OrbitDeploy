package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/OrbitDeploy/OrbitDeploy/config"
	"github.com/OrbitDeploy/OrbitDeploy/http_service"
	"github.com/OrbitDeploy/OrbitDeploy/models"
	"github.com/OrbitDeploy/OrbitDeploy/services"
	"github.com/opentdp/go-helper/dborm"
)

//go:embed all:frontend/dist
var frontendAssets embed.FS

//go:embed script/install_last_podman.sh
var podmanInstallScript string

//go:embed script/install_caddy_github.sh
var caddyInstallScript string

func main() {
	// Parse command line flags
	cfg := config.Load()

	// Initialize logger
	log.Printf("Starting Echo server on %s with database %s", cfg.ServerAddr, cfg.DBPath)

	// Initialize database
	if err := initDB(cfg.DBPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dborm.Destroy()

	db := dborm.Db

	projectManager, err := services.NewManager("/var/lib/orbitdeploy/projects")
	if err != nil {
		log.Fatalf("Failed to create project manager: %v", err)
	}

	podmanService := services.NewPodmanService()

	databaseService := services.NewDatabaseService(db)

	databaseOrchestrator := services.NewDatabaseOrchestrator(databaseService, podmanService)

	appService := services.NewApplicationService(db, podmanService)

	buildService := services.NewBuildService()
	envService := services.NewDeploymentEnvironmentService()
	deploymentOrchestrator := services.NewDeploymentOrchestrator(buildService, envService, podmanService)

	http_service.SetInstallationScripts(
		func() string { return podmanInstallScript },
		func() string { return caddyInstallScript },
	)

	// Create Echo server (replaces Gin setup)
	e := http_service.NewEchoServerWithDependencies(frontendAssets, appService, deploymentOrchestrator, projectManager, databaseService, databaseOrchestrator)

	// 为处理器设置 Docker 构建队列服务 暂时不用。注释掉

	// Initialize docker build queue service
	// queueSvc := services.NewDockerBuildQueueService(3, 2*time.Second)
	// if err := queueSvc.InitDB(); err != nil {
	// 	log.Fatalf("Failed to initialize docker build queue: %v", err)
	// }
	// if err := queueSvc.RecoverTasks(); err != nil {
	// 	log.Fatalf("Failed to recover docker build tasks: %v", err)
	// }
	// http_service.SetDockerBuildQueueService(queueSvc)

	// Create context for graceful shutdown
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// var wg sync.WaitGroup
	// queueSvc.StartWorkers(ctx, &wg)

	// Start deployment controller
	// deploymentController := services.NewDeploymentController(ctx)
	// go deploymentController.Start()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start Echo server in a goroutine
	go func() {
		log.Printf("Echo server starting on %s", cfg.ServerAddr)
		if err := e.Start(cfg.ServerAddr); err != nil {
			log.Printf("Echo server stopped: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-quit
	log.Println("Shutting down server...")

	// Stop deployment controller
	// deploymentController.Stop()

	// Shutdown Echo server
	// if err := e.Shutdown(ctx); err != nil {
	// 	log.Printf("Echo server forced to shutdown: %v", err)
	// }

	log.Println("Waiting for docker build workers to finish...")
	// wg.Wait()

	log.Println("Server exited")
}

func initDB(dbPath string) error {
	// Get absolute path for database
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Connect to database
	config := &dborm.Config{
		Type:   "sqlite",
		DbName: absPath,
	}
	if dborm.Connect(config) == nil {
		return fmt.Errorf("failed to connect to database")
	}

	// Auto migrate models
	modelsToMigrate := []interface{}{
		&models.User{},
		&models.TwoFactorRecoveryCode{},
		&models.AuthToken{},
		&models.CLIDeviceCode{},

		&models.Project{},
		&models.Application{},
		&models.ProviderAuth{},
		&models.EnvironmentVariable{},
		&models.Deployment{},
		&models.DeploymentLog{},
		&models.Release{},
		&models.Routing{},

		&models.GitHubToken{},
		&models.ProjectCredential{},
		&models.DockerBuildTask{},

		&models.SSHHost{},
		&models.ApplicationToken{},
		&models.SelfHostedDatabase{},
		&models.SystemSetting{},
	}
	if err := dborm.Db.AutoMigrate(modelsToMigrate...); err != nil {
		return fmt.Errorf("failed to auto migrate models: %w", err)
	}

	return nil
}
