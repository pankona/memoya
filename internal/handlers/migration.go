package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pankona/memoya/internal/storage"
)

// MigrationService defines the interface for migration operations
type MigrationService interface {
	CheckMigrationStatus(ctx context.Context) (*storage.MigrationStatus, error)
	PerformMigration(ctx context.Context, defaultUserGoogleID string) (*storage.MigrationStatus, error)
	CleanupLegacyCollections(ctx context.Context) error
}

type MigrationHandler struct {
	migration MigrationService
}

func NewMigrationHandler(migration MigrationService) *MigrationHandler {
	return &MigrationHandler{
		migration: migration,
	}
}

// MigrationStatusArgs represents arguments for checking migration status
type MigrationStatusArgs struct {
	// No arguments needed
}

// MigrationStatusResult represents the result of migration status check
type MigrationStatusResult struct {
	Success bool                     `json:"success"`
	Status  *storage.MigrationStatus `json:"status"`
	Message string                   `json:"message"`
}

// CheckMigrationStatus checks the current migration status
func (h *MigrationHandler) CheckMigrationStatus(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[MigrationStatusArgs]) (*mcp.CallToolResultFor[MigrationStatusResult], error) {
	if h.migration == nil {
		return nil, fmt.Errorf("migration service not initialized")
	}

	status, err := h.migration.CheckMigrationStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check migration status: %w", err)
	}

	var message string
	if status.IsCompleted {
		message = "Migration completed successfully"
	} else if status.LegacyMemosCount > 0 || status.LegacyTodosCount > 0 {
		message = fmt.Sprintf("Migration needed: %d legacy memos, %d legacy todos found",
			status.LegacyMemosCount, status.LegacyTodosCount)
	} else {
		message = "No migration needed"
	}

	result := MigrationStatusResult{
		Success: true,
		Status:  status,
		Message: message,
	}

	// Convert to JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[MigrationStatusResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}

// PerformMigrationArgs represents arguments for performing migration
type PerformMigrationArgs struct {
	DefaultUserGoogleID string `json:"default_user_google_id,omitempty"`
	Confirm             bool   `json:"confirm"`
}

// PerformMigrationResult represents the result of migration
type PerformMigrationResult struct {
	Success bool                     `json:"success"`
	Status  *storage.MigrationStatus `json:"status"`
	Message string                   `json:"message"`
}

// PerformMigration performs the data migration from legacy format to user-isolated format
func (h *MigrationHandler) PerformMigration(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[PerformMigrationArgs]) (*mcp.CallToolResultFor[PerformMigrationResult], error) {
	args := params.Arguments

	if h.migration == nil {
		return nil, fmt.Errorf("migration service not initialized")
	}

	// Require explicit confirmation
	if !args.Confirm {
		return nil, fmt.Errorf("migration requires explicit confirmation")
	}

	// Check if migration is needed
	currentStatus, err := h.migration.CheckMigrationStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check migration status: %w", err)
	}

	if currentStatus.IsCompleted {
		result := PerformMigrationResult{
			Success: true,
			Status:  currentStatus,
			Message: "Migration already completed",
		}

		jsonBytes, _ := json.Marshal(result)
		return &mcp.CallToolResultFor[PerformMigrationResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(jsonBytes)},
			},
		}, nil
	}

	if currentStatus.LegacyMemosCount == 0 && currentStatus.LegacyTodosCount == 0 {
		result := PerformMigrationResult{
			Success: true,
			Status:  currentStatus,
			Message: "No legacy data to migrate",
		}

		jsonBytes, _ := json.Marshal(result)
		return &mcp.CallToolResultFor[PerformMigrationResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(jsonBytes)},
			},
		}, nil
	}

	// Perform migration
	status, err := h.migration.PerformMigration(ctx, args.DefaultUserGoogleID)
	if err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	message := fmt.Sprintf("Migration completed successfully: migrated %d memos and %d todos to user %s",
		status.MigratedMemos, status.MigratedTodos, status.DefaultUserID)

	result := PerformMigrationResult{
		Success: true,
		Status:  status,
		Message: message,
	}

	// Convert to JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[PerformMigrationResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}

// CleanupMigrationArgs represents arguments for cleaning up after migration
type CleanupMigrationArgs struct {
	Confirm bool `json:"confirm"`
}

// CleanupMigrationResult represents the result of migration cleanup
type CleanupMigrationResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// CleanupMigration cleans up legacy collections after successful migration
func (h *MigrationHandler) CleanupMigration(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[CleanupMigrationArgs]) (*mcp.CallToolResultFor[CleanupMigrationResult], error) {
	args := params.Arguments

	if h.migration == nil {
		return nil, fmt.Errorf("migration service not initialized")
	}

	// Require explicit confirmation
	if !args.Confirm {
		return nil, fmt.Errorf("cleanup requires explicit confirmation")
	}

	err := h.migration.CleanupLegacyCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("cleanup failed: %w", err)
	}

	result := CleanupMigrationResult{
		Success: true,
		Message: "Legacy collections cleanup completed successfully",
	}

	// Convert to JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[CleanupMigrationResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}
