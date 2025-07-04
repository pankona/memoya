package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pankona/memoya/internal/models"
)

func TestTodoHandler_Create(t *testing.T) {
	mockStorage := NewMockStorage()
	handler := NewTodoHandlerWithStorage(mockStorage)

	args := TodoCreateArgs{
		Title:       "Test Todo",
		Description: "Test Description",
		Status:      "todo",
		Priority:    "high",
		Tags:        []string{"test", "urgent"},
		ParentID:    "parent-123",
	}

	params := &mcp.CallToolParamsFor[TodoCreateArgs]{
		Arguments: args,
	}

	result, err := handler.Create(context.Background(), nil, params)

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

	if len(mockStorage.todos) != 1 {
		t.Fatalf("Expected 1 todo in storage, got %d", len(mockStorage.todos))
	}

	for _, todo := range mockStorage.todos {
		if todo.Title != args.Title {
			t.Errorf("Expected title %s, got %s", args.Title, todo.Title)
		}
		if todo.Description != args.Description {
			t.Errorf("Expected description %s, got %s", args.Description, todo.Description)
		}
		if string(todo.Status) != args.Status {
			t.Errorf("Expected status %s, got %s", args.Status, string(todo.Status))
		}
		if string(todo.Priority) != args.Priority {
			t.Errorf("Expected priority %s, got %s", args.Priority, string(todo.Priority))
		}
		if len(todo.Tags) != len(args.Tags) {
			t.Errorf("Expected %d tags, got %d", len(args.Tags), len(todo.Tags))
		}
		if todo.ParentID != args.ParentID {
			t.Errorf("Expected parent ID %s, got %s", args.ParentID, todo.ParentID)
		}
	}
}

func TestTodoHandler_CreateWithDefaults(t *testing.T) {
	mockStorage := NewMockStorage()
	handler := NewTodoHandlerWithStorage(mockStorage)

	args := TodoCreateArgs{
		Title: "Test Todo",
	}

	params := &mcp.CallToolParamsFor[TodoCreateArgs]{
		Arguments: args,
	}

	result, err := handler.Create(context.Background(), nil, params)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	for _, todo := range mockStorage.todos {
		if todo.Status != models.StatusBacklog {
			t.Errorf("Expected default status %s, got %s", models.StatusBacklog, todo.Status)
		}
		if todo.Priority != models.PriorityNormal {
			t.Errorf("Expected default priority %s, got %s", models.PriorityNormal, todo.Priority)
		}
	}
}

