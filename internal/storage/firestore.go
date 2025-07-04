package storage

import (
	"context"
	"strings"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/pankona/memoya/internal/models"
	"google.golang.org/api/iterator"
)

// FirestoreStorage implements the Storage interface using Firestore
type FirestoreStorage struct {
	client *firestore.Client
}

// NewFirestoreStorage creates a new Firestore storage instance
func NewFirestoreStorage(ctx context.Context, projectID string) (*FirestoreStorage, error) {
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: projectID,
	})
	if err != nil {
		return nil, err
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, err
	}

	return &FirestoreStorage{
		client: client,
	}, nil
}

// Close closes the Firestore client
func (fs *FirestoreStorage) Close() error {
	return fs.client.Close()
}

// Todo operations
func (fs *FirestoreStorage) CreateTodo(ctx context.Context, todo *models.Todo) error {
	_, err := fs.client.Collection("todos").Doc(todo.ID).Set(ctx, todo)
	return err
}

func (fs *FirestoreStorage) GetTodo(ctx context.Context, id string) (*models.Todo, error) {
	doc, err := fs.client.Collection("todos").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}

	var todo models.Todo
	if err := doc.DataTo(&todo); err != nil {
		return nil, err
	}

	return &todo, nil
}

func (fs *FirestoreStorage) UpdateTodo(ctx context.Context, todo *models.Todo) error {
	_, err := fs.client.Collection("todos").Doc(todo.ID).Set(ctx, todo)
	return err
}

func (fs *FirestoreStorage) DeleteTodo(ctx context.Context, id string) error {
	_, err := fs.client.Collection("todos").Doc(id).Delete(ctx)
	return err
}

func (fs *FirestoreStorage) ListTodos(ctx context.Context, filters TodoFilters) ([]*models.Todo, error) {
	query := fs.client.Collection("todos").Query

	// Apply filters
	if filters.Status != nil {
		query = query.Where("status", "==", string(*filters.Status))
	}

	if filters.Priority != nil {
		query = query.Where("priority", "==", string(*filters.Priority))
	}

	if filters.ParentID != nil {
		query = query.Where("parent_id", "==", *filters.ParentID)
	}

	// Note: Firestore doesn't support array-contains-any with other filters
	// For tags filtering, we'll need to do it in-memory for now
	iter := query.Documents(ctx)
	defer iter.Stop()

	var todos []*models.Todo
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var todo models.Todo
		if err := doc.DataTo(&todo); err != nil {
			return nil, err
		}

		// Apply tag filtering in-memory
		if len(filters.Tags) > 0 {
			hasMatchingTag := false
			for _, filterTag := range filters.Tags {
				for _, todoTag := range todo.Tags {
					if todoTag == filterTag {
						hasMatchingTag = true
						break
					}
				}
				if hasMatchingTag {
					break
				}
			}
			if !hasMatchingTag {
				continue
			}
		}

		todos = append(todos, &todo)
	}

	return todos, nil
}

// Memo operations
func (fs *FirestoreStorage) CreateMemo(ctx context.Context, memo *models.Memo) error {
	_, err := fs.client.Collection("memos").Doc(memo.ID).Set(ctx, memo)
	return err
}

func (fs *FirestoreStorage) GetMemo(ctx context.Context, id string) (*models.Memo, error) {
	doc, err := fs.client.Collection("memos").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}

	var memo models.Memo
	if err := doc.DataTo(&memo); err != nil {
		return nil, err
	}

	return &memo, nil
}

func (fs *FirestoreStorage) UpdateMemo(ctx context.Context, memo *models.Memo) error {
	_, err := fs.client.Collection("memos").Doc(memo.ID).Set(ctx, memo)
	return err
}

func (fs *FirestoreStorage) DeleteMemo(ctx context.Context, id string) error {
	_, err := fs.client.Collection("memos").Doc(id).Delete(ctx)
	return err
}

func (fs *FirestoreStorage) ListMemos(ctx context.Context, filters MemoFilters) ([]*models.Memo, error) {
	query := fs.client.Collection("memos").Query

	iter := query.Documents(ctx)
	defer iter.Stop()

	var memos []*models.Memo
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var memo models.Memo
		if err := doc.DataTo(&memo); err != nil {
			return nil, err
		}

		// Apply tag filtering in-memory
		if len(filters.Tags) > 0 {
			hasMatchingTag := false
			for _, filterTag := range filters.Tags {
				for _, memoTag := range memo.Tags {
					if memoTag == filterTag {
						hasMatchingTag = true
						break
					}
				}
				if hasMatchingTag {
					break
				}
			}
			if !hasMatchingTag {
				continue
			}
		}

		memos = append(memos, &memo)
	}

	return memos, nil
}

