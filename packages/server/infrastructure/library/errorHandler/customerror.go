package errorHandler

import "encoding/json"

// Herror defines an HOF error structure
type Herror struct {
	ErrorCode     int    `json:"code"`
	TypeError     string `json:"error_type"`
	ErrorMessage  string `json:"message"`
	ErrorStatus   int    `json:"status"`
	ErrorDetails  string `json:"detail"`
	ErrorTraceID  string `json:"trace_id"`
	ErrorInstance string `json:"instance"`
	ErrorHelp     string `json:"help"`
}

// Ensure Herror implements error interface
var _ error = &Herror{}

// Code getter
func (t *Herror) Code() int {
	return t.ErrorCode
}

// ErrorType getter
func (t *Herror) ErrorType() string {
	return t.TypeError
}

// Message getter
func (t *Herror) Message() string {
	return t.ErrorMessage
}

// Status getter
func (t *Herror) Status() int {
	return t.ErrorStatus
}

// Detail getter
func (t *Herror) Detail() string {
	return t.ErrorDetails
}

// TraceID getter
func (t *Herror) TraceID() string {
	return t.ErrorTraceID
}

// Instance getter
func (t *Herror) Instance() string {
	return t.ErrorInstance
}

// Help getter
func (t *Herror) Help() string {
	return t.ErrorHelp
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
			Code:      t.ErrorCode,
			ErrorType: t.TypeError,
			Message:   t.ErrorMessage,
			Status:    t.ErrorStatus,
			Detail:    t.ErrorDetails,
			TraceID:   t.ErrorTraceID,
			Instance:  t.ErrorInstance,
			Help:      t.ErrorHelp,
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
		ErrorCode:    code,
		TypeError:    errorType,
		ErrorMessage: message,
		ErrorDetails: detail,
	}
	return herror
}
