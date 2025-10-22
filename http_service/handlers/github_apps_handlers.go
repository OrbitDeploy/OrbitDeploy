package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv" // Added for parsing installation_id
	"time"

	"github.com/labstack/echo/v4"
	"github.com/youfun/OrbitDeploy/models"
)

// GitHubAppManifest represents the GitHub App manifest structure
type GitHubAppManifest struct {
	Name               string                  `json:"name"`
	URL                string                  `json:"url"`
	HookAttributes     GitHubAppHookAttributes `json:"hook_attributes"`
	Public             bool                    `json:"public"`
	DefaultPermissions map[string]string       `json:"default_permissions"`
	DefaultEvents      []string                `json:"default_events"`
	Description        string                  `json:"description"`
	RedirectURL        string                  `json:"redirect_url"` // Added for GitHub redirect after registration
}

type GitHubAppHookAttributes struct {
	URL string `json:"url"`
}

// GitHubAppManifestRequest represents the request for generating manifest URL
type GitHubAppManifestRequest struct {
	ServerURL   string `json:"serverUrl" query:"serverUrl"`     // Base URL of the OrbitDeploy server
	CallbackURL string `json:"callbackUrl" query:"callbackUrl"` // Frontend callback URL after app creation
}

// GitHubAppManifestResponse represents the response containing the manifest URL
type GitHubAppManifestResponse struct {
	ManifestURL string            `json:"manifestUrl"`
	Manifest    GitHubAppManifest `json:"manifest"`
}

// GitHubAppCallbackRequest represents the callback data from GitHub
type GitHubAppCallbackRequest struct {
	Code           string `json:"code"`
	InstallationID uint   `json:"installationId,omitempty"`
	// Removed UserID, as it's not required (model doesn't use it)
}

// GitHubAppCreationResponse represents GitHub's response after app creation
type GitHubAppCreationResponse struct {
	ID     int    `json:"id"`
	Slug   string `json:"slug"`
	NodeID string `json:"node_id"`
	Owner  struct {
		Login string `json:"login"`
		ID    int    `json:"id"`
	} `json:"owner"`
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	ExternalURL   string            `json:"external_url"`
	HTMLUrl       string            `json:"html_url"`
	CreatedAt     string            `json:"created_at"`
	UpdatedAt     string            `json:"updated_at"`
	Permissions   map[string]string `json:"permissions"`
	Events        []string          `json:"events"`
	ClientID      string            `json:"client_id"`
	ClientSecret  string            `json:"client_secret"`
	WebhookSecret string            `json:"webhook_secret"`
	PEM           string            `json:"pem"`
}

// GenerateGitHubAppManifest generates a GitHub App manifest URL
func GenerateGitHubAppManifest(c echo.Context) error {
	var req GitHubAppManifestRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format: "+err.Error())
	}

	// Validate required fields
	if req.ServerURL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Server URL is required")
	}
	if req.CallbackURL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Callback URL is required")
	}

	// Auto-generate app name with timestamp
	appName := fmt.Sprintf("OrbitDeploy-%s", time.Now().Format("20060102150405"))

	// Create the manifest
	manifest := GitHubAppManifest{
		Name:        appName,
		URL:         req.ServerURL,
		Public:      false,
		Description: fmt.Sprintf("OrbitDeploy GitHub App for %s", appName),
		HookAttributes: GitHubAppHookAttributes{
			URL: fmt.Sprintf("%s/api/providers/github/webhook", req.ServerURL),
		},
		DefaultPermissions: map[string]string{
			"contents":      "read",
			"metadata":      "read",
			"pull_requests": "read",
		},
		DefaultEvents: []string{
			"push",
			"pull_request",
		},
		// Fixed: Set redirect_url to clean callback URL without query params (state is handled in manifest URL)
		RedirectURL: fmt.Sprintf("%s/api/providers/github/app-callback", req.ServerURL),
	}

	// Convert manifest to JSON
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate manifest: "+err.Error())
	}

	// Create the GitHub App manifest URL
	manifestURL := fmt.Sprintf("https://github.com/settings/apps/manifest?manifest=%s&state=%s",
		url.QueryEscape(string(manifestJSON)),
		url.QueryEscape(req.CallbackURL))

	// Add logging for debugging
	fmt.Printf("Generated manifest URL: %s\n", manifestURL)
	fmt.Printf("Manifest JSON: %s\n", string(manifestJSON))

	response := GitHubAppManifestResponse{
		ManifestURL: manifestURL,
		Manifest:    manifest,
	}

	return SendSuccess(c, response)
}

