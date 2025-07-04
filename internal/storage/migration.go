package storage

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/pankona/memoya/internal/models"
	"google.golang.org/api/iterator"
)

// Migration handles data migration from legacy format to user-isolated format
type Migration struct {
	client   *firestore.Client
	targetFS *FirestoreStorage
}

// NewMigration creates a new migration instance
func NewMigration(client *firestore.Client, targetFS *FirestoreStorage) *Migration {
	return &Migration{
		client:   client,
		targetFS: targetFS,
	}
}

// MigrationStatus represents the current migration status
type MigrationStatus struct {
	IsCompleted      bool      `json:"is_completed"`
	LegacyMemosCount int       `json:"legacy_memos_count"`
	LegacyTodosCount int       `json:"legacy_todos_count"`
	MigratedMemos    int       `json:"migrated_memos"`
	MigratedTodos    int       `json:"migrated_todos"`
	DefaultUserID    string    `json:"default_user_id,omitempty"`
	StartedAt        time.Time `json:"started_at,omitempty"`
	CompletedAt      time.Time `json:"completed_at,omitempty"`
	Error            string    `json:"error,omitempty"`
}

// CheckMigrationStatus checks if migration is needed and returns current status
func (m *Migration) CheckMigrationStatus(ctx context.Context) (*MigrationStatus, error) {
	status := &MigrationStatus{}

	// Check for legacy memos (direct in /memos collection)
	legacyMemos := m.client.Collection("memos")
	memoIter := legacyMemos.Documents(ctx)
	defer memoIter.Stop()

	legacyMemoCount := 0
	for {
		_, err := memoIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to count legacy memos: %w", err)
		}
		legacyMemoCount++
	}

	// Check for legacy todos (direct in /todos collection)
	legacyTodos := m.client.Collection("todos")
	todoIter := legacyTodos.Documents(ctx)
	defer todoIter.Stop()

	legacyTodoCount := 0
	for {
		_, err := todoIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to count legacy todos: %w", err)
		}
		legacyTodoCount++
	}

	status.LegacyMemosCount = legacyMemoCount
	status.LegacyTodosCount = legacyTodoCount
	status.IsCompleted = (legacyMemoCount == 0 && legacyTodoCount == 0)

	// Check migration metadata
	migrationDoc, err := m.client.Collection("system").Doc("migration").Get(ctx)
	if err == nil {
		migrationDoc.DataTo(status)
	}

	return status, nil
}

// PerformMigration migrates all legacy data to user-isolated format
func (m *Migration) PerformMigration(ctx context.Context, defaultUserGoogleID string) (*MigrationStatus, error) {
	status := &MigrationStatus{
		StartedAt: time.Now(),
	}

	// Create or get default user
	defaultUser, err := m.getOrCreateDefaultUser(ctx, defaultUserGoogleID)
	if err != nil {
		status.Error = fmt.Sprintf("failed to create default user: %v", err)
		return status, err
	}
	status.DefaultUserID = defaultUser.ID

	// Migrate memos
	err = m.migrateLegacyMemos(ctx, defaultUser.ID, status)
	if err != nil {
		status.Error = fmt.Sprintf("failed to migrate memos: %v", err)
		return status, err
	}

	// Migrate todos
	err = m.migrateLegacyTodos(ctx, defaultUser.ID, status)
	if err != nil {
		status.Error = fmt.Sprintf("failed to migrate todos: %v", err)
		return status, err
	}

	// Mark migration as completed
	status.IsCompleted = true
	status.CompletedAt = time.Now()

	// Save migration status
	_, err = m.client.Collection("system").Doc("migration").Set(ctx, status)
	if err != nil {
		return status, fmt.Errorf("failed to save migration status: %w", err)
	}

	return status, nil
}

