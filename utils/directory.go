package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// DirectoryManager handles creation and validation of directories
type DirectoryManager struct{}

// NewDirectoryManager creates a new DirectoryManager instance
func NewDirectoryManager() *DirectoryManager {
	return &DirectoryManager{}
}

// EnsureDirectoriesExist checks if directories exist and creates them if they don't
// Returns a list of directories that were created
func (dm *DirectoryManager) EnsureDirectoriesExist(directories []string) ([]string, error) {
	var createdDirs []string
	
	for _, dir := range directories {
		if dir == "" {
			continue
		}
		
		// Clean the path
		cleanDir := filepath.Clean(dir)
		
		// Check if directory exists
		if _, err := os.Stat(cleanDir); os.IsNotExist(err) {
			// Directory doesn't exist, create it
			if err := os.MkdirAll(cleanDir, 0755); err != nil {
				return createdDirs, fmt.Errorf("failed to create directory %s: %w", cleanDir, err)
			}
			createdDirs = append(createdDirs, cleanDir)
		} else if err != nil {
			// Some other error occurred while checking
			return createdDirs, fmt.Errorf("failed to check directory %s: %w", cleanDir, err)
		}
		// Directory already exists, nothing to do
	}
	
	return createdDirs, nil
}

// ValidateDirectories checks if all specified directories are valid and accessible
func (dm *DirectoryManager) ValidateDirectories(directories []string) error {
	for _, dir := range directories {
		if dir == "" {
			continue
		}
		
		cleanDir := filepath.Clean(dir)
		
		// Check if the path is absolute
		if !filepath.IsAbs(cleanDir) {
			return fmt.Errorf("directory path must be absolute: %s", cleanDir)
		}
		
		// Check if we can access the parent directory (for creation permission)
		parentDir := filepath.Dir(cleanDir)
		if parentDir != "/" && parentDir != "." {
			if _, err := os.Stat(parentDir); err != nil {
				return fmt.Errorf("cannot access parent directory of %s: %w", cleanDir, err)
			}
		}
	}
	
	return nil
}

// CheckDirectoryExists checks if a single directory exists
func (dm *DirectoryManager) CheckDirectoryExists(directory string) (bool, error) {
	if directory == "" {
		return false, fmt.Errorf("directory path cannot be empty")
	}
	
	cleanDir := filepath.Clean(directory)
	
	info, err := os.Stat(cleanDir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check directory %s: %w", cleanDir, err)
	}
	
	if !info.IsDir() {
		return false, fmt.Errorf("path exists but is not a directory: %s", cleanDir)
	}
	
	return true, nil
}

// CreateDirectory creates a single directory with proper permissions
func (dm *DirectoryManager) CreateDirectory(directory string) error {
	if directory == "" {
		return fmt.Errorf("directory path cannot be empty")
	}
	
	cleanDir := filepath.Clean(directory)
	
	// Check if directory already exists
	if exists, err := dm.CheckDirectoryExists(cleanDir); err != nil {
		return err
	} else if exists {
		return nil // Directory already exists
	}
	
	// Create the directory
	if err := os.MkdirAll(cleanDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", cleanDir, err)
	}
	
	return nil
}

// ValidateFilePermissions validates if a file path can be read/written
func (dm *DirectoryManager) ValidateFilePermissions(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}
	
	cleanPath := filepath.Clean(filePath)
	
	// Check if the path is absolute
	if !filepath.IsAbs(cleanPath) {
		return fmt.Errorf("file path must be absolute: %s", cleanPath)
	}
	
	// Check if file exists
	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, check if we can create it by validating parent directory
			parentDir := filepath.Dir(cleanPath)
			return dm.ValidateDirectoryWritePermissions(parentDir)
		}
		return fmt.Errorf("failed to check file %s: %w", cleanPath, err)
	}
	
	// File exists, check if it's actually a file
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", cleanPath)
	}
	
	// Check read permission
	if err := dm.checkFileReadable(cleanPath); err != nil {
		return fmt.Errorf("file is not readable: %s (%w)", cleanPath, err)
	}
	
	// Check write permission
	if err := dm.checkFileWritable(cleanPath); err != nil {
		return fmt.Errorf("file is not writable: %s (%w)", cleanPath, err)
	}
	
	return nil
}

// ValidateDirectoryWritePermissions validates if a directory can be written to
func (dm *DirectoryManager) ValidateDirectoryWritePermissions(dirPath string) error {
	if dirPath == "" {
		return fmt.Errorf("directory path cannot be empty")
	}
	
	cleanDir := filepath.Clean(dirPath)
	
	// Check if directory exists
	info, err := os.Stat(cleanDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist, check parent
			parentDir := filepath.Dir(cleanDir)
			if parentDir == cleanDir || parentDir == "/" {
				return fmt.Errorf("cannot create directory %s: no valid parent", cleanDir)
			}
			return dm.ValidateDirectoryWritePermissions(parentDir)
		}
		return fmt.Errorf("failed to check directory %s: %w", cleanDir, err)
	}
	
	if !info.IsDir() {
		return fmt.Errorf("path exists but is not a directory: %s", cleanDir)
	}
	
	// Test write permissions by trying to create a temporary file
	return dm.testDirectoryWritable(cleanDir)
}

// checkFileReadable tests if a file can be read
func (dm *DirectoryManager) checkFileReadable(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Try to read a small amount to test read permission
	buffer := make([]byte, 1)
	_, err = file.Read(buffer)
	if err != nil && err.Error() != "EOF" {
		return err
	}
	
	return nil
}

// checkFileWritable tests if a file can be written to
func (dm *DirectoryManager) checkFileWritable(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()
	
	return nil
}

// testDirectoryWritable tests if a directory can be written to
func (dm *DirectoryManager) testDirectoryWritable(dirPath string) error {
	tempFile := filepath.Join(dirPath, ".write_test_temp")
	
	// Try to create a temporary file
	file, err := os.Create(tempFile)
	if err != nil {
		return err
	}
	
	// Close and remove the temporary file
	file.Close()
	os.Remove(tempFile)
	
	return nil
}