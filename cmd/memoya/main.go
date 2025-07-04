package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pankona/memoya/config"
	"github.com/pankona/memoya/internal/handlers"
	"github.com/pankona/memoya/internal/storage"
)

func main() {
	ctx := context.Background()

	// Load .env file if exists
	_ = godotenv.Load()

	// Load configuration
	cfg := config.Load()

	// Initialize storage
	var store storage.Storage
	if cfg.FirebaseProjectID != "" {
		firestoreStorage, err := storage.NewFirestoreStorage(ctx, cfg.FirebaseProjectID)
		if err != nil {
			log.Fatalf("Failed to initialize Firestore: %v", err)
		}
		defer firestoreStorage.Close()
		store = firestoreStorage
	} else {
		log.Println("Warning: No FIREBASE_PROJECT_ID provided, using mock storage")
		// For now, we'll still use handlers without storage
		// TODO: Implement in-memory storage for development
	}

	// Create MCP server
	server := mcp.NewServer("memoya", "0.1.0", nil)

	// Create handlers
	todoHandler := handlers.NewTodoHandlerWithStorage(store)
	memoHandler := handlers.NewMemoHandlerWithStorage(store)
	searchHandler := handlers.NewSearchHandler(store)

	// Register todo tools
	server.AddTools(
		mcp.NewServerTool(
			"todo_create",
			"Create a new todo item",
			todoHandler.Create,
			mcp.Input(
				mcp.Property("title", mcp.Description("Todo title"), mcp.Required(true)),
				mcp.Property("description", mcp.Description("Todo description")),
				mcp.Property("status", mcp.Description("Todo status (backlog, todo, in_progress, done)")),
				mcp.Property("priority", mcp.Description("Todo priority (high, normal)")),
				mcp.Property("tags", mcp.Description("Tags for the todo")),
				mcp.Property("parent_id", mcp.Description("Parent todo ID for hierarchical structure")),
			),
		),
		mcp.NewServerTool(
			"todo_list",
			"List todo items with optional filters",
			todoHandler.List,
			mcp.Input(
				mcp.Property("status", mcp.Description("Filter by status")),
				mcp.Property("tags", mcp.Description("Filter by tags")),
				mcp.Property("priority", mcp.Description("Filter by priority")),
			),
		),
		mcp.NewServerTool(
			"todo_update",
			"Update an existing todo item",
			todoHandler.Update,
			mcp.Input(
				mcp.Property("id", mcp.Description("Todo ID to update"), mcp.Required(true)),
				mcp.Property("title", mcp.Description("New title")),
				mcp.Property("description", mcp.Description("New description")),
				mcp.Property("status", mcp.Description("New status")),
				mcp.Property("priority", mcp.Description("New priority")),
				mcp.Property("tags", mcp.Description("New tags")),
			),
		),
		mcp.NewServerTool(
			"todo_delete",
			"Delete a todo item",
			todoHandler.Delete,
			mcp.Input(
				mcp.Property("id", mcp.Description("Todo ID to delete"), mcp.Required(true)),
			),
		),
	)

	// Register memo tools
	server.AddTools(
		mcp.NewServerTool(
			"memo_create",
			"Create a new memo",
			memoHandler.Create,
			mcp.Input(
				mcp.Property("title", mcp.Description("Memo title"), mcp.Required(true)),
				mcp.Property("description", mcp.Description("Memo description")),
				mcp.Property("tags", mcp.Description("Tags for the memo")),
				mcp.Property("linked_todos", mcp.Description("IDs of linked todos")),
			),
		),
		mcp.NewServerTool(
			"memo_list",
			"List memos with optional filters",
			memoHandler.List,
			mcp.Input(
				mcp.Property("tags", mcp.Description("Filter by tags")),
			),
		),
		mcp.NewServerTool(
			"memo_update",
			"Update an existing memo",
			memoHandler.Update,
			mcp.Input(
				mcp.Property("id", mcp.Description("Memo ID to update"), mcp.Required(true)),
				mcp.Property("title", mcp.Description("New title")),
				mcp.Property("description", mcp.Description("New description")),
				mcp.Property("tags", mcp.Description("New tags")),
				mcp.Property("linked_todos", mcp.Description("New linked todo IDs")),
			),
		),
		mcp.NewServerTool(
			"memo_delete",
			"Delete a memo",
			memoHandler.Delete,
			mcp.Input(
				mcp.Property("id", mcp.Description("Memo ID to delete"), mcp.Required(true)),
			),
		),
	)

	// Register search tool
	server.AddTools(
		mcp.NewServerTool(
			"search",
			"Search todos and memos by keyword or tags",
			searchHandler.Search,
			mcp.Input(
				mcp.Property("query", mcp.Description("Search query")),
				mcp.Property("tags", mcp.Description("Filter by tags")),
				mcp.Property("type", mcp.Description("Filter by type (todo, memo, all)")),
			),
		),
	)

	// Run server with stdio transport
	transport := mcp.NewStdioTransport()
	if err := server.Run(ctx, transport); err != nil {
		log.Fatal(err)
	}
}