// getOrCreateDefaultUser creates a default user for legacy data migration
func (m *Migration) getOrCreateDefaultUser(ctx context.Context, googleID string) (*models.User, error) {
	// Try to find existing user by GoogleID
	if googleID != "" {
		user, err := m.targetFS.GetUserByGoogleID(ctx, googleID)
		if err == nil {
			return user, nil
		}
	}

	// Create new default user
	defaultUser := &models.User{
		ID:        uuid.New().String(),
		GoogleID:  googleID,
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	err := m.targetFS.CreateUser(ctx, defaultUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create default user: %w", err)
	}

	return defaultUser, nil
}

// migrateLegacyMemos migrates all memos from /memos to /users/{userID}/memos
func (m *Migration) migrateLegacyMemos(ctx context.Context, userID string, status *MigrationStatus) error {
	// Get all legacy memos
	legacyMemos := m.client.Collection("memos")
	memoIter := legacyMemos.Documents(ctx)
	defer memoIter.Stop()

	batch := m.client.Batch()
	batchCount := 0

	for {
		doc, err := memoIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate legacy memos: %w", err)
		}

		// Parse legacy memo
		var legacyMemo struct {
			ID           string    `firestore:"id"`
			Title        string    `firestore:"title"`
			Description  string    `firestore:"description"`
			Tags         []string  `firestore:"tags"`
			LinkedTodos  []string  `firestore:"linked_todos"`
			CreatedAt    time.Time `firestore:"created_at"`
			LastModified time.Time `firestore:"last_modified"`
			ClosedAt     time.Time `firestore:"closed_at"`
		}

		err = doc.DataTo(&legacyMemo)
		if err != nil {
			return fmt.Errorf("failed to parse legacy memo %s: %w", doc.Ref.ID, err)
		}

		// Create new memo with UserID
		newMemo := &models.Memo{
			ID:           legacyMemo.ID,
			UserID:       userID,
			Title:        legacyMemo.Title,
			Description:  legacyMemo.Description,
			Tags:         legacyMemo.Tags,
			LinkedTodos:  legacyMemo.LinkedTodos,
			CreatedAt:    legacyMemo.CreatedAt,
			LastModified: legacyMemo.LastModified,
		}

		if !legacyMemo.ClosedAt.IsZero() {
			newMemo.ClosedAt = &legacyMemo.ClosedAt
		}

		// Add to batch: create new location, delete old location
		newDocRef := m.client.Collection("users").Doc(userID).Collection("memos").Doc(legacyMemo.ID)
		batch.Set(newDocRef, newMemo)
		batch.Delete(doc.Ref)

		batchCount++
		status.MigratedMemos++

		// Commit batch if it gets too large
		if batchCount >= 500 {
			_, err = batch.Commit(ctx)
			if err != nil {
				return fmt.Errorf("failed to commit memo migration batch: %w", err)
			}
			batch = m.client.Batch()
			batchCount = 0
		}
	}

	// Commit remaining items
	if batchCount > 0 {
		_, err := batch.Commit(ctx)
		if err != nil {
			return fmt.Errorf("failed to commit final memo migration batch: %w", err)
		}
	}

	return nil
}

// migrateLegacyTodos migrates all todos from /todos to /users/{userID}/todos
func (m *Migration) migrateLegacyTodos(ctx context.Context, userID string, status *MigrationStatus) error {
	// Get all legacy todos
	legacyTodos := m.client.Collection("todos")
	todoIter := legacyTodos.Documents(ctx)
	defer todoIter.Stop()

	batch := m.client.Batch()
	batchCount := 0

	for {
		doc, err := todoIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate legacy todos: %w", err)
		}

		// Parse legacy todo
		var legacyTodo struct {
			ID           string    `firestore:"id"`
			Title        string    `firestore:"title"`
			Description  string    `firestore:"description"`
			Status       string    `firestore:"status"`
			Priority     string    `firestore:"priority"`
			Tags         []string  `firestore:"tags"`
			ParentID     string    `firestore:"parent_id"`
			CreatedAt    time.Time `firestore:"created_at"`
			LastModified time.Time `firestore:"last_modified"`
			ClosedAt     time.Time `firestore:"closed_at"`
		}

		err = doc.DataTo(&legacyTodo)
		if err != nil {
			return fmt.Errorf("failed to parse legacy todo %s: %w", doc.Ref.ID, err)
		}

		// Create new todo with UserID
		newTodo := &models.Todo{
			ID:           legacyTodo.ID,
			UserID:       userID,
			Title:        legacyTodo.Title,
			Description:  legacyTodo.Description,
			Status:       models.TodoStatus(legacyTodo.Status),
			Priority:     models.TodoPriority(legacyTodo.Priority),
			Tags:         legacyTodo.Tags,
			ParentID:     legacyTodo.ParentID,
			CreatedAt:    legacyTodo.CreatedAt,
			LastModified: legacyTodo.LastModified,
		}

		if !legacyTodo.ClosedAt.IsZero() {
			newTodo.ClosedAt = &legacyTodo.ClosedAt
		}

		// Add to batch: create new location, delete old location
		newDocRef := m.client.Collection("users").Doc(userID).Collection("todos").Doc(legacyTodo.ID)
		batch.Set(newDocRef, newTodo)
		batch.Delete(doc.Ref)

		batchCount++
		status.MigratedTodos++

		// Commit batch if it gets too large
		if batchCount >= 500 {
			_, err = batch.Commit(ctx)
			if err != nil {
				return fmt.Errorf("failed to commit todo migration batch: %w", err)
			}
			batch = m.client.Batch()
			batchCount = 0
		}
	}

	// Commit remaining items
	if batchCount > 0 {
		_, err := batch.Commit(ctx)
		if err != nil {
			return fmt.Errorf("failed to commit final todo migration batch: %w", err)
		}
	}

	return nil
}

// CleanupLegacyCollections removes empty legacy collections after successful migration
func (m *Migration) CleanupLegacyCollections(ctx context.Context) error {
	// Check if migration is completed
	status, err := m.CheckMigrationStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if !status.IsCompleted {
		return fmt.Errorf("migration not completed, cannot cleanup legacy collections")
	}

	if status.LegacyMemosCount > 0 || status.LegacyTodosCount > 0 {
		return fmt.Errorf("legacy data still exists, cannot cleanup")
	}

	// Legacy collections should already be empty due to migration process
	// This is just a safety check and log
	return nil
}
