package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/OrbitDeploy/OrbitDeploy/models"
	"github.com/OrbitDeploy/OrbitDeploy/services"
	"github.com/labstack/echo/v4"
	"github.com/opentdp/go-helper/logman"
	"github.com/tmaxmax/go-sse"
)

// RunningDeploymentSummary represents a summary of running deployments
type RunningDeploymentSummary struct {
	TotalRunning int                       `json:"total_running"`
	LastUpdated  time.Time                 `json:"last_updated"`
	Deployments  []RunningDeploymentDetail `json:"deployments"`
}

// RunningDeploymentDetail represents details of a single running deployment
type RunningDeploymentDetail struct {
	AppName string `json:"app_name"`
	Version string `json:"version"`
}

// SSE server and publisher for running deployments
var runningDeploymentsSSEServer *sse.Server
var runningDeploymentsSSEOnce sync.Once
var runningDeploymentsPublisherOnce sync.Once

func getRunningDeploymentsSSEServer() *sse.Server {
	runningDeploymentsSSEOnce.Do(func() {
		joe := &sse.Joe{}
		runningDeploymentsSSEServer = &sse.Server{
			Provider: joe,
			OnSession: func(w http.ResponseWriter, r *http.Request) (topics []string, allowed bool) {
				logman.Info("Running deployments SSE connected", "remote", r.RemoteAddr)
				return []string{"running_deployments"}, true
			},
		}
		logman.Info("Running deployments SSE server initialized")
	})
	return runningDeploymentsSSEServer
}

// startRunningDeploymentsPublisher ensures a single publisher goroutine sends updates periodically
func startRunningDeploymentsPublisher() {
	runningDeploymentsPublisherOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(10 * time.Second) // Update every 10 seconds
			defer ticker.Stop()
			for {
				summary := getRunningDeploymentsSummary()
				payload, _ := json.Marshal(summary)
				msg := &sse.Message{Type: sse.Type("running-deployments")}
				msg.AppendData(string(payload))
				server := getRunningDeploymentsSSEServer()
				if err := server.Publish(msg, "running_deployments"); err != nil {
					logman.Warn("Failed to publish running deployments", "error", err)
				}
				<-ticker.C
			}
		}()
	})
}

// getRunningDeploymentsSummary gets the count of running deployments
func getRunningDeploymentsSummary() *RunningDeploymentSummary {
	// Create PodmanService to check running status
	ps := services.NewPodmanService()

	// Get all deployments (preloaded with Application and Release)
	deployments, err := models.GetAllRunningDeployments()
	if err != nil {
		logman.Error("Failed to get all deployments", "error", err)
		return &RunningDeploymentSummary{
			TotalRunning: 0,
			LastUpdated:  time.Now(),
			Deployments:  []RunningDeploymentDetail{},
		}
	}

	// Filter for actually running deployments
	var runningDeployments []*models.Deployment
	for _, deployment := range deployments {
		if ps.CheckContainerRunningWithQuadlet(deployment.ServiceName) {
			runningDeployments = append(runningDeployments, deployment)
		}
	}

	details := make([]RunningDeploymentDetail, len(runningDeployments))
	for i, dep := range runningDeployments {
		version := ""
		if dep.Release.Version != "" {
			version = dep.Release.Version
		}
		details[i] = RunningDeploymentDetail{
			AppName: dep.Application.Name, // Assuming Application has a Name field
			Version: version,
		}
	}

	return &RunningDeploymentSummary{
		TotalRunning: len(runningDeployments),
		LastUpdated:  time.Now(),
		Deployments:  details,
	}
}

// RunningDeploymentsHandler serves the SSE endpoint for running deployments
func RunningDeploymentsHandler(c echo.Context) error {
	startRunningDeploymentsPublisher()
	server := getRunningDeploymentsSSEServer()
	server.ServeHTTP(c.Response().Writer, c.Request())
	return nil
}
