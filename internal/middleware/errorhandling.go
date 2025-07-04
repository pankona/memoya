package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/pankona/memoya/internal/errors"
	"github.com/pankona/memoya/internal/logging"
)

// ErrorResponse represents the standardized error response format
type ErrorResponse struct {
	Error   ErrorDetail `json:"error"`
	Success bool        `json:"success"`
}

// ErrorDetail contains detailed error information
type ErrorDetail struct {
	Code    errors.ErrorCode       `json:"code"`
	Message string                 `json:"message"`
	Details string                 `json:"details,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// GetLoggerFromContext extracts the logger from request context
func GetLoggerFromContext(ctx context.Context) *logging.Logger {
	if logger, ok := ctx.Value("logger").(*logging.Logger); ok {
		return logger
	}
	return logging.Default()
}

// ErrorHandler middleware catches panics and formats errors consistently
func ErrorHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger := GetLoggerFromContext(r.Context())
					logger.Error("Panic occurred",
						slog.Any("panic", err),
						slog.String("stack_trace", string(debug.Stack())),
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
					)

					// Create internal error for panic
					appErr := errors.New(errors.ErrorCodeInternal, "Internal server error")
					writeErrorResponse(w, appErr)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// WriteErrorResponse writes a standardized error response
func WriteErrorResponse(w http.ResponseWriter, err error) {
	writeErrorResponse(w, err)
}

// writeErrorResponse is the internal implementation
func writeErrorResponse(w http.ResponseWriter, err error) {
	var appErr *errors.AppError
	var ok bool

	// Convert to AppError if not already one
	if appErr, ok = errors.AsAppError(err); !ok {
		appErr = errors.NewInternalError(err)
	}

	// Create error response
	response := ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
			Context: appErr.Context,
		},
	}

	// Set headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.StatusCode)

	// Write JSON response
	if jsonBytes, err := json.Marshal(response); err != nil {
		// Use default logger if we can't get from context
		logging.Error("Failed to marshal error response", slog.Any("error", err))
		// Fallback to simple error message
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success":false,"error":{"code":"INTERNAL_ERROR","message":"Internal server error"}}`))
	} else {
		w.Write(jsonBytes)
	}

	// Log error for debugging (don't expose sensitive information)
	if appErr.StatusCode >= 500 {
		logging.Error("Server error occurred",
			slog.String("code", string(appErr.Code)),
			slog.String("message", appErr.Message),
			slog.Any("context", appErr.Context),
			slog.Int("status_code", appErr.StatusCode),
		)
	} else if appErr.StatusCode >= 400 {
		logging.Warn("Client error occurred",
			slog.String("code", string(appErr.Code)),
			slog.String("message", appErr.Message),
			slog.Any("context", appErr.Context),
			slog.Int("status_code", appErr.StatusCode),
		)
	}
}

// ErrorMiddleware is an alternative error handling middleware that can be used in HTTP handlers
func ErrorMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a custom ResponseWriter that can capture errors
			ew := &errorResponseWriter{
				ResponseWriter: w,
				request:        r,
			}

			next.ServeHTTP(ew, r)
		})
	}
}

// errorResponseWriter wraps http.ResponseWriter to intercept errors
type errorResponseWriter struct {
	http.ResponseWriter
	request     *http.Request
	wroteHeader bool
	statusCode  int
}

func (ew *errorResponseWriter) WriteHeader(statusCode int) {
	if !ew.wroteHeader {
		ew.statusCode = statusCode
		ew.wroteHeader = true

		// If it's an error status code, we might want to handle it
		if statusCode >= 400 {
			// For now, just proceed normally
			// In the future, we could intercept and format the response
		}

		ew.ResponseWriter.WriteHeader(statusCode)
	}
}

func (ew *errorResponseWriter) Write(data []byte) (int, error) {
	if !ew.wroteHeader {
		ew.WriteHeader(http.StatusOK)
	}
	return ew.ResponseWriter.Write(data)
}

// HandlerFunc represents an HTTP handler that can return an error
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// ErrorHandlerFunc wraps a HandlerFunc to handle errors automatically
func ErrorHandlerFunc(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			WriteErrorResponse(w, err)
		}
	}
}
