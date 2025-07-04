package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// RateLimiter manages rate limiting for multiple clients
type RateLimiter struct {
	mu       sync.RWMutex
	clients  map[string]*Client
	rate     int           // requests per window
	window   time.Duration // time window
	cleanup  time.Duration // cleanup interval
	stopChan chan struct{}
}

// Client represents a client's rate limit state
type Client struct {
	mu         sync.RWMutex
	requests   []time.Time
	lastAccess time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerWindow int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients:  make(map[string]*Client),
		rate:     requestsPerWindow,
		window:   window,
		cleanup:  window * 2, // cleanup inactive clients every 2 windows
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanupRoutine()

	return rl
}

// Allow checks if a request is allowed for the given client ID
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	client, exists := rl.clients[clientID]
	if !exists {
		client = &Client{
			requests:   make([]time.Time, 0),
			lastAccess: time.Now(),
		}
		rl.clients[clientID] = client
	}
	rl.mu.Unlock()

	client.mu.Lock()
	defer client.mu.Unlock()

	now := time.Now()
	client.lastAccess = now

	// Remove old requests outside the window
	cutoff := now.Add(-rl.window)
	validRequests := make([]time.Time, 0)
	for _, reqTime := range client.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	client.requests = validRequests

	// Check if we can add a new request
	if len(client.requests) < rl.rate {
		client.requests = append(client.requests, now)
		return true
	}

	return false
}

// cleanupRoutine removes inactive clients
func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanupInactiveClients()
		case <-rl.stopChan:
			return
		}
	}
}

// cleanupInactiveClients removes clients that haven't made requests recently
func (rl *RateLimiter) cleanupInactiveClients() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rl.cleanup)
	for clientID, client := range rl.clients {
		client.mu.RLock()
		lastAccess := client.lastAccess
		client.mu.RUnlock()

		if lastAccess.Before(cutoff) {
			delete(rl.clients, clientID)
		}
	}
}

// Stop stops the rate limiter cleanup routine
func (rl *RateLimiter) Stop() {
	close(rl.stopChan)
}

// RateLimitMiddleware creates a Chi middleware for rate limiting
func RateLimitMiddleware(requestsPerWindow int, window time.Duration) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(requestsPerWindow, window)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client identifier (IP address)
			clientID := GetClientIP(r)

			if !limiter.Allow(clientID) {
				// Rate limit exceeded
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(requestsPerWindow))
				w.Header().Set("X-RateLimit-Window", window.String())
				w.Header().Set("Retry-After", strconv.Itoa(int(window.Seconds())))

				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"rate limit exceeded","message":"too many requests"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetClientIP extracts the client IP address from the request
func GetClientIP(r *http.Request) string {
	// Try to get real IP from headers (for reverse proxy scenarios)
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		return forwardedFor
	}

	// Fall back to remote address
	return middleware.GetReqID(r.Context())
}

// PerUserRateLimitMiddleware creates a rate limiter based on authenticated user ID
func PerUserRateLimitMiddleware(requestsPerWindow int, window time.Duration) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(requestsPerWindow, window)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from context (set by auth middleware)
			userID := r.Context().Value("user_id")
			clientID := GetClientIP(r) // fallback to IP if no user ID

			if userID != nil {
				if uid, ok := userID.(string); ok && uid != "" {
					clientID = fmt.Sprintf("user:%s", uid)
				}
			}

			if !limiter.Allow(clientID) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(requestsPerWindow))
				w.Header().Set("X-RateLimit-Window", window.String())
				w.Header().Set("Retry-After", strconv.Itoa(int(window.Seconds())))

				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"rate limit exceeded","message":"too many requests"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
