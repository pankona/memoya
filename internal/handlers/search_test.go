package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pankona/memoya/internal/auth"
)

func TestSearchHandler_Search(t *testing.T) {
	mockStorage := NewMockStorage()
	mockStorage.SetupTestData()
	handler := NewSearchHandler(mockStorage)

	args := SearchArgs{
		Query: "Test",
		Tags:  []string{},
		Type:  "all",
	}

	params := &mcp.CallToolParamsFor[SearchArgs]{
		Arguments: args,
	}

	// Create context with test user ID
	ctx := context.WithValue(context.Background(), auth.UserIDKey, "test-user-1")
	result, err := handler.Search(ctx, nil, params)

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

	var searchResult SearchResult
	err = json.Unmarshal([]byte(textContent.Text), &searchResult)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if !searchResult.Success {
		t.Error("Expected success to be true")
	}

	if searchResult.Query != args.Query {
		t.Errorf("Expected query %s, got %s", args.Query, searchResult.Query)
	}

	if searchResult.Type != args.Type {
		t.Errorf("Expected type %s, got %s", args.Type, searchResult.Type)
	}

	if searchResult.Message == "" {
		t.Error("Expected non-empty message")
	}

	totalResults := len(searchResult.Results.Todos) + len(searchResult.Results.Memos)
	if totalResults == 0 {
		t.Error("Expected at least some search results")
	}
}

func TestSearchHandler_SearchTodosOnly(t *testing.T) {
	mockStorage := NewMockStorage()
	mockStorage.SetupTestData()
	handler := NewSearchHandler(mockStorage)

	args := SearchArgs{
		Query: "Test",
		Tags:  []string{},
		Type:  "todo",
	}

	params := &mcp.CallToolParamsFor[SearchArgs]{
		Arguments: args,
	}

	// Create context with test user ID
	ctx := context.WithValue(context.Background(), auth.UserIDKey, "test-user-1")
	result, err := handler.Search(ctx, nil, params)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var searchResult SearchResult
	textContent := result.Content[0].(*mcp.TextContent)
	json.Unmarshal([]byte(textContent.Text), &searchResult)

	if len(searchResult.Results.Todos) == 0 {
		t.Error("Expected at least one todo result")
	}

	if len(searchResult.Results.Memos) != 0 {
		t.Error("Expected no memo results when searching for todos only")
	}
}

func TestSearchHandler_SearchMemosOnly(t *testing.T) {
	mockStorage := NewMockStorage()
	mockStorage.SetupTestData()
	handler := NewSearchHandler(mockStorage)

	args := SearchArgs{
		Query: "Test",
		Tags:  []string{},
		Type:  "memo",
	}

	params := &mcp.CallToolParamsFor[SearchArgs]{
		Arguments: args,
	}

	// Create context with test user ID
	ctx := context.WithValue(context.Background(), auth.UserIDKey, "test-user-1")
	result, err := handler.Search(ctx, nil, params)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var searchResult SearchResult
	textContent := result.Content[0].(*mcp.TextContent)
	json.Unmarshal([]byte(textContent.Text), &searchResult)

	if len(searchResult.Results.Memos) == 0 {
		t.Error("Expected at least one memo result")
	}

	if len(searchResult.Results.Todos) != 0 {
		t.Error("Expected no todo results when searching for memos only")
	}
}

