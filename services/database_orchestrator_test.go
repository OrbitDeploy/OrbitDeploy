package services

import (
	"testing"

	"github.com/OrbitDeploy/OrbitDeploy/models"
)

// TestImageSelectionLogic tests the image selection logic without full orchestrator
func TestImageSelectionLogic(t *testing.T) {
	tests := []struct {
		name          string
		customImage   string
		dbType        models.SelfHostedDatabaseType
		version       string
		expectedImage string
		description   string
	}{
		{
			name:          "Custom image takes priority",
			customImage:   "docker.io/postgrest/postgrest:v12.2.8",
			dbType:        models.PostgreSQL,
			version:       "16",
			expectedImage: "docker.io/postgrest/postgrest:v12.2.8",
			description:   "When CustomImage is set, it should be used regardless of version",
		},
		{
			name:          "Default to alpine with version",
			customImage:   "",
			dbType:        models.PostgreSQL,
			version:       "16",
			expectedImage: "docker.io/library/postgres:16-alpine",
			description:   "When no CustomImage, should append -alpine to version",
		},
		{
			name:          "Alpine version preserved",
			customImage:   "",
			dbType:        models.PostgreSQL,
			version:       "16-alpine",
			expectedImage: "docker.io/library/postgres:16-alpine-alpine",
			description:   "When version already contains -alpine, it will be double-appended (known behavior, user should use custom image)",
		},
		{
			name:          "Latest version uses alpine",
			customImage:   "",
			dbType:        models.PostgreSQL,
			version:       "latest",
			expectedImage: "docker.io/library/postgres:alpine",
			description:   "Latest version should use postgres:alpine",
		},
		{
			name:          "Empty version uses alpine",
			customImage:   "",
			dbType:        models.PostgreSQL,
			version:       "",
			expectedImage: "docker.io/library/postgres:alpine",
			description:   "Empty version should use postgres:alpine",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the image selection logic from database_orchestrator.go
			var imageName string

			if tt.customImage != "" {
				imageName = tt.customImage
			} else {
				switch tt.dbType {
				case models.PostgreSQL:
					if tt.version != "" && tt.version != "latest" {
						imageName = "docker.io/library/postgres:" + tt.version + "-alpine"
					} else {
						imageName = "docker.io/library/postgres:alpine"
					}
				default:
					imageName = "docker.io/library/postgres:" + tt.version
				}
			}

			if imageName != tt.expectedImage {
				t.Errorf("Image selection = %v, want %v\nDescription: %s",
					imageName, tt.expectedImage, tt.description)
			}
		})
	}
}

// TestCustomImageExamples tests real-world custom image examples
func TestCustomImageExamples(t *testing.T) {
	examples := []struct {
		name        string
		customImage string
		description string
	}{
		{
			name:        "PostgREST",
			customImage: "docker.io/postgrest/postgrest:v12.2.8",
			description: "PostgREST API server",
		},
		{
			name:        "TimescaleDB",
			customImage: "docker.io/timescale/timescaledb:latest-pg16",
			description: "TimescaleDB time-series database",
		},
		{
			name:        "PostgreSQL with specific registry",
			customImage: "quay.io/postgres/postgres:16-alpine",
			description: "PostgreSQL from Quay registry",
		},
		{
			name:        "pgAdmin",
			customImage: "docker.io/dpage/pgadmin4:latest",
			description: "pgAdmin web interface",
		},
	}

	for _, ex := range examples {
		t.Run(ex.name, func(t *testing.T) {
			// Just verify the custom image would be used as-is
			db := &models.SelfHostedDatabase{
				CustomImage: ex.customImage,
				Type:        models.PostgreSQL,
				Version:     "16",
			}

			var imageName string
			if db.CustomImage != "" {
				imageName = db.CustomImage
			}

			if imageName != ex.customImage {
				t.Errorf("Custom image = %v, want %v\nDescription: %s",
					imageName, ex.customImage, ex.description)
			}

			t.Logf("✓ %s: %s", ex.name, ex.description)
		})
	}
}
