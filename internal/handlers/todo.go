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

type TodoHandler struct {
	storage storage.Storage
}

func NewTodoHandler() *TodoHandler {
	// TODO: Initialize with actual storage
	return &TodoHandler{}
}

func NewTodoHandlerWithStorage(storage storage.Storage) *TodoHandler {
	return &TodoHandler{
		storage: storage,
	}
}

// TodoCreateArgs represents arguments for creating a todo
type TodoCreateArgs struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status,omitempty"`
	Priority    string   `json:"priority,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	ParentID    string   `json:"parent_id,omitempty"`
}

// TodoResult represents the result of todo operations
type TodoResult struct {
	Success bool         `json:"success"`
	Todo    *models.Todo `json:"todo"`
	Message string       `json:"message"`
}

func (h *TodoHandler) Create(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[TodoCreateArgs]) (*mcp.CallToolResultFor[TodoResult], error) {
	args := params.Arguments

	// Get user ID from context (set by auth middleware)
	userID, err := auth.RequireAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	todo := &models.Todo{
		ID:           uuid.New().String(),
		UserID:       userID,
		Title:        args.Title,
		Description:  args.Description,
		Tags:         args.Tags,
		ParentID:     args.ParentID,
		CreatedAt:    time.Now(),
		LastModified: time.Now(),
	}

	// Set status with default
	if args.Status != "" {
		todo.Status = models.TodoStatus(args.Status)
	} else {
		todo.Status = models.StatusBacklog
	}

	// Set priority with default
	if args.Priority != "" {
		todo.Priority = models.TodoPriority(args.Priority)
	} else {
		todo.Priority = models.PriorityNormal
	}

	// Save to storage
	if h.storage != nil {
		err := h.storage.CreateTodo(ctx, todo)
		if err != nil {
			return nil, fmt.Errorf("failed to create todo: %w", err)
		}
	}

	// Create result
	result := TodoResult{
		Success: true,
		Todo:    todo,
		Message: fmt.Sprintf("Todo '%s' created successfully with ID: %s", todo.Title, todo.ID),
	}

	// Convert to JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[TodoResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}

type TodoListArgs struct {
	Status   string   `json:"status,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Priority string   `json:"priority,omitempty"`
}

type TodoListResult struct {
	Success bool           `json:"success"`
	Todos   []*models.Todo `json:"todos"`
	Message string         `json:"message"`
}

func (h *TodoHandler) List(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[TodoListArgs]) (*mcp.CallToolResultFor[TodoListResult], error) {
	args := params.Arguments

	// Get user ID from context (set by auth middleware)
	userID, err := auth.RequireAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	if h.storage != nil {
		// Build filters from arguments with user isolation
		filters := storage.TodoFilters{
			UserID: userID,
		}

		if args.Status != "" {
			status := models.TodoStatus(args.Status)
			filters.Status = &status
		}

		if args.Priority != "" {
			priority := models.TodoPriority(args.Priority)
			filters.Priority = &priority
		}

		if len(args.Tags) > 0 {
			filters.Tags = args.Tags
		}

		todos, err := h.storage.ListTodos(ctx, filters)
		if err != nil {
			return &mcp.CallToolResultFor[TodoListResult]{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to list todos: %v", err)},
				},
				IsError: true,
			}, nil
		}

		result := TodoListResult{
			Success: true,
			Todos:   todos,
			Message: fmt.Sprintf("Found %d todos", len(todos)),
		}

		jsonBytes, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}

		return &mcp.CallToolResultFor[TodoListResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(jsonBytes)},
			},
		}, nil
	}

	return &mcp.CallToolResultFor[TodoListResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Todo listing (using mock storage - no todos)"},
		},
	}, nil
}

type TodoUpdateArgs struct {
	ID          string   `json:"id"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status,omitempty"`
	Priority    string   `json:"priority,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

func (h *TodoHandler) Update(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[TodoUpdateArgs]) (*mcp.CallToolResultFor[TodoResult], error) {
	args := params.Arguments

	if h.storage == nil {
		return nil, fmt.Errorf("storage not initialized")
	}

	// Get user ID from context (set by auth middleware)
	userID, err := auth.RequireAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	// Fetch existing todo from storage
	todo, err := h.storage.GetTodo(ctx, args.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	// Check ownership
	if todo.UserID != userID {
		return nil, fmt.Errorf("access denied: todo belongs to different user")
	}

	// Update fields
	if args.Title != "" {
		todo.Title = args.Title
	}

	if args.Description != "" {
		todo.Description = args.Description
	}

	if args.Status != "" {
		todo.Status = models.TodoStatus(args.Status)
		if args.Status == "done" && todo.ClosedAt == nil {
			now := time.Now()
			todo.ClosedAt = &now
		}
	}

	if args.Priority != "" {
		todo.Priority = models.TodoPriority(args.Priority)
	}

	if len(args.Tags) > 0 {
		todo.Tags = args.Tags
	}

	todo.LastModified = time.Now()

	// Save to storage
	err = h.storage.UpdateTodo(ctx, todo)
	if err != nil {
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}

	// Create result
	result := TodoResult{
		Success: true,
		Todo:    todo,
		Message: fmt.Sprintf("Todo '%s' updated successfully", todo.Title),
	}

	// Convert to JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[TodoResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}

type TodoDeleteArgs struct {
	ID string `json:"id"`
}

type DeleteResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (h *TodoHandler) Delete(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[TodoDeleteArgs]) (*mcp.CallToolResultFor[DeleteResult], error) {
	args := params.Arguments

	if h.storage == nil {
		return nil, fmt.Errorf("storage not initialized")
	}

	// Get user ID from context (set by auth middleware)
	userID, err := auth.RequireAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	// Fetch existing todo to check ownership
	todo, err := h.storage.GetTodo(ctx, args.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	// Check ownership
	if todo.UserID != userID {
		return nil, fmt.Errorf("access denied: todo belongs to different user")
	}

	// Delete from storage
	err = h.storage.DeleteTodo(ctx, args.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete todo: %w", err)
	}

	result := DeleteResult{
		Success: true,
		Message: fmt.Sprintf("Todo %s deleted successfully", args.ID),
	}

	// Convert to JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[DeleteResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}
