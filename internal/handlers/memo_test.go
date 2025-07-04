package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pankona/memoya/internal/auth"
)

func TestMemoHandler_Create(t *testing.T) {
	mockStorage := NewMockStorage()
	handler := NewMemoHandlerWithStorage(mockStorage)

	args := MemoCreateArgs{
		Title:       "Test Memo",
		Description: "Test Description",
		Tags:        []string{"test", "memo"},
		LinkedTodos: []string{"todo-1", "todo-2"},
	}

	params := &mcp.CallToolParamsFor[MemoCreateArgs]{
		Arguments: args,
	}

	// Create context with test user ID
	ctx := context.WithValue(context.Background(), auth.UserIDKey, "test-user-1")
	result, err := handler.Create(ctx, nil, params)

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

	if textContent.Text == "" {
		t.Fatal("Expected non-empty text content")
	}

	if len(mockStorage.memos) != 1 {
		t.Fatalf("Expected 1 memo in storage, got %d", len(mockStorage.memos))
	}

	for _, memo := range mockStorage.memos {
		if memo.Title != args.Title {
			t.Errorf("Expected title %s, got %s", args.Title, memo.Title)
		}
		if memo.Description != args.Description {
			t.Errorf("Expected description %s, got %s", args.Description, memo.Description)
		}
		if len(memo.Tags) != len(args.Tags) {
			t.Errorf("Expected %d tags, got %d", len(args.Tags), len(memo.Tags))
		}
		if len(memo.LinkedTodos) != len(args.LinkedTodos) {
			t.Errorf("Expected %d linked todos, got %d", len(args.LinkedTodos), len(memo.LinkedTodos))
		}
	}
}

func TestMemoHandler_List(t *testing.T) {
	mockStorage := NewMockStorage()
	mockStorage.SetupTestData()
	handler := NewMemoHandlerWithStorage(mockStorage)

	args := MemoListArgs{
		Tags: []string{},
	}

	params := &mcp.CallToolParamsFor[MemoListArgs]{
		Arguments: args,
	}

	result, err := handler.List(context.Background(), nil, params)

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

	var memoListResult MemoListResult
	err = json.Unmarshal([]byte(textContent.Text), &memoListResult)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if !memoListResult.Success {
		t.Error("Expected success to be true")
	}

	if len(memoListResult.Memos) != 2 {
		t.Errorf("Expected 2 memos, got %d", len(memoListResult.Memos))
	}

	if memoListResult.Message == "" {
		t.Error("Expected non-empty message")
	}
}

func TestMemoHandler_List_WithTagFilter(t *testing.T) {
	mockStorage := NewMockStorage()
	mockStorage.SetupTestData()
	handler := NewMemoHandlerWithStorage(mockStorage)

	args := MemoListArgs{
		Tags: []string{"work"},
	}

	params := &mcp.CallToolParamsFor[MemoListArgs]{
		Arguments: args,
	}

	result, err := handler.List(context.Background(), nil, params)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var memoListResult MemoListResult
	textContent := result.Content[0].(*mcp.TextContent)
	json.Unmarshal([]byte(textContent.Text), &memoListResult)

	if len(memoListResult.Memos) != 1 {
		t.Errorf("Expected 1 memo with 'work' tag, got %d", len(memoListResult.Memos))
	}
}

func TestMemoHandler_Update(t *testing.T) {
	mockStorage := NewMockStorage()
	mockStorage.SetupTestData()
	handler := NewMemoHandlerWithStorage(mockStorage)

	args := MemoUpdateArgs{
		ID:          "test-memo-1",
		Title:       "Updated Title",
		Description: "Updated Description",
		Tags:        []string{"updated", "test"},
		LinkedTodos: []string{"new-todo"},
	}

	params := &mcp.CallToolParamsFor[MemoUpdateArgs]{
		Arguments: args,
	}

	result, err := handler.Update(context.Background(), nil, params)

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

	if textContent.Text == "" {
		t.Fatal("Expected non-empty text content")
	}

	updatedMemo, err := mockStorage.GetMemo(context.Background(), "test-memo-1")
	if err != nil {
		t.Fatalf("Failed to get updated memo: %v", err)
	}

	if updatedMemo.Title != args.Title {
		t.Errorf("Expected title %s, got %s", args.Title, updatedMemo.Title)
	}

	if updatedMemo.Description != args.Description {
		t.Errorf("Expected description %s, got %s", args.Description, updatedMemo.Description)
	}

	if len(updatedMemo.Tags) != len(args.Tags) {
		t.Errorf("Expected %d tags, got %d", len(args.Tags), len(updatedMemo.Tags))
	}

	if len(updatedMemo.LinkedTodos) != len(args.LinkedTodos) {
		t.Errorf("Expected %d linked todos, got %d", len(args.LinkedTodos), len(updatedMemo.LinkedTodos))
	}
}

func TestMemoHandler_Delete(t *testing.T) {
	mockStorage := NewMockStorage()
	mockStorage.SetupTestData()
	handler := NewMemoHandlerWithStorage(mockStorage)

	args := MemoDeleteArgs{
		ID: "test-memo-1",
	}

	params := &mcp.CallToolParamsFor[MemoDeleteArgs]{
		Arguments: args,
	}

	result, err := handler.Delete(context.Background(), nil, params)

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

	if textContent.Text == "" {
		t.Fatal("Expected non-empty text content")
	}

	if len(mockStorage.memos) != 1 {
		t.Errorf("Expected 1 memo remaining in storage, got %d", len(mockStorage.memos))
	}

	_, err = mockStorage.GetMemo(context.Background(), "test-memo-1")
	if err == nil {
		t.Error("Expected error when getting deleted memo, got nil")
	}
}

func TestMemoHandler_CreateWithoutStorage(t *testing.T) {
	handler := NewMemoHandler()

	args := MemoCreateArgs{
		Title: "Test Memo",
	}

	params := &mcp.CallToolParamsFor[MemoCreateArgs]{
		Arguments: args,
	}

	_, err := handler.Create(context.Background(), nil, params)

	if err == nil {
		t.Fatal("Expected error when storage is not initialized, got nil")
	}
}

func TestMemoHandler_ListWithoutStorage(t *testing.T) {
	handler := NewMemoHandler()

	args := MemoListArgs{}

	params := &mcp.CallToolParamsFor[MemoListArgs]{
		Arguments: args,
	}

	_, err := handler.List(context.Background(), nil, params)

	if err == nil {
		t.Fatal("Expected error when storage is not initialized, got nil")
	}
}

func TestMemoHandler_UpdateWithoutStorage(t *testing.T) {
	handler := NewMemoHandler()

	args := MemoUpdateArgs{
		ID:    "test-id",
		Title: "Test Title",
	}

	params := &mcp.CallToolParamsFor[MemoUpdateArgs]{
		Arguments: args,
	}

	_, err := handler.Update(context.Background(), nil, params)

	if err == nil {
		t.Fatal("Expected error when storage is not initialized, got nil")
	}
}

func TestMemoHandler_DeleteWithoutStorage(t *testing.T) {
	handler := NewMemoHandler()

	args := MemoDeleteArgs{
		ID: "test-id",
	}

	params := &mcp.CallToolParamsFor[MemoDeleteArgs]{
		Arguments: args,
	}

	_, err := handler.Delete(context.Background(), nil, params)

	if err == nil {
		t.Fatal("Expected error when storage is not initialized, got nil")
	}
}
