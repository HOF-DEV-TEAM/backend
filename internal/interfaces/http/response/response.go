// Package response provides standard JSON HTTP response helpers.
package response

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/hofng/hofApp/internal/domain/shared"
)

// envelope is the standard JSON wrapper for all API responses.
type envelope struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Total   *int64 `json:"total,omitempty"`
}

// JSON writes a success JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, data any) {
	write(w, status, envelope{Success: true, Data: data})
}

// JSONList writes a paginated JSON response.
func JSONList(w http.ResponseWriter, status int, data any, total int64) {
	write(w, status, envelope{Success: true, Data: data, Total: &total})
}

// Error writes a JSON error response, mapping domain errors to HTTP status codes.
func Error(w http.ResponseWriter, err error) {
	status, message := classify(err)
	write(w, status, envelope{Success: false, Error: message})
}

// BadRequest writes a 400 JSON error.
func BadRequest(w http.ResponseWriter, msg string) {
	write(w, http.StatusBadRequest, envelope{Success: false, Error: msg})
}

// Unauthorized writes a 401 JSON error.
func Unauthorized(w http.ResponseWriter) {
	write(w, http.StatusUnauthorized, envelope{Success: false, Error: "unauthorized"})
}

// NotFound writes a 404 JSON error.
func NotFound(w http.ResponseWriter) {
	write(w, http.StatusNotFound, envelope{Success: false, Error: "not found"})
}

func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func classify(err error) (status int, message string) {
	switch {
	case shared.IsNotFound(err):
		return http.StatusNotFound, err.Error()
	case shared.IsAlreadyExists(err):
		return http.StatusConflict, err.Error()
	case shared.IsInvalidInput(err):
		return http.StatusBadRequest, err.Error()
	case shared.IsUnauthorized(err):
		return http.StatusUnauthorized, err.Error()
	case shared.IsForbidden(err):
		return http.StatusForbidden, err.Error()
	default:
		return http.StatusInternalServerError, "an unexpected error occurred"
	}
}
