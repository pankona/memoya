package validation

import (
	"github.com/pankona/memoya/internal/models"
)

// MemoValidator validates memo inputs
type MemoValidator struct{}

// NewMemoValidator creates a new memo validator
func NewMemoValidator() *MemoValidator {
	return &MemoValidator{}
}

// ValidateCreate validates memo creation input
func (v *MemoValidator) ValidateCreate(title, description string, tags []string) error {
	var errors ValidationErrors

	// Validate title
	if err := ValidateString("title", title, 1, MaxMemoTitleLength, true); err != nil {
		errors = append(errors, err.(ValidationError))
	}

	// Validate description
	if err := ValidateString("description", description, 0, MaxMemoDescriptionLength, false); err != nil {
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

// ValidateUpdate validates memo update input
func (v *MemoValidator) ValidateUpdate(memo *models.Memo) error {
	if memo == nil {
		return ValidationError{Field: "memo", Message: "is required"}
	}

	return v.ValidateCreate(memo.Title, memo.Description, memo.Tags)
}

// SanitizeMemoInput sanitizes memo input data
func SanitizeMemoInput(title, description string, tags []string) (string, string, []string) {
	return SanitizeString(title), SanitizeString(description), SanitizeTags(tags)
}
