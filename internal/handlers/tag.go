package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pankona/memoya/internal/storage"
)

type TagHandler struct {
	storage storage.Storage
}

func NewTagHandler(storage storage.Storage) *TagHandler {
	return &TagHandler{
		storage: storage,
	}
}

// TagListArgs represents arguments for listing tags
type TagListArgs struct {
	// No arguments needed for listing all tags
}

// TagListResult represents the result of tag list operation
type TagListResult struct {
	Success bool     `json:"success"`
	Tags    []string `json:"tags"`
	Count   int      `json:"count"`
	Message string   `json:"message"`
}

func (h *TagHandler) List(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[TagListArgs]) (*mcp.CallToolResultFor[TagListResult], error) {
	if h.storage == nil {
		return nil, fmt.Errorf("storage not initialized")
	}

	// Get all tags from storage
	tags, err := h.storage.GetAllTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	result := TagListResult{
		Success: true,
		Tags:    tags,
		Count:   len(tags),
		Message: fmt.Sprintf("Found %d unique tags", len(tags)),
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[TagListResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}
