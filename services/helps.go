package services

import (
	"fmt"
	"os"
	"strings"
)

// ResolveSystemdPath resolves systemd path specifiers to actual filesystem paths
// For backward compatibility, this function only handles %h 
func ResolveSystemdPath(pathWithSpecifier string) (string, error) {
	result := pathWithSpecifier
	
	// Handle %h - user home directory
	if strings.Contains(result, "%h") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not get user home directory: %w", err)
		}
		result = strings.ReplaceAll(result, "%h", homeDir)
	}
	
	return result, nil
}

// ResolveSystemdPathWithContext resolves systemd path specifiers with container context
func ResolveSystemdPathWithContext(pathWithSpecifier string, containerName string) (string, error) {
	result := pathWithSpecifier
	
	// Handle %h - user home directory
	if strings.Contains(result, "%h") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not get user home directory: %w", err)
		}
		result = strings.ReplaceAll(result, "%h", homeDir)
	}
	
	// Handle %s - service/container name (in systemd context)
	if strings.Contains(result, "%s") && containerName != "" {
		result = strings.ReplaceAll(result, "%s", containerName)
	}
	
	// Handle %i - instance name (similar to %s, use container name)
	if strings.Contains(result, "%i") && containerName != "" {
		result = strings.ReplaceAll(result, "%i", containerName)
	}
	
	return result, nil
}
