package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/youfun/OrbitDeploy/config"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeSuccess NotificationType = "success"
	NotificationTypeError   NotificationType = "error"
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeInfo    NotificationType = "info"
)

// WebhookNotification represents the structure of a webhook notification
type WebhookNotification struct {
	Type        NotificationType `json:"type"`
	Title       string           `json:"title"`
	Message     string           `json:"message"`
	Service     string           `json:"service,omitempty"`
	Timestamp   time.Time        `json:"timestamp"`
	Details     interface{}      `json:"details,omitempty"`
	ServerAddr  string           `json:"server_addr,omitempty"`
	Environment string           `json:"environment,omitempty"`
}

// WebhookConfig holds the configuration for webhook notifications
type WebhookConfig struct {
	URL        string
	Token      string
	MaxRetries int
	Timeout    time.Duration
	RetryDelay time.Duration
}

// DefaultWebhookConfig returns a webhook config with sensible defaults
func DefaultWebhookConfig() *WebhookConfig {
	cfg := config.Load()
	return &WebhookConfig{
		URL:        cfg.WebhookURL,
		Token:      cfg.WebhookToken,
		MaxRetries: 3,
		Timeout:    30 * time.Second,
		RetryDelay: 2 * time.Second,
	}
}

// SendWebhookNotification sends a notification to a webhook endpoint
// This is the main function that other parts of the application should call
func SendWebhookNotification(notificationType NotificationType, title, message string, opts ...WebhookOption) error {
	config := DefaultWebhookConfig()

	// If no webhook URL is configured, log a warning and return
	if config.URL == "" {
		log.Printf("Webhook notification not sent: WEBHOOK_URL not configured (type: %s, title: %s)", notificationType, title)
		return nil
	}

	notification := &WebhookNotification{
		Type:      notificationType,
		Title:     title,
		Message:   message,
		Timestamp: time.Now(),
	}

	// Apply options
	for _, opt := range opts {
		opt(notification)
	}

	return sendNotificationWithRetry(config, notification)
}

// WebhookOption is a function type for configuring webhook notifications
type WebhookOption func(*WebhookNotification)

// WithService sets the service name for the notification
func WithService(service string) WebhookOption {
	return func(n *WebhookNotification) {
		n.Service = service
	}
}

// WithDetails adds additional details to the notification
func WithDetails(details interface{}) WebhookOption {
	return func(n *WebhookNotification) {
		n.Details = details
	}
}

// WithEnvironment sets the environment for the notification
func WithEnvironment(env string) WebhookOption {
	return func(n *WebhookNotification) {
		n.Environment = env
	}
}

// WithServerAddr sets the server address for the notification
func WithServerAddr(addr string) WebhookOption {
	return func(n *WebhookNotification) {
		n.ServerAddr = addr
	}
}

// sendNotificationWithRetry sends the notification with retry logic
func sendNotificationWithRetry(config *WebhookConfig, notification *WebhookNotification) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Webhook notification retry attempt %d/%d", attempt, config.MaxRetries)
			time.Sleep(config.RetryDelay)
		}

		err := sendNotification(config, notification)
		if err == nil {
			if attempt > 0 {
				log.Printf("Webhook notification sent successfully after %d retries", attempt)
			}
			return nil
		}

		lastErr = err
		log.Printf("Webhook notification attempt %d failed: %v", attempt+1, err)
	}

	return fmt.Errorf("webhook notification failed after %d attempts: %w", config.MaxRetries+1, lastErr)
}

// sendNotification sends a single notification attempt
func sendNotification(config *WebhookConfig, notification *WebhookNotification) error {
	// Convert notification to JSON
	jsonData, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Create HTTP request
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", config.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "web-deploy/1.0")

	// Add authorization header if token is provided
	if config.Token != "" {
		// Support different token formats
		if strings.HasPrefix(config.Token, "Bearer ") {
			req.Header.Set("Authorization", config.Token)
		} else {
			req.Header.Set("Authorization", "Bearer "+config.Token)
		}
	}

	// Send request
	client := &http.Client{
		Timeout: config.Timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook endpoint returned status %d", resp.StatusCode)
	}

	log.Printf("Webhook notification sent successfully (type: %s, title: %s, status: %d)",
		notification.Type, notification.Title, resp.StatusCode)

	return nil
}

// Convenience functions for common notification types

// SendSuccessNotification sends a success notification
func SendSuccessNotification(title, message string, opts ...WebhookOption) error {
	return SendWebhookNotification(NotificationTypeSuccess, title, message, opts...)
}

// SendErrorNotification sends an error notification
func SendErrorNotification(title, message string, opts ...WebhookOption) error {
	return SendWebhookNotification(NotificationTypeError, title, message, opts...)
}

// SendWarningNotification sends a warning notification
func SendWarningNotification(title, message string, opts ...WebhookOption) error {
	return SendWebhookNotification(NotificationTypeWarning, title, message, opts...)
}

// SendInfoNotification sends an info notification
func SendInfoNotification(title, message string, opts ...WebhookOption) error {
	return SendWebhookNotification(NotificationTypeInfo, title, message, opts...)
}

// SendServiceRestartNotification is a convenience function for service restart notifications
func SendServiceRestartNotification(serviceName, status string, details interface{}) error {
	var notificationType NotificationType
	var title string

	switch strings.ToLower(status) {
	case "success", "completed":
		notificationType = NotificationTypeSuccess
		title = fmt.Sprintf("Service %s Restarted Successfully", serviceName)
	case "failed", "error":
		notificationType = NotificationTypeError
		title = fmt.Sprintf("Service %s Restart Failed", serviceName)
	case "warning":
		notificationType = NotificationTypeWarning
		title = fmt.Sprintf("Service %s Restart Warning", serviceName)
	default:
		notificationType = NotificationTypeInfo
		title = fmt.Sprintf("Service %s Restart Status: %s", serviceName, status)
	}

	message := fmt.Sprintf("Service %s restart operation completed with status: %s", serviceName, status)

	return SendWebhookNotification(notificationType, title, message,
		WithService(serviceName),
		WithDetails(details),
	)
}

// SendDeploymentNotification is a convenience function for deployment notifications
func SendDeploymentNotification(containerName, stage, status string, details interface{}) error {
	var notificationType NotificationType
	var title string

	switch strings.ToLower(status) {
	case "success", "completed":
		notificationType = NotificationTypeSuccess
		title = fmt.Sprintf("Deployment %s - %s Completed", containerName, stage)
	case "failed", "error":
		notificationType = NotificationTypeError
		title = fmt.Sprintf("Deployment %s - %s Failed", containerName, stage)
	case "warning":
		notificationType = NotificationTypeWarning
		title = fmt.Sprintf("Deployment %s - %s Warning", containerName, stage)
	default:
		notificationType = NotificationTypeInfo
		title = fmt.Sprintf("Deployment %s - %s: %s", containerName, stage, status)
	}

	message := fmt.Sprintf("Container %s deployment stage %s completed with status: %s", containerName, stage, status)

	return SendWebhookNotification(notificationType, title, message,
		WithService(containerName),
		WithDetails(details),
	)
}
