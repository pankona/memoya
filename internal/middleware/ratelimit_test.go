package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	// Create a rate limiter: 3 requests per second
	limiter := NewRateLimiter(3, time.Second)
	defer limiter.Stop()

	clientID := "test-client"

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		if !limiter.Allow(clientID) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied
	if limiter.Allow(clientID) {
		t.Error("4th request should be denied")
	}

	// Wait for window to reset
	time.Sleep(time.Second + 10*time.Millisecond)

	// Request should be allowed again
	if !limiter.Allow(clientID) {
		t.Error("Request after window reset should be allowed")
	}
}

func TestRateLimiter_MultipleClients(t *testing.T) {
	limiter := NewRateLimiter(2, time.Second)
	defer limiter.Stop()

	client1 := "client1"
	client2 := "client2"

	// Both clients should be able to make 2 requests
	for i := 0; i < 2; i++ {
		if !limiter.Allow(client1) {
			t.Errorf("Client1 request %d should be allowed", i+1)
		}
		if !limiter.Allow(client2) {
			t.Errorf("Client2 request %d should be allowed", i+1)
		}
	}

	// Both clients should be denied on 3rd request
	if limiter.Allow(client1) {
		t.Error("Client1 3rd request should be denied")
	}
	if limiter.Allow(client2) {
		t.Error("Client2 3rd request should be denied")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap with rate limit middleware: 2 requests per second
	middleware := RateLimitMiddleware(2, time.Second)
	wrappedHandler := middleware(handler)

	// Create test server
	server := httptest.NewServer(wrappedHandler)
	defer server.Close()

	client := &http.Client{}

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		resp, err := client.Get(server.URL)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, resp.StatusCode)
		}
		resp.Body.Close()
	}

	// 3rd request should be rate limited
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Rate limited request failed: %v", err)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Rate limited request: expected status 429, got %d", resp.StatusCode)
	}

	// Check rate limit headers
	if resp.Header.Get("X-RateLimit-Limit") != "2" {
		t.Errorf("Expected X-RateLimit-Limit: 2, got %s", resp.Header.Get("X-RateLimit-Limit"))
	}
	if resp.Header.Get("Retry-After") != "1" {
		t.Errorf("Expected Retry-After: 1, got %s", resp.Header.Get("Retry-After"))
	}
	resp.Body.Close()
}

func TestRateLimiter_Cleanup(t *testing.T) {
	// Create rate limiter with short cleanup interval
	limiter := NewRateLimiter(10, 100*time.Millisecond)
	defer limiter.Stop()

	// Add some clients
	limiter.Allow("client1")
	limiter.Allow("client2")
	limiter.Allow("client3")

	// Check clients exist
	limiter.mu.RLock()
	initialCount := len(limiter.clients)
	limiter.mu.RUnlock()

	if initialCount != 3 {
		t.Errorf("Expected 3 clients, got %d", initialCount)
	}

	// Wait for cleanup to run
	time.Sleep(300 * time.Millisecond)

	// Clients should be cleaned up
	limiter.mu.RLock()
	finalCount := len(limiter.clients)
	limiter.mu.RUnlock()

	if finalCount != 0 {
		t.Errorf("Expected 0 clients after cleanup, got %d", finalCount)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name           string
		headers        map[string]string
		expectedPrefix string
	}{
		{
			name: "X-Real-IP header",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.1",
			},
			expectedPrefix: "192.168.1.1",
		},
		{
			name: "X-Forwarded-For header",
			headers: map[string]string{
				"X-Forwarded-For": "10.0.0.1, 172.16.0.1",
			},
			expectedPrefix: "10.0.0.1, 172.16.0.1",
		},
		{
			name:           "No special headers",
			headers:        map[string]string{},
			expectedPrefix: "", // Should fall back to middleware.GetReqID
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			ip := GetClientIP(req)
			if tt.expectedPrefix != "" && ip != tt.expectedPrefix {
				t.Errorf("GetClientIP() = %v, want %v", ip, tt.expectedPrefix)
			}
		})
	}
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	limiter := NewRateLimiter(1000, time.Second)
	defer limiter.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		clientID := "bench-client"
		for pb.Next() {
			limiter.Allow(clientID)
		}
	})
}
