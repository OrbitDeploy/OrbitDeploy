package handlers

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/youfun/OrbitDeploy/models"
	"github.com/youfun/OrbitDeploy/services"
)

// toDatabaseResponse converts a SelfHostedDatabase model to a DatabaseResponse, omitting sensitive data.
func toDatabaseResponse(db *models.SelfHostedDatabase) DatabaseResponse {
	var sshHostUid *string
	if db.SSHHostID != nil {
		uid := EncodeFriendlyID(PrefixSSHHost, *db.SSHHostID)
		sshHostUid = &uid
	}

	return DatabaseResponse{
		Uid:          EncodeFriendlyID(PrefixDatabase, db.ID),
		Name:         db.Name,
		Type:         db.Type,
		Version:      db.Version,
		CustomImage:  db.CustomImage,
		Status:       db.Status,
		Port:         db.Port,
		InternalPort: db.InternalPort,
		Username:     db.Username,
		DatabaseName: db.DatabaseName,
		DataPath:     db.DataPath,
		ConfigPath:   db.ConfigPath,
		IsRemote:     db.IsRemote,
		SSHHostUid:   sshHostUid,
		ExtraConfig:  db.ExtraConfig,
		LastCheckAt:  db.LastCheckAt,
		CreatedAt:    db.CreatedAt,
		UpdatedAt:    db.UpdatedAt,
	}
}

// NewListDatabasesHandler creates a handler for listing databases.
func NewListDatabasesHandler(dbService *services.DatabaseService) echo.HandlerFunc {
	return func(c echo.Context) error {
		databases, err := dbService.GetAllDatabases()
		if err != nil {
			return SendError(c, http.StatusInternalServerError, "Failed to retrieve databases")
		}

		responses := make([]DatabaseResponse, len(databases))
		for i, db := range databases {
			responses[i] = toDatabaseResponse(&db)
		}

		return SendSuccess(c, responses)
	}
}

// NewCreateDatabaseHandler creates a handler for creating a database.
func NewCreateDatabaseHandler(dbService *services.DatabaseService) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req CreateDatabaseRequest
		if err := c.Bind(&req); err != nil {
			return SendError(c, http.StatusBadRequest, "Invalid request body")
		}

		// TODO: Add validation for the request

		var sshHostID *uuid.UUID
		if req.SSHHostUid != nil {
			id, err := DecodeFriendlyID(PrefixSSHHost, *req.SSHHostUid)
			if err != nil {
				return SendError(c, http.StatusBadRequest, "Invalid SSH Host UID")
			}
			sshHostID = &id
		}

		dbModel := &models.SelfHostedDatabase{
			Name:         req.Name,
			Type:         req.Type,
			Version:      req.Version,
			CustomImage:  req.CustomImage,
			Port:         req.Port,
			InternalPort: req.InternalPort,
			Username:     req.Username,
			Password:     req.Password,
			DatabaseName: req.DatabaseName,
			DataPath:     req.DataPath,
			ConfigPath:   req.ConfigPath,
			IsRemote:     req.IsRemote,
			SSHHostID:    sshHostID,
			ExtraConfig:  req.ExtraConfig,
		}

		database, err := dbService.CreateDatabase(dbModel)
		if err != nil {
			return SendError(c, http.StatusInternalServerError, "Failed to create database")
		}

		return SendCreated(c, toDatabaseResponse(database))
	}
}

// NewGetDatabaseHandler creates a handler for getting a single database.
func NewGetDatabaseHandler(dbService *services.DatabaseService) echo.HandlerFunc {
	return func(c echo.Context) error {
		dbUID := c.Param("id")
		dbID, err := DecodeFriendlyID(PrefixDatabase, dbUID)
		if err != nil {
			return SendError(c, http.StatusBadRequest, "Invalid database UID")
		}

		database, err := dbService.GetDatabaseByID(dbID)
		if err != nil {
			return SendError(c, http.StatusNotFound, "Database not found")
		}

		return SendSuccess(c, toDatabaseResponse(database))
	}
}

// NewUpdateDatabaseHandler creates a handler for updating a database.
func NewUpdateDatabaseHandler(dbService *services.DatabaseService) echo.HandlerFunc {
	return func(c echo.Context) error {
		dbUID := c.Param("id")
		dbID, err := DecodeFriendlyID(PrefixDatabase, dbUID)
		if err != nil {
			return SendError(c, http.StatusBadRequest, "Invalid database UID")
		}

		var req UpdateDatabaseRequest
		if err := c.Bind(&req); err != nil {
			return SendError(c, http.StatusBadRequest, "Invalid request body")
		}

		database, err := dbService.GetDatabaseByID(dbID)
		if err != nil {
			return SendError(c, http.StatusNotFound, "Database not found")
		}

		// Update fields if provided
		if req.Port != nil {
			database.Port = *req.Port
		}
		if req.Username != nil {
			database.Username = *req.Username
		}
		if req.Password != nil {
			database.Password = *req.Password
		}
		if req.DataPath != nil {
			database.DataPath = *req.DataPath
		}
		if req.ConfigPath != nil {
			database.ConfigPath = *req.ConfigPath
		}
		if req.ExtraConfig != nil {
			database.ExtraConfig = *req.ExtraConfig
		}

		if err := dbService.UpdateDatabase(database); err != nil {
			return SendError(c, http.StatusInternalServerError, "Failed to update database")
		}

		return SendSuccess(c, toDatabaseResponse(database))
	}
}