// Search operations
func (fs *FirestoreStorage) Search(ctx context.Context, query string, filters SearchFilters) (*SearchResults, error) {
	results := &SearchResults{
		Todos: []*models.Todo{},
		Memos: []*models.Memo{},
	}

	// Search todos if needed
	if filters.Type == "todo" || filters.Type == "all" {
		todos, err := fs.searchTodos(ctx, query, filters.Tags)
		if err != nil {
			return nil, err
		}
		results.Todos = todos
	}

	// Search memos if needed
	if filters.Type == "memo" || filters.Type == "all" {
		memos, err := fs.searchMemos(ctx, query, filters.Tags)
		if err != nil {
			return nil, err
		}
		results.Memos = memos
	}

	return results, nil
}

func (fs *FirestoreStorage) searchTodos(ctx context.Context, query string, tags []string) ([]*models.Todo, error) {
	iter := fs.client.Collection("todos").Documents(ctx)
	defer iter.Stop()

	var todos []*models.Todo
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var todo models.Todo
		if err := doc.DataTo(&todo); err != nil {
			return nil, err
		}

		// Simple text search in title and description
		if query != "" {
			lowerQuery := strings.ToLower(query)
			if !strings.Contains(strings.ToLower(todo.Title), lowerQuery) &&
				!strings.Contains(strings.ToLower(todo.Description), lowerQuery) {
				continue
			}
		}

		// Apply tag filtering
		if len(tags) > 0 {
			hasMatchingTag := false
			for _, filterTag := range tags {
				for _, todoTag := range todo.Tags {
					if todoTag == filterTag {
						hasMatchingTag = true
						break
					}
				}
				if hasMatchingTag {
					break
				}
			}
			if !hasMatchingTag {
				continue
			}
		}

		todos = append(todos, &todo)
	}

	return todos, nil
}

func (fs *FirestoreStorage) searchMemos(ctx context.Context, query string, tags []string) ([]*models.Memo, error) {
	iter := fs.client.Collection("memos").Documents(ctx)
	defer iter.Stop()

	var memos []*models.Memo
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var memo models.Memo
		if err := doc.DataTo(&memo); err != nil {
			return nil, err
		}

		// Simple text search in title and description
		if query != "" {
			lowerQuery := strings.ToLower(query)
			if !strings.Contains(strings.ToLower(memo.Title), lowerQuery) &&
				!strings.Contains(strings.ToLower(memo.Description), lowerQuery) {
				continue
			}
		}

		// Apply tag filtering
		if len(tags) > 0 {
			hasMatchingTag := false
			for _, filterTag := range tags {
				for _, memoTag := range memo.Tags {
					if memoTag == filterTag {
						hasMatchingTag = true
						break
					}
				}
				if hasMatchingTag {
					break
				}
			}
			if !hasMatchingTag {
				continue
			}
		}

		memos = append(memos, &memo)
	}

	return memos, nil
}

// GetAllTags retrieves all unique tags from both todos and memos
func (fs *FirestoreStorage) GetAllTags(ctx context.Context) ([]string, error) {
	tagSet := make(map[string]bool)

	// Get tags from todos
	todoIter := fs.client.Collection("todos").Documents(ctx)
	for {
		doc, err := todoIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var todo models.Todo
		if err := doc.DataTo(&todo); err != nil {
			continue // Skip documents that can't be parsed
		}

		for _, tag := range todo.Tags {
			if tag != "" {
				tagSet[tag] = true
			}
		}
	}
	todoIter.Stop()

	// Get tags from memos
	memoIter := fs.client.Collection("memos").Documents(ctx)
	for {
		doc, err := memoIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var memo models.Memo
		if err := doc.DataTo(&memo); err != nil {
			continue // Skip documents that can't be parsed
		}

		for _, tag := range memo.Tags {
			if tag != "" {
				tagSet[tag] = true
			}
		}
	}
	memoIter.Stop()

	// Convert map to sorted slice
	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	return tags, nil
}
