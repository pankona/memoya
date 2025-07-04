package storage

import (
	"context"

	"github.com/pankona/memoya/internal/models"
)

// Storage defines the interface for data persistence
type Storage interface {
	// User operations
	CreateUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, id string) (*models.User, error)
	GetUserByGoogleID(ctx context.Context, googleID string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id string) error

	// Device auth operations
	CreateDeviceAuthSession(ctx context.Context, session *models.DeviceAuthSession) error
	GetDeviceAuthSession(ctx context.Context, deviceCode string) (*models.DeviceAuthSession, error)
	UpdateDeviceAuthSession(ctx context.Context, session *models.DeviceAuthSession) error
	DeleteDeviceAuthSession(ctx context.Context, deviceCode string) error

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
	GetAllTags(ctx context.Context, userID string) ([]string, error)
}

type TodoFilters struct {
	UserID   string // Required for user isolation
	Status   *models.TodoStatus
	Priority *models.TodoPriority
	Tags     []string
	ParentID *string
}

type MemoFilters struct {
	UserID string // Required for user isolation
	Tags   []string
}

type SearchFilters struct {
	UserID string // Required for user isolation
	Tags   []string
	Type   string // "todo", "memo", or "all"
}

type SearchResults struct {
	Todos []*models.Todo
	Memos []*models.Memo
}
