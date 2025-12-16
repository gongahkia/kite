package errors

import (
	"errors"
	"fmt"
)

// Common error variables
var (
	// Scraping Errors
	ErrNetworkFailure    = errors.New("network request failed")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrParsingFailure    = errors.New("failed to parse response")
	ErrRobotsDisallowed  = errors.New("robots.txt disallows scraping")
	ErrTimeout           = errors.New("request timeout")
	ErrInvalidResponse   = errors.New("invalid response from server")

	// Validation Errors
	ErrValidationFailed  = errors.New("validation failed")
	ErrInvalidData       = errors.New("invalid data")
	ErrMissingRequired   = errors.New("missing required field")
	ErrDuplicateEntry    = errors.New("duplicate entry detected")

	// Storage Errors
	ErrNotFound          = errors.New("resource not found")
	ErrAlreadyExists     = errors.New("resource already exists")
	ErrStorageFailure    = errors.New("storage operation failed")
	ErrDatabaseConnection = errors.New("database connection failed")

	// Queue Errors
	ErrQueueFull         = errors.New("queue is full")
	ErrQueueEmpty        = errors.New("queue is empty")
	ErrJobFailed         = errors.New("job execution failed")
	ErrWorkerUnavailable = errors.New("no workers available")

	// Authentication Errors
	ErrUnauthorized      = errors.New("unauthorized access")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired      = errors.New("authentication token expired")

	// Citation Errors
	ErrInvalidCitation   = errors.New("invalid citation format")
	ErrCitationNotFound  = errors.New("citation not found")

	// Configuration Errors
	ErrInvalidConfig     = errors.New("invalid configuration")
	ErrConfigNotFound    = errors.New("configuration not found")
)

// KiteError represents a custom error with additional context
type KiteError struct {
	Code    string
	Message string
	Err     error
	Context map[string]interface{}
}

// Error implements the error interface
func (e *KiteError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *KiteError) Unwrap() error {
	return e.Err
}

// WithContext adds context to the error
func (e *KiteError) WithContext(key string, value interface{}) *KiteError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// NewKiteError creates a new KiteError
func NewKiteError(code, message string, err error) *KiteError {
	return &KiteError{
		Code:    code,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// NetworkError creates a network-related error
func NetworkError(message string, err error) *KiteError {
	return NewKiteError("NETWORK_ERROR", message, err)
}

// RateLimitError creates a rate limit error
func RateLimitError(message string) *KiteError {
	return NewKiteError("RATE_LIMIT_ERROR", message, ErrRateLimitExceeded)
}

// ParsingError creates a parsing error
func ParsingError(message string, err error) *KiteError {
	return NewKiteError("PARSING_ERROR", message, err)
}

// ValidationError creates a validation error
func ValidationError(message string, err error) *KiteError {
	return NewKiteError("VALIDATION_ERROR", message, err)
}

// StorageError creates a storage error
func StorageError(message string, err error) *KiteError {
	return NewKiteError("STORAGE_ERROR", message, err)
}

// QueueError creates a queue error
func QueueError(message string, err error) *KiteError {
	return NewKiteError("QUEUE_ERROR", message, err)
}

// AuthError creates an authentication error
func AuthError(message string, err error) *KiteError {
	return NewKiteError("AUTH_ERROR", message, err)
}

// CitationError creates a citation-related error
func CitationError(message string, err error) *KiteError {
	return NewKiteError("CITATION_ERROR", message, err)
}

// ConfigError creates a configuration error
func ConfigError(message string, err error) *KiteError {
	return NewKiteError("CONFIG_ERROR", message, err)
}
