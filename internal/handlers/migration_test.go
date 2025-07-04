package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pankona/memoya/internal/storage"
)

// MockMigration implements a mock migration service for testing
type MockMigration struct {
	status        *storage.MigrationStatus
	shouldSucceed bool
}

func NewMockMigration(shouldSucceed bool) *MockMigration {
	return &MockMigration{
		status: &storage.MigrationStatus{
			IsCompleted:      false,
			LegacyMemosCount: 5,
			LegacyTodosCount: 3,
			MigratedMemos:    0,
			MigratedTodos:    0,
		},
		shouldSucceed: shouldSucceed,
	}
}

func (m *MockMigration) CheckMigrationStatus(ctx context.Context) (*storage.MigrationStatus, error) {
	if !m.shouldSucceed {
		return nil, fmt.Errorf("mock migration status check failed")
	}
	return m.status, nil
}

func (m *MockMigration) PerformMigration(ctx context.Context, defaultUserGoogleID string) (*storage.MigrationStatus, error) {
	if !m.shouldSucceed {
		return nil, fmt.Errorf("mock migration failed")
	}

	// Simulate successful migration
	m.status.IsCompleted = true
	m.status.MigratedMemos = m.status.LegacyMemosCount
	m.status.MigratedTodos = m.status.LegacyTodosCount
	m.status.LegacyMemosCount = 0
	m.status.LegacyTodosCount = 0
	m.status.DefaultUserID = "migrated-user-123"

	return m.status, nil
}

func (m *MockMigration) CleanupLegacyCollections(ctx context.Context) error {
	if !m.shouldSucceed {
		return fmt.Errorf("mock cleanup failed")
	}
	return nil
}

func TestMigrationHandler_CheckMigrationStatus(t *testing.T) {
	mockMigration := NewMockMigration(true)
	handler := NewMigrationHandler(mockMigration)

	args := MigrationStatusArgs{}
	params := &mcp.CallToolParamsFor[MigrationStatusArgs]{
		Arguments: args,
	}

	result, err := handler.CheckMigrationStatus(context.Background(), nil, params)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if len(result.Content) == 0 {
		t.Fatal("Expected content, got empty")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}

	var statusResult MigrationStatusResult
	err = json.Unmarshal([]byte(textContent.Text), &statusResult)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if !statusResult.Success {
		t.Error("Expected success to be true")
	}

	if statusResult.Status == nil {
		t.Fatal("Expected status to be present")
	}

	if statusResult.Status.LegacyMemosCount != 5 {
		t.Errorf("Expected 5 legacy memos, got %d", statusResult.Status.LegacyMemosCount)
	}

	if statusResult.Status.LegacyTodosCount != 3 {
		t.Errorf("Expected 3 legacy todos, got %d", statusResult.Status.LegacyTodosCount)
	}

	if statusResult.Message == "" {
		t.Error("Expected non-empty message")
	}
}

func TestMigrationHandler_CheckMigrationStatus_Error(t *testing.T) {
	mockMigration := NewMockMigration(false)
	handler := NewMigrationHandler(mockMigration)

	args := MigrationStatusArgs{}
	params := &mcp.CallToolParamsFor[MigrationStatusArgs]{
		Arguments: args,
	}

	result, err := handler.CheckMigrationStatus(context.Background(), nil, params)

	if err == nil {
		t.Fatal("Expected error when migration service fails")
	}

	if result != nil {
		t.Error("Expected nil result when error occurs")
	}
}

func TestMigrationHandler_PerformMigration(t *testing.T) {
	mockMigration := NewMockMigration(true)
	handler := NewMigrationHandler(mockMigration)

	args := PerformMigrationArgs{
		DefaultUserGoogleID: "google-123",
		Confirm:             true,
	}
	params := &mcp.CallToolParamsFor[PerformMigrationArgs]{
		Arguments: args,
	}

	result, err := handler.PerformMigration(context.Background(), nil, params)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}

	var migrationResult PerformMigrationResult
	err = json.Unmarshal([]byte(textContent.Text), &migrationResult)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if !migrationResult.Success {
		t.Error("Expected success to be true")
	}

	if migrationResult.Status == nil {
		t.Fatal("Expected status to be present")
	}

	if !migrationResult.Status.IsCompleted {
		t.Error("Expected migration to be completed")
	}

	if migrationResult.Status.MigratedMemos != 5 {
		t.Errorf("Expected 5 migrated memos, got %d", migrationResult.Status.MigratedMemos)
	}

	if migrationResult.Status.MigratedTodos != 3 {
		t.Errorf("Expected 3 migrated todos, got %d", migrationResult.Status.MigratedTodos)
	}

	if migrationResult.Status.DefaultUserID != "migrated-user-123" {
		t.Errorf("Expected default user ID 'migrated-user-123', got %s", migrationResult.Status.DefaultUserID)
	}
}

func TestMigrationHandler_PerformMigration_WithoutConfirm(t *testing.T) {
	mockMigration := NewMockMigration(true)
	handler := NewMigrationHandler(mockMigration)

	args := PerformMigrationArgs{
		DefaultUserGoogleID: "google-123",
		Confirm:             false, // No confirmation
	}
	params := &mcp.CallToolParamsFor[PerformMigrationArgs]{
		Arguments: args,
	}

	result, err := handler.PerformMigration(context.Background(), nil, params)

	if err == nil {
		t.Fatal("Expected error when confirmation is missing")
	}

	if result != nil {
		t.Error("Expected nil result when confirmation is missing")
	}
}

func TestMigrationHandler_CleanupMigration(t *testing.T) {
	mockMigration := NewMockMigration(true)
	handler := NewMigrationHandler(mockMigration)

	args := CleanupMigrationArgs{
		Confirm: true,
	}
	params := &mcp.CallToolParamsFor[CleanupMigrationArgs]{
		Arguments: args,
	}

	result, err := handler.CleanupMigration(context.Background(), nil, params)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}

	var cleanupResult CleanupMigrationResult
	err = json.Unmarshal([]byte(textContent.Text), &cleanupResult)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if !cleanupResult.Success {
		t.Error("Expected success to be true")
	}

	if cleanupResult.Message == "" {
		t.Error("Expected non-empty message")
	}
}

func TestMigrationHandler_CleanupMigration_WithoutConfirm(t *testing.T) {
	mockMigration := NewMockMigration(true)
	handler := NewMigrationHandler(mockMigration)

	args := CleanupMigrationArgs{
		Confirm: false, // No confirmation
	}
	params := &mcp.CallToolParamsFor[CleanupMigrationArgs]{
		Arguments: args,
	}

	result, err := handler.CleanupMigration(context.Background(), nil, params)

	if err == nil {
		t.Fatal("Expected error when confirmation is missing")
	}

	if result != nil {
		t.Error("Expected nil result when confirmation is missing")
	}
}

func TestMigrationHandler_WithoutMigrationService(t *testing.T) {
	handler := NewMigrationHandler(nil)

	args := MigrationStatusArgs{}
	params := &mcp.CallToolParamsFor[MigrationStatusArgs]{
		Arguments: args,
	}

	result, err := handler.CheckMigrationStatus(context.Background(), nil, params)

	if err == nil {
		t.Fatal("Expected error when migration service is not initialized")
	}

	if result != nil {
		t.Error("Expected nil result when migration service is not initialized")
	}
}
