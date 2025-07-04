package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPClient_Ping(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","timestamp":"2024-01-01T10:00:00Z"}`))
	}))
	defer server.Close()

	// Create client
	client := NewHTTPClient(server.URL)

	// Test ping
	err := client.Ping(context.Background())
	if err != nil {
		t.Errorf("Expected ping to succeed, got error: %v", err)
	}
}

func TestHTTPClient_Ping_ServerDown(t *testing.T) {
	// Create client with invalid URL
	client := NewHTTPClient("http://localhost:99999")

	// Test ping
	err := client.Ping(context.Background())
	if err == nil {
		t.Error("Expected ping to fail when server is down")
	}
}

func TestHTTPClient_SetAuthToken(t *testing.T) {
	client := NewHTTPClient("http://example.com")

	// Initially no token
	if client.authToken != "" {
		t.Error("Expected no auth token initially")
	}

	// Set token
	testToken := "test-jwt-token"
	client.SetAuthToken(testToken)

	if client.authToken != testToken {
		t.Errorf("Expected auth token %s, got %s", testToken, client.authToken)
	}
}

func TestHTTPClient_CallTool_MemoCreate(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mcp/memo_create" {
			http.NotFound(w, r)
			return
		}
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check headers
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Invalid content type", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"data":{"id":"memo-123","title":"Test Memo"},"message":"memo created successfully"}`))
	}))
	defer server.Close()

	// Create client
	client := NewHTTPClient(server.URL)

	// Test call
	args := map[string]interface{}{
		"title":       "Test Memo",
		"description": "Test description",
	}

	response, err := client.CallTool(context.Background(), "memo_create", args)
	if err != nil {
		t.Errorf("Expected call to succeed, got error: %v", err)
	}

	expected := `{"success":true,"data":{"id":"memo-123","title":"Test Memo"},"message":"memo created successfully"}`
	if string(response) != expected {
		t.Errorf("Expected response %s, got %s", expected, string(response))
	}
}

func TestHTTPClient_CallTool_AuthEndpoint(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/device_start" {
			http.NotFound(w, r)
			return
		}
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"device_code":"ABCD-1234","user_code":"WXYZ-5678"}`))
	}))
	defer server.Close()

	// Create client
	client := NewHTTPClient(server.URL)

	// Test auth endpoint call
	args := map[string]interface{}{}
	response, err := client.CallTool(context.Background(), "device_auth_start", args)
	if err != nil {
		t.Errorf("Expected call to succeed, got error: %v", err)
	}

	expected := `{"success":true,"device_code":"ABCD-1234","user_code":"WXYZ-5678"}`
	if string(response) != expected {
		t.Errorf("Expected response %s, got %s", expected, string(response))
	}
}

func TestHTTPClient_CallTool_UserInfoGET(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/user" {
			http.NotFound(w, r)
			return
		}
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check auth header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"user":{"id":"user-123","is_active":true}}`))
	}))
	defer server.Close()

	// Create client with auth token
	client := NewHTTPClient(server.URL)
	client.SetAuthToken("test-token")

	// Test user info call (should use GET)
	response, err := client.CallTool(context.Background(), "user_info", nil)
	if err != nil {
		t.Errorf("Expected call to succeed, got error: %v", err)
	}

	expected := `{"success":true,"user":{"id":"user-123","is_active":true}}`
	if string(response) != expected {
		t.Errorf("Expected response %s, got %s", expected, string(response))
	}
}

func TestHTTPClient_CallTool_ErrorResponse(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success":false,"error":"Invalid request","code":"BAD_REQUEST"}`))
	}))
	defer server.Close()

	// Create client
	client := NewHTTPClient(server.URL)

	// Test call that should fail
	args := map[string]interface{}{"invalid": "data"}
	_, err := client.CallTool(context.Background(), "memo_create", args)
	if err == nil {
		t.Error("Expected call to fail with error response")
	}

	expectedError := "server error [BAD_REQUEST]: Invalid request"
	if err.Error() != expectedError {
		t.Errorf("Expected error %s, got %s", expectedError, err.Error())
	}
}
