package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/pankona/memoya/internal/models"
	"github.com/pankona/memoya/internal/storage"
)

type MockStorage struct {
	todos map[string]*models.Todo
	memos map[string]*models.Memo
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		todos: make(map[string]*models.Todo),
		memos: make(map[string]*models.Memo),
	}
}

func (m *MockStorage) CreateTodo(ctx context.Context, todo *models.Todo) error {
	m.todos[todo.ID] = todo
	return nil
}

func (m *MockStorage) GetTodo(ctx context.Context, id string) (*models.Todo, error) {
	todo, exists := m.todos[id]
	if !exists {
		return nil, fmt.Errorf("todo not found")
	}
	return todo, nil
}

func (m *MockStorage) UpdateTodo(ctx context.Context, todo *models.Todo) error {
	if _, exists := m.todos[todo.ID]; !exists {
		return fmt.Errorf("todo not found")
	}
	m.todos[todo.ID] = todo
	return nil
}

func (m *MockStorage) DeleteTodo(ctx context.Context, id string) error {
	if _, exists := m.todos[id]; !exists {
		return fmt.Errorf("todo not found")
	}
	delete(m.todos, id)
	return nil
}

func (m *MockStorage) ListTodos(ctx context.Context, filters storage.TodoFilters) ([]*models.Todo, error) {
	var result []*models.Todo
	for _, todo := range m.todos {
		if m.matchesTodo(todo, filters) {
			result = append(result, todo)
		}
	}
	return result, nil
}

func (m *MockStorage) CreateMemo(ctx context.Context, memo *models.Memo) error {
	m.memos[memo.ID] = memo
	return nil
}

func (m *MockStorage) GetMemo(ctx context.Context, id string) (*models.Memo, error) {
	memo, exists := m.memos[id]
	if !exists {
		return nil, fmt.Errorf("memo not found")
	}
	return memo, nil
}

func (m *MockStorage) UpdateMemo(ctx context.Context, memo *models.Memo) error {
	if _, exists := m.memos[memo.ID]; !exists {
		return fmt.Errorf("memo not found")
	}
	m.memos[memo.ID] = memo
	return nil
}

func (m *MockStorage) DeleteMemo(ctx context.Context, id string) error {
	if _, exists := m.memos[id]; !exists {
		return fmt.Errorf("memo not found")
	}
	delete(m.memos, id)
	return nil
}

func (m *MockStorage) ListMemos(ctx context.Context, filters storage.MemoFilters) ([]*models.Memo, error) {
	var result []*models.Memo
	for _, memo := range m.memos {
		if m.matchesMemo(memo, filters) {
			result = append(result, memo)
		}
	}
	return result, nil
}

func (m *MockStorage) Search(ctx context.Context, query string, filters storage.SearchFilters) (*storage.SearchResults, error) {
	results := &storage.SearchResults{
		Todos: []*models.Todo{},
		Memos: []*models.Memo{},
	}

	if filters.Type == "todo" || filters.Type == "all" || filters.Type == "" {
		for _, todo := range m.todos {
			if m.matchesSearch(todo.Title, todo.Description, todo.Tags, query, filters.Tags) {
				results.Todos = append(results.Todos, todo)
			}
		}
	}

	if filters.Type == "memo" || filters.Type == "all" || filters.Type == "" {
		for _, memo := range m.memos {
			if m.matchesSearch(memo.Title, memo.Description, memo.Tags, query, filters.Tags) {
				results.Memos = append(results.Memos, memo)
			}
		}
	}

	return results, nil
}

func (m *MockStorage) GetAllTags(ctx context.Context) ([]string, error) {
	tagSet := make(map[string]bool)

	for _, todo := range m.todos {
		for _, tag := range todo.Tags {
			tagSet[tag] = true
		}
	}

	for _, memo := range m.memos {
		for _, tag := range memo.Tags {
			tagSet[tag] = true
		}
	}

	var tags []string
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	return tags, nil
}

func (m *MockStorage) matchesTodo(todo *models.Todo, filters storage.TodoFilters) bool {
	if filters.Status != nil && todo.Status != *filters.Status {
		return false
	}
	if filters.Priority != nil && todo.Priority != *filters.Priority {
		return false
	}
	if len(filters.Tags) > 0 {
		for _, filterTag := range filters.Tags {
			found := false
			for _, todoTag := range todo.Tags {
				if todoTag == filterTag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	return true
}

func (m *MockStorage) matchesMemo(memo *models.Memo, filters storage.MemoFilters) bool {
	if len(filters.Tags) > 0 {
		for _, filterTag := range filters.Tags {
			found := false
			for _, memoTag := range memo.Tags {
				if memoTag == filterTag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	return true
}

func (m *MockStorage) matchesSearch(title, description string, tags []string, query string, searchTags []string) bool {
	if query != "" {
		if !contains(title, query) && !contains(description, query) {
			queryFound := false
			for _, tag := range tags {
				if contains(tag, query) {
					queryFound = true
					break
				}
			}
			if !queryFound {
				return false
			}
		}
	}

	if len(searchTags) > 0 {
		for _, searchTag := range searchTags {
			found := false
			for _, tag := range tags {
				if tag == searchTag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsInMiddle(s, substr))))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (m *MockStorage) SetupTestData() {
	now := time.Now()

	todo1 := &models.Todo{
		ID:           "test-todo-1",
		Title:        "Test Todo 1",
		Description:  "Test description 1",
		Status:       "todo",
		Priority:     "high",
		Tags:         []string{"work", "urgent"},
		CreatedAt:    now,
		LastModified: now,
	}

	todo2 := &models.Todo{
		ID:           "test-todo-2",
		Title:        "Test Todo 2",
		Description:  "Test description 2",
		Status:       "in_progress",
		Priority:     "normal",
		Tags:         []string{"personal"},
		CreatedAt:    now,
		LastModified: now,
	}

	memo1 := &models.Memo{
		ID:           "test-memo-1",
		Title:        "Test Memo 1",
		Description:  "Test memo description 1",
		Tags:         []string{"work", "notes"},
		LinkedTodos:  []string{"test-todo-1"},
		CreatedAt:    now,
		LastModified: now,
	}

	memo2 := &models.Memo{
		ID:           "test-memo-2",
		Title:        "Test Memo 2",
		Description:  "Test memo description 2",
		Tags:         []string{"personal", "ideas"},
		LinkedTodos:  []string{},
		CreatedAt:    now,
		LastModified: now,
	}

	m.todos[todo1.ID] = todo1
	m.todos[todo2.ID] = todo2
	m.memos[memo1.ID] = memo1
	m.memos[memo2.ID] = memo2
}