// HandleGitHubAppCallback processes the callback after GitHub App creation
func HandleGitHubAppCallback(c echo.Context) error {
	// Read from query parameters (GET request from GitHub redirect)
	code := c.QueryParam("code")
	state := c.QueryParam("state") // Frontend URL to redirect to after processing

	if code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Authorization code is required")
	}

	// Parse installation_id if provided (optional)
	installationIDStr := c.QueryParam("installation_id")
	installationID := uint(0)
	if installationIDStr != "" {
		if id, err := strconv.ParseUint(installationIDStr, 10, 32); err == nil {
			installationID = uint(id)
		}
	} else {
		// Log warning if installation_id is missing
		fmt.Printf("Warning: installation_id not provided in callback. InstallationID set to 0. User may need to update it manually.\n")
	}

	// Exchange the code for app credentials
	appCredentials, err := exchangeCodeForAppCredentials(code)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to exchange code for credentials: "+err.Error())
	}

	// Log the credentials for debugging
	fmt.Printf("AppCredentials: ID=%d, Slug=%s, PEM length=%d, WebhookSecret=%s\n", appCredentials.ID, appCredentials.Slug, len(appCredentials.PEM), appCredentials.WebhookSecret)

	// Create a ProviderAuth record for the GitHub App (no userId needed)
	providerAuth, err := models.CreateProviderAuth(
		"github",
		"",                                   // ClientID not used for GitHub Apps
		"",                                   // ClientSecret not used for GitHub Apps
		"",                                   // RedirectURI not used for GitHub Apps
		"",                                   // Username not used for GitHub Apps
		"",                                   // AppPassword not used for GitHub Apps
		fmt.Sprintf("%d", appCredentials.ID), // AppID
		appCredentials.Slug,                  // Slug
		appCredentials.PEM,                   // PrivateKey
		appCredentials.WebhookSecret,         // WebhookSecret
		installationID,                       // InstallationID (if provided)
		"contents:read,metadata:read,pull_requests:read", // Scopes
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save GitHub App credentials: "+err.Error())
	}

	// Return the created provider auth info
	response := &ProviderAuthResponse{
		Uid:            EncodeFriendlyID(PrefixProviderAuth, providerAuth.ID),
		Platform:       providerAuth.Platform,
		AppID:          providerAuth.AppID,
		WebhookSecret:  providerAuth.WebhookSecret,
		InstallationID: providerAuth.InstallationID,
		Scopes:         providerAuth.Scopes,
		IsActive:       providerAuth.IsActive,
		CreatedAt:      providerAuth.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      providerAuth.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// If state is provided, decode and redirect to frontend after processing
	if state != "" {
		decodedState, err := url.QueryUnescape(state)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid state parameter: "+err.Error())
		}
		return c.Redirect(http.StatusFound, decodedState)
	}

	// Fallback: Return JSON response if no state
	return SendSuccess(c, response)
}

// exchangeCodeForAppCredentials exchanges the GitHub authorization code for app credentials
func exchangeCodeForAppCredentials(code string) (*GitHubAppCreationResponse, error) {
	// Create the request to GitHub's API
	exchangeURL := "https://api.github.com/app-manifests/" + code + "/conversions"

	// Make the POST request to GitHub
	resp, err := http.Post(exchangeURL, "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		// Read response body for detailed error
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status: %d, body: %s", resp.StatusCode, string(body))
	}

	var appResponse GitHubAppCreationResponse
	if err := json.NewDecoder(resp.Body).Decode(&appResponse); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub response: %v", err)
	}

	return &appResponse, nil
}

