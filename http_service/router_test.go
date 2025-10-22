package http_service

import (
	"embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed test_assets/*
var testAssets embed.FS

// TestEnvironmentRoutesArePublic verifies that environment routes do not require authentication
func TestEnvironmentRoutesArePublic(t *testing.T) {
	// Create test assets directory structure
	e := NewEchoServer(testAssets)

	testCases := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "CheckDeploymentEnvironment is public",
			method: http.MethodGet,
			path:   "/api/environment/check",
		},
		{
			name:   "InstallPodman is public",
			method: http.MethodPost,
			path:   "/api/environment/install-podman",
		},
		{
			name:   "InstallCaddy is public",
			method: http.MethodPost,
			path:   "/api/environment/install-caddy",
		},
		{
			name:   "ContainerCheckEnv is public",
			method: http.MethodPost,
			path:   "/api/containers/check-env",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			// These routes should NOT return 401 Unauthorized
			// They may return other errors (like 500 if scripts not set), but not 401
			assert.NotEqual(t, http.StatusUnauthorized, rec.Code,
				"Route %s should be accessible without authentication but returned %d",
				tc.path, rec.Code)
		})
	}
}

// TestProtectedRoutesRequireAuth verifies that protected routes still require authentication
func TestProtectedRoutesRequireAuth(t *testing.T) {
	e := NewEchoServer(testAssets)

	testCases := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "SystemMonitor requires auth",
			method: http.MethodGet,
			path:   "/api/system/monitor",
		},
		{
			name:   "ListProjects requires auth",
			method: http.MethodGet,
			path:   "/api/projects",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			// These routes SHOULD return 401 Unauthorized when no auth token is provided
			assert.Equal(t, http.StatusUnauthorized, rec.Code,
				"Route %s should require authentication but returned %d",
				tc.path, rec.Code)
		})
	}
}
