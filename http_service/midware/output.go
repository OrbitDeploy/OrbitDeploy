package midware

import (
	"encoding/json"
	"log"
	"net/http"
)

// SuccessResponse sends a unified success response
func SuccessResponse(w http.ResponseWriter, data interface{}, message ...string) {
	w.Header().Set("Content-Type", "application/json")
	
	response := map[string]interface{}{
		"success": true,
	}
	
	if data != nil {
		response["data"] = data
	}
	
	if len(message) > 0 && message[0] != "" {
		response["message"] = message[0]
	}
	
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode success response: %v", err)
	}
}

// ErrorResponse sends a unified error response
func ErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"success": false,
		"message": message,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

// CreatedResponse sends a unified response for created resources
func CreatedResponse(w http.ResponseWriter, data interface{}, message ...string) {
	w.Header().Set("Content-Type", "application/json")
	
	response := map[string]interface{}{
		"success": true,
	}
	
	if data != nil {
		response["data"] = data
	}
	
	if len(message) > 0 && message[0] != "" {
		response["message"] = message[0]
	}
	
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode created response: %v", err)
	}
}

// AuthSuccessResponse sends a unified success response for authentication
func AuthSuccessResponse(w http.ResponseWriter, user interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"success": true,
		"message": message,
	}
	
	if user != nil {
		response["user"] = user
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode auth success response: %v", err)
	}
}

// AuthCreatedResponse sends a unified response for user creation
func AuthCreatedResponse(w http.ResponseWriter, user interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	
	response := map[string]interface{}{
		"success": true,
		"message": message,
	}
	
	if user != nil {
		response["user"] = user
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode auth created response: %v", err)
	}
}
