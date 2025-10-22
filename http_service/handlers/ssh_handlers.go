package handlers

import (
	"fmt"
	"log"
	"net/http"

	"strings"
	"time"

	"github.com/OrbitDeploy/OrbitDeploy/models"
	"github.com/labstack/echo/v4"
	"github.com/opentdp/go-helper/logman"
	"github.com/opentdp/go-helper/webssh"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/websocket"
)

// SSH Host Management Handlers - Echo compatible

// convertSSHHostToResponse converts database model to API response model
func convertSSHHostToResponse(host *models.SSHHost) SSHHostResponse {
	return SSHHostResponse{
		Uid:         EncodeFriendlyID(PrefixSSHHost, host.ID),
		Name:        host.Name,
		Addr:        host.Addr,
		Port:        host.Port,
		User:        host.User,
		Description: host.Description,
		Status:      host.Status,
		Region:      host.Region,
		CPUCores:    host.CPUCores,
		MemoryGB:    host.MemoryGB,
		DiskGB:      host.DiskGB,
		IsActive:    host.IsActive,
		CreatedAt:   host.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   host.UpdatedAt.Format(time.RFC3339),
	}
}

// convertSSHHostsToResponse converts slice of database models to API response models
func convertSSHHostsToResponse(hosts []*models.SSHHost) []SSHHostResponse {
	responses := make([]SSHHostResponse, len(hosts))
	for i, host := range hosts {
		responses[i] = convertSSHHostToResponse(host)
	}
	return responses
}

// ListSSHHosts returns all SSH host configurations
func ListSSHHosts(c echo.Context) error {
	hosts, err := models.GetAllSSHHosts()
	if err != nil {
		log.Printf("Failed to get SSH hosts: %v", err)
		return SendError(c, http.StatusInternalServerError, "Failed to get SSH hosts")
	}

	return SendSuccess(c, convertSSHHostsToResponse(hosts))
}

// CreateSSHHost creates a new SSH host configuration
func CreateSSHHost(c echo.Context) error {
	var req SSHHostRequest
	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid JSON payload")
	}

	// Validate required fields
	if req.Name == "" || req.Addr == "" || req.User == "" {
		return SendError(c, http.StatusBadRequest, "Name, address, and user are required")
	}

	// Validate that either password or private key is provided
	if req.Password == "" && req.PrivateKey == "" {
		return SendError(c, http.StatusBadRequest, "Either password or private key must be provided")
	}

	host, err := models.CreateSSHHost(
		req.Name,
		req.Addr,
		req.User,
		req.Password,
		req.PrivateKey,
		req.Description,
		req.Port,
	)
	if err != nil {
		log.Printf("Failed to create SSH host: %v", err)
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return SendError(c, http.StatusBadRequest, "SSH host name already exists")
		}
		return SendError(c, http.StatusInternalServerError, "Failed to create SSH host")
	}

	return SendCreated(c, convertSSHHostToResponse(host))
}

// GetSSHHost retrieves a specific SSH host by ID
func GetSSHHost(c echo.Context) error {
	idStr := c.Param("uid")
	hostID, err := DecodeFriendlyID(PrefixSSHHost, idStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid SSH host ID")
	}

	host, err := models.GetSSHHostByID(hostID)
	if err != nil {
		log.Printf("Failed to get SSH host: %v", err)
		return SendError(c, http.StatusNotFound, "SSH host not found")
	}

	return SendSuccess(c, convertSSHHostToResponse(host))
}

// UpdateSSHHost updates an existing SSH host configuration
func UpdateSSHHost(c echo.Context) error {
	idStr := c.Param("uid")
	hostID, err := DecodeFriendlyID(PrefixSSHHost, idStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid SSH host ID")
	}

	var req SSHHostRequest
	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid JSON payload")
	}

	// Validate required fields
	if req.Name == "" || req.Addr == "" || req.User == "" {
		return SendError(c, http.StatusBadRequest, "Name, address, and user are required")
	}

	// Validate that either password or private key is provided
	if req.Password == "" && req.PrivateKey == "" {
		return SendError(c, http.StatusBadRequest, "Either password or private key must be provided")
	}

	host, err := models.UpdateSSHHost(
		hostID,
		req.Name,
		req.Addr,
		req.User,
		req.Password,
		req.PrivateKey,
		req.Description,
		req.Port,
	)
	if err != nil {
		log.Printf("Failed to update SSH host: %v", err)
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return SendError(c, http.StatusBadRequest, "SSH host name already exists")
		}
		return SendError(c, http.StatusInternalServerError, "Failed to update SSH host")
	}

	return SendSuccess(c, convertSSHHostToResponse(host))
}

