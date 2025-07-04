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

// User operations
func (fs *FirestoreStorage) CreateUser(ctx context.Context, user *models.User) error {
	_, err := fs.client.Collection("users").Doc(user.ID).Set(ctx, user)
	return err
}

func (fs *FirestoreStorage) GetUser(ctx context.Context, id string) (*models.User, error) {
	doc, err := fs.client.Collection("users").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (fs *FirestoreStorage) GetUserByGoogleID(ctx context.Context, googleID string) (*models.User, error) {
	iter := fs.client.Collection("users").Where("google_id", "==", googleID).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (fs *FirestoreStorage) UpdateUser(ctx context.Context, user *models.User) error {
	_, err := fs.client.Collection("users").Doc(user.ID).Set(ctx, user)
	return err
}

func (fs *FirestoreStorage) DeleteUser(ctx context.Context, id string) error {
	// Delete all user data including memos and todos
	batch := fs.client.Batch()

	// Delete user document
	userDoc := fs.client.Collection("users").Doc(id)
	batch.Delete(userDoc)

	// Delete user's memos
	memoIter := fs.client.Collection("users").Doc(id).Collection("memos").Documents(ctx)
	for {
		doc, err := memoIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		batch.Delete(doc.Ref)
	}
	memoIter.Stop()

	// Delete user's todos
	todoIter := fs.client.Collection("users").Doc(id).Collection("todos").Documents(ctx)
	for {
		doc, err := todoIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		batch.Delete(doc.Ref)
	}
	todoIter.Stop()

	// Commit batch
	_, err := batch.Commit(ctx)
	return err
}

// Device auth operations
func (fs *FirestoreStorage) CreateDeviceAuthSession(ctx context.Context, session *models.DeviceAuthSession) error {
	_, err := fs.client.Collection("device_auth_sessions").Doc(session.DeviceCode).Set(ctx, session)
	return err
}

func (fs *FirestoreStorage) GetDeviceAuthSession(ctx context.Context, deviceCode string) (*models.DeviceAuthSession, error) {
	doc, err := fs.client.Collection("device_auth_sessions").Doc(deviceCode).Get(ctx)
	if err != nil {
		return nil, err
	}

	var session models.DeviceAuthSession
	if err := doc.DataTo(&session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (fs *FirestoreStorage) UpdateDeviceAuthSession(ctx context.Context, session *models.DeviceAuthSession) error {
	_, err := fs.client.Collection("device_auth_sessions").Doc(session.DeviceCode).Set(ctx, session)
	return err
}

func (fs *FirestoreStorage) DeleteDeviceAuthSession(ctx context.Context, deviceCode string) error {
	_, err := fs.client.Collection("device_auth_sessions").Doc(deviceCode).Delete(ctx)
	return err
}

// Todo operations (updated for user isolation)
func (fs *FirestoreStorage) CreateTodo(ctx context.Context, todo *models.Todo) error {
	_, err := fs.client.Collection("users").Doc(todo.UserID).Collection("todos").Doc(todo.ID).Set(ctx, todo)
	return err
}

func (fs *FirestoreStorage) GetTodo(ctx context.Context, id string) (*models.Todo, error) {
	// Note: We need to search across all users since we only have the todo ID
	// In practice, this should be called with user context to avoid this
	iter := fs.client.CollectionGroup("todos").Where("id", "==", id).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
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
	_, err := fs.client.Collection("users").Doc(todo.UserID).Collection("todos").Doc(todo.ID).Set(ctx, todo)
	return err
}

func (fs *FirestoreStorage) DeleteTodo(ctx context.Context, id string) error {
	// Find the todo first to get the userID
	todo, err := fs.GetTodo(ctx, id)
	if err != nil {
		return err
	}

	_, err = fs.client.Collection("users").Doc(todo.UserID).Collection("todos").Doc(id).Delete(ctx)
	return err
}

func (fs *FirestoreStorage) ListTodos(ctx context.Context, filters TodoFilters) ([]*models.Todo, error) {
	// User isolation: query within user's todos collection
	query := fs.client.Collection("users").Doc(filters.UserID).Collection("todos").Query

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

// Memo operations (updated for user isolation)
func (fs *FirestoreStorage) CreateMemo(ctx context.Context, memo *models.Memo) error {
	_, err := fs.client.Collection("users").Doc(memo.UserID).Collection("memos").Doc(memo.ID).Set(ctx, memo)
	return err
}

func (fs *FirestoreStorage) GetMemo(ctx context.Context, id string) (*models.Memo, error) {
	// Note: We need to search across all users since we only have the memo ID
	// In practice, this should be called with user context to avoid this
	iter := fs.client.CollectionGroup("memos").Where("id", "==", id).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
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
	_, err := fs.client.Collection("users").Doc(memo.UserID).Collection("memos").Doc(memo.ID).Set(ctx, memo)
	return err
}

func (fs *FirestoreStorage) DeleteMemo(ctx context.Context, id string) error {
	// Find the memo first to get the userID
	memo, err := fs.GetMemo(ctx, id)
	if err != nil {
		return err
	}

	_, err = fs.client.Collection("users").Doc(memo.UserID).Collection("memos").Doc(id).Delete(ctx)
	return err
}

func (fs *FirestoreStorage) ListMemos(ctx context.Context, filters MemoFilters) ([]*models.Memo, error) {
	// User isolation: query within user's memos collection
	query := fs.client.Collection("users").Doc(filters.UserID).Collection("memos").Query

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

// Search operations (updated for user isolation)
func (fs *FirestoreStorage) Search(ctx context.Context, query string, filters SearchFilters) (*SearchResults, error) {
	results := &SearchResults{
		Todos: []*models.Todo{},
		Memos: []*models.Memo{},
	}

	// Search todos if needed
	if filters.Type == "todo" || filters.Type == "all" {
		todos, err := fs.searchTodos(ctx, query, filters.UserID, filters.Tags)
		if err != nil {
			return nil, err
		}
		results.Todos = todos
	}

	// Search memos if needed
	if filters.Type == "memo" || filters.Type == "all" {
		memos, err := fs.searchMemos(ctx, query, filters.UserID, filters.Tags)
		if err != nil {
			return nil, err
		}
		results.Memos = memos
	}

	return results, nil
}

func (fs *FirestoreStorage) searchTodos(ctx context.Context, query string, userID string, tags []string) ([]*models.Todo, error) {
	// User isolation: search within user's todos collection only
	iter := fs.client.Collection("users").Doc(userID).Collection("todos").Documents(ctx)
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

func (fs *FirestoreStorage) searchMemos(ctx context.Context, query string, userID string, tags []string) ([]*models.Memo, error) {
	// User isolation: search within user's memos collection only
	iter := fs.client.Collection("users").Doc(userID).Collection("memos").Documents(ctx)
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

// GetAllTags retrieves all unique tags from both todos and memos for a specific user
func (fs *FirestoreStorage) GetAllTags(ctx context.Context, userID string) ([]string, error) {
	tagSet := make(map[string]bool)

	// Get tags from user's todos
	todoIter := fs.client.Collection("users").Doc(userID).Collection("todos").Documents(ctx)
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

	// Get tags from user's memos
	memoIter := fs.client.Collection("users").Doc(userID).Collection("memos").Documents(ctx)
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
