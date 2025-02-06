package utils

import (
	"banking-system/internal/database/models"
	"encoding/json"
	"net/http"
)

// WriteJSONError writes an error response in JSON format
func WriteJSONError(w http.ResponseWriter, statusCode int, message string, err error) {
	response := models.Response{
		StatusCode: statusCode,
		Success:    false,
		Message:    message,
		Data: map[string]interface{}{
			"error": err.Error(),
		},
	}
	WriteJSONResponse(w, statusCode, message, response)
}

// WriteJSONResponse writes a JSON response with the given status code, message, and data
func WriteJSONResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	response := models.Response{
		StatusCode: statusCode,
		Success:    statusCode < 400,
		Message:    message,
		Data:       data,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
} 