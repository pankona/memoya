package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pankona/memoya/internal/handlers"
)

func TestMCPBridge_MemoCreate(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mcp/memo_create" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"memo":{"id":"memo-123","title":"Test Memo"},"message":"memo created successfully"}`))
	}))
	defer server.Close()

	// Create bridge
	httpClient := NewHTTPClient(server.URL)
	bridge := NewMCPBridge(httpClient)

	// Test memo create
	args := handlers.MemoCreateArgs{
		Title:       "Test Memo",
		Description: "Test description",
		Tags:        []string{"test"},
	}
	params := &mcp.CallToolParamsFor[handlers.MemoCreateArgs]{Arguments: args}

	result, err := bridge.MemoCreate(context.Background(), nil, params)
	if err != nil {
		t.Errorf("Expected memo create to succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if len(result.Content) == 0 {
		t.Fatal("Expected content, got empty")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}

	expected := `{"success":true,"memo":{"id":"memo-123","title":"Test Memo"},"message":"memo created successfully"}`
	if textContent.Text != expected {
		t.Errorf("Expected content %s, got %s", expected, textContent.Text)
	}
}

func TestMCPBridge_TodoList(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mcp/todo_list" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"todos":[{"id":"todo-1","title":"Test Todo"}],"message":"Found 1 todos"}`))
	}))
	defer server.Close()

	// Create bridge
	httpClient := NewHTTPClient(server.URL)
	bridge := NewMCPBridge(httpClient)

	// Test todo list
	args := handlers.TodoListArgs{
		Status: "todo",
		Tags:   []string{"work"},
	}
	params := &mcp.CallToolParamsFor[handlers.TodoListArgs]{Arguments: args}

	result, err := bridge.TodoList(context.Background(), nil, params)
	if err != nil {
		t.Errorf("Expected todo list to succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if len(result.Content) == 0 {
		t.Fatal("Expected content, got empty")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}

	expected := `{"success":true,"todos":[{"id":"todo-1","title":"Test Todo"}],"message":"Found 1 todos"}`
	if textContent.Text != expected {
		t.Errorf("Expected content %s, got %s", expected, textContent.Text)
	}
}

func TestMCPBridge_DeviceAuthStart(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/device_start" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"device_code":"ABCD-1234","user_code":"WXYZ-5678","verification_uri":"https://accounts.google.com/device"}`))
	}))
	defer server.Close()

	// Create bridge
	httpClient := NewHTTPClient(server.URL)
	bridge := NewMCPBridge(httpClient)

	// Test device auth start
	args := handlers.DeviceAuthStartArgs{}
	params := &mcp.CallToolParamsFor[handlers.DeviceAuthStartArgs]{Arguments: args}

	result, err := bridge.DeviceAuthStart(context.Background(), nil, params)
	if err != nil {
		t.Errorf("Expected device auth start to succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if len(result.Content) == 0 {
		t.Fatal("Expected content, got empty")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}

	expected := `{"success":true,"device_code":"ABCD-1234","user_code":"WXYZ-5678","verification_uri":"https://accounts.google.com/device"}`
	if textContent.Text != expected {
		t.Errorf("Expected content %s, got %s", expected, textContent.Text)
	}
}

func TestMCPBridge_GetUserInfo(t *testing.T) {
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"user":{"id":"user-123","is_active":true},"message":"User info retrieved"}`))
	}))
	defer server.Close()

	// Create bridge
	httpClient := NewHTTPClient(server.URL)
	bridge := NewMCPBridge(httpClient)

	// Test get user info
	args := handlers.UserInfoArgs{}
	params := &mcp.CallToolParamsFor[handlers.UserInfoArgs]{Arguments: args}

	result, err := bridge.GetUserInfo(context.Background(), nil, params)
	if err != nil {
		t.Errorf("Expected get user info to succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if len(result.Content) == 0 {
		t.Fatal("Expected content, got empty")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}

	expected := `{"success":true,"user":{"id":"user-123","is_active":true},"message":"User info retrieved"}`
	if textContent.Text != expected {
		t.Errorf("Expected content %s, got %s", expected, textContent.Text)
	}
}

func TestMCPBridge_DeleteAccount(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/delete_account" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"message":"Account deleted successfully"}`))
	}))
	defer server.Close()

	// Create bridge
	httpClient := NewHTTPClient(server.URL)
	bridge := NewMCPBridge(httpClient)

	// Test delete account
	args := handlers.AccountDeleteArgs{
		Confirm: true,
	}
	params := &mcp.CallToolParamsFor[handlers.AccountDeleteArgs]{Arguments: args}

	result, err := bridge.DeleteAccount(context.Background(), nil, params)
	if err != nil {
		t.Errorf("Expected delete account to succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if len(result.Content) == 0 {
		t.Fatal("Expected content, got empty")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}

	expected := `{"success":true,"message":"Account deleted successfully"}`
	if textContent.Text != expected {
		t.Errorf("Expected content %s, got %s", expected, textContent.Text)
	}
}

func TestMCPBridge_HTTPError(t *testing.T) {
	// Create test server that returns 500 error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success":false,"error":"Internal server error","code":"INTERNAL_ERROR"}`))
	}))
	defer server.Close()

	// Create bridge
	httpClient := NewHTTPClient(server.URL)
	bridge := NewMCPBridge(httpClient)

	// Test memo create with error
	args := handlers.MemoCreateArgs{
		Title: "Test Memo",
	}
	params := &mcp.CallToolParamsFor[handlers.MemoCreateArgs]{Arguments: args}

	result, err := bridge.MemoCreate(context.Background(), nil, params)
	if err == nil {
		t.Error("Expected error when server returns 500")
	}

	if result != nil {
		t.Error("Expected nil result when error occurs")
	}

	expectedError := "memo_create failed: server error [INTERNAL_ERROR]: Internal server error"
	if err.Error() != expectedError {
		t.Errorf("Expected error %s, got %s", expectedError, err.Error())
	}
}
