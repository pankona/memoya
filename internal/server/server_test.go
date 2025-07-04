package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pankona/memoya/internal/auth"
	"github.com/pankona/memoya/internal/handlers"
)

func TestServer_HealthCheck(t *testing.T) {
	// Mock storage setup
	mockStorage := handlers.NewMockStorage()
	server := NewServer(context.Background(), mockStorage)

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HealthCheck)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check content type
	expected := "application/json"
	if ct := rr.Header().Get("Content-Type"); ct != expected {
		t.Errorf("handler returned wrong content type: got %v want %v", ct, expected)
	}

	// Check response body structure
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Could not parse response JSON: %v", err)
	}

	if status, ok := response["status"]; !ok || status != "ok" {
		t.Errorf("Expected status 'ok', got %v", status)
	}

	if _, ok := response["timestamp"]; !ok {
		t.Errorf("Expected timestamp field in response")
	}
}

func TestServer_CreateMemo(t *testing.T) {
	// Mock storage setup with test data
	mockStorage := handlers.NewMockStorage()
	mockStorage.SetupTestData()
	server := NewServer(context.Background(), mockStorage)

	// Test case: valid memo creation
	reqBody := map[string]interface{}{
		"title":       "Test Memo",
		"description": "Test Description",
		"tags":        []string{"test", "memo"},
	}

	reqJSON, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/mcp/memo_create", bytes.NewBuffer(reqJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Add authentication context
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "test-user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.CreateMemo)
	handler.ServeHTTP(rr, req)

	// Debug: print actual response
	t.Logf("Response status: %d", rr.Code)
	t.Logf("Response body: %s", rr.Body.String())

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		return
	}

	// Check content type
	expected := "application/json"
	if ct := rr.Header().Get("Content-Type"); ct != expected {
		t.Errorf("handler returned wrong content type: got %v want %v", ct, expected)
	}

	// Parse response
	var response handlers.MemoResult
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Could not parse response JSON: %v", err)
		return
	}

	// Validate response structure
	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	if response.Memo == nil {
		t.Errorf("Expected memo field to be present")
		return
	}

	if response.Memo.Title != "Test Memo" {
		t.Errorf("Expected memo title 'Test Memo', got %v", response.Memo.Title)
	}
}

func TestServer_CreateMemo_InvalidJSON(t *testing.T) {
	mockStorage := handlers.NewMockStorage()
	server := NewServer(context.Background(), mockStorage)

	// Test case: invalid JSON
	req, err := http.NewRequest("POST", "/mcp/memo_create", bytes.NewBuffer([]byte("invalid json")))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Add authentication context
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "test-user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.CreateMemo)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	// Parse error response
	var errorResp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &errorResp); err != nil {
		t.Errorf("Could not parse error response JSON: %v", err)
	}

	if success, ok := errorResp["success"]; !ok || success != false {
		t.Errorf("Expected success=false in error response, got %v", success)
	}
}

func TestServer_ListMemos(t *testing.T) {
	// Mock storage setup with test data
	mockStorage := handlers.NewMockStorage()
	mockStorage.SetupTestData()
	server := NewServer(context.Background(), mockStorage)

	// Test case: list all memos
	reqBody := map[string]interface{}{}
	reqJSON, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/mcp/memo_list", bytes.NewBuffer(reqJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Add authentication context
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "test-user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.ListMemos)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Parse response
	var response handlers.MemoListResult
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Could not parse response JSON: %v", err)
	}

	// Validate response structure
	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	if response.Memos == nil {
		t.Errorf("Expected memos field to be present")
	}

	if len(response.Memos) == 0 {
		t.Errorf("Expected at least one memo in test data")
	}
}

