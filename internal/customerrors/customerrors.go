package customerrors

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrKeyRequired           = new(ErrKeyRequiredCode, "key is required")
	ErrValueRequired         = new(ErrValueRequiredCode, "value is required")
	ErrMalformedSearchString = new(ErrMalformedSearchStringCode, "malformed search string")
	ErrKeyNotFound           = new(ErrKeyNotFoundCode, "key not found")
)

var statusCodeMap = map[error]int{
	ErrKeyRequired:           http.StatusBadRequest,
	ErrValueRequired:         http.StatusBadRequest,
	ErrMalformedSearchString: http.StatusBadRequest,
	ErrKeyNotFound:           http.StatusNotFound,
}

const (
	ErrKeyRequiredCode           = "KEY_REQUIRED"
	ErrValueRequiredCode         = "VALUE_REQUIRED"
	ErrMalformedSearchStringCode = "MALFORMED_SEARCH_STRING"
	ErrKeyNotFoundCode           = "KEY_NOT_FOUND"
)

// InternalError represents a domain error
type InternalError struct {
	Code    string // Machine-readable error code
	Message string // Human-readable error message
	Op      string // Logical operation name
	Err     error  // Underlying error
}

func (e *InternalError) Error() string {
	if e.Err == nil {
		return e.DisplayError()
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Err.Error())
}

func (e *InternalError) DisplayError() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// New creates a new InternalError
func new(code string, message string) *InternalError {
	return &InternalError{
		Code:    code,
		Message: message,
	}
}

func HTTPStatusFromErr(err error) int {
	for e, status := range statusCodeMap {
		if errors.Is(err, e) {
			return status
		}
	}
	return http.StatusInternalServerError
}
