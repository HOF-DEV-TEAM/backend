package errorHandler

import "encoding/json"

// Herror defines an HOF error structure
type Herror struct {
	code      int
	errorType string
	message   string
	status    int
	detail    string
	traceID   string
	instance  string
	help      string
}

// Ensure Herror implements error interface
var _ error = &Herror{}

// Code getter
func (t *Herror) Code() int {
	return t.code
}

// ErrorType getter
func (t *Herror) ErrorType() string {
	return t.errorType
}

// Message getter
func (t *Herror) Message() string {
	return t.message
}

// Status getter
func (t *Herror) Status() int {
	return t.status
}

// Detail getter
func (t *Herror) Detail() string {
	return t.detail
}

// TraceID getter
func (t *Herror) TraceID() string {
	return t.traceID
}

// Instance getter
func (t *Herror) Instance() string {
	return t.instance
}

// Help getter
func (t *Herror) Help() string {
	return t.help
}

// herrorError represents the json sharable model of a HOF Herror

type herrorError struct {
	Code      int    `json:"code"`
	ErrorType string `json:"type"`
	Message   string `json:"message"`
	Status    int    `json:"status,omitempty"`
	Detail    string `json:"detail"`
	TraceID   string `json:"trace_id,omitempty"`
	Instance  string `json:"instance,omitempty"`
	Help      string `json:"help,omitempty"`
}

// Error returns a json string representation of the HOF Herror
func (t *Herror) Error() string {
	// Get model json string
	jsonBytes, _ := json.Marshal(
		herrorError{
			Code:      t.code,
			ErrorType: t.errorType,
			Message:   t.message,
			Status:    t.status,
			Detail:    t.detail,
			TraceID:   t.traceID,
			Instance:  t.instance,
			Help:      t.help,
		})

	return string(jsonBytes)
}

// NewHerror returns an instance of a Herror with the given attributes
func NewHerror(
	code int,
	errorType string,
	message string,
	detail string,
) *Herror {
	herror := &Herror{
		code:      code,
		errorType: errorType,
		message:   message,
		detail:    detail,
	}
	return herror
}
