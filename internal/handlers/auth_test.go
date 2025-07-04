package handlers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pankona/memoya/internal/auth"
	"github.com/pankona/memoya/internal/models"
)

func TestAuthHandler_GetUserInfo(t *testing.T) {
	mockStorage := NewMockStorage()
	handler := NewAuthHandler(mockStorage, "test-client", "test-secret")

	// Create test user
	testUser := &models.User{
		ID:        "test-user-1",
		GoogleID:  "google-123",
		CreatedAt: time.Now(),
		IsActive:  true,
	}
	err := mockStorage.CreateUser(context.Background(), testUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	args := UserInfoArgs{}
	params := &mcp.CallToolParamsFor[UserInfoArgs]{
		Arguments: args,
	}

	// Create context with test user ID
	ctx := context.WithValue(context.Background(), auth.UserIDKey, "test-user-1")
	result, err := handler.GetUserInfo(ctx, nil, params)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
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

	var userInfoResult UserInfoResult
	err = json.Unmarshal([]byte(textContent.Text), &userInfoResult)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if !userInfoResult.Success {
		t.Error("Expected success to be true")
	}

	if userInfoResult.User == nil {
		t.Fatal("Expected user data, got nil")
	}

	if userInfoResult.User.ID != "test-user-1" {
		t.Errorf("Expected user ID %s, got %s", "test-user-1", userInfoResult.User.ID)
	}

	if !userInfoResult.User.IsActive {
		t.Error("Expected user to be active")
	}

	if userInfoResult.Message == "" {
		t.Error("Expected non-empty message")
	}
}

func TestAuthHandler_GetUserInfo_WithoutAuth(t *testing.T) {
	mockStorage := NewMockStorage()
	handler := NewAuthHandler(mockStorage, "test-client", "test-secret")

	args := UserInfoArgs{}
	params := &mcp.CallToolParamsFor[UserInfoArgs]{
		Arguments: args,
	}

	// Use context without authentication
	result, err := handler.GetUserInfo(context.Background(), nil, params)

	if err == nil {
		t.Fatal("Expected authentication error, got nil")
	}

	if result != nil {
		t.Error("Expected nil result when authentication fails")
	}
}

func TestAuthHandler_DeleteAccount(t *testing.T) {
	mockStorage := NewMockStorage()
	mockStorage.SetupTestData() // This creates test data for "test-user-1"
	handler := NewAuthHandler(mockStorage, "test-client", "test-secret")

	// Verify initial data exists
	if len(mockStorage.memos) == 0 {
		t.Fatal("Expected test memos to exist")
	}
	if len(mockStorage.todos) == 0 {
		t.Fatal("Expected test todos to exist")
	}

	args := AccountDeleteArgs{
		Confirm: true,
	}
	params := &mcp.CallToolParamsFor[AccountDeleteArgs]{
		Arguments: args,
	}

	// Create context with test user ID
	ctx := context.WithValue(context.Background(), auth.UserIDKey, "test-user-1")
	result, err := handler.DeleteAccount(ctx, nil, params)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}

	var deleteResult AccountDeleteResult
	err = json.Unmarshal([]byte(textContent.Text), &deleteResult)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if !deleteResult.Success {
		t.Error("Expected success to be true")
	}

	if deleteResult.Message == "" {
		t.Error("Expected non-empty message")
	}

	// Verify all user data was deleted
	if len(mockStorage.users) != 0 {
		t.Errorf("Expected all users to be deleted, got %d users", len(mockStorage.users))
	}

	// Check that user's memos were deleted
	userMemoCount := 0
	for _, memo := range mockStorage.memos {
		if memo.UserID == "test-user-1" {
			userMemoCount++
		}
	}
	if userMemoCount != 0 {
		t.Errorf("Expected user's memos to be deleted, found %d memos", userMemoCount)
	}

	// Check that user's todos were deleted
	userTodoCount := 0
	for _, todo := range mockStorage.todos {
		if todo.UserID == "test-user-1" {
			userTodoCount++
		}
	}
	if userTodoCount != 0 {
		t.Errorf("Expected user's todos to be deleted, found %d todos", userTodoCount)
	}
}

func TestAuthHandler_DeleteAccount_WithoutConfirm(t *testing.T) {
	mockStorage := NewMockStorage()
	handler := NewAuthHandler(mockStorage, "test-client", "test-secret")

	args := AccountDeleteArgs{
		Confirm: false, // No confirmation
	}
	params := &mcp.CallToolParamsFor[AccountDeleteArgs]{
		Arguments: args,
	}

	// Create context with test user ID
	ctx := context.WithValue(context.Background(), auth.UserIDKey, "test-user-1")
	result, err := handler.DeleteAccount(ctx, nil, params)

	if err == nil {
		t.Fatal("Expected confirmation error, got nil")
	}

	if result != nil {
		t.Error("Expected nil result when confirmation is missing")
	}
}

func TestAuthHandler_DeleteAccount_WithoutAuth(t *testing.T) {
	mockStorage := NewMockStorage()
	handler := NewAuthHandler(mockStorage, "test-client", "test-secret")

	args := AccountDeleteArgs{
		Confirm: true,
	}
	params := &mcp.CallToolParamsFor[AccountDeleteArgs]{
		Arguments: args,
	}

	// Use context without authentication
	result, err := handler.DeleteAccount(context.Background(), nil, params)

	if err == nil {
		t.Fatal("Expected authentication error, got nil")
	}

	if result != nil {
		t.Error("Expected nil result when authentication fails")
	}
}

func TestAuthHandler_GetUserInfo_WithoutStorage(t *testing.T) {
	handler := NewAuthHandler(nil, "test-client", "test-secret")

	args := UserInfoArgs{}
	params := &mcp.CallToolParamsFor[UserInfoArgs]{
		Arguments: args,
	}

	// Create context with test user ID
	ctx := context.WithValue(context.Background(), auth.UserIDKey, "test-user-1")
	result, err := handler.GetUserInfo(ctx, nil, params)

	if err == nil {
		t.Fatal("Expected storage error, got nil")
	}

	if result != nil {
		t.Error("Expected nil result when storage is not initialized")
	}
}

func TestAuthHandler_DeleteAccount_WithoutStorage(t *testing.T) {
	handler := NewAuthHandler(nil, "test-client", "test-secret")

	args := AccountDeleteArgs{
		Confirm: true,
	}
	params := &mcp.CallToolParamsFor[AccountDeleteArgs]{
		Arguments: args,
	}

	// Create context with test user ID
	ctx := context.WithValue(context.Background(), auth.UserIDKey, "test-user-1")
	result, err := handler.DeleteAccount(ctx, nil, params)

	if err == nil {
		t.Fatal("Expected storage error, got nil")
	}

	if result != nil {
		t.Error("Expected nil result when storage is not initialized")
	}
}
