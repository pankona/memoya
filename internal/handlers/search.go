package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pankona/memoya/internal/models"
	"github.com/pankona/memoya/internal/storage"
)

// SearchArgs represents arguments for search
type SearchArgs struct {
	Query string   `json:"query,omitempty"`
	Tags  []string `json:"tags,omitempty"`
	Type  string   `json:"type,omitempty"`
}

// SearchResult represents the result of search operation
type SearchResult struct {
	Success bool        `json:"success"`
	Query   string      `json:"query"`
	Tags    []string    `json:"tags"`
	Type    string      `json:"type"`
	Results SearchItems `json:"results"`
	Message string      `json:"message"`
}

// SearchItems represents search results
type SearchItems struct {
	Todos []*models.Todo `json:"todos"`
	Memos []*models.Memo `json:"memos"`
}

type SearchHandler struct {
	storage storage.Storage
}

func NewSearchHandler(storage storage.Storage) *SearchHandler {
	return &SearchHandler{
		storage: storage,
	}
}

func (h *SearchHandler) Search(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[SearchArgs]) (*mcp.CallToolResultFor[SearchResult], error) {
	args := params.Arguments

	if h.storage == nil {
		return nil, fmt.Errorf("storage not initialized")
	}

	// Default to "all" if type not specified
	searchType := args.Type
	if searchType == "" {
		searchType = "all"
	}

	// Create search filters
	filters := storage.SearchFilters{
		Type: searchType,
		Tags: args.Tags,
	}

	// Perform search
	results, err := h.storage.Search(ctx, args.Query, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	return &mcp.CallToolResultFor[SearchResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Found %d todos and %d memos", len(results.Todos), len(results.Memos))},
		},
	}, nil
}

// Keep the old function for backward compatibility
func Search(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[SearchArgs]) (*mcp.CallToolResultFor[SearchResult], error) {
	// This function is kept for backward compatibility but should not be used
	return &mcp.CallToolResultFor[SearchResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Search not fully implemented yet - please use SearchHandler"},
		},
	}, nil
}
