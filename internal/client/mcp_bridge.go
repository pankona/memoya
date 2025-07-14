package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pankona/memoya/internal/handlers"
)

// MCPBridge converts MCP tool calls to HTTP requests
type MCPBridge struct {
	httpClient *HTTPClient
}

// NewMCPBridge creates a new MCP bridge instance
func NewMCPBridge(httpClient *HTTPClient) *MCPBridge {
	return &MCPBridge{
		httpClient: httpClient,
	}
}

// handleError converts HTTP errors into structured JSON responses for MCP
func (b *MCPBridge) handleError(err error) []byte {
	errorMsg := err.Error()

	// Check if it's an authentication error
	if strings.Contains(errorMsg, "AUTHENTICATION_REQUIRED") {
		response := map[string]interface{}{
			"success":       false,
			"authenticated": false,
			"error":         "Authentication required",
			"message":       "Authentication is required to use memoya. Please run the 'auth_start' tool to authenticate.",
			"suggestion":    "Use the auth_start tool to begin the authentication process.",
		}
		jsonBytes, _ := json.Marshal(response)
		return jsonBytes
	}

	// Check if it's an OAuth configuration error
	if strings.Contains(errorMsg, "OAuth") && strings.Contains(errorMsg, "not found") {
		response := map[string]interface{}{
			"success": false,
			"error":   "OAuth configuration missing",
			"message": "The memoya server is not properly configured. Please contact the administrator to set up OAuth credentials.",
			"details": "This is a server-side configuration issue that requires administrator attention.",
		}
		jsonBytes, _ := json.Marshal(response)
		return jsonBytes
	}

	// Generic error response
	response := map[string]interface{}{
		"success": false,
		"error":   errorMsg,
		"message": fmt.Sprintf("Operation failed: %s", errorMsg),
	}
	jsonBytes, _ := json.Marshal(response)
	return jsonBytes
}

// Memo operations

func (b *MCPBridge) MemoCreate(ctx context.Context, ss *mcp.ServerSession,
	params *mcp.CallToolParamsFor[handlers.MemoCreateArgs]) (*mcp.CallToolResultFor[handlers.MemoResult], error) {

	respData, err := b.httpClient.CallTool(ctx, "memo_create", params.Arguments)
	if err != nil {
		// Return structured error response instead of failing
		errorData := b.handleError(err)
		return &mcp.CallToolResultFor[handlers.MemoResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(errorData)},
			},
		}, nil
	}

	return &mcp.CallToolResultFor[handlers.MemoResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(respData)},
		},
	}, nil
}

func (b *MCPBridge) MemoList(ctx context.Context, ss *mcp.ServerSession,
	params *mcp.CallToolParamsFor[handlers.MemoListArgs]) (*mcp.CallToolResultFor[handlers.MemoListResult], error) {

	respData, err := b.httpClient.CallTool(ctx, "memo_list", params.Arguments)
	if err != nil {
		errorData := b.handleError(err)
		return &mcp.CallToolResultFor[handlers.MemoListResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(errorData)},
			},
		}, nil
	}

	return &mcp.CallToolResultFor[handlers.MemoListResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(respData)},
		},
	}, nil
}

func (b *MCPBridge) MemoUpdate(ctx context.Context, ss *mcp.ServerSession,
	params *mcp.CallToolParamsFor[handlers.MemoUpdateArgs]) (*mcp.CallToolResultFor[handlers.MemoResult], error) {

	respData, err := b.httpClient.CallTool(ctx, "memo_update", params.Arguments)
	if err != nil {
		errorData := b.handleError(err)
		return &mcp.CallToolResultFor[handlers.MemoResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(errorData)},
			},
		}, nil
	}

	return &mcp.CallToolResultFor[handlers.MemoResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(respData)},
		},
	}, nil
}