// DeleteSSHHost deletes an SSH host configuration
func DeleteSSHHost(c echo.Context) error {
	idStr := c.Param("uid")
	hostID, err := DecodeFriendlyID(PrefixSSHHost, idStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid SSH host ID")
	}

	err = models.DeleteSSHHost(hostID)
	if err != nil {
		log.Printf("Failed to delete SSH host: %v", err)
		return SendError(c, http.StatusInternalServerError, "Failed to delete SSH host")
	}

	return SendSuccess(c, map[string]string{"message": "SSH host deleted successfully"})
}

// TestSSHConnection tests SSH connection to a host
func TestSSHConnection(c echo.Context) error {
	idStr := c.Param("uid")
	hostID, err := DecodeFriendlyID(PrefixSSHHost, idStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid SSH host ID")
	}

	host, err := models.GetSSHHostByID(hostID)
	if err != nil {
		return SendError(c, http.StatusNotFound, "SSH host not found")
	}

	// Test SSH connection using go-helper library
	opt := createSSHClientOption(*host)
	client, err := webssh.NewSSHClient(opt)
	if err != nil {
		return SendSuccess(c, map[string]interface{}{
			"connected": false,
			"message":   fmt.Sprintf("SSH connection failed: %v", err),
		})
	}
	defer client.Close()

	// Test with a simple command
	session, err := client.NewSession()
	if err != nil {
		return SendSuccess(c, map[string]interface{}{
			"connected": false,
			"message":   fmt.Sprintf("SSH session failed: %v", err),
		})
	}
	defer session.Close()

	// Execute a simple command to verify connection
	_, err = session.CombinedOutput("echo 'connection test'")
	if err != nil {
		return SendSuccess(c, map[string]interface{}{
			"connected": false,
			"message":   fmt.Sprintf("SSH command execution failed: %v", err),
		})
	}

	return SendSuccess(c, map[string]interface{}{
		"connected": true,
		"message":   "SSH connection successful",
	})
}

// SSH Client Utilities

// createSSHClientOption creates SSH client options for the given host
func createSSHClientOption(host models.SSHHost) *webssh.SSHClientOption {
	addr := host.Addr
	if host.Port != 22 {
		addr = fmt.Sprintf("%s:%d", host.Addr, host.Port)
	} else if !strings.Contains(addr, ":") {
		addr = addr + ":22"
	}

	return &webssh.SSHClientOption{
		Addr:       addr,
		User:       host.User,
		Password:   host.Password,
		PrivateKey: host.PrivateKey,
	}
}

// createSSHClient creates an SSH client for the given host using webssh library
func createSSHClient(host models.SSHHost) (*ssh.Client, error) {
	opt := createSSHClientOption(host)
	return webssh.NewSSHClient(opt)
}

// ExecuteSSHCommand executes a command on remote host via SSH using webssh library
func ExecuteSSHCommand(host models.SSHHost, cmd string) (string, error) {
	// Create SSH client using webssh library
	client, err := createSSHClient(host)
	if err != nil {
		return "", fmt.Errorf("failed to create SSH client: %v", err)
	}
	defer client.Close()

	// Create session
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	// Execute command
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", fmt.Errorf("command execution failed: %v", err)
	}

	return string(output), nil
}

