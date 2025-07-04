package validation

import (
	"strings"
	"testing"
)

func TestValidateString(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		minLength int
		maxLength int
		required  bool
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid string",
			field:     "title",
			value:     "Test Title",
			minLength: 1,
			maxLength: 100,
			required:  true,
			wantErr:   false,
		},
		{
			name:      "empty required string",
			field:     "title",
			value:     "",
			minLength: 1,
			maxLength: 100,
			required:  true,
			wantErr:   true,
			errMsg:    "is required",
		},
		{
			name:      "empty optional string",
			field:     "description",
			value:     "",
			minLength: 0,
			maxLength: 100,
			required:  false,
			wantErr:   false,
		},
		{
			name:      "string too short",
			field:     "title",
			value:     "a",
			minLength: 5,
			maxLength: 100,
			required:  true,
			wantErr:   true,
			errMsg:    "must be at least 5 characters",
		},
		{
			name:      "string too long",
			field:     "title",
			value:     strings.Repeat("a", 101),
			minLength: 1,
			maxLength: 100,
			required:  true,
			wantErr:   true,
			errMsg:    "must be at most 100 characters",
		},
		{
			name:      "unicode string length",
			field:     "title",
			value:     "テスト", // 3 characters in Japanese
			minLength: 1,
			maxLength: 5,
			required:  true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateString(tt.field, tt.value, tt.minLength, tt.maxLength, tt.required)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateString() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateString() error message = %v, want to contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateTags(t *testing.T) {
	tests := []struct {
		name    string
		tags    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid tags",
			tags:    []string{"tag1", "tag2", "tag3"},
			wantErr: false,
		},
		{
			name:    "empty tags",
			tags:    []string{},
			wantErr: false,
		},
		{
			name:    "too many tags",
			tags:    make([]string, MaxTagsCount+1),
			wantErr: true,
			errMsg:  "must have at most",
		},
		{
			name:    "tag too long",
			tags:    []string{strings.Repeat("a", MaxTagLength+1)},
			wantErr: true,
			errMsg:  "must be at most",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize tags for "too many tags" test
			if tt.name == "too many tags" {
				for i := range tt.tags {
					tt.tags[i] = "tag"
				}
			}

			err := ValidateTags(tt.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTags() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateTags() error message = %v, want to contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normal string",
			input: "Hello World",
			want:  "Hello World",
		},
		{
			name:  "string with spaces",
			input: "  Hello World  ",
			want:  "Hello World",
		},
		{
			name:  "string with HTML",
			input: "<script>alert('xss')</script>",
			want:  "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:  "string with quotes",
			input: `"Hello" & 'World'`,
			want:  `&#34;Hello&#34; &amp; &#39;World&#39;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeString(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizeTags(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "normal tags",
			input: []string{"tag1", "tag2", "tag3"},
			want:  []string{"tag1", "tag2", "tag3"},
		},
		{
			name:  "tags with spaces",
			input: []string{"  tag1  ", "tag2", "  tag3"},
			want:  []string{"tag1", "tag2", "tag3"},
		},
		{
			name:  "duplicate tags",
			input: []string{"tag1", "TAG1", "tag1", "tag2"},
			want:  []string{"tag1", "tag2"},
		},
		{
			name:  "empty tags",
			input: []string{"tag1", "", "  ", "tag2"},
			want:  []string{"tag1", "tag2"},
		},
		{
			name:  "tags with HTML",
			input: []string{"<tag>", "normal", "&tag"},
			want:  []string{"&lt;tag&gt;", "normal", "&amp;tag"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeTags(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("SanitizeTags() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i, tag := range got {
				if tag != tt.want[i] {
					t.Errorf("SanitizeTags()[%d] = %v, want %v", i, tag, tt.want[i])
				}
			}
		})
	}
}
