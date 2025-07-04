package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// LogLevel represents logging levels
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
}

// Config represents logger configuration
type Config struct {
	Level  LogLevel `json:"level"`
	Format string   `json:"format"` // "json" or "text"
}

var (
	defaultLogger *Logger
)

// init initializes the default logger
func init() {
	defaultLogger = NewLogger(Config{
		Level:  LevelInfo,
		Format: "json",
	})
}

// NewLogger creates a new structured logger
func NewLogger(config Config) *Logger {
	var level slog.Level
	switch config.Level {
	case LevelDebug:
		level = slog.LevelDebug
	case LevelInfo:
		level = slog.LevelInfo
	case LevelWarn:
		level = slog.LevelWarn
	case LevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var handler slog.Handler
	if config.Format == "text" {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// NewLoggerFromEnv creates a logger from environment variables
func NewLoggerFromEnv() *Logger {
	level := LevelInfo
	if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
		switch strings.ToLower(envLevel) {
		case "debug":
			level = LevelDebug
		case "info":
			level = LevelInfo
		case "warn", "warning":
			level = LevelWarn
		case "error":
			level = LevelError
		}
	}

	format := "json"
	if envFormat := os.Getenv("LOG_FORMAT"); envFormat != "" {
		if strings.ToLower(envFormat) == "text" {
			format = "text"
		}
	}

	return NewLogger(Config{
		Level:  level,
		Format: format,
	})
}

// Default returns the default logger instance
func Default() *Logger {
	return defaultLogger
}

// SetDefault sets the default logger
func SetDefault(logger *Logger) {
	defaultLogger = logger
}

// WithContext returns a logger with context fields
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extract common context values
	var args []any

	// Add request ID if available
	if reqID := ctx.Value("request_id"); reqID != nil {
		if id, ok := reqID.(string); ok {
			args = append(args, "request_id", id)
		}
	}

	// Add user ID if available
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			args = append(args, "user_id", id)
		}
	}

	if len(args) > 0 {
		return &Logger{
			Logger: l.Logger.With(args...),
		}
	}

	return l
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	var args []any
	for k, v := range fields {
		args = append(args, k, v)
	}

	return &Logger{
		Logger: l.Logger.With(args...),
	}
}

// WithError returns a logger with error information
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.Logger.With("error", err),
	}
}

// Operation logs the start and completion of an operation
func (l *Logger) Operation(ctx context.Context, operation string) *OperationLogger {
	logger := l.WithContext(ctx)
	logger.Info("Operation started", slog.String("operation", operation))

	return &OperationLogger{
		logger:    logger,
		operation: operation,
	}
}

// OperationLogger tracks the lifecycle of an operation
type OperationLogger struct {
	logger    *Logger
	operation string
}

// Success logs successful completion of an operation
func (ol *OperationLogger) Success(message string, args ...any) {
	finalArgs := append(args, "operation", ol.operation, "status", "success")
	ol.logger.Info(message, finalArgs...)
}

// Error logs failed completion of an operation
func (ol *OperationLogger) Error(message string, err error, args ...any) {
	finalArgs := append(args, "operation", ol.operation, "status", "error", "error", err)
	ol.logger.Error(message, finalArgs...)
}

// Package-level convenience functions using default logger

// Debug logs a debug message
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// WithContext returns a logger with context fields using default logger
func WithContext(ctx context.Context) *Logger {
	return defaultLogger.WithContext(ctx)
}

// WithFields returns a logger with additional fields using default logger
func WithFields(fields map[string]interface{}) *Logger {
	return defaultLogger.WithFields(fields)
}

// WithError returns a logger with error information using default logger
func WithError(err error) *Logger {
	return defaultLogger.WithError(err)
}
