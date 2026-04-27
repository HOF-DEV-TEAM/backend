package shared_test

import (
	"testing"

	"bitbucket.org/hofng/hofApp/internal/domain/shared"
)

// ── ErrNotFound ───────────────────────────────────────────────────────────────

func TestErrNotFound_Error(t *testing.T) {
	err := shared.ErrNotFound{Resource: "user", ID: "abc-123"}
	want := "user with id 'abc-123' not found"
	if err.Error() != want {
		t.Errorf("ErrNotFound.Error() = %q, want %q", err.Error(), want)
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"ErrNotFound", shared.ErrNotFound{Resource: "plan", ID: "x"}, true},
		{"ErrInvalidInput", shared.ErrInvalidInput{Message: "bad"}, false},
		{"ErrAlreadyExists", shared.ErrAlreadyExists{Resource: "user", Field: "email", Value: "x"}, false},
		{"ErrUnauthorized", shared.ErrUnauthorized{Message: "no"}, false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shared.IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// ── ErrAlreadyExists ─────────────────────────────────────────────────────────

func TestErrAlreadyExists_Error(t *testing.T) {
	err := shared.ErrAlreadyExists{Resource: "user", Field: "email", Value: "test@example.com"}
	want := "user with email 'test@example.com' already exists"
	if err.Error() != want {
		t.Errorf("ErrAlreadyExists.Error() = %q, want %q", err.Error(), want)
	}
}

func TestIsAlreadyExists(t *testing.T) {
	if !shared.IsAlreadyExists(shared.ErrAlreadyExists{Resource: "user", Field: "email", Value: "x"}) {
		t.Error("IsAlreadyExists should return true for ErrAlreadyExists")
	}
	if shared.IsAlreadyExists(shared.ErrNotFound{Resource: "user", ID: "x"}) {
		t.Error("IsAlreadyExists should return false for ErrNotFound")
	}
}

// ── ErrInvalidInput ───────────────────────────────────────────────────────────

func TestErrInvalidInput_Error_WithField(t *testing.T) {
	err := shared.ErrInvalidInput{Field: "email", Message: "must be valid"}
	want := "invalid input for field 'email': must be valid"
	if err.Error() != want {
		t.Errorf("ErrInvalidInput.Error() = %q, want %q", err.Error(), want)
	}
}

func TestErrInvalidInput_Error_NoField(t *testing.T) {
	err := shared.ErrInvalidInput{Message: "passwords do not match"}
	want := "invalid input: passwords do not match"
	if err.Error() != want {
		t.Errorf("ErrInvalidInput.Error() = %q, want %q", err.Error(), want)
	}
}

func TestIsInvalidInput(t *testing.T) {
	if !shared.IsInvalidInput(shared.ErrInvalidInput{Message: "bad"}) {
		t.Error("IsInvalidInput should return true for ErrInvalidInput")
	}
	if shared.IsInvalidInput(shared.ErrNotFound{Resource: "x", ID: "y"}) {
		t.Error("IsInvalidInput should return false for ErrNotFound")
	}
}

// ── ErrUnauthorized ───────────────────────────────────────────────────────────

func TestErrUnauthorized_Error_WithMessage(t *testing.T) {
	err := shared.ErrUnauthorized{Message: "token expired"}
	if err.Error() != "token expired" {
		t.Errorf("ErrUnauthorized.Error() = %q, want %q", err.Error(), "token expired")
	}
}

func TestErrUnauthorized_Error_NoMessage(t *testing.T) {
	err := shared.ErrUnauthorized{}
	if err.Error() != "unauthorized" {
		t.Errorf("ErrUnauthorized.Error() = %q, want %q", err.Error(), "unauthorized")
	}
}

func TestIsUnauthorized(t *testing.T) {
	if !shared.IsUnauthorized(shared.ErrUnauthorized{Message: "no"}) {
		t.Error("IsUnauthorized should return true for ErrUnauthorized")
	}
	if shared.IsUnauthorized(shared.ErrForbidden{Message: "no"}) {
		t.Error("IsUnauthorized should return false for ErrForbidden")
	}
}

// ── ErrForbidden ──────────────────────────────────────────────────────────────

func TestErrForbidden_Error_WithMessage(t *testing.T) {
	err := shared.ErrForbidden{Message: "admin only"}
	if err.Error() != "admin only" {
		t.Errorf("ErrForbidden.Error() = %q, want %q", err.Error(), "admin only")
	}
}

func TestErrForbidden_Error_NoMessage(t *testing.T) {
	err := shared.ErrForbidden{}
	if err.Error() != "forbidden" {
		t.Errorf("ErrForbidden.Error() = %q, want %q", err.Error(), "forbidden")
	}
}

func TestIsForbidden(t *testing.T) {
	if !shared.IsForbidden(shared.ErrForbidden{Message: "no"}) {
		t.Error("IsForbidden should return true for ErrForbidden")
	}
	if shared.IsForbidden(shared.ErrUnauthorized{Message: "no"}) {
		t.Error("IsForbidden should return false for ErrUnauthorized")
	}
}

// ── ErrConflict ───────────────────────────────────────────────────────────────

func TestErrConflict_Error(t *testing.T) {
	err := shared.ErrConflict{Message: "state conflict"}
	if err.Error() != "state conflict" {
		t.Errorf("ErrConflict.Error() = %q, want %q", err.Error(), "state conflict")
	}
}
