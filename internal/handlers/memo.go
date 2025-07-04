package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pankona/memoya/internal/auth"
	"github.com/pankona/memoya/internal/errors"
	"github.com/pankona/memoya/internal/models"
	"github.com/pankona/memoya/internal/storage"
	"github.com/pankona/memoya/internal/validation"
)

type MemoHandler struct {
	storage   storage.Storage
	validator *validation.MemoValidator
}

func NewMemoHandler() *MemoHandler {
	// TODO: Initialize with actual storage
	return &MemoHandler{
		validator: validation.NewMemoValidator(),
	}
}

func NewMemoHandlerWithStorage(storage storage.Storage) *MemoHandler {
	return &MemoHandler{
		storage:   storage,
		validator: validation.NewMemoValidator(),
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
		return nil, errors.NewInternalError(fmt.Errorf("storage not initialized"))
	}

	// Get user ID from context (set by auth middleware)
	userID, err := auth.RequireAuth(ctx)
	if err != nil {
		return nil, errors.NewAuthenticationError("authentication required")
	}

	// Validate input
	if err := h.validator.ValidateCreate(args.Title, args.Description, args.Tags); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// Sanitize input
	title, description, tags := validation.SanitizeMemoInput(args.Title, args.Description, args.Tags)

	memo := &models.Memo{
		ID:           uuid.New().String(),
		UserID:       userID,
		Title:        title,
		Description:  description,
		Tags:         tags,
		LinkedTodos:  args.LinkedTodos,
		CreatedAt:    time.Now(),
		LastModified: time.Now(),
	}

	// Save to storage
	err = h.storage.CreateMemo(ctx, memo)
	if err != nil {
		return nil, errors.NewDatabaseError(err).WithContext("operation", "create_memo").WithContext("memo_id", memo.ID)
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
		return nil, errors.NewInternalError(err).WithContext("operation", "marshal_result")
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
		return nil, errors.NewInternalError(fmt.Errorf("storage not initialized"))
	}

	// Get user ID from context (set by auth middleware)
	userID, err := auth.RequireAuth(ctx)
	if err != nil {
		return nil, errors.NewAuthenticationError("authentication required")
	}

	// Create filters with user isolation
	filters := storage.MemoFilters{
		UserID: userID,
		Tags:   args.Tags,
	}

	// Fetch from storage
	memos, err := h.storage.ListMemos(ctx, filters)
	if err != nil {
		return nil, errors.NewDatabaseError(err).WithContext("operation", "list_memos").WithContext("user_id", userID)
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
		return nil, errors.NewInternalError(fmt.Errorf("storage not initialized"))
	}

	// Get user ID from context (set by auth middleware)
	userID, err := auth.RequireAuth(ctx)
	if err != nil {
		return nil, errors.NewAuthenticationError("authentication required")
	}

	// Fetch existing memo from storage
	memo, err := h.storage.GetMemo(ctx, args.ID)
	if err != nil {
		return nil, errors.NewDatabaseError(err).WithContext("operation", "get_memo").WithContext("memo_id", args.ID)
	}

	// Check ownership
	if memo.UserID != userID {
		return nil, errors.NewAuthorizationError("access denied: memo belongs to different user").WithContext("memo_id", args.ID).WithContext("user_id", userID)
	}

	// Update fields with validation
	if args.Title != "" {
		if err := validation.ValidateString("title", args.Title, 1, validation.MaxMemoTitleLength, true); err != nil {
			return nil, errors.NewValidationError(err.Error()).WithContext("field", "title")
		}
		memo.Title = validation.SanitizeString(args.Title)
	}

	if args.Description != "" {
		if err := validation.ValidateString("description", args.Description, 0, validation.MaxMemoDescriptionLength, false); err != nil {
			return nil, errors.NewValidationError(err.Error()).WithContext("field", "description")
		}
		memo.Description = validation.SanitizeString(args.Description)
	}

	if len(args.Tags) > 0 {
		if err := validation.ValidateTags(args.Tags); err != nil {
			return nil, errors.NewValidationError(err.Error()).WithContext("field", "tags")
		}
		memo.Tags = validation.SanitizeTags(args.Tags)
	}

	if len(args.LinkedTodos) > 0 {
		memo.LinkedTodos = args.LinkedTodos
	}

	memo.LastModified = time.Now()

	// Save to storage
	err = h.storage.UpdateMemo(ctx, memo)
	if err != nil {
		return nil, errors.NewDatabaseError(err).WithContext("operation", "update_memo").WithContext("memo_id", memo.ID)
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
		return nil, errors.NewInternalError(err).WithContext("operation", "marshal_result")
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
		return nil, errors.NewInternalError(fmt.Errorf("storage not initialized"))
	}

	// Get user ID from context (set by auth middleware)
	userID, err := auth.RequireAuth(ctx)
	if err != nil {
		return nil, errors.NewAuthenticationError("authentication required")
	}

	// Fetch existing memo to check ownership
	memo, err := h.storage.GetMemo(ctx, args.ID)
	if err != nil {
		return nil, errors.NewDatabaseError(err).WithContext("operation", "get_memo").WithContext("memo_id", args.ID)
	}

	// Check ownership
	if memo.UserID != userID {
		return nil, errors.NewAuthorizationError("access denied: memo belongs to different user").WithContext("memo_id", args.ID).WithContext("user_id", userID)
	}

	// Delete from storage
	err = h.storage.DeleteMemo(ctx, args.ID)
	if err != nil {
		return nil, errors.NewDatabaseError(err).WithContext("operation", "delete_memo").WithContext("memo_id", args.ID)
	}

	result := MemoDeleteResult{
		Success: true,
		Message: fmt.Sprintf("Memo %s deleted successfully", args.ID),
	}

	// Convert to JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, errors.NewInternalError(err).WithContext("operation", "marshal_result")
	}

	return &mcp.CallToolResultFor[MemoDeleteResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}
