package storage

import (
	"context"

	"github.com/pankona/memoya/internal/models"
)

// Storage defines the interface for data persistence
type Storage interface {
	// Todo operations
	CreateTodo(ctx context.Context, todo *models.Todo) error
	GetTodo(ctx context.Context, id string) (*models.Todo, error)
	UpdateTodo(ctx context.Context, todo *models.Todo) error
	DeleteTodo(ctx context.Context, id string) error
	ListTodos(ctx context.Context, filters TodoFilters) ([]*models.Todo, error)

	// Memo operations
	CreateMemo(ctx context.Context, memo *models.Memo) error
	GetMemo(ctx context.Context, id string) (*models.Memo, error)
	UpdateMemo(ctx context.Context, memo *models.Memo) error
	DeleteMemo(ctx context.Context, id string) error
	ListMemos(ctx context.Context, filters MemoFilters) ([]*models.Memo, error)

	// Search operations
	Search(ctx context.Context, query string, filters SearchFilters) (*SearchResults, error)

	// Tag operations
	GetAllTags(ctx context.Context) ([]string, error)
}

type TodoFilters struct {
	Status   *models.TodoStatus
	Priority *models.TodoPriority
	Tags     []string
	ParentID *string
}

type MemoFilters struct {
	Tags []string
}

type SearchFilters struct {
	Tags []string
	Type string // "todo", "memo", or "all"
}

type SearchResults struct {
	Todos []*models.Todo
	Memos []*models.Memo
}
