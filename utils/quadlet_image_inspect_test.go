package utils

import (
	"strings"
	"testing"
)

// TestAddPermissionsToQuadlet tests the AddPermissionsToQuadlet function
func TestAddPermissionsToQuadlet(t *testing.T) {
	// Test case 1: Basic functionality
	t.Run("Basic permission addition", func(t *testing.T) {
		quadletContent := `[Unit]
Description=Test Container

[Container]
Image=nginx:latest
Volume=/host/data:/container/data
PublishPort=8080:80

[Install]
WantedBy=default.target`

		userInfo := &ImageUserInfo{
			UID:     "1000",
			GID:     "1000",
			UserStr: "1000:1000",
		}

		hostPaths := []string{"/host/data"}

		result, err := AddPermissionsToQuadlet(quadletContent, userInfo, hostPaths)
		if err != nil {
			t.Fatalf("AddPermissionsToQuadlet failed: %v", err)
		}

		// Check that Service section was added
		if !strings.Contains(result, "[Service]") {
			t.Error("Expected [Service] section to be added")
		}

		// Check that ExecStartPre commands were added
		if !strings.Contains(result, "ExecStartPre=-/bin/mkdir -p /host/data") {
			t.Error("Expected mkdir command to be added")
		}

		if !strings.Contains(result, "ExecStartPre=-/bin/chown 1000:1000 /host/data") {
			t.Error("Expected chown command to be added")
		}

		// Check that Service section comes before Container section
		servicePos := strings.Index(result, "[Service]")
		containerPos := strings.Index(result, "[Container]")
		if servicePos >= containerPos {
			t.Error("Expected [Service] section to come before [Container] section")
		}
	})

	// Test case 2: Service section already exists
	t.Run("Service section already exists", func(t *testing.T) {
		quadletContent := `[Unit]
Description=Test Container

[Service]
Type=notify
ExecStartPre=-/bin/echo "Custom pre-start"

[Container]
Image=nginx:latest
Volume=/host/data:/container/data

[Install]
WantedBy=default.target`

		userInfo := &ImageUserInfo{
			UID:     "1000",
			GID:     "1000",
			UserStr: "1000:1000",
		}

		hostPaths := []string{"/host/data"}

		result, err := AddPermissionsToQuadlet(quadletContent, userInfo, hostPaths)
		if err != nil {
			t.Fatalf("AddPermissionsToQuadlet failed: %v", err)
		}

		// Content should remain unchanged
		if result != quadletContent {
			t.Error("Expected content to remain unchanged when [Service] section already exists")
		}
	})

	// Test case 3: No host paths
	t.Run("No host paths", func(t *testing.T) {
		quadletContent := `[Unit]
Description=Test Container

[Container]
Image=nginx:latest

[Install]
WantedBy=default.target`

		userInfo := &ImageUserInfo{
			UID:     "1000",
			GID:     "1000",
			UserStr: "1000:1000",
		}

		hostPaths := []string{}

		result, err := AddPermissionsToQuadlet(quadletContent, userInfo, hostPaths)
		if err != nil {
			t.Fatalf("AddPermissionsToQuadlet failed: %v", err)
		}

		// Content should remain unchanged
		if result != quadletContent {
			t.Error("Expected content to remain unchanged when no host paths")
		}
	})

	// Test case 4: Multiple host paths
	t.Run("Multiple host paths", func(t *testing.T) {
		quadletContent := `[Unit]
Description=Test Container

[Container]
Image=nginx:latest
Volume=/host/data:/container/data
Volume=/host/config:/container/config

[Install]
WantedBy=default.target`

		userInfo := &ImageUserInfo{
			UID:     "1000",
			GID:     "1000",
			UserStr: "1000:1000",
		}

		hostPaths := []string{"/host/data", "/host/config"}

		result, err := AddPermissionsToQuadlet(quadletContent, userInfo, hostPaths)
		if err != nil {
			t.Fatalf("AddPermissionsToQuadlet failed: %v", err)
		}

		// Check that both paths are handled
		if !strings.Contains(result, "ExecStartPre=-/bin/mkdir -p /host/data") {
			t.Error("Expected mkdir command for /host/data")
		}

		if !strings.Contains(result, "ExecStartPre=-/bin/chown 1000:1000 /host/data") {
			t.Error("Expected chown command for /host/data")
		}

		if !strings.Contains(result, "ExecStartPre=-/bin/mkdir -p /host/config") {
			t.Error("Expected mkdir command for /host/config")
		}

		if !strings.Contains(result, "ExecStartPre=-/bin/chown 1000:1000 /host/config") {
			t.Error("Expected chown command for /host/config")
		}
	})
}

// TestIsNumeric tests the isNumeric helper function
func TestIsNumeric(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"0", true},
		{"1000", true},
		{"abc", false},
		{"12.34", false},
		{"", false},
		{"1000:1000", false},
		{"user", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := isNumeric(tc.input)
			if result != tc.expected {
				t.Errorf("isNumeric(%q) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}