// NewDeleteDatabaseHandler creates a handler for deleting a database.
func NewDeleteDatabaseHandler(dbService *services.DatabaseService) echo.HandlerFunc {
	return func(c echo.Context) error {
		dbUID := c.Param("id")
		dbID, err := DecodeFriendlyID(PrefixDatabase, dbUID)
		if err != nil {
			return SendError(c, http.StatusBadRequest, "Invalid database UID")
		}

		if err := dbService.DeleteDatabase(dbID); err != nil {
			return SendError(c, http.StatusInternalServerError, "Failed to delete database")
		}

		return SendSuccess(c, nil)
	}
}

// NewDeployDatabaseHandler creates a handler for deploying a database.
func NewDeployDatabaseHandler(orchestrator *services.DatabaseOrchestrator) echo.HandlerFunc {
	return func(c echo.Context) error {
		dbUID := c.Param("id")
		dbID, err := DecodeFriendlyID(PrefixDatabase, dbUID)
		if err != nil {
			return SendError(c, http.StatusBadRequest, "Invalid database UID")
		}

		go func() {
			if err := orchestrator.DeployDatabase(dbID); err != nil {
				// Log the error, as we can't return an HTTP error in a goroutine
				// In a real app, you'd use a more robust logging/error handling mechanism
				println("Failed to deploy database asynchronously:", err.Error())
			}
		}()

		return SendSuccess(c, map[string]string{"message": "Database deployment started"})
	}
}

// NewStartDatabaseHandler creates a handler for starting a database service.
func NewStartDatabaseHandler(orchestrator *services.DatabaseOrchestrator) echo.HandlerFunc {
	return func(c echo.Context) error {
		dbUID := c.Param("id")
		dbID, err := DecodeFriendlyID(PrefixDatabase, dbUID)
		if err != nil {
			return SendError(c, http.StatusBadRequest, "Invalid database UID")
		}

		if err := orchestrator.StartDatabase(dbID); err != nil {
			return SendError(c, http.StatusInternalServerError, "Failed to start database")
		}

		return SendSuccess(c, map[string]string{"message": "Database started successfully"})
	}
}

// NewStopDatabaseHandler creates a handler for stopping a database service.
func NewStopDatabaseHandler(orchestrator *services.DatabaseOrchestrator) echo.HandlerFunc {
	return func(c echo.Context) error {
		dbUID := c.Param("id")
		dbID, err := DecodeFriendlyID(PrefixDatabase, dbUID)
		if err != nil {
			return SendError(c, http.StatusBadRequest, "Invalid database UID")
		}

		if err := orchestrator.StopDatabase(dbID); err != nil {
			return SendError(c, http.StatusInternalServerError, "Failed to stop database")
		}

		return SendSuccess(c, map[string]string{"message": "Database stopped successfully"})
	}
}

// NewRestartDatabaseHandler creates a handler for restarting a database service.
func NewRestartDatabaseHandler(orchestrator *services.DatabaseOrchestrator) echo.HandlerFunc {
	return func(c echo.Context) error {
		dbUID := c.Param("id")
		dbID, err := DecodeFriendlyID(PrefixDatabase, dbUID)
		if err != nil {
			return SendError(c, http.StatusBadRequest, "Invalid database UID")
		}

		if err := orchestrator.RestartDatabase(dbID); err != nil {
			return SendError(c, http.StatusInternalServerError, "Failed to restart database")
		}

		return SendSuccess(c, map[string]string{"message": "Database restarted successfully"})
	}
}

// NewGetDatabaseConnectionInfoHandler creates a handler for getting database connection information.
func NewGetDatabaseConnectionInfoHandler(dbService *services.DatabaseService) echo.HandlerFunc {
	return func(c echo.Context) error {
		dbUID := c.Param("id")
		dbID, err := DecodeFriendlyID(PrefixDatabase, dbUID)
		if err != nil {
			return SendError(c, http.StatusBadRequest, "Invalid database UID")
		}

		database, err := dbService.GetDatabaseByID(dbID)
		if err != nil {
			return SendError(c, http.StatusNotFound, "Database not found")
		}

		unmask := c.QueryParam("unmask") == "true"

		connInfo := DatabaseConnectionInfoResponse{
			Host:     "localhost", // For local databases
			Port:     database.Port,
			User:     database.Username,
			Database: database.DatabaseName,
		}

		// Only include password if explicitly requested
		if unmask {
			connInfo.Password = database.Password
		}

		// Generate connection string based on database type
		if database.Type == models.PostgreSQL {
			if unmask {
				connInfo.ConnectionString = fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
					database.Username, database.Password, connInfo.Host, database.Port, database.DatabaseName)
			} else {
				connInfo.ConnectionString = fmt.Sprintf("postgresql://%s:****@%s:%d/%s",
					database.Username, connInfo.Host, database.Port, database.DatabaseName)
			}
		}

		return SendSuccess(c, connInfo)
	}
}
