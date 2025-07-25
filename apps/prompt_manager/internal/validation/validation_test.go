package validation

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateSessionID(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		expectErr bool
	}{
		{"valid session ID", "session-123_abc", false},
		{"empty session ID", "", true},
		{"too long session ID", strings.Repeat("a", MaxSessionIDLength+1), true},
		{"invalid characters", "session@123", true},
		{"valid with underscores", "session_123", false},
		{"valid with hyphens", "session-123", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSessionID(tt.sessionID)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateSessionID() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestValidateTitle(t *testing.T) {
	tests := []struct {
		name      string
		title     *string
		expectErr bool
	}{
		{"nil title", nil, false},
		{"valid title", stringPtr("My Conversation"), false},
		{"too long title", stringPtr(strings.Repeat("a", MaxTitleLength+1)), true},
		{"empty title", stringPtr(""), false},
		{"valid UTF-8", stringPtr("Hello 世界"), false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTitle(tt.title)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateTitle() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestValidateContent(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		expectErr bool
	}{
		{"valid content", "Hello world", false},
		{"empty content", "", true},
		{"too long content", strings.Repeat("a", MaxContentLength+1), true},
		{"valid UTF-8", "Hello 世界", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateContent(tt.content)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateContent() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestValidateRating(t *testing.T) {
	tests := []struct {
		name      string
		rating    int
		expectErr bool
	}{
		{"valid rating 1", 1, false},
		{"valid rating 5", 5, false},
		{"valid rating 3", 3, false},
		{"invalid rating 0", 0, true},
		{"invalid rating 6", 6, true},
		{"invalid rating negative", -1, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRating(tt.rating)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateRating() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		fieldName string
		expectErr bool
	}{
		{"valid ID", 123, "id", false},
		{"zero ID", 0, "id", true},
		{"negative ID", -1, "id", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateID(tt.id, tt.fieldName)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateID() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestValidatePagination(t *testing.T) {
	tests := []struct {
		name      string
		page      int
		perPage   int
		expectErr bool
	}{
		{"valid pagination", 1, 20, false},
		{"max valid pagination", MaxPageNumber, MaxPageSize, false},
		{"page too small", 0, 20, true},
		{"page too large", MaxPageNumber + 1, 20, true},
		{"per_page too small", 1, 0, true},
		{"per_page too large", 1, MaxPageSize + 1, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePagination(tt.page, tt.perPage)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidatePagination() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestParseAndValidateID(t *testing.T) {
	tests := []struct {
		name      string
		idStr     string
		fieldName string
		expected  int
		expectErr bool
	}{
		{"valid ID string", "123", "id", 123, false},
		{"empty string", "", "id", 0, true},
		{"invalid string", "abc", "id", 0, true},
		{"zero ID", "0", "id", 0, true},
		{"negative ID", "-1", "id", 0, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAndValidateID(tt.idStr, tt.fieldName)
			if (err != nil) != tt.expectErr {
				t.Errorf("ParseAndValidateID() error = %v, expectErr %v", err, tt.expectErr)
			}
			if !tt.expectErr && result != tt.expected {
				t.Errorf("ParseAndValidateID() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestParseAndValidatePage(t *testing.T) {
	tests := []struct {
		name         string
		pageStr      string
		perPageStr   string
		expectedPage int
		expectedPer  int
		expectErr    bool
	}{
		{"default values", "", "", 1, 20, false},
		{"valid values", "2", "50", 2, 50, false},
		{"invalid page", "abc", "20", 0, 0, true},
		{"invalid per_page", "1", "abc", 0, 0, true},
		{"page too large", "99999", "20", 0, 0, true},
		{"per_page too large", "1", "999", 0, 0, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, perPage, err := ParseAndValidatePage(tt.pageStr, tt.perPageStr)
			if (err != nil) != tt.expectErr {
				t.Errorf("ParseAndValidatePage() error = %v, expectErr %v", err, tt.expectErr)
			}
			if !tt.expectErr {
				if page != tt.expectedPage {
					t.Errorf("ParseAndValidatePage() page = %v, expected %v", page, tt.expectedPage)
				}
				if perPage != tt.expectedPer {
					t.Errorf("ParseAndValidatePage() perPage = %v, expected %v", perPage, tt.expectedPer)
				}
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{"normal string", "hello world", 100, "hello world"},
		{"string with null bytes", "hello\x00world", 100, "helloworld"},
		{"string too long", "hello world", 5, "hello"},
		{"string with tabs and newlines", "hello\tworld\n", 100, "hello\tworld"},
		{"string with whitespace", "  hello world  ", 100, "hello world"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input, tt.maxLength)
			if result != tt.expected {
				t.Errorf("SanitizeString() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestIsValidationError(t *testing.T) {
	validationErr := &ValidationError{Field: "test", Message: "test error"}
	regularErr := errors.New("regular error")
	
	if !IsValidationError(validationErr) {
		t.Error("Expected validation error to be identified as ValidationError")
	}
	
	if IsValidationError(regularErr) {
		t.Error("Expected regular error to not be identified as ValidationError")
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}