func TestServer_ListMemos_WithTagFilter(t *testing.T) {
	// Mock storage setup with test data
	mockStorage := handlers.NewMockStorage()
	mockStorage.SetupTestData()
	server := NewServer(context.Background(), mockStorage)

	// Test case: list memos with tag filter
	reqBody := map[string]interface{}{
		"tags": []string{"work"},
	}
	reqJSON, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/mcp/memo_list", bytes.NewBuffer(reqJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Add authentication context
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "test-user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.ListMemos)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Parse response
	var response handlers.MemoListResult
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Could not parse response JSON: %v", err)
	}

	// Validate response structure
	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	// Check that filtered results only contain memos with "work" tag
	for _, memo := range response.Memos {
		hasWorkTag := false
		for _, tag := range memo.Tags {
			if tag == "work" {
				hasWorkTag = true
				break
			}
		}
		if !hasWorkTag {
			t.Errorf("Expected memo to have 'work' tag, but it doesn't: %+v", memo)
		}
	}
}

func TestServer_CreateTodo(t *testing.T) {
	// Mock storage setup with test data
	mockStorage := handlers.NewMockStorage()
	mockStorage.SetupTestData()
	server := NewServer(context.Background(), mockStorage)

	// Test case: valid todo creation
	reqBody := map[string]interface{}{
		"title":       "Test Todo",
		"description": "Test Description",
		"status":      "todo",
		"priority":    "high",
		"tags":        []string{"test", "todo"},
	}

	reqJSON, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/mcp/todo_create", bytes.NewBuffer(reqJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Add authentication context
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "test-user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.CreateTodo)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Parse response
	var response handlers.TodoResult
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Could not parse response JSON: %v", err)
	}

	// Validate response structure
	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	if response.Todo == nil {
		t.Errorf("Expected todo field to be present")
	}

	if response.Todo.Title != "Test Todo" {
		t.Errorf("Expected todo title 'Test Todo', got %v", response.Todo.Title)
	}

	if response.Todo.Status != "todo" {
		t.Errorf("Expected todo status 'todo', got %v", response.Todo.Status)
	}

	if response.Todo.Priority != "high" {
		t.Errorf("Expected todo priority 'high', got %v", response.Todo.Priority)
	}
}

func TestServer_ListTodos(t *testing.T) {
	// Mock storage setup with test data
	mockStorage := handlers.NewMockStorage()
	mockStorage.SetupTestData()
	server := NewServer(context.Background(), mockStorage)

	// Test case: list all todos
	reqBody := map[string]interface{}{}
	reqJSON, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/mcp/todo_list", bytes.NewBuffer(reqJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Add authentication context
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "test-user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.ListTodos)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Parse response
	var response handlers.TodoListResult
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Could not parse response JSON: %v", err)
	}

	// Validate response structure
	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	if response.Todos == nil {
		t.Errorf("Expected todos field to be present")
	}

	if len(response.Todos) == 0 {
		t.Errorf("Expected at least one todo in test data")
	}
}

func TestServer_Search(t *testing.T) {
	// Mock storage setup with test data
	mockStorage := handlers.NewMockStorage()
	mockStorage.SetupTestData()
	server := NewServer(context.Background(), mockStorage)

	// Test case: search with query
	reqBody := map[string]interface{}{
		"query": "Project",
		"type":  "all",
	}
	reqJSON, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/mcp/search", bytes.NewBuffer(reqJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Add authentication context
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "test-user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.Search)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Parse response
	var response handlers.SearchResult
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Could not parse response JSON: %v", err)
	}

	// Validate response structure
	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	if response.Results.Memos == nil && response.Results.Todos == nil {
		t.Errorf("Expected results field to be present")
	}
}

func TestServer_ListTags(t *testing.T) {
	// Mock storage setup with test data
	mockStorage := handlers.NewMockStorage()
	mockStorage.SetupTestData()
	server := NewServer(context.Background(), mockStorage)

	// Test case: list all tags
	reqBody := map[string]interface{}{}
	reqJSON, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/mcp/tag_list", bytes.NewBuffer(reqJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Add authentication context
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "test-user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.ListTags)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Parse response
	var response handlers.TagListResult
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Could not parse response JSON: %v", err)
	}

	// Validate response structure
	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	if response.Tags == nil {
		t.Errorf("Expected tags field to be present")
	}

	if len(response.Tags) == 0 {
		t.Errorf("Expected at least one tag in test data")
	}
}

