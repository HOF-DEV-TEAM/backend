// Package response provides standard JSON HTTP response helpers.
package response

import (
	"encoding/json"
	"errors"
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

// Forbidden writes a 403 JSON error.
func Forbidden(w http.ResponseWriter) {
	write(w, http.StatusForbidden, envelope{Success: false, Error: "forbidden"})
}

func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func classify(err error) (status int, message string) {
	if v, ok := errors.AsType[shared.ErrNotFound](err); ok {
		return http.StatusNotFound, v.Error()
	}
	if v, ok := errors.AsType[shared.ErrAlreadyExists](err); ok {
		return http.StatusConflict, v.Error()
	}
	if v, ok := errors.AsType[shared.ErrConflict](err); ok {
		return http.StatusConflict, v.Error()
	}
	if v, ok := errors.AsType[shared.ErrInvalidInput](err); ok {
		return http.StatusBadRequest, v.Error()
	}
	if v, ok := errors.AsType[shared.ErrUnauthorized](err); ok {
		return http.StatusUnauthorized, v.Error()
	}
	if v, ok := errors.AsType[shared.ErrForbidden](err); ok {
		return http.StatusForbidden, v.Error()
	}
	return http.StatusInternalServerError, "an unexpected error occurred"
}
