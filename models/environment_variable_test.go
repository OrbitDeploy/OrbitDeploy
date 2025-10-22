package models

import (
	"path/filepath"
	"testing"

	"github.com/opentdp/go-helper/dborm"
	"github.com/stretchr/testify/assert"
)

func setupEnvironmentVariableTestDB(t *testing.T) {
	// Create a temporary database for testing
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_env_var.db")

	config := &dborm.Config{
		Type:   "sqlite",
		DbName: dbPath,
	}

	if dborm.Connect(config) == nil {
		t.Fatal("Failed to connect to test database")
	}

	// Migrate test models
	err := dborm.Db.AutoMigrate(&Application{}, &EnvironmentVariable{})
	if err != nil {
		t.Fatalf("Failed to migrate test models: %v", err)
	}
}

func teardownEnvironmentVariableTestDB() {
	if dborm.Db != nil {
		dborm.Destroy()
	}
}

func TestEnvironmentVariableCRUD(t *testing.T) {
	setupEnvironmentVariableTestDB(t)
	defer teardownEnvironmentVariableTestDB()

	// Setup - Create test application
	application := &Application{
		ProjectID:   1,
		Name:        "test-app",
		Description: "Test application",
		TargetPort:  8080,
		Status:      "stopped",
	}
	err := dborm.Db.Create(application).Error
	assert.NoError(t, err)

	t.Run("CreateEnvironmentVariable", func(t *testing.T) {
		// Create plain environment variable
		envVar1, err := CreateEnvironmentVariable(application.ID, "TEST_KEY", "test_value", false)
		assert.NoError(t, err)
		assert.NotNil(t, envVar1)
		assert.Equal(t, "TEST_KEY", envVar1.Key)
		assert.Equal(t, application.ID, envVar1.ApplicationID)
		assert.False(t, envVar1.IsEncrypted)

		// Create encrypted environment variable
		envVar2, err := CreateEnvironmentVariable(application.ID, "SECRET_KEY", "secret_value", true)
		assert.NoError(t, err)
		assert.NotNil(t, envVar2)
		assert.Equal(t, "SECRET_KEY", envVar2.Key)
		assert.True(t, envVar2.IsEncrypted)
		assert.NotEqual(t, "secret_value", envVar2.Value) // Should be encrypted
	})

	t.Run("GetDecryptedValue", func(t *testing.T) {
		// Test plain value
		envVar1, err := CreateEnvironmentVariable(application.ID, "PLAIN_KEY", "plain_value", false)
		assert.NoError(t, err)
		value, err := envVar1.GetDecryptedValue()
		assert.NoError(t, err)
		assert.Equal(t, "plain_value", value)

		// Test encrypted value
		envVar2, err := CreateEnvironmentVariable(application.ID, "ENCRYPTED_KEY", "encrypted_value", true)
		assert.NoError(t, err)
		value, err = envVar2.GetDecryptedValue()
		assert.NoError(t, err)
		assert.Equal(t, "encrypted_value", value)
	})

	t.Run("ListEnvironmentVariablesByApplicationID", func(t *testing.T) {
		// Create second application for isolation
		application2 := &Application{
			ProjectID:   1,
			Name:        "test-app-2",
			Description: "Test application 2",
			TargetPort:  8081,
			Status:      "stopped",
		}
		err := dborm.Db.Create(application2).Error
		assert.NoError(t, err)

		// Create environment variables for both applications
		_, err = CreateEnvironmentVariable(application.ID, "KEY1", "value1", false)
		assert.NoError(t, err)
		_, err = CreateEnvironmentVariable(application.ID, "KEY2", "value2", true)
		assert.NoError(t, err)
		_, err = CreateEnvironmentVariable(application2.ID, "KEY3", "value3", false)
		assert.NoError(t, err)

		// List variables for first application
		envVars1, err := ListEnvironmentVariablesByApplicationID(application.ID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(envVars1), 2) // At least the 2 we just created

		// List variables for second application
		envVars2, err := ListEnvironmentVariablesByApplicationID(application2.ID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(envVars2), 1) // At least the 1 we just created
	})

	t.Run("UpdateEnvironmentVariable", func(t *testing.T) {
		envVar, err := CreateEnvironmentVariable(application.ID, "UPDATE_KEY", "original_value", false)
		assert.NoError(t, err)

		// Update the environment variable
		updatedVar, err := UpdateEnvironmentVariable(envVar.ID, "UPDATED_KEY", "updated_value", true)
		assert.NoError(t, err)
		assert.Equal(t, "UPDATED_KEY", updatedVar.Key)
		assert.True(t, updatedVar.IsEncrypted)

		// Verify the value was encrypted and can be decrypted
		value, err := updatedVar.GetDecryptedValue()
		assert.NoError(t, err)
		assert.Equal(t, "updated_value", value)
	})

	t.Run("DeleteEnvironmentVariable", func(t *testing.T) {
		envVar, err := CreateEnvironmentVariable(application.ID, "DELETE_KEY", "delete_value", false)
		assert.NoError(t, err)

		// Delete the environment variable
		err = DeleteEnvironmentVariable(envVar.ID)
		assert.NoError(t, err)

		// Verify it's deleted
		_, err = GetEnvironmentVariableByID(envVar.ID)
		assert.Error(t, err) // Should return error since record is deleted
	})

	t.Run("GenerateEnvFileContent", func(t *testing.T) {
		// Create environment variables
		_, err := CreateEnvironmentVariable(application.ID, "ENV1", "value1", false)
		assert.NoError(t, err)
		_, err = CreateEnvironmentVariable(application.ID, "ENV2", "value2", true)
		assert.NoError(t, err)

		// Generate env file content
		content, err := GenerateEnvFileContent(application.ID)
		assert.NoError(t, err)
		assert.Contains(t, content, "ENV1=value1")
		assert.Contains(t, content, "ENV2=value2")
	})

	t.Run("CreateSnapshotForDeployment", func(t *testing.T) {
		// Create environment variables
		_, err := CreateEnvironmentVariable(application.ID, "SNAP1", "snapvalue1", false)
		assert.NoError(t, err)
		_, err = CreateEnvironmentVariable(application.ID, "SNAP2", "snapvalue2", true)
		assert.NoError(t, err)

		// Create snapshot
		snapshot, err := CreateSnapshotForDeployment(application.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, snapshot)
		assert.Contains(t, snapshot, "SNAP1")
		assert.Contains(t, snapshot, "SNAP2")
		assert.Contains(t, snapshot, "snapvalue1")
		assert.Contains(t, snapshot, "snapvalue2")
	})
}