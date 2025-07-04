package client

import (
	"context"
	"fmt"

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

// Memo operations

func (b *MCPBridge) MemoCreate(ctx context.Context, ss *mcp.ServerSession,
	params *mcp.CallToolParamsFor[handlers.MemoCreateArgs]) (*mcp.CallToolResultFor[handlers.MemoResult], error) {

	respData, err := b.httpClient.CallTool(ctx, "memo_create", params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("memo_create failed: %w", err)
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
		return nil, fmt.Errorf("memo_list failed: %w", err)
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
		return nil, fmt.Errorf("memo_update failed: %w", err)
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
		return nil, fmt.Errorf("memo_delete failed: %w", err)
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
		return nil, fmt.Errorf("todo_create failed: %w", err)
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
		return nil, fmt.Errorf("todo_list failed: %w", err)
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
		return nil, fmt.Errorf("todo_update failed: %w", err)
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
		return nil, fmt.Errorf("todo_delete failed: %w", err)
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
		return nil, fmt.Errorf("search failed: %w", err)
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
		return nil, fmt.Errorf("tag_list failed: %w", err)
	}

	return &mcp.CallToolResultFor[handlers.TagListResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(respData)},
		},
	}, nil
}
