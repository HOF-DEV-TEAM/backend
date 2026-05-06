package response_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"bitbucket.org/hofng/hofApp/internal/domain/shared"
	"bitbucket.org/hofng/hofApp/internal/interfaces/http/response"
)

type envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Total   *int64      `json:"total,omitempty"`
}

func decode(t *testing.T, w *httptest.ResponseRecorder) envelope {
	t.Helper()
	var e envelope
	if err := json.NewDecoder(w.Body).Decode(&e); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	return e
}

// ── JSON ──────────────────────────────────────────────────────────────────────

func TestJSON_Success(t *testing.T) {
	w := httptest.NewRecorder()
	response.JSON(w, http.StatusOK, map[string]string{"message": "ok"})

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	e := decode(t, w)
	if !e.Success {
		t.Error("expected success:true")
	}
	if e.Error != "" {
		t.Errorf("expected no error, got %q", e.Error)
	}
}

func TestJSON_Created(t *testing.T) {
	w := httptest.NewRecorder()
	response.JSON(w, http.StatusCreated, map[string]string{"id": "123"})
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
}

func TestJSON_ContentType(t *testing.T) {
	w := httptest.NewRecorder()
	response.JSON(w, http.StatusOK, nil)
	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
}

// ── JSONList ──────────────────────────────────────────────────────────────────

func TestJSONList(t *testing.T) {
	w := httptest.NewRecorder()
	var total int64 = 42
	response.JSONList(w, http.StatusOK, []string{"a", "b"}, total)

	e := decode(t, w)
	if !e.Success {
		t.Error("expected success:true")
	}
	if e.Total == nil || *e.Total != 42 {
		t.Errorf("expected total=42, got %v", e.Total)
	}
}

// ── Error mapping ─────────────────────────────────────────────────────────────

func TestError_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	response.Error(w, shared.ErrNotFound{Resource: "user", ID: "xyz"})

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
	e := decode(t, w)
	if e.Success {
		t.Error("expected success:false")
	}
	if e.Error == "" {
		t.Error("expected non-empty error message")
	}
}

func TestError_AlreadyExists(t *testing.T) {
	w := httptest.NewRecorder()
	response.Error(w, shared.ErrAlreadyExists{Resource: "user", Field: "email", Value: "x@x.com"})
	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", w.Code, http.StatusConflict)
	}
}

func TestError_InvalidInput(t *testing.T) {
	w := httptest.NewRecorder()
	response.Error(w, shared.ErrInvalidInput{Message: "passwords do not match"})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestError_Unauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	response.Error(w, shared.ErrUnauthorized{Message: "token expired"})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestError_Forbidden(t *testing.T) {
	w := httptest.NewRecorder()
	response.Error(w, shared.ErrForbidden{Message: "admin only"})
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestError_Conflict(t *testing.T) {
	w := httptest.NewRecorder()
	response.Error(w, shared.ErrConflict{Message: "some internal conflict"})
	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", w.Code, http.StatusConflict)
	}
	e := decode(t, w)
	if e.Error != "some internal conflict" {
		t.Errorf("error = %q, want %q", e.Error, "some internal conflict")
	}
}

func TestError_UnknownError_Maps500(t *testing.T) {
	w := httptest.NewRecorder()
	response.Error(w, fmt.Errorf("raw untyped error"))
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	e := decode(t, w)
	if e.Error != "an unexpected error occurred" {
		t.Errorf("error = %q, want %q", e.Error, "an unexpected error occurred")
	}
}

// ── Helper responses ──────────────────────────────────────────────────────────

func TestBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	response.BadRequest(w, "invalid request body")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	e := decode(t, w)
	if e.Error != "invalid request body" {
		t.Errorf("error = %q, want %q", e.Error, "invalid request body")
	}
}

func TestUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	response.Unauthorized(w)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	response.NotFound(w)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
