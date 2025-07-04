package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents application-specific error codes
type ErrorCode string

const (
	// Client errors (4xx)
	ErrorCodeValidation     ErrorCode = "VALIDATION_ERROR"
	ErrorCodeAuthentication ErrorCode = "AUTHENTICATION_ERROR"
	ErrorCodeAuthorization  ErrorCode = "AUTHORIZATION_ERROR"
	ErrorCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrorCodeConflict       ErrorCode = "CONFLICT"
	ErrorCodeRateLimit      ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrorCodeBadRequest     ErrorCode = "BAD_REQUEST"

	// Server errors (5xx)
	ErrorCodeInternal        ErrorCode = "INTERNAL_ERROR"
	ErrorCodeDatabase        ErrorCode = "DATABASE_ERROR"
	ErrorCodeExternalService ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrorCodeTimeout         ErrorCode = "TIMEOUT"
	ErrorCodeUnavailable     ErrorCode = "SERVICE_UNAVAILABLE"
)

// AppError represents a structured application error
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	StatusCode int                    `json:"-"`
	Cause      error                  `json:"-"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithContext adds context information to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithDetails adds detailed information to the error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// New creates a new AppError
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: getDefaultStatusCode(code),
	}
}

// Wrap wraps an existing error with application error information
func Wrap(err error, code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: getDefaultStatusCode(code),
		Cause:      err,
	}
}

// WrapWithStatus wraps an error with a custom HTTP status code
func WrapWithStatus(err error, code ErrorCode, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Cause:      err,
	}
}

// getDefaultStatusCode returns the default HTTP status code for an error code
func getDefaultStatusCode(code ErrorCode) int {
	switch code {
	case ErrorCodeValidation, ErrorCodeBadRequest:
		return http.StatusBadRequest
	case ErrorCodeAuthentication:
		return http.StatusUnauthorized
	case ErrorCodeAuthorization:
		return http.StatusForbidden
	case ErrorCodeNotFound:
		return http.StatusNotFound
	case ErrorCodeConflict:
		return http.StatusConflict
	case ErrorCodeRateLimit:
		return http.StatusTooManyRequests
	case ErrorCodeTimeout:
		return http.StatusRequestTimeout
	case ErrorCodeUnavailable:
		return http.StatusServiceUnavailable
	case ErrorCodeDatabase, ErrorCodeExternalService, ErrorCodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// Predefined error constructors for common scenarios

// NewValidationError creates a validation error
func NewValidationError(message string) *AppError {
	return New(ErrorCodeValidation, message)
}

// NewAuthenticationError creates an authentication error
func NewAuthenticationError(message string) *AppError {
	return New(ErrorCodeAuthentication, message)
}

// NewAuthorizationError creates an authorization error
func NewAuthorizationError(message string) *AppError {
	return New(ErrorCodeAuthorization, message)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return New(ErrorCodeNotFound, fmt.Sprintf("%s not found", resource))
}

// NewDatabaseError wraps a database error
func NewDatabaseError(err error) *AppError {
	return Wrap(err, ErrorCodeDatabase, "Database operation failed")
}

// NewInternalError wraps an internal error
func NewInternalError(err error) *AppError {
	return Wrap(err, ErrorCodeInternal, "Internal server error")
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// AsAppError attempts to convert an error to AppError
func AsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}

// GetStatusCode extracts HTTP status code from an error
func GetStatusCode(err error) int {
	if appErr, ok := AsAppError(err); ok {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}
