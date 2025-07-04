package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pankona/memoya/internal/generated/server"
	"github.com/pankona/memoya/internal/handlers"
	"github.com/pankona/memoya/internal/storage"
)

// Server implements the generated ServerInterface
type Server struct {
	memoHandler   *handlers.MemoHandler
	todoHandler   *handlers.TodoHandler
	searchHandler *handlers.SearchHandler
	tagHandler    *handlers.TagHandler
}

// NewServer creates a new server instance
func NewServer(storage storage.Storage) *Server {
	return &Server{
		memoHandler:   handlers.NewMemoHandlerWithStorage(storage),
		todoHandler:   handlers.NewTodoHandlerWithStorage(storage),
		searchHandler: handlers.NewSearchHandler(storage),
		tagHandler:    handlers.NewTagHandler(storage),
	}
}

// writeErrorResponse writes an error response
func writeErrorResponse(w http.ResponseWriter, statusCode int, message, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	successFlag := false
	errorResp := server.Error{
		Success: &successFlag,
		Error:   &message,
		Code:    &code,
	}
	
	json.NewEncoder(w).Encode(errorResp)
}

// writeSuccessResponse writes the MCP handler result as JSON
func writeSuccessResponse(w http.ResponseWriter, result interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	// Extract JSON from TextContent using type switch
	switch r := result.(type) {
	case *mcp.CallToolResultFor[handlers.MemoResult]:
		if len(r.Content) > 0 {
			if textContent, ok := r.Content[0].(*mcp.TextContent); ok {
				w.Write([]byte(textContent.Text))
				return nil
			}
		}
	case *mcp.CallToolResultFor[handlers.MemoListResult]:
		if len(r.Content) > 0 {
			if textContent, ok := r.Content[0].(*mcp.TextContent); ok {
				w.Write([]byte(textContent.Text))
				return nil
			}
		}
	case *mcp.CallToolResultFor[handlers.MemoDeleteResult]:
		if len(r.Content) > 0 {
			if textContent, ok := r.Content[0].(*mcp.TextContent); ok {
				w.Write([]byte(textContent.Text))
				return nil
			}
		}
	case *mcp.CallToolResultFor[handlers.TodoResult]:
		if len(r.Content) > 0 {
			if textContent, ok := r.Content[0].(*mcp.TextContent); ok {
				w.Write([]byte(textContent.Text))
				return nil
			}
		}
	case *mcp.CallToolResultFor[handlers.TodoListResult]:
		if len(r.Content) > 0 {
			if textContent, ok := r.Content[0].(*mcp.TextContent); ok {
				w.Write([]byte(textContent.Text))
				return nil
			}
		}
	case *mcp.CallToolResultFor[handlers.DeleteResult]:
		if len(r.Content) > 0 {
			if textContent, ok := r.Content[0].(*mcp.TextContent); ok {
				w.Write([]byte(textContent.Text))
				return nil
			}
		}
	case *mcp.CallToolResultFor[handlers.SearchResult]:
		if len(r.Content) > 0 {
			if textContent, ok := r.Content[0].(*mcp.TextContent); ok {
				w.Write([]byte(textContent.Text))
				return nil
			}
		}
	case *mcp.CallToolResultFor[handlers.TagListResult]:
		if len(r.Content) > 0 {
			if textContent, ok := r.Content[0].(*mcp.TextContent); ok {
				w.Write([]byte(textContent.Text))
				return nil
			}
		}
	}
	
	return fmt.Errorf("invalid response format")
}

// HealthCheck implements the health check endpoint
func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateMemo implements POST /mcp/memo_create
func (s *Server) CreateMemo(w http.ResponseWriter, r *http.Request) {
	var req server.MemoCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", "BAD_REQUEST")
		return
	}

	// Convert to MCP handler arguments
	args := handlers.MemoCreateArgs{
		Title:       req.Title,
		Description: getStringValue(req.Description),
		Tags:        getStringSliceValue(req.Tags),
		LinkedTodos: getStringSliceValue(req.LinkedTodos),
	}

	params := &mcp.CallToolParamsFor[handlers.MemoCreateArgs]{Arguments: args}
	result, err := s.memoHandler.Create(r.Context(), nil, params)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	if err := writeSuccessResponse(w, result); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to encode response", "INTERNAL_ERROR")
	}
}

// ListMemos implements POST /mcp/memo_list
func (s *Server) ListMemos(w http.ResponseWriter, r *http.Request) {
	var req server.MemoListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", "BAD_REQUEST")
		return
	}

	args := handlers.MemoListArgs{
		Tags: getStringSliceValue(req.Tags),
	}

	params := &mcp.CallToolParamsFor[handlers.MemoListArgs]{Arguments: args}
	result, err := s.memoHandler.List(r.Context(), nil, params)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	if err := writeSuccessResponse(w, result); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to encode response", "INTERNAL_ERROR")
	}
}

// UpdateMemo implements POST /mcp/memo_update
func (s *Server) UpdateMemo(w http.ResponseWriter, r *http.Request) {
	var req server.MemoUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", "BAD_REQUEST")
		return
	}

	args := handlers.MemoUpdateArgs{
		ID:          req.Id,
		Title:       getStringValue(req.Title),
		Description: getStringValue(req.Description),
		Tags:        getStringSliceValue(req.Tags),
		LinkedTodos: getStringSliceValue(req.LinkedTodos),
	}

	params := &mcp.CallToolParamsFor[handlers.MemoUpdateArgs]{Arguments: args}
	result, err := s.memoHandler.Update(r.Context(), nil, params)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	if err := writeSuccessResponse(w, result); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to encode response", "INTERNAL_ERROR")
	}
}

