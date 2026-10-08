package domain

// AppError represents a business logic error with an HTTP status code.
type AppError struct {
	Code    int         `json:"-"`                 // HTTP status code (not serialized to JSON)
	Message string      `json:"message"`           // Human-readable error message for the client
	Details interface{} `json:"details,omitempty"` // Additional error details (omitted if empty)
}

// Error implements the error interface for AppError.
func (e *AppError) Error() string {
	return e.Message
}