func TestTodoHandler_List(t *testing.T) {
	mockStorage := NewMockStorage()
	mockStorage.SetupTestData()
	handler := NewTodoHandlerWithStorage(mockStorage)

	args := TodoListArgs{}

	params := &mcp.CallToolParamsFor[TodoListArgs]{
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

	var todoListResult TodoListResult
	err = json.Unmarshal([]byte(textContent.Text), &todoListResult)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if !todoListResult.Success {
		t.Error("Expected success to be true")
	}

	if len(todoListResult.Todos) != 2 {
		t.Errorf("Expected 2 todos, got %d", len(todoListResult.Todos))
	}

	if todoListResult.Message == "" {
		t.Error("Expected non-empty message")
	}
}

func TestTodoHandler_ListWithStatusFilter(t *testing.T) {
	mockStorage := NewMockStorage()
	mockStorage.SetupTestData()
	handler := NewTodoHandlerWithStorage(mockStorage)

	args := TodoListArgs{
		Status: "todo",
	}

	params := &mcp.CallToolParamsFor[TodoListArgs]{
		Arguments: args,
	}

	result, err := handler.List(context.Background(), nil, params)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var todoListResult TodoListResult
	textContent := result.Content[0].(*mcp.TextContent)
	json.Unmarshal([]byte(textContent.Text), &todoListResult)

	if len(todoListResult.Todos) != 1 {
		t.Errorf("Expected 1 todo with status 'todo', got %d", len(todoListResult.Todos))
	}
}

func TestTodoHandler_ListWithPriorityFilter(t *testing.T) {
	mockStorage := NewMockStorage()
	mockStorage.SetupTestData()
	handler := NewTodoHandlerWithStorage(mockStorage)

	args := TodoListArgs{
		Priority: "high",
	}

	params := &mcp.CallToolParamsFor[TodoListArgs]{
		Arguments: args,
	}

	result, err := handler.List(context.Background(), nil, params)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var todoListResult TodoListResult
	textContent := result.Content[0].(*mcp.TextContent)
	json.Unmarshal([]byte(textContent.Text), &todoListResult)

	if len(todoListResult.Todos) != 1 {
		t.Errorf("Expected 1 todo with priority 'high', got %d", len(todoListResult.Todos))
	}
}

func TestTodoHandler_ListWithTagFilter(t *testing.T) {
	mockStorage := NewMockStorage()
	mockStorage.SetupTestData()
	handler := NewTodoHandlerWithStorage(mockStorage)

	args := TodoListArgs{
		Tags: []string{"work"},
	}

	params := &mcp.CallToolParamsFor[TodoListArgs]{
		Arguments: args,
	}

	result, err := handler.List(context.Background(), nil, params)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var todoListResult TodoListResult
	textContent := result.Content[0].(*mcp.TextContent)
	json.Unmarshal([]byte(textContent.Text), &todoListResult)

	if len(todoListResult.Todos) != 1 {
		t.Errorf("Expected 1 todo with 'work' tag, got %d", len(todoListResult.Todos))
	}
}

func TestTodoHandler_Update(t *testing.T) {
	mockStorage := NewMockStorage()
	mockStorage.SetupTestData()
	handler := NewTodoHandlerWithStorage(mockStorage)

	args := TodoUpdateArgs{
		ID:          "test-todo-1",
		Title:       "Updated Title",
		Description: "Updated Description",
		Status:      "done",
		Priority:    "normal",
		Tags:        []string{"updated", "test"},
	}

	params := &mcp.CallToolParamsFor[TodoUpdateArgs]{
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

	updatedTodo, err := mockStorage.GetTodo(context.Background(), "test-todo-1")
	if err != nil {
		t.Fatalf("Failed to get updated todo: %v", err)
	}

	if updatedTodo.Title != args.Title {
		t.Errorf("Expected title %s, got %s", args.Title, updatedTodo.Title)
	}

	if updatedTodo.Description != args.Description {
		t.Errorf("Expected description %s, got %s", args.Description, updatedTodo.Description)
	}

	if string(updatedTodo.Status) != args.Status {
		t.Errorf("Expected status %s, got %s", args.Status, string(updatedTodo.Status))
	}

	if string(updatedTodo.Priority) != args.Priority {
		t.Errorf("Expected priority %s, got %s", args.Priority, string(updatedTodo.Priority))
	}

	if len(updatedTodo.Tags) != len(args.Tags) {
		t.Errorf("Expected %d tags, got %d", len(args.Tags), len(updatedTodo.Tags))
	}

	if updatedTodo.ClosedAt == nil {
		t.Error("Expected ClosedAt to be set when status is 'done'")
	}
}

func TestTodoHandler_Delete(t *testing.T) {
	mockStorage := NewMockStorage()
	mockStorage.SetupTestData()
	handler := NewTodoHandlerWithStorage(mockStorage)

	args := TodoDeleteArgs{
		ID: "test-todo-1",
	}

	params := &mcp.CallToolParamsFor[TodoDeleteArgs]{
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

	if len(mockStorage.todos) != 1 {
		t.Errorf("Expected 1 todo remaining in storage, got %d", len(mockStorage.todos))
	}

	_, err = mockStorage.GetTodo(context.Background(), "test-todo-1")
	if err == nil {
		t.Error("Expected error when getting deleted todo, got nil")
	}
}

func TestTodoHandler_CreateWithoutStorage(t *testing.T) {
	handler := NewTodoHandler()

	args := TodoCreateArgs{
		Title: "Test Todo",
	}

	params := &mcp.CallToolParamsFor[TodoCreateArgs]{
		Arguments: args,
	}

	result, err := handler.Create(context.Background(), nil, params)

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
}

func TestTodoHandler_ListWithoutStorage(t *testing.T) {
	handler := NewTodoHandler()

	args := TodoListArgs{}

	params := &mcp.CallToolParamsFor[TodoListArgs]{
		Arguments: args,
	}

	result, err := handler.List(context.Background(), nil, params)

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
}

func TestTodoHandler_UpdateWithoutStorage(t *testing.T) {
	handler := NewTodoHandler()

	args := TodoUpdateArgs{
		ID:    "test-id",
		Title: "Test Title",
	}

	params := &mcp.CallToolParamsFor[TodoUpdateArgs]{
		Arguments: args,
	}

	_, err := handler.Update(context.Background(), nil, params)

	if err == nil {
		t.Fatal("Expected error when storage is not initialized, got nil")
	}
}

func TestTodoHandler_DeleteWithoutStorage(t *testing.T) {
	handler := NewTodoHandler()

	args := TodoDeleteArgs{
		ID: "test-id",
	}

	params := &mcp.CallToolParamsFor[TodoDeleteArgs]{
		Arguments: args,
	}

	_, err := handler.Delete(context.Background(), nil, params)

	if err == nil {
		t.Fatal("Expected error when storage is not initialized, got nil")
	}
}
