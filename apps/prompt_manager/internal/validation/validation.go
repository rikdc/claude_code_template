package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Constants for validation limits
const (
	MaxTitleLength       = 200
	MaxContentLength     = 100000 // 100KB of text
	MaxCommentLength     = 1000
	MaxPathLength        = 1000
	MaxSessionIDLength   = 100
	MaxToolCallLength    = 50000 // 50KB for tool calls JSON
	MinRating           = 1
	MaxRating           = 5
	MaxPageSize         = 100
	MinPageSize         = 1
	MaxPageNumber       = 10000
)

// Regular expressions for validation
var (
	sessionIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	pathRegex      = regexp.MustCompile(`^[a-zA-Z0-9._/\\:-]+$`)
)

// ValidationError represents input validation errors
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// SanitizeString removes potentially dangerous characters and limits length
func SanitizeString(input string, maxLength int) string {
	// Remove null bytes and other control characters except newlines and tabs
	cleaned := strings.Map(func(r rune) rune {
		if r == 0 || (r < 32 && r != '\n' && r != '\t' && r != '\r') {
			return -1 // Remove character
		}
		return r
	}, input)
	
	// Trim whitespace
	cleaned = strings.TrimSpace(cleaned)
	
	// Limit length
	if len(cleaned) > maxLength {
		return cleaned[:maxLength]
	}
	
	return cleaned
}

// ValidateSessionID validates a session ID
func ValidateSessionID(sessionID string) error {
	if sessionID == "" {
		return &ValidationError{Field: "session_id", Message: "cannot be empty"}
	}
	
	if len(sessionID) > MaxSessionIDLength {
		return &ValidationError{
			Field:   "session_id",
			Value:   sessionID,
			Message: fmt.Sprintf("cannot exceed %d characters", MaxSessionIDLength),
		}
	}
	
	if !sessionIDRegex.MatchString(sessionID) {
		return &ValidationError{
			Field:   "session_id",
			Message: "can only contain letters, numbers, underscores, and hyphens",
		}
	}
	
	return nil
}

// ValidateTitle validates a conversation title
func ValidateTitle(title *string) error {
	if title == nil {
		return nil // Title is optional
	}
	
	if len(*title) > MaxTitleLength {
		return &ValidationError{
			Field:   "title",
			Value:   *title,
			Message: fmt.Sprintf("cannot exceed %d characters", MaxTitleLength),
		}
	}
	
	if !utf8.ValidString(*title) {
		return &ValidationError{
			Field:   "title",
			Message: "must be valid UTF-8",
		}
	}
	
	return nil
}

// ValidateContent validates message content
func ValidateContent(content string) error {
	if content == "" {
		return &ValidationError{Field: "content", Message: "cannot be empty"}
	}
	
	if len(content) > MaxContentLength {
		return &ValidationError{
			Field:   "content",
			Message: fmt.Sprintf("cannot exceed %d characters", MaxContentLength),
		}
	}
	
	if !utf8.ValidString(content) {
		return &ValidationError{
			Field:   "content",
			Message: "must be valid UTF-8",
		}
	}
	
	return nil
}

// ValidateComment validates rating comments
func ValidateComment(comment *string) error {
	if comment == nil {
		return nil // Comment is optional
	}
	
	if len(*comment) > MaxCommentLength {
		return &ValidationError{
			Field:   "comment",
			Value:   *comment,
			Message: fmt.Sprintf("cannot exceed %d characters", MaxCommentLength),
		}
	}
	
	if !utf8.ValidString(*comment) {
		return &ValidationError{
			Field:   "comment",
			Message: "must be valid UTF-8",
		}
	}
	
	return nil
}

// ValidatePath validates file paths
func ValidatePath(path *string) error {
	if path == nil {
		return nil // Path is optional
	}
	
	if len(*path) > MaxPathLength {
		return &ValidationError{
			Field:   "path",
			Value:   *path,
			Message: fmt.Sprintf("cannot exceed %d characters", MaxPathLength),
		}
	}
	
	if !pathRegex.MatchString(*path) {
		return &ValidationError{
			Field:   "path",
			Message: "contains invalid characters",
		}
	}
	
	return nil
}

// ValidateRating validates rating values
func ValidateRating(rating int) error {
	if rating < MinRating || rating > MaxRating {
		return &ValidationError{
			Field:   "rating",
			Value:   rating,
			Message: fmt.Sprintf("must be between %d and %d", MinRating, MaxRating),
		}
	}
	
	return nil
}

// ValidateID validates positive integer IDs
func ValidateID(id int, fieldName string) error {
	if id <= 0 {
		return &ValidationError{
			Field:   fieldName,
			Value:   id,
			Message: "must be a positive integer",
		}
	}
	
	return nil
}

// ValidatePagination validates pagination parameters
func ValidatePagination(page, perPage int) error {
	if page < 1 {
		return &ValidationError{
			Field:   "page",
			Value:   page,
			Message: "must be at least 1",
		}
	}
	
	if page > MaxPageNumber {
		return &ValidationError{
			Field:   "page",
			Value:   page,
			Message: fmt.Sprintf("cannot exceed %d", MaxPageNumber),
		}
	}
	
	if perPage < MinPageSize {
		return &ValidationError{
			Field:   "per_page",
			Value:   perPage,
			Message: fmt.Sprintf("must be at least %d", MinPageSize),
		}
	}
	
	if perPage > MaxPageSize {
		return &ValidationError{
			Field:   "per_page",
			Value:   perPage,
			Message: fmt.Sprintf("cannot exceed %d", MaxPageSize),
		}
	}
	
	return nil
}

// ParseAndValidateID safely parses and validates an ID from a string
func ParseAndValidateID(idStr, fieldName string) (int, error) {
	if idStr == "" {
		return 0, &ValidationError{
			Field:   fieldName,
			Message: "cannot be empty",
		}
	}
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, &ValidationError{
			Field:   fieldName,
			Value:   idStr,
			Message: "must be a valid integer",
		}
	}
	
	if err := ValidateID(id, fieldName); err != nil {
		return 0, err
	}
	
	return id, nil
}

// ParseAndValidatePage safely parses pagination parameters
func ParseAndValidatePage(pageStr, perPageStr string) (int, int, error) {
	page := 1
	perPage := 20 // Default page size
	
	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil {
			return 0, 0, &ValidationError{
				Field:   "page",
				Value:   pageStr,
				Message: "must be a valid integer",
			}
		}
		page = p
	}
	
	if perPageStr != "" {
		pp, err := strconv.Atoi(perPageStr)
		if err != nil {
			return 0, 0, &ValidationError{
				Field:   "per_page",
				Value:   perPageStr,
				Message: "must be a valid integer",
			}
		}
		perPage = pp
	}
	
	if err := ValidatePagination(page, perPage); err != nil {
		return 0, 0, err
	}
	
	return page, perPage, nil
}

// IsValidationError checks if an error is a ValidationError
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}