func ListRemoteContainers(c echo.Context) error {
	idStr := c.Param("uid")
	hostID, err := DecodeFriendlyID(PrefixSSHHost, idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid SSH host ID")
	}

	// Get SSH host
	host, err := models.GetSSHHostByID(hostID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "SSH host not found")
	}

	// For now, return mock data. In actual implementation, this would:
	// - Connect to remote host via SSH
	// - Query containers using podman/docker commands
	// - Return real container information
	mockContainers := []map[string]interface{}{
		{
			"name":   "nginx-web",
			"status": "running",
			"image":  "nginx:latest",
			"ports":  []string{"80:8080"},
		},
		{
			"name":   "postgres-db",
			"status": "running",
			"image":  "postgres:15",
			"ports":  []string{"5432:5432"},
		},
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"host_uid":   EncodeFriendlyID(PrefixSSHHost, host.ID),
			"host_name":  host.Name,
			"containers": mockContainers,
		},
	})
}

// ControlRemoteContainer controls a container on a remote SSH host
func ControlRemoteContainer(c echo.Context) error {
	idStr := c.Param("uid")
	hostID, err := DecodeFriendlyID(PrefixSSHHost, idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid SSH host ID")
	}

	containerName := c.Param("name")
	if containerName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Container name is required")
	}

	// Get SSH host
	host, err := models.GetSSHHostByID(hostID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "SSH host not found")
	}

	// Determine action from URL path
	action := "unknown"
	if c.Request().URL.Path[len(c.Request().URL.Path)-5:] == "start" {
		action = "start"
	} else if c.Request().URL.Path[len(c.Request().URL.Path)-4:] == "stop" {
		action = "stop"
	} else if c.Request().URL.Path[len(c.Request().URL.Path)-7:] == "restart" {
		action = "restart"
	}

	// For now, return success. In actual implementation, this would:
	// - Connect to remote host via SSH
	// - Execute container control commands
	// - Return actual operation status
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Remote container " + action + " initiated",
		"data": map[string]interface{}{
			"host_uid":       EncodeFriendlyID(PrefixSSHHost, host.ID),
			"container_name": containerName,
			"action":         action,
			"status":         "executing",
		},
	})
}

// EchoGetRemoteContainerLogs gets logs from a container on a remote SSH host

func ConnectSSHHandler(c echo.Context) error {
	logman.Info("收到 SSH 连接请求")
	websocket.Handler(ConnectSSH).ServeHTTP(c.Response().Writer, c.Request())
	return nil
}

// ConnectSSH handles WebSocket SSH connections using webssh library
func ConnectSSH(ws *websocket.Conn) {
	defer ws.Close()

	// Read SSH host ID from the query parameter
	hostIDStr := ws.Request().URL.Query().Get("host_id")
	if hostIDStr == "" {
		ws.Write([]byte("Error: host_id parameter is required\r\n"))
		return
	}

	hostUUID, err := DecodeFriendlyID(PrefixSSHHost, hostIDStr)
	if err != nil {
		ws.Write([]byte("Error: invalid host_id parameter\r\n"))
		return
	}

	// Get SSH host configuration
	host, err := models.GetSSHHostByID(hostUUID)
	if err != nil {
		ws.Write([]byte(fmt.Sprintf("Error: SSH host not found: %v\r\n", err)))
		return
	}

	// Create SSH client options using webssh library
	opt := createSSHClientOption(*host)

	// Start SSH connection using webssh.Connect
	if err := webssh.Connect(ws, opt); err != nil {
		log.Printf("SSH connection failed: %v", err)
		ws.Write([]byte(fmt.Sprintf("SSH connection failed: %v\r\n", err)))
	}
}

// ListPodmanConnections lists podman connections on a remote host
func ListPodmanConnections(c echo.Context) error {
	idStr := c.Param("uid")
	hostID, err := DecodeFriendlyID(PrefixSSHHost, idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid SSH host ID")
	}

	// Get SSH host
	host, err := models.GetSSHHostByID(hostID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "SSH host not found")
	}

	// For now, return mock data. In actual implementation, this would:
	// - Connect to remote host via SSH
	// - Query podman connections
	// - Return real connection information
	mockConnections := []map[string]interface{}{
		{
			"name":    "default",
			"uri":     "unix:///run/user/1000/podman/podman.sock",
			"default": true,
			"status":  "active",
		},
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"host_uid":    EncodeFriendlyID(PrefixSSHHost, host.ID),
			"host_name":   host.Name,
			"connections": mockConnections,
		},
	})
}
