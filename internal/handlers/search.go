package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pankona/memoya/internal/auth"
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

	// Get user ID from context (set by auth middleware)
	userID, err := auth.RequireAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	// Default to "all" if type not specified
	searchType := args.Type
	if searchType == "" {
		searchType = "all"
	}

	// Create search filters with user isolation
	filters := storage.SearchFilters{
		UserID: userID,
		Type:   searchType,
		Tags:   args.Tags,
	}

	// Perform search
	results, err := h.storage.Search(ctx, args.Query, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	searchResult := SearchResult{
		Success: true,
		Query:   args.Query,
		Tags:    args.Tags,
		Type:    searchType,
		Results: SearchItems{
			Todos: results.Todos,
			Memos: results.Memos,
		},
		Message: fmt.Sprintf("Found %d todos and %d memos", len(results.Todos), len(results.Memos)),
	}

	jsonBytes, err := json.Marshal(searchResult)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[SearchResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
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