func TestServer_UpdateMemo(t *testing.T) {
	// Mock storage setup with test data
	mockStorage := handlers.NewMockStorage()
	mockStorage.SetupTestData()
	server := NewServer(context.Background(), mockStorage)

	// Test case: valid memo update
	reqBody := map[string]interface{}{
		"id":          "test-memo-1",
		"title":       "Updated Memo Title",
		"description": "Updated description",
		"tags":        []string{"updated", "memo"},
	}

	reqJSON, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/mcp/memo_update", bytes.NewBuffer(reqJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Add authentication context
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "test-user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.UpdateMemo)
	handler.ServeHTTP(rr, req)

	// Debug: print actual response
	t.Logf("Update Response status: %d", rr.Code)
	t.Logf("Update Response body: %s", rr.Body.String())

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		return
	}

	// Parse response
	var response handlers.MemoResult
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Could not parse response JSON: %v", err)
	}

	// Validate response structure
	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	if response.Memo == nil {
		t.Errorf("Expected memo field to be present")
	}

	if response.Memo.Title != "Updated Memo Title" {
		t.Errorf("Expected updated memo title 'Updated Memo Title', got %v", response.Memo.Title)
	}
}

func TestServer_DeleteMemo(t *testing.T) {
	// Mock storage setup with test data
	mockStorage := handlers.NewMockStorage()
	mockStorage.SetupTestData()
	server := NewServer(context.Background(), mockStorage)

	// Test case: valid memo deletion
	reqBody := map[string]interface{}{
		"id": "test-memo-1",
	}

	reqJSON, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/mcp/memo_delete", bytes.NewBuffer(reqJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Add authentication context
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "test-user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.DeleteMemo)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Parse response
	var response handlers.MemoDeleteResult
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Could not parse response JSON: %v", err)
	}

	// Validate response structure
	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	if response.Message == "" {
		t.Errorf("Expected message field to be present")
	}
}

func TestServer_GetUserInfo(t *testing.T) {
	// Mock storage setup with test data
	mockStorage := handlers.NewMockStorage()
	mockStorage.SetupTestData()
	server := NewServer(context.Background(), mockStorage)

	req, err := http.NewRequest("GET", "/auth/user", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Add authentication context
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "test-user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.GetUserInfo)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Parse response
	var response handlers.UserInfoResult
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Could not parse response JSON: %v", err)
	}

	// Validate response structure
	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	if response.User == nil {
		t.Error("Expected user field to be present")
	}

	if response.User.ID != "test-user-1" {
		t.Errorf("Expected user ID %s, got %s", "test-user-1", response.User.ID)
	}
}

func TestServer_DeleteAccount(t *testing.T) {
	// Mock storage setup with test data
	mockStorage := handlers.NewMockStorage()
	mockStorage.SetupTestData()
	server := NewServer(context.Background(), mockStorage)

	// Verify initial data exists
	if len(mockStorage.GetUsers()) == 0 {
		t.Fatal("Expected test user to exist")
	}

	reqBody := map[string]interface{}{
		"confirm": true,
	}
	reqJSON, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/auth/delete_account", bytes.NewBuffer(reqJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Add authentication context
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "test-user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.DeleteAccount)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Parse response
	var response handlers.AccountDeleteResult
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Could not parse response JSON: %v", err)
	}

	// Validate response structure
	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	if response.Message == "" {
		t.Error("Expected message field to be present")
	}

	// Verify user was deleted
	if len(mockStorage.GetUsers()) != 0 {
		t.Error("Expected user to be deleted")
	}
}
