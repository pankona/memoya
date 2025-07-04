package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	generatedServer "github.com/pankona/memoya/internal/generated/server"
	"github.com/pankona/memoya/internal/logging"
	rateLimitMiddleware "github.com/pankona/memoya/internal/middleware"
	"github.com/pankona/memoya/internal/server"
	"github.com/pankona/memoya/internal/storage"
)

func main() {
	// Initialize structured logging
	logger := logging.NewLoggerFromEnv()
	logging.SetDefault(logger)

	// Initialize context
	ctx := context.Background()

	// Initialize Firestore storage
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		logger.Error("PROJECT_ID environment variable is required")
		os.Exit(1)
	}

	storage, err := storage.NewFirestoreStorage(ctx, projectID)
	if err != nil {
		logger.Error("Failed to initialize Firestore", slog.Any("error", err), slog.String("project_id", projectID))
		os.Exit(1)
	}
	defer storage.Close()

	logger.Info("Firestore storage initialized", slog.String("project_id", projectID))

	// Create server implementation
	serverImpl := server.NewServer(ctx, storage)

	// Create router
	r := chi.NewRouter()

	// Rate limiting configuration
	rateLimitRequests := 100 // default: 100 requests per minute
	rateLimitWindow := time.Minute

	if rateLimit := os.Getenv("RATE_LIMIT_REQUESTS"); rateLimit != "" {
		if parsed, err := strconv.Atoi(rateLimit); err == nil && parsed > 0 {
			rateLimitRequests = parsed
		}
	}

	if window := os.Getenv("RATE_LIMIT_WINDOW"); window != "" {
		if parsed, err := time.ParseDuration(window); err == nil && parsed > 0 {
			rateLimitWindow = parsed
		}
	}

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(rateLimitMiddleware.StructuredLoggingMiddleware(logger))
	r.Use(middleware.Recoverer)
	r.Use(rateLimitMiddleware.RateLimitMiddleware(rateLimitRequests, rateLimitWindow))
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS configuration
	allowedOrigins := []string{"*"} // Default to allow all for development
	if corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); corsOrigins != "" {
		// Parse comma-separated origins for production
		allowedOrigins = strings.Split(corsOrigins, ",")
		for i, origin := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(origin)
		}
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Add OpenAPI routes
	generatedServer.HandlerFromMux(serverImpl, r)

	// Health check endpoint (if not already included)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
	})

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server",
			slog.String("port", port),
			slog.String("address", srv.Addr),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start server", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutdown signal received, starting graceful shutdown...")

	// Create a deadline for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("Server shutdown completed successfully")
}
