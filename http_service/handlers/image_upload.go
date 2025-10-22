package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/opentdp/go-helper/logman"
)

// Define the response structure for JSON outputs.
type ImageResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Image   string      `json:"image_name,omitempty"`
}

// ImageUploadResponse represents the response for image operations
type ImageUploadResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	ImageName string      `json:"image_name,omitempty"`
}

// UploadImageArchive handles uploading a podman image archive (.tar)
func UploadImageArchive(c echo.Context) error {
	// Parse multipart form (limit to 2GB)
	file, err := c.FormFile("image_file")
	if err != nil {
		logman.Error("Failed to get uploaded file", "error", err)
		return c.JSON(http.StatusBadRequest, &ImageResponse{
			Success: false,
			Message: "No image file provided",
		})
	}

	// Validate file extension
	if !strings.HasSuffix(file.Filename, ".tar") && !strings.HasSuffix(file.Filename, ".tar.gz") {
		return c.JSON(http.StatusBadRequest, &ImageResponse{
			Success: false,
			Message: "Only .tar and .tar.gz files are supported",
		})
	}

	// Get app name and image tag from form data
	appName := c.FormValue("app_name")
	if appName == "" {
		return c.JSON(http.StatusBadRequest, &ImageResponse{
			Success: false,
			Message: "App name is required",
		})
	}

	imageTag := c.FormValue("image_tag")
	if imageTag == "" {
		imageTag = "latest"
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		logman.Error("Failed to open uploaded file", "error", err)
		return c.JSON(http.StatusInternalServerError, &ImageResponse{
			Success: false,
			Message: "Failed to open uploaded file",
		})
	}
	defer src.Close()

	// Create temporary directory for upload
	tempDir := "/tmp/image-uploads"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		logman.Error("Failed to create temp directory", "error", err)
		return c.JSON(http.StatusInternalServerError, &ImageResponse{
			Success: false,
			Message: "Internal server error",
		})
	}

	// Save uploaded file to temporary location
	tempFilePath := filepath.Join(tempDir, fmt.Sprintf("%s_%d_%s", appName, time.Now().Unix(), file.Filename))
	dst, err := os.Create(tempFilePath)
	if err != nil {
		logman.Error("Failed to create temp file", "error", err)
		return c.JSON(http.StatusInternalServerError, &ImageResponse{
			Success: false,
			Message: "Failed to save file",
		})
	}
	defer dst.Close()
	defer os.Remove(tempFilePath) // Clean up temp file

	// Copy uploaded file to temp file
	if _, err := io.Copy(dst, src); err != nil {
		logman.Error("Failed to copy uploaded file", "error", err)
		return c.JSON(http.StatusInternalServerError, &ImageResponse{
			Success: false,
			Message: "Failed to save file",
		})
	}

	// Load the image using podman
	imageName := fmt.Sprintf("%s:%s", appName, imageTag)
	logman.Info("Loading image into podman", "image_name", imageName, "file", tempFilePath)

	cmd := exec.Command("podman", "load", "-i", tempFilePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logman.Error("Failed to load image with podman", "error", err, "output", string(output))
		return c.JSON(http.StatusInternalServerError, &ImageResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to load image: %s", string(output)),
		})
	}

	// Parse podman load output to get the actual image name
	loadedImageName := parseLoadedImageName(string(output))
	if loadedImageName == "" {
		loadedImageName = imageName
	}

	// Tag the image with our desired name if it's different
	if loadedImageName != imageName {
		tagCmd := exec.Command("podman", "tag", loadedImageName, imageName)
		if tagOutput, tagErr := tagCmd.CombinedOutput(); tagErr != nil {
			logman.Warn("Failed to tag image", "error", tagErr, "output", string(tagOutput))
			imageName = loadedImageName
		}
	}

	logman.Info("Image uploaded and loaded successfully", "app_name", appName, "image_name", imageName)
	return c.JSON(http.StatusOK, &ImageResponse{
		Success: true,
		Message: "Image uploaded successfully",
		Data: map[string]interface{}{
			"app_name":      appName,
			"image_name":    imageName,
			"upload_time":   time.Now(),
			"original_file": file.Filename,
		},
		Image: imageName,
	})
}

