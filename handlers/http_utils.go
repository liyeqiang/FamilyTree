package handlers

import (
	"encoding/json"
	"net/http"
)

// WriteErrorResponse 写入错误响应
func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"error":  true,
		"message": message,
		"status": statusCode,
	}
	
	json.NewEncoder(w).Encode(response)
}

// WriteJSONResponse 写入JSON响应
func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// WriteSuccessResponse 写入成功响应
func WriteSuccessResponse(w http.ResponseWriter, message string) {
	WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": message,
	})
} 