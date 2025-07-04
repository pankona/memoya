package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/pankona/memoya/internal/logging"
)

// StructuredLoggingMiddleware replaces the default chi logger with structured logging
func StructuredLoggingMiddleware(logger *logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get request ID from context (set by RequestID middleware)
			reqID := middleware.GetReqID(r.Context())

			// Create a structured logger with request context
			ctx := context.WithValue(r.Context(), "request_id", reqID)
			requestLogger := logger.WithContext(ctx)

			// Add logger to context for use in handlers
			ctx = context.WithValue(ctx, "logger", requestLogger)
			r = r.WithContext(ctx)

			// Create a wrapped response writer to capture status and size
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Log request start
			requestLogger.Info("Request started",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", reqID),
			)

			defer func() {
				// Log request completion
				duration := time.Since(start)
				status := ww.Status()
				size := ww.BytesWritten()

				// Determine log level based on status code
				logLevel := slog.LevelInfo
				if status >= 400 && status < 500 {
					logLevel = slog.LevelWarn
				} else if status >= 500 {
					logLevel = slog.LevelError
				}

				requestLogger.Log(r.Context(), logLevel, "Request completed",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Int("status", status),
					slog.Int("size", size),
					slog.Duration("duration", duration),
					slog.String("request_id", reqID),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

// LoggingMiddleware provides a simple logging middleware using the default logger
func LoggingMiddleware() func(http.Handler) http.Handler {
	return StructuredLoggingMiddleware(logging.Default())
}

// RequestLoggingMiddleware logs HTTP requests with custom fields
func RequestLoggingMiddleware(logger *logging.Logger, includeBody bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := middleware.GetReqID(r.Context())

			// Create enriched context
			ctx := context.WithValue(r.Context(), "request_id", reqID)
			requestLogger := logger.WithContext(ctx)

			// Log detailed request information
			args := []any{
				"method", r.Method,
				"url", r.URL.String(),
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"referer", r.Referer(),
				"request_id", reqID,
				"content_length", r.ContentLength,
			}

			// Add headers if in debug mode
			if logger.Enabled(r.Context(), slog.LevelDebug) {
				headers := make(map[string]interface{})
				for k, v := range r.Header {
					// Exclude sensitive headers
					if k != "Authorization" && k != "Cookie" {
						headers[k] = v
					}
				}
				args = append(args, "headers", headers)
			}

			requestLogger.Info("HTTP request received", args...)

			// Add logger to context
			ctx = context.WithValue(ctx, "logger", requestLogger)
			r = r.WithContext(ctx)

			// Wrap response writer
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				duration := time.Since(start)
				status := ww.Status()
				size := ww.BytesWritten()

				// Response logging arguments
				responseArgs := []any{
					"method", r.Method,
					"path", r.URL.Path,
					"status", status,
					"response_size", size,
					"duration", duration,
					"request_id", reqID,
				}

				// Log based on status code
				if status >= 400 {
					if status >= 500 {
						requestLogger.Error("HTTP request failed", responseArgs...)
					} else {
						requestLogger.Warn("HTTP request client error", responseArgs...)
					}
				} else {
					requestLogger.Info("HTTP request completed", responseArgs...)
				}
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