// DeleteMemo implements POST /mcp/memo_delete
func (s *Server) DeleteMemo(w http.ResponseWriter, r *http.Request) {
	var req server.MemoDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", "BAD_REQUEST")
		return
	}

	args := handlers.MemoDeleteArgs{
		ID: req.Id,
	}

	params := &mcp.CallToolParamsFor[handlers.MemoDeleteArgs]{Arguments: args}
	result, err := s.memoHandler.Delete(r.Context(), nil, params)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	if err := writeSuccessResponse(w, result); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to encode response", "INTERNAL_ERROR")
	}
}

// CreateTodo implements POST /mcp/todo_create
func (s *Server) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var req server.TodoCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", "BAD_REQUEST")
		return
	}

	args := handlers.TodoCreateArgs{
		Title:       req.Title,
		Description: getStringValue(req.Description),
		Status:      getStatusValue(req.Status),
		Priority:    getPriorityValue(req.Priority),
		Tags:        getStringSliceValue(req.Tags),
		ParentID:    getStringValue(req.ParentId),
	}

	params := &mcp.CallToolParamsFor[handlers.TodoCreateArgs]{Arguments: args}
	result, err := s.todoHandler.Create(r.Context(), nil, params)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	if err := writeSuccessResponse(w, result); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to encode response", "INTERNAL_ERROR")
	}
}

// ListTodos implements POST /mcp/todo_list
func (s *Server) ListTodos(w http.ResponseWriter, r *http.Request) {
	var req server.TodoListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", "BAD_REQUEST")
		return
	}

	args := handlers.TodoListArgs{
		Status:   getListStatusValue(req.Status),
		Priority: getListPriorityValue(req.Priority),
		Tags:     getStringSliceValue(req.Tags),
	}

	params := &mcp.CallToolParamsFor[handlers.TodoListArgs]{Arguments: args}
	result, err := s.todoHandler.List(r.Context(), nil, params)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	if err := writeSuccessResponse(w, result); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to encode response", "INTERNAL_ERROR")
	}
}

// UpdateTodo implements POST /mcp/todo_update
func (s *Server) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	var req server.TodoUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", "BAD_REQUEST")
		return
	}

	args := handlers.TodoUpdateArgs{
		ID:          req.Id,
		Title:       getStringValue(req.Title),
		Description: getStringValue(req.Description),
		Status:      getUpdateStatusValue(req.Status),
		Priority:    getUpdatePriorityValue(req.Priority),
		Tags:        getStringSliceValue(req.Tags),
	}

	params := &mcp.CallToolParamsFor[handlers.TodoUpdateArgs]{Arguments: args}
	result, err := s.todoHandler.Update(r.Context(), nil, params)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	if err := writeSuccessResponse(w, result); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to encode response", "INTERNAL_ERROR")
	}
}

// DeleteTodo implements POST /mcp/todo_delete
func (s *Server) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	var req server.TodoDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", "BAD_REQUEST")
		return
	}

	args := handlers.TodoDeleteArgs{
		ID: req.Id,
	}

	params := &mcp.CallToolParamsFor[handlers.TodoDeleteArgs]{Arguments: args}
	result, err := s.todoHandler.Delete(r.Context(), nil, params)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	if err := writeSuccessResponse(w, result); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to encode response", "INTERNAL_ERROR")
	}
}

// Search implements POST /mcp/search
func (s *Server) Search(w http.ResponseWriter, r *http.Request) {
	var req server.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", "BAD_REQUEST")
		return
	}

	args := handlers.SearchArgs{
		Query: getStringValue(req.Query),
		Tags:  getStringSliceValue(req.Tags),
		Type:  getSearchTypeValue(req.Type),
	}

	params := &mcp.CallToolParamsFor[handlers.SearchArgs]{Arguments: args}
	result, err := s.searchHandler.Search(r.Context(), nil, params)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	if err := writeSuccessResponse(w, result); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to encode response", "INTERNAL_ERROR")
	}
}

// ListTags implements POST /mcp/tag_list
func (s *Server) ListTags(w http.ResponseWriter, r *http.Request) {
	var req server.TagListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", "BAD_REQUEST")
		return
	}

	args := handlers.TagListArgs{}

	params := &mcp.CallToolParamsFor[handlers.TagListArgs]{Arguments: args}
	result, err := s.tagHandler.List(r.Context(), nil, params)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	if err := writeSuccessResponse(w, result); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to encode response", "INTERNAL_ERROR")
	}
}

// Helper functions to handle optional values
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func getStringSliceValue(ptr *[]string) []string {
	if ptr == nil {
		return []string{}
	}
	return *ptr
}

// Helper functions for generated enum types
func getStatusValue(ptr *server.TodoCreateRequestStatus) string {
	if ptr == nil {
		return ""
	}
	return string(*ptr)
}

func getPriorityValue(ptr *server.TodoCreateRequestPriority) string {
	if ptr == nil {
		return ""
	}
	return string(*ptr)
}

func getListStatusValue(ptr *server.TodoListRequestStatus) string {
	if ptr == nil {
		return ""
	}
	return string(*ptr)
}

func getListPriorityValue(ptr *server.TodoListRequestPriority) string {
	if ptr == nil {
		return ""
	}
	return string(*ptr)
}

func getUpdateStatusValue(ptr *server.TodoUpdateRequestStatus) string {
	if ptr == nil {
		return ""
	}
	return string(*ptr)
}

func getUpdatePriorityValue(ptr *server.TodoUpdateRequestPriority) string {
	if ptr == nil {
		return ""
	}
	return string(*ptr)
}

func getSearchTypeValue(ptr *server.SearchRequestType) string {
	if ptr == nil {
		return ""
	}
	return string(*ptr)
}