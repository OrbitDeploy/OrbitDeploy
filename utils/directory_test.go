package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoryManager_EnsureDirectoriesExist(t *testing.T) {
	dm := NewDirectoryManager()
	
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	testDirs := []string{
		filepath.Join(tempDir, "test1"),
		filepath.Join(tempDir, "test2", "nested"),
		filepath.Join(tempDir, "test3"),
	}
	
	// Test creating directories
	createdDirs, err := dm.EnsureDirectoriesExist(testDirs)
	if err != nil {
		t.Fatalf("EnsureDirectoriesExist failed: %v", err)
	}
	
	// All directories should be created
	if len(createdDirs) != len(testDirs) {
		t.Errorf("Expected %d created directories, got %d", len(testDirs), len(createdDirs))
	}
	
	// Verify directories exist
	for _, dir := range testDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Directory %s was not created", dir)
		}
	}
	
	// Test with existing directories (should not create again)
	createdDirs2, err := dm.EnsureDirectoriesExist(testDirs)
	if err != nil {
		t.Fatalf("EnsureDirectoriesExist failed on second call: %v", err)
	}
	
	// No directories should be created this time
	if len(createdDirs2) != 0 {
		t.Errorf("Expected 0 created directories on second call, got %d", len(createdDirs2))
	}
}

func TestDirectoryManager_CheckDirectoryExists(t *testing.T) {
	dm := NewDirectoryManager()
	
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	existingDir := filepath.Join(tempDir, "existing")
	os.Mkdir(existingDir, 0755)
	
	nonExistingDir := filepath.Join(tempDir, "nonexisting")
	
	// Test existing directory
	exists, err := dm.CheckDirectoryExists(existingDir)
	if err != nil {
		t.Fatalf("CheckDirectoryExists failed: %v", err)
	}
	if !exists {
		t.Error("Expected existing directory to return true")
	}
	
	// Test non-existing directory
	exists, err = dm.CheckDirectoryExists(nonExistingDir)
	if err != nil {
		t.Fatalf("CheckDirectoryExists failed: %v", err)
	}
	if exists {
		t.Error("Expected non-existing directory to return false")
	}
	
	// Test empty path
	_, err = dm.CheckDirectoryExists("")
	if err == nil {
		t.Error("Expected error for empty directory path")
	}
}

func TestDirectoryManager_ValidateDirectories(t *testing.T) {
	dm := NewDirectoryManager()
	
	tempDir := t.TempDir()
	
	tests := []struct {
		name        string
		directories []string
		expectError bool
	}{
		{
			name:        "valid absolute paths",
			directories: []string{filepath.Join(tempDir, "valid1"), filepath.Join(tempDir, "valid2")},
			expectError: false,
		},
		{
			name:        "relative path should fail",
			directories: []string{"relative/path"},
			expectError: true,
		},
		{
			name:        "empty paths should be skipped",
			directories: []string{"", filepath.Join(tempDir, "valid")},
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dm.ValidateDirectories(tt.directories)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}