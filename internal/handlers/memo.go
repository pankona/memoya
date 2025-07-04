package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pankona/memoya/internal/auth"
	"github.com/pankona/memoya/internal/models"
	"github.com/pankona/memoya/internal/storage"
)

type MemoHandler struct {
	storage storage.Storage
}

func NewMemoHandler() *MemoHandler {
	// TODO: Initialize with actual storage
	return &MemoHandler{}
}

func NewMemoHandlerWithStorage(storage storage.Storage) *MemoHandler {
	return &MemoHandler{
		storage: storage,
	}
}

// MemoCreateArgs represents arguments for creating a memo
type MemoCreateArgs struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	LinkedTodos []string `json:"linked_todos,omitempty"`
}

// MemoResult represents the result of memo operations
type MemoResult struct {
	Success bool         `json:"success"`
	Memo    *models.Memo `json:"memo"`
	Message string       `json:"message"`
}

func (h *MemoHandler) Create(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[MemoCreateArgs]) (*mcp.CallToolResultFor[MemoResult], error) {
	args := params.Arguments

	if h.storage == nil {
		return nil, fmt.Errorf("storage not initialized")
	}

	// Get user ID from context (set by auth middleware)
	userID, err := auth.RequireAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	memo := &models.Memo{
		ID:           uuid.New().String(),
		UserID:       userID,
		Title:        args.Title,
		Description:  args.Description,
		Tags:         args.Tags,
		LinkedTodos:  args.LinkedTodos,
		CreatedAt:    time.Now(),
		LastModified: time.Now(),
	}

	// Save to storage
	err = h.storage.CreateMemo(ctx, memo)
	if err != nil {
		return nil, fmt.Errorf("failed to create memo: %w", err)
	}

	// Create result
	result := MemoResult{
		Success: true,
		Memo:    memo,
		Message: fmt.Sprintf("Memo '%s' created successfully with ID: %s", memo.Title, memo.ID),
	}

	// Convert to JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[MemoResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}

type MemoListArgs struct {
	Tags []string `json:"tags,omitempty"`
}

type MemoListResult struct {
	Success bool           `json:"success"`
	Memos   []*models.Memo `json:"memos"`
	Message string         `json:"message"`
}

func (h *MemoHandler) List(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[MemoListArgs]) (*mcp.CallToolResultFor[MemoListResult], error) {
	args := params.Arguments

	if h.storage == nil {
		return nil, fmt.Errorf("storage not initialized")
	}

	// Get user ID from context (set by auth middleware)
	userID, err := auth.RequireAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	// Create filters with user isolation
	filters := storage.MemoFilters{
		UserID: userID,
		Tags:   args.Tags,
	}

	// Fetch from storage
	memos, err := h.storage.ListMemos(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list memos: %w", err)
	}

	result := MemoListResult{
		Success: true,
		Memos:   memos,
		Message: fmt.Sprintf("Found %d memos", len(memos)),
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[MemoListResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}

type MemoUpdateArgs struct {
	ID          string   `json:"id"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	LinkedTodos []string `json:"linked_todos,omitempty"`
}

func (h *MemoHandler) Update(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[MemoUpdateArgs]) (*mcp.CallToolResultFor[MemoResult], error) {
	args := params.Arguments

	if h.storage == nil {
		return nil, fmt.Errorf("storage not initialized")
	}

	// Get user ID from context (set by auth middleware)
	userID, err := auth.RequireAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	// Fetch existing memo from storage
	memo, err := h.storage.GetMemo(ctx, args.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get memo: %w", err)
	}

	// Check ownership
	if memo.UserID != userID {
		return nil, fmt.Errorf("access denied: memo belongs to different user")
	}

	// Update fields
	if args.Title != "" {
		memo.Title = args.Title
	}

	if args.Description != "" {
		memo.Description = args.Description
	}

	if len(args.Tags) > 0 {
		memo.Tags = args.Tags
	}

	if len(args.LinkedTodos) > 0 {
		memo.LinkedTodos = args.LinkedTodos
	}

	memo.LastModified = time.Now()

	// Save to storage
	err = h.storage.UpdateMemo(ctx, memo)
	if err != nil {
		return nil, fmt.Errorf("failed to update memo: %w", err)
	}

	// Create result
	result := MemoResult{
		Success: true,
		Memo:    memo,
		Message: fmt.Sprintf("Memo '%s' updated successfully", memo.Title),
	}

	// Convert to JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[MemoResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}

type MemoDeleteArgs struct {
	ID string `json:"id"`
}

type MemoDeleteResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (h *MemoHandler) Delete(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[MemoDeleteArgs]) (*mcp.CallToolResultFor[MemoDeleteResult], error) {
	args := params.Arguments

	if h.storage == nil {
		return nil, fmt.Errorf("storage not initialized")
	}

	// Get user ID from context (set by auth middleware)
	userID, err := auth.RequireAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	// Fetch existing memo to check ownership
	memo, err := h.storage.GetMemo(ctx, args.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get memo: %w", err)
	}

	// Check ownership
	if memo.UserID != userID {
		return nil, fmt.Errorf("access denied: memo belongs to different user")
	}

	// Delete from storage
	err = h.storage.DeleteMemo(ctx, args.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete memo: %w", err)
	}

	result := MemoDeleteResult{
		Success: true,
		Message: fmt.Sprintf("Memo %s deleted successfully", args.ID),
	}

	// Convert to JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[MemoDeleteResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}
