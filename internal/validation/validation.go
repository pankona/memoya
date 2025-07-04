package validation

import (
	"fmt"
	"html"
	"strings"
	"unicode/utf8"
)

const (
	// Memo constraints
	MaxMemoTitleLength       = 200
	MaxMemoDescriptionLength = 10000
	MaxTagLength             = 50
	MaxTagsCount             = 20

	// Todo constraints
	MaxTodoTitleLength       = 200
	MaxTodoDescriptionLength = 5000
)

// ValidationError represents a validation error with field information
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// ValidateString validates a string field with length constraints
func ValidateString(field, value string, minLength, maxLength int, required bool) error {
	length := utf8.RuneCountInString(value)

	if required && length == 0 {
		return ValidationError{Field: field, Message: "is required"}
	}

	if length > 0 && length < minLength {
		return ValidationError{Field: field, Message: fmt.Sprintf("must be at least %d characters", minLength)}
	}

	if maxLength > 0 && length > maxLength {
		return ValidationError{Field: field, Message: fmt.Sprintf("must be at most %d characters", maxLength)}
	}

	return nil
}

// ValidateTags validates a slice of tags
func ValidateTags(tags []string) error {
	if len(tags) > MaxTagsCount {
		return ValidationError{Field: "tags", Message: fmt.Sprintf("must have at most %d tags", MaxTagsCount)}
	}

	for i, tag := range tags {
		if err := ValidateString(fmt.Sprintf("tags[%d]", i), tag, 1, MaxTagLength, false); err != nil {
			return err
		}
	}

	return nil
}

// SanitizeString performs basic sanitization on a string
func SanitizeString(s string) string {
	// Trim whitespace
	s = strings.TrimSpace(s)
	// Escape HTML to prevent XSS
	s = html.EscapeString(s)
	return s
}

// SanitizeTags sanitizes a slice of tags
func SanitizeTags(tags []string) []string {
	sanitized := make([]string, 0, len(tags))
	seen := make(map[string]bool)

	for _, tag := range tags {
		tag = SanitizeString(tag)
		// Remove duplicates
		if tag != "" && !seen[strings.ToLower(tag)] {
			sanitized = append(sanitized, tag)
			seen[strings.ToLower(tag)] = true
		}
	}

	return sanitized
}
