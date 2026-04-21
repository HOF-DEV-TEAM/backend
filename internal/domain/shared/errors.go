package shared

import "fmt"

// ErrNotFound is returned when a requested resource does not exist.
type ErrNotFound struct {
	Resource string
	ID       string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s with id '%s' not found", e.Resource, e.ID)
}

// ErrAlreadyExists is returned when a resource being created already exists.
type ErrAlreadyExists struct {
	Resource string
	Field    string
	Value    string
}

func (e ErrAlreadyExists) Error() string {
	return fmt.Sprintf("%s with %s '%s' already exists", e.Resource, e.Field, e.Value)
}

// ErrInvalidInput is returned when the caller provides invalid data.
type ErrInvalidInput struct {
	Field   string
	Message string
}

func (e ErrInvalidInput) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("invalid input for field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("invalid input: %s", e.Message)
}

// ErrUnauthorized is returned when an operation is attempted without valid credentials.
type ErrUnauthorized struct {
	Message string
}

func (e ErrUnauthorized) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "unauthorized"
}

// ErrForbidden is returned when the caller lacks permission for the operation.
type ErrForbidden struct {
	Message string
}

func (e ErrForbidden) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "forbidden"
}

// ErrConflict is returned when a state conflict prevents the operation.
type ErrConflict struct {
	Message string
}

func (e ErrConflict) Error() string {
	return e.Message
}

// IsNotFound reports whether err is an ErrNotFound.
func IsNotFound(err error) bool {
	_, ok := err.(ErrNotFound)
	return ok
}

// IsAlreadyExists reports whether err is an ErrAlreadyExists.
func IsAlreadyExists(err error) bool {
	_, ok := err.(ErrAlreadyExists)
	return ok
}

// IsInvalidInput reports whether err is an ErrInvalidInput.
func IsInvalidInput(err error) bool {
	_, ok := err.(ErrInvalidInput)
	return ok
}

// IsUnauthorized reports whether err is an ErrUnauthorized.
func IsUnauthorized(err error) bool {
	_, ok := err.(ErrUnauthorized)
	return ok
}

// IsForbidden reports whether err is an ErrForbidden.
func IsForbidden(err error) bool {
	_, ok := err.(ErrForbidden)
	return ok
}