// BuildImageFromContext builds an image from uploaded build context
func BuildImageFromContext(c echo.Context) error {
	// Parse multipart form
	file, err := c.FormFile("context_file")
	if err != nil {
		logman.Error("Failed to get uploaded context file", "error", err)
		return c.JSON(http.StatusBadRequest, &ImageResponse{
			Success: false,
			Message: "No context file provided",
		})
	}

	// Get parameters
	appName := c.FormValue("app_name")
	if appName == "" {
		return c.JSON(http.StatusBadRequest, &ImageResponse{
			Success: false,
			Message: "App name is required",
		})
	}

	imageTag := c.FormValue("image_tag")
	if imageTag == "" {
		imageTag = "latest"
	}

	dockerfilePath := c.FormValue("dockerfile")
	if dockerfilePath == "" {
		dockerfilePath = "Dockerfile"
	}

	// Open the uploaded context file
	src, err := file.Open()
	if err != nil {
		logman.Error("Failed to open uploaded file", "error", err)
		return c.JSON(http.StatusInternalServerError, &ImageResponse{
			Success: false,
			Message: "Failed to open uploaded file",
		})
	}
	defer src.Close()

	// Create temporary directory for build context
	tempDir := "/tmp/build-contexts"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		logman.Error("Failed to create temp directory", "error", err)
		return c.JSON(http.StatusInternalServerError, &ImageResponse{
			Success: false,
			Message: "Internal server error",
		})
	}

	buildDir := filepath.Join(tempDir, fmt.Sprintf("%s_%d", appName, time.Now().Unix()))
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		logman.Error("Failed to create build directory", "error", err)
		return c.JSON(http.StatusInternalServerError, &ImageResponse{
			Success: false,
			Message: "Failed to create build directory",
		})
	}
	defer os.RemoveAll(buildDir) // Clean up build directory

	// Save and extract context file
	contextFilePath := filepath.Join(buildDir, file.Filename)
	dst, err := os.Create(contextFilePath)
	if err != nil {
		logman.Error("Failed to create context file", "error", err)
		return c.JSON(http.StatusInternalServerError, &ImageResponse{
			Success: false,
			Message: "Failed to save context file",
		})
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		logman.Error("Failed to copy context file", "error", err)
		return c.JSON(http.StatusInternalServerError, &ImageResponse{
			Success: false,
			Message: "Failed to save context file",
		})
	}

	// Extract the context file
	extractDir := filepath.Join(buildDir, "context")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		logman.Error("Failed to create extract directory", "error", err)
		return c.JSON(http.StatusInternalServerError, &ImageResponse{
			Success: false,
			Message: "Failed to create extract directory",
		})
	}

	var extractCmd *exec.Cmd
	if strings.HasSuffix(file.Filename, ".tar.gz") || strings.HasSuffix(file.Filename, ".tgz") {
		extractCmd = exec.Command("tar", "-xzf", contextFilePath, "-C", extractDir)
	} else if strings.HasSuffix(file.Filename, ".tar") {
		extractCmd = exec.Command("tar", "-xf", contextFilePath, "-C", extractDir)
	} else if strings.HasSuffix(file.Filename, ".zip") {
		extractCmd = exec.Command("unzip", "-q", contextFilePath, "-d", extractDir)
	} else {
		return c.JSON(http.StatusBadRequest, &ImageResponse{
			Success: false,
			Message: "Unsupported archive format. Use .tar, .tar.gz, or .zip",
		})
	}

	if extractOutput, extractErr := extractCmd.CombinedOutput(); extractErr != nil {
		logman.Error("Failed to extract context", "error", extractErr, "output", string(extractOutput))
		return c.JSON(http.StatusBadRequest, &ImageResponse{
			Success: false,
			Message: "Failed to extract build context",
		})
	}

	// Build the image using podman
	imageName := fmt.Sprintf("%s:%s", appName, imageTag)
	logman.Info("Building image with podman", "image_name", imageName, "dockerfile", dockerfilePath)

	buildCmd := exec.Command("podman", "build", "-t", imageName, "-f", dockerfilePath, extractDir)
	buildOutput, buildErr := buildCmd.CombinedOutput()
	if buildErr != nil {
		logman.Error("Failed to build image", "error", buildErr, "output", string(buildOutput))
		return c.JSON(http.StatusInternalServerError, &ImageResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to build image: %s", string(buildOutput)),
		})
	}

	logman.Info("Image built successfully", "app_name", appName, "image_name", imageName)
	return c.JSON(http.StatusOK, &ImageResponse{
		Success: true,
		Message: "Image built successfully",
		Data: map[string]interface{}{
			"app_name":     appName,
			"image_name":   imageName,
			"build_time":   time.Now(),
			"dockerfile":   dockerfilePath,
			"context_file": file.Filename,
		},
		Image: imageName,
	})
}

// ListLocalImages lists available local images
func ListLocalImages(c echo.Context) error {
	cmd := exec.Command("podman", "images", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		logman.Error("Failed to list images", "error", err)
		return c.JSON(http.StatusInternalServerError, &ImageResponse{
			Success: false,
			Message: "Failed to list images",
		})
	}

	// Parse JSON output
	var images []map[string]interface{}
	if err := json.Unmarshal(output, &images); err != nil {
		logman.Error("Failed to parse images JSON", "error", err)
		return c.JSON(http.StatusInternalServerError, &ImageResponse{
			Success: false,
			Message: "Failed to parse images",
		})
	}

	return c.JSON(http.StatusOK, &ImageResponse{
		Success: true,
		Message: "Images retrieved successfully",
		Data:    images,
	})
}

// DeleteLocalImage deletes a local image
func DeleteLocalImage(c echo.Context) error {
	imageName := c.Param("imageName")
	if imageName == "" {
		return c.JSON(http.StatusBadRequest, &ImageResponse{
			Success: false,
			Message: "Image name is required in URL",
		})
	}

	cmd := exec.Command("podman", "rmi", imageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logman.Error("Failed to delete image", "image", imageName, "error", err, "output", string(output))
		return c.JSON(http.StatusInternalServerError, &ImageResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to delete image: %s", string(output)),
		})
	}

	logman.Info("Image deleted successfully", "image", imageName)
	return c.JSON(http.StatusOK, &ImageResponse{
		Success: true,
		Message: "Image deleted successfully",
	})
}

// parseLoadedImageName extracts the image name from podman load output
func parseLoadedImageName(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Loaded image") {
			parts := strings.Split(line, ": ")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}
func sendImageUploadResponse(w http.ResponseWriter, statusCode int, success bool, message string, data interface{}, imageName string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ImageUploadResponse{
		Success:   success,
		Message:   message,
		Data:      data,
		ImageName: imageName,
	}

	json.NewEncoder(w).Encode(response)
}