func (b *MCPBridge) MemoDelete(ctx context.Context, ss *mcp.ServerSession,
	params *mcp.CallToolParamsFor[handlers.MemoDeleteArgs]) (*mcp.CallToolResultFor[handlers.MemoDeleteResult], error) {

	respData, err := b.httpClient.CallTool(ctx, "memo_delete", params.Arguments)
	if err != nil {
		errorData := b.handleError(err)
		return &mcp.CallToolResultFor[handlers.MemoDeleteResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(errorData)},
			},
		}, nil
	}

	return &mcp.CallToolResultFor[handlers.MemoDeleteResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(respData)},
		},
	}, nil
}

// Todo operations

func (b *MCPBridge) TodoCreate(ctx context.Context, ss *mcp.ServerSession,
	params *mcp.CallToolParamsFor[handlers.TodoCreateArgs]) (*mcp.CallToolResultFor[handlers.TodoResult], error) {

	respData, err := b.httpClient.CallTool(ctx, "todo_create", params.Arguments)
	if err != nil {
		errorData := b.handleError(err)
		return &mcp.CallToolResultFor[handlers.TodoResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(errorData)},
			},
		}, nil
	}

	return &mcp.CallToolResultFor[handlers.TodoResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(respData)},
		},
	}, nil
}

func (b *MCPBridge) TodoList(ctx context.Context, ss *mcp.ServerSession,
	params *mcp.CallToolParamsFor[handlers.TodoListArgs]) (*mcp.CallToolResultFor[handlers.TodoListResult], error) {

	respData, err := b.httpClient.CallTool(ctx, "todo_list", params.Arguments)
	if err != nil {
		errorData := b.handleError(err)
		return &mcp.CallToolResultFor[handlers.TodoListResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(errorData)},
			},
		}, nil
	}

	return &mcp.CallToolResultFor[handlers.TodoListResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(respData)},
		},
	}, nil
}

func (b *MCPBridge) TodoUpdate(ctx context.Context, ss *mcp.ServerSession,
	params *mcp.CallToolParamsFor[handlers.TodoUpdateArgs]) (*mcp.CallToolResultFor[handlers.TodoResult], error) {

	respData, err := b.httpClient.CallTool(ctx, "todo_update", params.Arguments)
	if err != nil {
		errorData := b.handleError(err)
		return &mcp.CallToolResultFor[handlers.TodoResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(errorData)},
			},
		}, nil
	}

	return &mcp.CallToolResultFor[handlers.TodoResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(respData)},
		},
	}, nil
}

func (b *MCPBridge) TodoDelete(ctx context.Context, ss *mcp.ServerSession,
	params *mcp.CallToolParamsFor[handlers.TodoDeleteArgs]) (*mcp.CallToolResultFor[handlers.DeleteResult], error) {

	respData, err := b.httpClient.CallTool(ctx, "todo_delete", params.Arguments)
	if err != nil {
		errorData := b.handleError(err)
		return &mcp.CallToolResultFor[handlers.DeleteResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(errorData)},
			},
		}, nil
	}

	return &mcp.CallToolResultFor[handlers.DeleteResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(respData)},
		},
	}, nil
}

// Search operations

func (b *MCPBridge) Search(ctx context.Context, ss *mcp.ServerSession,
	params *mcp.CallToolParamsFor[handlers.SearchArgs]) (*mcp.CallToolResultFor[handlers.SearchResult], error) {

	respData, err := b.httpClient.CallTool(ctx, "search", params.Arguments)
	if err != nil {
		errorData := b.handleError(err)
		return &mcp.CallToolResultFor[handlers.SearchResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(errorData)},
			},
		}, nil
	}

	return &mcp.CallToolResultFor[handlers.SearchResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(respData)},
		},
	}, nil
}

// Tag operations

func (b *MCPBridge) TagList(ctx context.Context, ss *mcp.ServerSession,
	params *mcp.CallToolParamsFor[handlers.TagListArgs]) (*mcp.CallToolResultFor[handlers.TagListResult], error) {

	respData, err := b.httpClient.CallTool(ctx, "tag_list", params.Arguments)
	if err != nil {
		errorData := b.handleError(err)
		return &mcp.CallToolResultFor[handlers.TagListResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(errorData)},
			},
		}, nil
	}

	return &mcp.CallToolResultFor[handlers.TagListResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(respData)},
		},
	}, nil
}
