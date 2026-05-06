// Package shared provides domain-wide typed errors used across all layers.
package shared

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// StackTrace captures the call stack for error tracing
type StackTrace struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
}

// CaptureStackTrace captures the current call stack (skipping the first n frames)
func CaptureStackTrace(skip int) []StackTrace {
	var trace []StackTrace
	for i := skip; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		funcName := "unknown"
		if fn != nil {
			funcName = fn.Name()
			// Get just the function name without package path
			if idx := strings.LastIndex(funcName, "."); idx != -1 {
				funcName = funcName[idx+1:]
			}
		}
		trace = append(trace, StackTrace{
			File:     file,
			Line:     line,
			Function: funcName,
		})
	}
	return trace
}

// TracedError interface for errors with stack traces
type TracedError interface {
	error
	GetTrace() []StackTrace
}

// ErrorWithTrace wraps an error with stack trace information
type ErrorWithTrace struct {
	OriginalError error
	Trace         []StackTrace
	Context       map[string]interface{}
}

func (e *ErrorWithTrace) Error() string {
	return e.OriginalError.Error()
}

func (e *ErrorWithTrace) GetTrace() []StackTrace {
	return e.Trace
}

func (e *ErrorWithTrace) Unwrap() error {
	return e.OriginalError
}

// WrapWithTrace wraps an error with stack trace and context
func WrapWithTrace(err error, context map[string]interface{}) error {
	if err == nil {
		return nil
	}
	return &ErrorWithTrace{
		OriginalError: err,
		Trace:         CaptureStackTrace(2), // Skip WrapWithTrace and its caller
		Context:       context,
	}
}

// GetTrace retrieves stack trace from an error if available
func GetTrace(err error) []StackTrace {
	if traced, ok := err.(TracedError); ok {
		return traced.GetTrace()
	}
	return nil
}

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

// IsNotFound reports whether err or any error in its chain is an ErrNotFound.
func IsNotFound(err error) bool {
	var e ErrNotFound
	return errors.As(err, &e)
}

// IsAlreadyExists reports whether err or any error in its chain is an ErrAlreadyExists.
func IsAlreadyExists(err error) bool {
	var e ErrAlreadyExists
	return errors.As(err, &e)
}

// IsInvalidInput reports whether err or any error in its chain is an ErrInvalidInput.
func IsInvalidInput(err error) bool {
	var e ErrInvalidInput
	return errors.As(err, &e)
}

// IsUnauthorized reports whether err or any error in its chain is an ErrUnauthorized.
func IsUnauthorized(err error) bool {
	var e ErrUnauthorized
	return errors.As(err, &e)
}

// IsForbidden reports whether err or any error in its chain is an ErrForbidden.
func IsForbidden(err error) bool {
	var e ErrForbidden
	return errors.As(err, &e)
}

// IsConflict reports whether err or any error in its chain is an ErrConflict.
func IsConflict(err error) bool {
	var e ErrConflict
	return errors.As(err, &e)
}
