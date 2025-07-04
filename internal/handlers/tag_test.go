package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestTagHandler_List(t *testing.T) {
	mockStorage := NewMockStorage()
	mockStorage.SetupTestData()
	handler := NewTagHandler(mockStorage)

	args := TagListArgs{}

	params := &mcp.CallToolParamsFor[TagListArgs]{
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

	var tagListResult TagListResult
	err = json.Unmarshal([]byte(textContent.Text), &tagListResult)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if !tagListResult.Success {
		t.Error("Expected success to be true")
	}

	if len(tagListResult.Tags) == 0 {
		t.Error("Expected at least some tags")
	}

	if tagListResult.Count != len(tagListResult.Tags) {
		t.Errorf("Expected count %d to match tags length %d", tagListResult.Count, len(tagListResult.Tags))
	}

	if tagListResult.Message == "" {
		t.Error("Expected non-empty message")
	}

	expectedTags := map[string]bool{
		"work":     true,
		"urgent":   true,
		"personal": true,
		"notes":    true,
		"ideas":    true,
	}

	for _, tag := range tagListResult.Tags {
		if !expectedTags[tag] {
			t.Errorf("Unexpected tag: %s", tag)
		}
	}
}

func TestTagHandler_ListEmpty(t *testing.T) {
	mockStorage := NewMockStorage()
	handler := NewTagHandler(mockStorage)

	args := TagListArgs{}

	params := &mcp.CallToolParamsFor[TagListArgs]{
		Arguments: args,
	}

	result, err := handler.List(context.Background(), nil, params)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var tagListResult TagListResult
	textContent := result.Content[0].(*mcp.TextContent)
	json.Unmarshal([]byte(textContent.Text), &tagListResult)

	if !tagListResult.Success {
		t.Error("Expected success to be true")
	}

	if len(tagListResult.Tags) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(tagListResult.Tags))
	}

	if tagListResult.Count != 0 {
		t.Errorf("Expected count 0, got %d", tagListResult.Count)
	}
}

func TestTagHandler_ListWithoutStorage(t *testing.T) {
	handler := NewTagHandler(nil)

	args := TagListArgs{}

	params := &mcp.CallToolParamsFor[TagListArgs]{
		Arguments: args,
	}

	_, err := handler.List(context.Background(), nil, params)

	if err == nil {
		t.Fatal("Expected error when storage is not initialized, got nil")
	}
}

func TestTagHandler_ListUniqueTags(t *testing.T) {
	mockStorage := NewMockStorage()
	mockStorage.SetupTestData()
	handler := NewTagHandler(mockStorage)

	args := TagListArgs{}

	params := &mcp.CallToolParamsFor[TagListArgs]{
		Arguments: args,
	}

	result, err := handler.List(context.Background(), nil, params)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var tagListResult TagListResult
	textContent := result.Content[0].(*mcp.TextContent)
	json.Unmarshal([]byte(textContent.Text), &tagListResult)

	tagMap := make(map[string]int)
	for _, tag := range tagListResult.Tags {
		tagMap[tag]++
		if tagMap[tag] > 1 {
			t.Errorf("Tag %s appears multiple times, expected unique tags", tag)
		}
	}
}

func TestTagHandler_ListConsistentCount(t *testing.T) {
	mockStorage := NewMockStorage()
	mockStorage.SetupTestData()
	handler := NewTagHandler(mockStorage)

	args := TagListArgs{}

	params := &mcp.CallToolParamsFor[TagListArgs]{
		Arguments: args,
	}

	result, err := handler.List(context.Background(), nil, params)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var tagListResult TagListResult
	textContent := result.Content[0].(*mcp.TextContent)
	json.Unmarshal([]byte(textContent.Text), &tagListResult)

	if tagListResult.Count != len(tagListResult.Tags) {
		t.Errorf("Count field %d does not match actual tag count %d", tagListResult.Count, len(tagListResult.Tags))
	}
}
