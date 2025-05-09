// internal/util/response.go

package util

import (
	"encoding/json"
	"net/http"
)

// Response is the standard API response structure
type Response struct {
	Success     bool         `json:"success"`
	Data        interface{}  `json:"data,omitempty"`
	GameMessage *GameMessage `json:"gameMessage,omitempty"`
	Error       *ErrorInfo   `json:"error,omitempty"`
}

// ErrorInfo contains detailed error information
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// GameMessage is used for gameplay-related messages
type GameMessage struct {
	Type    string `json:"type"` // "success", "warning", "error", "info"
	Message string `json:"message"`
}

// RespondWithJSON sends a JSON response
func RespondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	response := Response{
		Success: statusCode >= 200 && statusCode < 300,
		Data:    payload,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// RespondWithError sends an error JSON response
func RespondWithError(w http.ResponseWriter, statusCode int, message string) {
	errorCode := "unknown_error"

	// Map common status codes to error codes
	switch statusCode {
	case http.StatusBadRequest:
		errorCode = "bad_request"
	case http.StatusUnauthorized:
		errorCode = "unauthorized"
	case http.StatusForbidden:
		errorCode = "forbidden"
	case http.StatusNotFound:
		errorCode = "not_found"
	case http.StatusConflict:
		errorCode = "conflict"
	case http.StatusInternalServerError:
		errorCode = "internal_error"
	}

	response := Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    errorCode,
			Message: message,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// RespondWithGameMessage sends a response with a game message
func RespondWithGameMessage(w http.ResponseWriter, statusCode int, data interface{}, messageType string, message string) {
	response := Response{
		Success: statusCode >= 200 && statusCode < 300,
		Data:    data,
	}

	if messageType != "" && message != "" {
		response.GameMessage = &GameMessage{
			Type:    messageType,
			Message: message,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
