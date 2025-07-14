package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pankona/memoya/internal/client"
	"github.com/pankona/memoya/internal/handlers"
)

func main() {
	ctx := context.Background()

	// Load .env file if exists
	_ = godotenv.Load()

	// Get Cloud Run URL from environment
	cloudRunURL := os.Getenv("MEMOYA_CLOUD_RUN_URL")
	if cloudRunURL == "" {
		// Default to production Cloud Run URL
		cloudRunURL = "https://memoya-server-152455917187.asia-northeast1.run.app"
		log.Printf("Using default Cloud Run URL: %s", cloudRunURL)
	} else {
		log.Printf("Using Cloud Run URL from environment: %s", cloudRunURL)
	}

	// Initialize HTTP client
	httpClient := client.NewHTTPClient(cloudRunURL)

	// Set auth token from environment or saved config
	authToken := os.Getenv("MEMOYA_AUTH_TOKEN")
	if authToken == "" {
		// Try to load from saved config
		if savedToken, err := handlers.GetAuthToken(); err == nil && savedToken != "" {
			authToken = savedToken
			log.Println("Using saved auth token")
		}
	} else {
		log.Println("Using auth token from environment")
	}

	if authToken != "" {
		httpClient.SetAuthToken(authToken)
		log.Println("Auth token configured")
	}

	// Test connectivity
	if err := httpClient.Ping(ctx); err != nil {
		log.Printf("Warning: Failed to ping server at %s: %v", cloudRunURL, err)
		log.Println("Continuing anyway - server might not be running yet")
	} else {
		log.Println("Successfully connected to Cloud Run server")
	}

	// Create MCP bridge
	bridge := client.NewMCPBridge(httpClient)

	// Create MCP server
	server := mcp.NewServer("memoya", "0.1.0", nil)

	// Create auth handler using HTTP client
	authHandler := handlers.NewAuthHandler(httpClient)

	// Register memo tools (HTTP-backed)
	server.AddTools(
		mcp.NewServerTool(
			"memo_create",
			"Create a new memo",
			bridge.MemoCreate,
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
			bridge.MemoList,
			mcp.Input(
				mcp.Property("tags", mcp.Description("Filter by tags")),
			),
		),
		mcp.NewServerTool(
			"memo_update",
			"Update an existing memo",
			bridge.MemoUpdate,
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
			bridge.MemoDelete,
			mcp.Input(
				mcp.Property("id", mcp.Description("Memo ID to delete"), mcp.Required(true)),
			),
		),
	)

	// Register todo tools (HTTP-backed)
	server.AddTools(
		mcp.NewServerTool(
			"todo_create",
			"Create a new todo item",
			bridge.TodoCreate,
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
			bridge.TodoList,
			mcp.Input(
				mcp.Property("status", mcp.Description("Filter by status")),
				mcp.Property("tags", mcp.Description("Filter by tags")),
				mcp.Property("priority", mcp.Description("Filter by priority")),
			),
		),
		mcp.NewServerTool(
			"todo_update",
			"Update an existing todo item",
			bridge.TodoUpdate,
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
			bridge.TodoDelete,
			mcp.Input(
				mcp.Property("id", mcp.Description("Todo ID to delete"), mcp.Required(true)),
			),
		),
	)

	// Register search tool (HTTP-backed)
	server.AddTools(
		mcp.NewServerTool(
			"search",
			"Search todos and memos by keyword or tags",
			bridge.Search,
			mcp.Input(
				mcp.Property("query", mcp.Description("Search query")),
				mcp.Property("tags", mcp.Description("Filter by tags")),
				mcp.Property("type", mcp.Description("Filter by type (todo, memo, all)")),
			),
		),
	)

	// Register tag tools (HTTP-backed)
	server.AddTools(
		mcp.NewServerTool(
			"tag_list",
			"List all unique tags from todos and memos",
			bridge.TagList,
			mcp.Input(),
		),
	)

	// Register auth tools (HTTP-backed)
	server.AddTools(
		mcp.NewServerTool(
			"auth_start",
			"Start authentication process for memoya",
			authHandler.Start,
			mcp.Input(),
		),
		mcp.NewServerTool(
			"auth_status",
			"Check authentication status and retrieve auth token",
			authHandler.Status,
			mcp.Input(),
		),
	)

	log.Println("Starting MCP client with HTTP transport to Cloud Run...")

	// Run server with stdio transport
	transport := mcp.NewStdioTransport()
	if err := server.Run(ctx, transport); err != nil {
		log.Fatal(err)
	}
}
