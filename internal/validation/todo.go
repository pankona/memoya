package validation

import (
	"fmt"

	"github.com/pankona/memoya/internal/models"
)

// TodoValidator validates todo inputs
type TodoValidator struct{}

// NewTodoValidator creates a new todo validator
func NewTodoValidator() *TodoValidator {
	return &TodoValidator{}
}

// ValidateCreate validates todo creation input
func (v *TodoValidator) ValidateCreate(title, description string, status models.TodoStatus, priority models.TodoPriority, tags []string) error {
	var errors ValidationErrors

	// Validate title
	if err := ValidateString("title", title, 1, MaxTodoTitleLength, true); err != nil {
		errors = append(errors, err.(ValidationError))
	}

	// Validate description
	if err := ValidateString("description", description, 0, MaxTodoDescriptionLength, false); err != nil {
		errors = append(errors, err.(ValidationError))
	}

	// Validate status
	if err := v.validateStatus(status); err != nil {
		errors = append(errors, err.(ValidationError))
	}

	// Validate priority
	if err := v.validatePriority(priority); err != nil {
		errors = append(errors, err.(ValidationError))
	}

	// Validate tags
	if err := ValidateTags(tags); err != nil {
		errors = append(errors, err.(ValidationError))
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// ValidateUpdate validates todo update input
func (v *TodoValidator) ValidateUpdate(todo *models.Todo) error {
	if todo == nil {
		return ValidationError{Field: "todo", Message: "is required"}
	}

	return v.ValidateCreate(todo.Title, todo.Description, todo.Status, todo.Priority, todo.Tags)
}

// validateStatus validates todo status
func (v *TodoValidator) validateStatus(status models.TodoStatus) error {
	switch status {
	case models.StatusBacklog, models.StatusTodo, models.StatusInProgress, models.StatusDone:
		return nil
	case "":
		// Empty status is allowed (will use default)
		return nil
	default:
		return ValidationError{
			Field:   "status",
			Message: fmt.Sprintf("invalid status '%s', must be one of: backlog, todo, in_progress, done", status),
		}
	}
}

// validatePriority validates todo priority
func (v *TodoValidator) validatePriority(priority models.TodoPriority) error {
	switch priority {
	case models.PriorityHigh, models.PriorityNormal:
		return nil
	case "":
		// Empty priority is allowed (will use default)
		return nil
	default:
		return ValidationError{
			Field:   "priority",
			Message: fmt.Sprintf("invalid priority '%s', must be one of: high, normal", priority),
		}
	}
}

// SanitizeTodoInput sanitizes todo input data
func SanitizeTodoInput(title, description string, tags []string) (string, string, []string) {
	return SanitizeString(title), SanitizeString(description), SanitizeTags(tags)
}
