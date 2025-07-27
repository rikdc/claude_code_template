package database

import "errors"

// Define sentinel errors for common database conditions
var (
	ErrConversationNotFound = errors.New("conversation not found")
	ErrRatingNotFound       = errors.New("rating not found")
)