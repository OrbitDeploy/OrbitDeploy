package services

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/command"
	"github.com/opentdp/go-helper/logman"
	"github.com/youfun/OrbitDeploy/models"
)

// DatabaseOrchestrator handles the deployment logic for self-hosted databases.
type DatabaseOrchestrator struct {
	dbService     *DatabaseService
	podmanService *PodmanService
	// sseLogSender can be added later for real-time logging
}

// NewDatabaseOrchestrator creates a new instance of DatabaseOrchestrator.
func NewDatabaseOrchestrator(dbService *DatabaseService, podmanService *PodmanService) *DatabaseOrchestrator {
	return &DatabaseOrchestrator{
		dbService:     dbService,
		podmanService: podmanService,
	}
}

// DeployDatabase orchestrates the deployment of a database.
func (o *DatabaseOrchestrator) DeployDatabase(dbID uuid.UUID) error {
	db, err := o.dbService.GetDatabaseByID(dbID)
	if err != nil {
		return fmt.Errorf("failed to get database details: %w", err)
	}

	logman.Info("Starting deployment for database", "db_name", db.Name, "db_id", db.ID)

	if db.IsRemote {
		return fmt.Errorf("remote deployment is not yet supported")
	}

	// 1. Generate environment file content
	envContent := o.generateEnvContent(db)

	// 2. Generate Quadlet .container file content
	quadletContent, err := o.generateQuadletContent(db)
	if err != nil {
		return fmt.Errorf("failed to generate quadlet content: %w", err)
	}

	// 3. Write files to the system
	if err := o.writeSystemFiles(db, quadletContent, envContent); err != nil {
		return fmt.Errorf("failed to write system files: %w", err)
	}

	// 4. Reload systemd and start the service
	if err := o.manageSystemdService(db); err != nil {
		models.UpdateDatabaseStatus(db.ID, models.DatabaseStatusFailed)
		return fmt.Errorf("failed to manage systemd service: %w", err)
	}

	// 5. Update database status
	if err := models.UpdateDatabaseStatus(db.ID, models.DatabaseStatusRunning); err != nil {
		logman.Error("Failed to update database status to running", "db_id", db.ID, "error", err)
		// Continue even if status update fails, as the service might be running
	}

	logman.Info("Database deployment completed successfully", "db_name", db.Name, "db_id", db.ID)
	return nil
}

func (o *DatabaseOrchestrator) generateEnvContent(db *models.SelfHostedDatabase) string {
	return fmt.Sprintf(
		"POSTGRES_USER=%s\nPOSTGRES_PASSWORD=%s\nPOSTGRES_DB=%s\n",
		db.Username, db.Password, db.DatabaseName,
	)
}

func (o *DatabaseOrchestrator) generateQuadletContent(db *models.SelfHostedDatabase) (string, error) {
	// This is a simplified version. A more robust implementation would use templates.
	var imageName string

	// Use custom image if provided, otherwise construct default image name
	if db.CustomImage != "" {
		imageName = db.CustomImage
	} else {
		// 默认使用基于 Alpine 的镜像以减小体积 (Default to using Alpine-based images to reduce size)
		switch db.Type {
		case models.PostgreSQL:
			// 直接使用提供的版本（例如 "16-alpine"、"latest"） (Directly use the provided version e.g., "16-alpine", "latest")
			if db.Version != "" {
				imageName = fmt.Sprintf("docker.io/library/postgres:%s", db.Version)
			} else {
				imageName = "docker.io/library/postgres:alpine"
			}
		default:
			// 对于其他数据库类型，使用标准版本 (For other database types, use the standard version)
			// Assuming a default behavior, adjust if necessary.
			imageName = fmt.Sprintf("docker.io/library/postgres:%s", db.Version)
		}
	}
	// The extra closing brace '}' was here, which has been removed.

	quadlet := fmt.Sprintf(`[Unit]
Description=PostgreSQL database: %s

[Container]
Image=%s
PublishPort=%d:%d
Volume=%s:/var/lib/postgresql/data
`, db.Name, imageName, db.Port, db.InternalPort, db.DataPath)

	envFileName := fmt.Sprintf("db-%s.env", db.ID.String())
	var envFilePath string
	if os.Geteuid() == 0 {
		envFilePath = filepath.Join("/etc/orbit-deploy/db-envs", envFileName)
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		envFilePath = filepath.Join(homeDir, ".config", "orbit-deploy", "db-envs", envFileName)
	}
	quadlet += fmt.Sprintf("EnvironmentFile=%s\n", envFilePath)

	quadlet += `
[Install]
WantedBy=default.target
`
	return quadlet, nil
}