func TestSearchHandler_SearchWithTags(t *testing.T) {
	mockStorage := NewMockStorage()
	mockStorage.SetupTestData()
	handler := NewSearchHandler(mockStorage)

	args := SearchArgs{
		Query: "",
		Tags:  []string{"work"},
		Type:  "all",
	}

	params := &mcp.CallToolParamsFor[SearchArgs]{
		Arguments: args,
	}

	// Create context with test user ID
	ctx := context.WithValue(context.Background(), auth.UserIDKey, "test-user-1")
	result, err := handler.Search(ctx, nil, params)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var searchResult SearchResult
	textContent := result.Content[0].(*mcp.TextContent)
	json.Unmarshal([]byte(textContent.Text), &searchResult)

	if len(searchResult.Results.Todos) == 0 && len(searchResult.Results.Memos) == 0 {
		t.Error("Expected at least some results with 'work' tag")
	}

	for _, todo := range searchResult.Results.Todos {
		found := false
		for _, tag := range todo.Tags {
			if tag == "work" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected todo to have 'work' tag")
		}
	}

	for _, memo := range searchResult.Results.Memos {
		found := false
		for _, tag := range memo.Tags {
			if tag == "work" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected memo to have 'work' tag")
		}
	}
}

func TestSearchHandler_SearchDefaultType(t *testing.T) {
	mockStorage := NewMockStorage()
	mockStorage.SetupTestData()
	handler := NewSearchHandler(mockStorage)

	args := SearchArgs{
		Query: "Test",
		Tags:  []string{},
		Type:  "",
	}

	params := &mcp.CallToolParamsFor[SearchArgs]{
		Arguments: args,
	}

	// Create context with test user ID
	ctx := context.WithValue(context.Background(), auth.UserIDKey, "test-user-1")
	result, err := handler.Search(ctx, nil, params)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var searchResult SearchResult
	textContent := result.Content[0].(*mcp.TextContent)
	json.Unmarshal([]byte(textContent.Text), &searchResult)

	if searchResult.Type != "all" {
		t.Errorf("Expected type to default to 'all', got %s", searchResult.Type)
	}
}

func TestSearchHandler_SearchWithoutStorage(t *testing.T) {
	handler := NewSearchHandler(nil)

	args := SearchArgs{
		Query: "Test",
	}

	params := &mcp.CallToolParamsFor[SearchArgs]{
		Arguments: args,
	}

	_, err := handler.Search(context.Background(), nil, params)

	if err == nil {
		t.Fatal("Expected error when storage is not initialized, got nil")
	}
}

func TestSearchHandler_SearchEmptyQuery(t *testing.T) {
	mockStorage := NewMockStorage()
	mockStorage.SetupTestData()
	handler := NewSearchHandler(mockStorage)

	args := SearchArgs{
		Query: "",
		Tags:  []string{},
		Type:  "all",
	}

	params := &mcp.CallToolParamsFor[SearchArgs]{
		Arguments: args,
	}

	// Create context with test user ID
	ctx := context.WithValue(context.Background(), auth.UserIDKey, "test-user-1")
	result, err := handler.Search(ctx, nil, params)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var searchResult SearchResult
	textContent := result.Content[0].(*mcp.TextContent)
	json.Unmarshal([]byte(textContent.Text), &searchResult)

	if searchResult.Query != "" {
		t.Errorf("Expected empty query, got %s", searchResult.Query)
	}

	totalResults := len(searchResult.Results.Todos) + len(searchResult.Results.Memos)
	if totalResults == 0 {
		t.Error("Expected some results even with empty query")
	}
}

func TestSearchHandler_SearchNoResults(t *testing.T) {
	mockStorage := NewMockStorage()
	handler := NewSearchHandler(mockStorage)

	args := SearchArgs{
		Query: "NonExistentContent",
		Tags:  []string{},
		Type:  "all",
	}

	params := &mcp.CallToolParamsFor[SearchArgs]{
		Arguments: args,
	}

	// Create context with test user ID
	ctx := context.WithValue(context.Background(), auth.UserIDKey, "test-user-1")
	result, err := handler.Search(ctx, nil, params)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var searchResult SearchResult
	textContent := result.Content[0].(*mcp.TextContent)
	json.Unmarshal([]byte(textContent.Text), &searchResult)

	if len(searchResult.Results.Todos) != 0 {
		t.Error("Expected no todo results")
	}

	if len(searchResult.Results.Memos) != 0 {
		t.Error("Expected no memo results")
	}
}
