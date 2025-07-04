package errors

import (
	"fmt"
	"net/http"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected string
	}{
		{
			name: "simple error",
			err: &AppError{
				Code:    ErrorCodeValidation,
				Message: "Invalid input",
			},
			expected: "VALIDATION_ERROR: Invalid input",
		},
		{
			name: "error with cause",
			err: &AppError{
				Code:    ErrorCodeDatabase,
				Message: "Database error",
				Cause:   fmt.Errorf("connection failed"),
			},
			expected: "DATABASE_ERROR: Database error (caused by: connection failed)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("AppError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAppError_WithContext(t *testing.T) {
	err := New(ErrorCodeValidation, "Test error")
	err.WithContext("field", "username").WithContext("value", "invalid")

	if err.Context["field"] != "username" {
		t.Errorf("Expected context field to be 'username', got %v", err.Context["field"])
	}
	if err.Context["value"] != "invalid" {
		t.Errorf("Expected context value to be 'invalid', got %v", err.Context["value"])
	}
}

func TestAppError_WithDetails(t *testing.T) {
	err := New(ErrorCodeValidation, "Test error")
	details := "Username must be at least 3 characters"
	err.WithDetails(details)

	if err.Details != details {
		t.Errorf("Expected details to be %q, got %q", details, err.Details)
	}
}

func TestNew(t *testing.T) {
	err := New(ErrorCodeValidation, "Test message")

	if err.Code != ErrorCodeValidation {
		t.Errorf("Expected code to be %v, got %v", ErrorCodeValidation, err.Code)
	}
	if err.Message != "Test message" {
		t.Errorf("Expected message to be 'Test message', got %q", err.Message)
	}
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code to be %d, got %d", http.StatusBadRequest, err.StatusCode)
	}
}

func TestWrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	wrappedErr := Wrap(originalErr, ErrorCodeDatabase, "Database operation failed")

	if wrappedErr.Code != ErrorCodeDatabase {
		t.Errorf("Expected code to be %v, got %v", ErrorCodeDatabase, wrappedErr.Code)
	}
	if wrappedErr.Cause != originalErr {
		t.Errorf("Expected cause to be original error, got %v", wrappedErr.Cause)
	}
	if wrappedErr.Unwrap() != originalErr {
		t.Errorf("Expected Unwrap() to return original error, got %v", wrappedErr.Unwrap())
	}
}

func TestGetDefaultStatusCode(t *testing.T) {
	tests := []struct {
		code       ErrorCode
		statusCode int
	}{
		{ErrorCodeValidation, http.StatusBadRequest},
		{ErrorCodeAuthentication, http.StatusUnauthorized},
		{ErrorCodeAuthorization, http.StatusForbidden},
		{ErrorCodeNotFound, http.StatusNotFound},
		{ErrorCodeConflict, http.StatusConflict},
		{ErrorCodeRateLimit, http.StatusTooManyRequests},
		{ErrorCodeTimeout, http.StatusRequestTimeout},
		{ErrorCodeUnavailable, http.StatusServiceUnavailable},
		{ErrorCodeDatabase, http.StatusInternalServerError},
		{ErrorCodeInternal, http.StatusInternalServerError},
		{"UNKNOWN_CODE", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			got := getDefaultStatusCode(tt.code)
			if got != tt.statusCode {
				t.Errorf("getDefaultStatusCode(%v) = %v, want %v", tt.code, got, tt.statusCode)
			}
		})
	}
}

func TestPredefinedErrorConstructors(t *testing.T) {
	tests := []struct {
		name     string
		createFn func() *AppError
		code     ErrorCode
		status   int
	}{
		{
			name:     "ValidationError",
			createFn: func() *AppError { return NewValidationError("test validation") },
			code:     ErrorCodeValidation,
			status:   http.StatusBadRequest,
		},
		{
			name:     "AuthenticationError",
			createFn: func() *AppError { return NewAuthenticationError("test auth") },
			code:     ErrorCodeAuthentication,
			status:   http.StatusUnauthorized,
		},
		{
			name:     "AuthorizationError",
			createFn: func() *AppError { return NewAuthorizationError("test authz") },
			code:     ErrorCodeAuthorization,
			status:   http.StatusForbidden,
		},
		{
			name:     "NotFoundError",
			createFn: func() *AppError { return NewNotFoundError("user") },
			code:     ErrorCodeNotFound,
			status:   http.StatusNotFound,
		},
		{
			name:     "DatabaseError",
			createFn: func() *AppError { return NewDatabaseError(fmt.Errorf("db error")) },
			code:     ErrorCodeDatabase,
			status:   http.StatusInternalServerError,
		},
		{
			name:     "InternalError",
			createFn: func() *AppError { return NewInternalError(fmt.Errorf("internal error")) },
			code:     ErrorCodeInternal,
			status:   http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.createFn()
			if err.Code != tt.code {
				t.Errorf("Expected code %v, got %v", tt.code, err.Code)
			}
			if err.StatusCode != tt.status {
				t.Errorf("Expected status %d, got %d", tt.status, err.StatusCode)
			}
		})
	}
}

func TestIsAppError(t *testing.T) {
	appErr := New(ErrorCodeValidation, "test")
	stdErr := fmt.Errorf("standard error")

	if !IsAppError(appErr) {
		t.Error("IsAppError should return true for AppError")
	}
	if IsAppError(stdErr) {
		t.Error("IsAppError should return false for standard error")
	}
}

func TestAsAppError(t *testing.T) {
	appErr := New(ErrorCodeValidation, "test")
	stdErr := fmt.Errorf("standard error")

	// Test with AppError
	converted, ok := AsAppError(appErr)
	if !ok {
		t.Error("AsAppError should return true for AppError")
	}
	if converted != appErr {
		t.Error("AsAppError should return the same AppError")
	}

	// Test with standard error
	_, ok = AsAppError(stdErr)
	if ok {
		t.Error("AsAppError should return false for standard error")
	}
}

func TestGetStatusCode(t *testing.T) {
	appErr := New(ErrorCodeValidation, "test")
	stdErr := fmt.Errorf("standard error")

	statusCode := GetStatusCode(appErr)
	if statusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, statusCode)
	}

	statusCode = GetStatusCode(stdErr)
	if statusCode != http.StatusInternalServerError {
		t.Errorf("Expected default status code %d, got %d", http.StatusInternalServerError, statusCode)
	}
}