func (o *DatabaseOrchestrator) writeSystemFiles(db *models.SelfHostedDatabase, quadletContent, envContent string) error {
	var systemdPath, envPath string
	if os.Geteuid() == 0 {
		systemdPath = "/etc/containers/systemd"
		envPath = "/etc/orbit-deploy/db-envs"
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		systemdPath = filepath.Join(homeDir, ".config", "containers", "systemd")
		envPath = filepath.Join(homeDir, ".config", "orbit-deploy", "db-envs")
	}

	// Ensure the database data path exists and has correct permissions
	if err := os.MkdirAll(db.DataPath, 0700); err != nil {
		return fmt.Errorf("failed to create database data directory %s: %w", db.DataPath, err)
	}
	if err := os.Chown(db.DataPath, 999, 999); err != nil {
		return fmt.Errorf("failed to change ownership of database data directory %s: %w", db.DataPath, err)
	}

	// Write Quadlet file
	if err := os.MkdirAll(systemdPath, 0755); err != nil {
		return err
	}
	containerFileName := fmt.Sprintf("db-%s.container", db.ID.String())
	if err := os.WriteFile(filepath.Join(systemdPath, containerFileName), []byte(quadletContent), 0644); err != nil {
		return err
	}

	// Write environment file
	if err := os.MkdirAll(envPath, 0755); err != nil {
		return err
	}
	envFileName := fmt.Sprintf("db-%s.env", db.ID.String())
	if err := os.WriteFile(filepath.Join(envPath, envFileName), []byte(envContent), 0644); err != nil {
		return err
	}

	return nil
}

func (o *DatabaseOrchestrator) manageSystemdService(db *models.SelfHostedDatabase) error {
	var cmdPrefix string
	if os.Geteuid() == 0 {
		cmdPrefix = "systemctl"
	} else {
		cmdPrefix = "systemctl --user"
	}

	logman.Info("Reloading systemd daemon")
	reloadCmd := fmt.Sprintf("%s daemon-reload", cmdPrefix)
	logman.Info("Executing command", "command", reloadCmd)
	output, err := command.Exec(&command.ExecPayload{Content: reloadCmd, CommandType: "SHELL", Timeout: 30})
	if err != nil {
		logman.Error("Failed to reload systemd daemon", "error", err, "output", output)
		return fmt.Errorf("failed to reload systemd daemon: %w", err)
	}
	logman.Info("Systemd daemon reloaded successfully", "output", output)

	serviceName := fmt.Sprintf("db-%s.service", db.ID.String())
	logman.Info("Starting systemd service", "service", serviceName)
	startCmd := fmt.Sprintf("%s start %s", cmdPrefix, serviceName)
	logman.Info("Executing command", "command", startCmd)
	output, err = command.Exec(&command.ExecPayload{Content: startCmd, CommandType: "SHELL", Timeout: 60})
	if err != nil {
		logman.Error("Failed to start service", "service", serviceName, "error", err, "output", output)
		return fmt.Errorf("failed to start service %s: %w", serviceName, err)
	}
	logman.Info("Service start command executed successfully", "service", serviceName, "output", output)

	logman.Info("Waiting for service to stabilize", "service", serviceName)
	time.Sleep(10 * time.Second) // Wait for the database to initialize

	logman.Info("Checking service health", "service", serviceName)
	healthCheckCmd := fmt.Sprintf("%s is-active %s", cmdPrefix, serviceName)
	logman.Info("Executing command", "command", healthCheckCmd)
	output, err = command.Exec(&command.ExecPayload{Content: healthCheckCmd, CommandType: "SHELL", Timeout: 30})
	if err != nil {
		logman.Error("Service health check failed", "service", serviceName, "error", err, "output", output)
		return fmt.Errorf("service %s is not active: %w", serviceName, err)
	}

	logman.Info("Service is active", "service", serviceName, "output", output)
	return nil
}