// InstallGitHubApp handles the GitHub App installation flow
func InstallGitHubApp(c echo.Context) error {
	// Get the provider auth ID from URL parameter
	providerAuthIDStr := c.Param("id")
	if providerAuthIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Provider auth ID is required")
	}

	paID, err := DecodeFriendlyID(PrefixProviderAuth, providerAuthIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid provider auth ID format")
	}

	// Get installation ID from request
	var req struct {
		InstallationID uint `json:"installationId"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format: "+err.Error())
	}

	if req.InstallationID == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Installation ID is required")
	}

	// Get the provider auth record
	providerAuth, err := models.GetProviderAuthByID(paID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Provider auth not found")
	}

	// Update the installation ID
	_, err = models.UpdateProviderAuth(
		paID,
		providerAuth.ClientID,
		providerAuth.ClientSecret,
		providerAuth.RedirectURI,
		providerAuth.Username,
		providerAuth.AppPassword,
		providerAuth.AppID,
		providerAuth.Slug, // Added missing slug parameter
		providerAuth.PrivateKey,
		providerAuth.WebhookSecret,
		req.InstallationID,
		providerAuth.Scopes,
		providerAuth.IsActive,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update installation ID: "+err.Error())
	}

	return SendSuccess(c, map[string]interface{}{
		"success":        true,
		"message":        "GitHub App installation completed successfully",
		"installationId": req.InstallationID,
	})
}

// HandleGitHubWebhook handles incoming webhooks from GitHub
func HandleGitHubWebhook(c echo.Context) error {
	// Get the webhook payload
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to read request body")
	}

	// Get the signature from headers
	signature := c.Request().Header.Get("X-Hub-Signature-256")
	if signature == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "Missing webhook signature")
	}

	// Get the event type
	eventType := c.Request().Header.Get("X-GitHub-Event")
	if eventType == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing event type")
	}

	// Find the provider auth by platform (assuming only one GitHub app for now)
	providerAuths, err := models.ListProviderAuthsByPlatform("github")
	if err != nil || len(providerAuths) == 0 {
		return echo.NewHTTPError(http.StatusInternalServerError, "No GitHub provider auth found")
	}
	providerAuth := providerAuths[0] // Use the first one

	// Verify the signature using WebhookSecret
	if !verifyWebhookSignature(body, signature, providerAuth.WebhookSecret) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid webhook signature")
	}

	// Parse the payload
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to parse payload")
	}

	// Handle installation event
	if eventType == "installation" {
		action, ok := payload["action"].(string)
		if !ok {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid installation payload")
		}

		fmt.Printf("Webhook event: action=%s, providerAuth ID=%d\n", action, providerAuth.ID)

		if action == "created" || action == "new_permissions_accepted" {
			installation, ok := payload["installation"].(map[string]interface{})
			if !ok {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid installation data")
			}

			installationIDFloat, ok := installation["id"].(float64)
			if !ok {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid installation ID")
			}
			installationID := uint(installationIDFloat)

			fmt.Printf("Extracted InstallationID=%d from webhook\n", installationID)

			// Update the InstallationID in database
			_, err := models.UpdateProviderAuth(
				providerAuth.ID,
				providerAuth.ClientID,
				providerAuth.ClientSecret,
				providerAuth.RedirectURI,
				providerAuth.Username,
				providerAuth.AppPassword,
				providerAuth.AppID,
				providerAuth.Slug,
				providerAuth.PrivateKey,
				providerAuth.WebhookSecret,
				installationID,
				providerAuth.Scopes,
				providerAuth.IsActive,
			)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update installation ID: "+err.Error())
			}

			fmt.Printf("Updated InstallationID to %d for provider auth ID %d\n", installationID, providerAuth.ID)
		}
	}

	// TODO: Handle other events like push, pull_request if needed

	return SendSuccess(c, map[string]string{"message": "Webhook processed"})
}

// verifyWebhookSignature verifies the HMAC signature of the webhook
func verifyWebhookSignature(payload []byte, signature, secret string) bool {
	if secret == "" {
		return false
	}

	// Remove "sha256=" prefix if present
	if len(signature) > 7 && signature[:7] == "sha256=" {
		signature = signature[7:]
	}

	// Compute HMAC
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))

}
