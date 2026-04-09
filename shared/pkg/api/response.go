package api

import (
	"encoding/json"
	"net/http"

	"github.com/quality-gateway/shared/pkg/errors"
)

// Response represents a standard API response.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo represents error information in API responses.
type ErrorInfo struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// JSON writes a JSON response.
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(Response{
		Success: statusCode >= 200 && statusCode < 300,
		Data:    data,
	})
}

// OK writes a successful JSON response with data.
func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

// Created writes a 201 Created response.
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// NoContent writes a 204 No Content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// BadRequest writes a 400 Bad Request response.
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized writes a 401 Unauthorized response.
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden writes a 403 Forbidden response.
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound writes a 404 Not Found response.
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, "NOT_FOUND", message)
}

// Conflict writes a 409 Conflict response.
func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, "CONFLICT", message)
}

// InternalError writes a 500 Internal Server Error response.
func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// Error writes an error response.
func Error(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	})
}

// FromAppError writes an error response from an AppError.
func FromAppError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		statusCode := statusCodeFromCode(appErr.Code)
		Error(w, statusCode, appErr.Code, appErr.Message)
		return
	}
	InternalError(w, "An internal error occurred")
}

// statusCodeFromCode maps error codes to HTTP status codes.
func statusCodeFromCode(code string) int {
	switch code {
	case "NOT_FOUND":
		return http.StatusNotFound
	case "ALREADY_EXISTS":
		return http.StatusConflict
	case "INVALID_INPUT":
		return http.StatusBadRequest
	case "UNAUTHORIZED":
		return http.StatusUnauthorized
	case "FORBIDDEN":
		return http.StatusForbidden
	case "CONFLICT":
		return http.StatusConflict
	case "TIMEOUT":
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

// HandlerFunc is a function that returns an error.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Handle wraps a HandlerFunc and handles errors.
func Handle(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			HandleError(w, r, err)
		}
	}
}

// HandleError handles an error and writes an appropriate response.
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		statusCode := statusCodeFromCode(appErr.Code)
		Error(w, statusCode, appErr.Code, appErr.Message)
		return
	}
	InternalError(w, "An internal error occurred")
}

// DecodeJSON decodes JSON from a request.
func DecodeJSON(r *http.Request, dest interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dest)
}

// WriteJSON writes JSON to a response.
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}