// StartDatabase starts a database service.
func (o *DatabaseOrchestrator) StartDatabase(dbID uuid.UUID) error {
	db, err := o.dbService.GetDatabaseByID(dbID)
	if err != nil {
		return fmt.Errorf("failed to get database details: %w", err)
	}

	var cmdPrefix string
	if os.Geteuid() == 0 {
		cmdPrefix = "systemctl"
	} else {
		cmdPrefix = "systemctl --user"
	}

	serviceName := fmt.Sprintf("db-%s.service", db.ID.String())
	logman.Info("Starting systemd service", "service", serviceName)
	_, err = command.Exec(&command.ExecPayload{Content: fmt.Sprintf("%s start %s", cmdPrefix, serviceName), CommandType: "SHELL", Timeout: 60})
	if err != nil {
		return fmt.Errorf("failed to start service %s: %w", serviceName, err)
	}

	if err := models.UpdateDatabaseStatus(db.ID, models.DatabaseStatusRunning); err != nil {
		logman.Error("Failed to update database status", "db_id", db.ID, "error", err)
	}

	return nil
}

// StopDatabase stops a database service.
func (o *DatabaseOrchestrator) StopDatabase(dbID uuid.UUID) error {
	db, err := o.dbService.GetDatabaseByID(dbID)
	if err != nil {
		return fmt.Errorf("failed to get database details: %w", err)
	}

	var cmdPrefix string
	if os.Geteuid() == 0 {
		cmdPrefix = "systemctl"
	} else {
		cmdPrefix = "systemctl --user"
	}

	serviceName := fmt.Sprintf("db-%s.service", db.ID.String())
	logman.Info("Stopping systemd service", "service", serviceName)
	_, err = command.Exec(&command.ExecPayload{Content: fmt.Sprintf("%s stop %s", cmdPrefix, serviceName), CommandType: "SHELL", Timeout: 60})
	if err != nil {
		return fmt.Errorf("failed to stop service %s: %w", serviceName, err)
	}

	if err := models.UpdateDatabaseStatus(db.ID, models.DatabaseStatusStopped); err != nil {
		logman.Error("Failed to update database status", "db_id", db.ID, "error", err)
	}

	return nil
}

// RestartDatabase restarts a database service.
func (o *DatabaseOrchestrator) RestartDatabase(dbID uuid.UUID) error {
	db, err := o.dbService.GetDatabaseByID(dbID)
	if err != nil {
		return fmt.Errorf("failed to get database details: %w", err)
	}

	var cmdPrefix string
	if os.Geteuid() == 0 {
		cmdPrefix = "systemctl"
	} else {
		cmdPrefix = "systemctl --user"
	}

	serviceName := fmt.Sprintf("db-%s.service", db.ID.String())
	logman.Info("Restarting systemd service", "service", serviceName)
	_, err = command.Exec(&command.ExecPayload{Content: fmt.Sprintf("%s restart %s", cmdPrefix, serviceName), CommandType: "SHELL", Timeout: 60})
	if err != nil {
		return fmt.Errorf("failed to restart service %s: %w", serviceName, err)
	}

	if err := models.UpdateDatabaseStatus(db.ID, models.DatabaseStatusRunning); err != nil {
		logman.Error("Failed to update database status", "db_id", db.ID, "error", err)
	}

	return nil
}
