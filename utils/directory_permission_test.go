package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoryManager_ValidateFilePermissions(t *testing.T) {
	dm := NewDirectoryManager()
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		setup       func() string
		expectError bool
	}{
		{
			name: "valid existing file with proper permissions",
			setup: func() string {
				testFile := filepath.Join(tempDir, "test_file.txt")
				err := os.WriteFile(testFile, []byte("test content"), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return testFile
			},
			expectError: false,
		},
		{
			name: "non-existent file in valid directory",
			setup: func() string {
				return filepath.Join(tempDir, "non_existent.txt")
			},
			expectError: false, // Should pass because parent directory is writable
		},
		{
			name: "relative path should fail",
			setup: func() string {
				return "relative/path/file.txt"
			},
			expectError: true,
		},
		{
			name: "empty path should fail",
			setup: func() string {
				return ""
			},
			expectError: true,
		},
		{
			name: "path to directory instead of file",
			setup: func() string {
				testDir := filepath.Join(tempDir, "test_dir")
				err := os.MkdirAll(testDir, 0755)
				if err != nil {
					t.Fatalf("Failed to create test directory: %v", err)
				}
				return testDir
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setup()
			err := dm.ValidateFilePermissions(filePath)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestDirectoryManager_ValidateDirectoryWritePermissions(t *testing.T) {
	dm := NewDirectoryManager()
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		setup       func() string
		expectError bool
	}{
		{
			name: "valid writable directory",
			setup: func() string {
				return tempDir
			},
			expectError: false,
		},
		{
			name: "non-existent directory with valid parent",
			setup: func() string {
				return filepath.Join(tempDir, "new_dir")
			},
			expectError: false,
		},
		{
			name: "empty path should fail",
			setup: func() string {
				return ""
			},
			expectError: true,
		},
		{
			name: "path to file instead of directory",
			setup: func() string {
				testFile := filepath.Join(tempDir, "test_file.txt")
				err := os.WriteFile(testFile, []byte("test"), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return testFile
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirPath := tt.setup()
			err := dm.ValidateDirectoryWritePermissions(dirPath)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestDirectoryManager_CheckFileReadable(t *testing.T) {
	dm := NewDirectoryManager()
	tempDir := t.TempDir()

	// Create a readable file
	readableFile := filepath.Join(tempDir, "readable.txt")
	err := os.WriteFile(readableFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create readable file: %v", err)
	}

	// Test readable file
	err = dm.checkFileReadable(readableFile)
	if err != nil {
		t.Errorf("Expected file to be readable, got error: %v", err)
	}

	// Test non-existent file
	err = dm.checkFileReadable(filepath.Join(tempDir, "non_existent.txt"))
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestDirectoryManager_CheckFileWritable(t *testing.T) {
	dm := NewDirectoryManager()
	tempDir := t.TempDir()

	// Create a writable file
	writableFile := filepath.Join(tempDir, "writable.txt")
	err := os.WriteFile(writableFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create writable file: %v", err)
	}

	// Test writable file
	err = dm.checkFileWritable(writableFile)
	if err != nil {
		t.Errorf("Expected file to be writable, got error: %v", err)
	}

	// Test non-existent file
	err = dm.checkFileWritable(filepath.Join(tempDir, "non_existent.txt"))
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestDirectoryManager_TestDirectoryWritable(t *testing.T) {
	dm := NewDirectoryManager()
	tempDir := t.TempDir()

	// Test writable directory
	err := dm.testDirectoryWritable(tempDir)
	if err != nil {
		t.Errorf("Expected directory to be writable, got error: %v", err)
	}

	// Test non-existent directory
	err = dm.testDirectoryWritable(filepath.Join(tempDir, "non_existent"))
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}