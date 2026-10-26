package errors

import "errors"

// AppError represents a business logic error with an HTTP status code.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`           // Human-readable error message for the client
	Details any    `json:"details,omitempty"` // Additional error details (omitted if empty)
	Err     error  `json:"-"`                 // Internal error for wrapping
}

// Error implements the error interface for AppError.
func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap allows errors.As/Is to work with wrapped errors.
func (e *AppError) Unwrap() error {
	return e.Err
}

// Wrap wraps an error with AppError, preserving the original error.
func Wrap(err error, code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Sentinel errors for domain-specific errors (no HTTP code)
var (
	ErrNotFound           = errors.New("not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrConflict           = errors.New("conflict")
	ErrValidation         = errors.New("validation failed")
	ErrBadGateway         = errors.New("bad gateway")
	ErrServiceUnavailable = errors.New("service unavailable")
)
