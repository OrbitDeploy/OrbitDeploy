package services

import (
	"testing"
)

// TestDeploymentOrchestratorSSELogSender tests that the SSE log sender is properly set and called
func TestDeploymentOrchestratorSSELogSender(t *testing.T) {
	// Create orchestrator
	buildService := NewBuildService()
	envService := NewDeploymentEnvironmentService()
	podmanService := NewPodmanService()
	orchestrator := NewDeploymentOrchestrator(buildService, envService, podmanService)

	// Test that SSE sender is initially nil
	if orchestrator.sseLogSender != nil {
		t.Error("Expected SSE log sender to be nil initially")
	}

	// Track calls to SSE sender
	var calledWith []struct {
		deploymentID uint
		message      string
	}

	// Set up mock SSE sender
	mockSender := func(deploymentID uint, message string) {
		calledWith = append(calledWith, struct {
			deploymentID uint
			message      string
		}{deploymentID, message})
	}

	orchestrator.SetSSELogSender(mockSender)

	// Test that sender is set
	if orchestrator.sseLogSender == nil {
		t.Error("Expected SSE log sender to be set")
	}

	// Test sendDeploymentLog function
	testDeploymentID := uint(123)
	testMessage := "测试部署日志消息"

	orchestrator.sendDeploymentLog(testDeploymentID, testMessage)

	// Verify the mock sender was called
	if len(calledWith) != 1 {
		t.Fatalf("Expected 1 call to SSE sender, got %d", len(calledWith))
	}

	call := calledWith[0]
	if call.deploymentID != testDeploymentID {
		t.Errorf("Expected deployment ID %d, got %d", testDeploymentID, call.deploymentID)
	}

	if call.message != testMessage {
		t.Errorf("Expected message '%s', got '%s'", testMessage, call.message)
	}
}

// TestDeploymentOrchestratorLogHelperHandlesNilSender tests that the log helper doesn't crash with nil sender
func TestDeploymentOrchestratorLogHelperHandlesNilSender(t *testing.T) {
	// Create orchestrator without setting SSE sender
	buildService := NewBuildService()
	envService := NewDeploymentEnvironmentService()
	podmanService := NewPodmanService()
	orchestrator := NewDeploymentOrchestrator(buildService, envService, podmanService)

	// This should not panic
	orchestrator.sendDeploymentLog(uint(123), "Test message")
}