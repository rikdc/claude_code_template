package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// Conversation represents a conversation thread with metadata
type Conversation struct {
	ID               int                     `json:"id"`
	SessionID        string                  `json:"session_id"`
	Title            *string                 `json:"title,omitempty"`
	CreatedAt        time.Time               `json:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
	PromptCount      int                     `json:"prompt_count"`
	TotalCharacters  int                     `json:"total_characters"`
	WorkingDirectory *string                 `json:"working_directory,omitempty"`
	TranscriptPath   *string                 `json:"transcript_path,omitempty"`
	Messages         []Message               `json:"messages,omitempty"`
	Ratings          []Rating                `json:"ratings,omitempty"`
	Tags             []Tag                   `json:"tags,omitempty"`
	Metadata         map[string]interface{}  `json:"metadata,omitempty"`
}

// Message represents an individual message (prompt or response) in a conversation
type Message struct {
	ID             int                    `json:"id"`
	ConversationID int                    `json:"conversation_id"`
	MessageType    MessageType            `json:"message_type"`
	Content        string                 `json:"content"`
	CharacterCount int                    `json:"character_count"`
	Timestamp      time.Time              `json:"timestamp"`
	ToolCalls      []ToolCall             `json:"tool_calls,omitempty"`
	ExecutionTime  *int                   `json:"execution_time,omitempty"` // milliseconds
	Ratings        []Rating               `json:"ratings,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// MessageType represents the type of message
type MessageType string

const (
	MessageTypePrompt   MessageType = "prompt"
	MessageTypeResponse MessageType = "response"
)

// ToolCall represents a tool call made during message processing
type ToolCall struct {
	Name       string                 `json:"name"`
	Arguments  map[string]interface{} `json:"arguments"`
	Result     *string                `json:"result,omitempty"`
	Error      *string                `json:"error,omitempty"`
	Duration   *int                   `json:"duration,omitempty"` // milliseconds
}

// Session represents a Claude Code session with aggregated metrics
type Session struct {
	ID                  int       `json:"id"`
	SessionID           string    `json:"session_id"`
	StartTime           time.Time `json:"start_time"`
	EndTime             *time.Time `json:"end_time,omitempty"`
	ConversationCount   int       `json:"conversation_count"`
	TotalPromptCount    int       `json:"total_prompt_count"`
	AvgResponseTime     int       `json:"avg_response_time"` // milliseconds
	WorkingDirectory    *string   `json:"working_directory,omitempty"`
	Status              SessionStatus `json:"status"`
	Conversations       []Conversation `json:"conversations,omitempty"`
}

// SessionStatus represents the status of a session
type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusArchived  SessionStatus = "archived"
)

// Rating represents a user rating for a conversation or message
type Rating struct {
	ID             int        `json:"id"`
	ConversationID *int       `json:"conversation_id,omitempty"`
	MessageID      *int       `json:"message_id,omitempty"`
	Rating         int        `json:"rating"` // 1-5 scale
	Comment        *string    `json:"comment,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Tag represents a tag that can be applied to conversations
type Tag struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Color       *string   `json:"color,omitempty"` // hex color code
	CreatedAt   time.Time `json:"created_at"`
	UsageCount  int       `json:"usage_count,omitempty"` // computed field
}

// ConversationTag represents the many-to-many relationship between conversations and tags
type ConversationTag struct {
	ConversationID int       `json:"conversation_id"`
	TagID          int       `json:"tag_id"`
	CreatedAt      time.Time `json:"created_at"`
}

// ConversationSummary provides aggregated information about a conversation
type ConversationSummary struct {
	ID              int       `json:"id"`
	SessionID       string    `json:"session_id"`
	Title           *string   `json:"title,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	PromptCount     int       `json:"prompt_count"`
	ResponseCount   int       `json:"response_count"`
	TotalCharacters int       `json:"total_characters"`
	AvgRating       *float64  `json:"avg_rating,omitempty"`
	TagCount        int       `json:"tag_count"`
	Tags            []Tag     `json:"tags,omitempty"`
}

// Validation methods

// Validate checks if the conversation model is valid
func (c *Conversation) Validate() error {
	if c.SessionID == "" {
		return fmt.Errorf("session_id is required")
	}
	
	if c.PromptCount < 0 {
		return fmt.Errorf("prompt_count cannot be negative")
	}
	
	if c.TotalCharacters < 0 {
		return fmt.Errorf("total_characters cannot be negative")
	}
	
	return nil
}

// Validate checks if the message model is valid
func (m *Message) Validate() error {
	if m.ConversationID <= 0 {
		return fmt.Errorf("conversation_id is required")
	}
	
	if m.MessageType != MessageTypePrompt && m.MessageType != MessageTypeResponse {
		return fmt.Errorf("message_type must be 'prompt' or 'response'")
	}
	
	if m.Content == "" {
		return fmt.Errorf("content is required")
	}
	
	if m.CharacterCount != len(m.Content) {
		return fmt.Errorf("character_count mismatch")
	}
	
	if m.ExecutionTime != nil && *m.ExecutionTime < 0 {
		return fmt.Errorf("execution_time cannot be negative")
	}
	
	return nil
}

// Validate checks if the rating model is valid
func (r *Rating) Validate() error {
	if r.ConversationID == nil && r.MessageID == nil {
		return fmt.Errorf("either conversation_id or message_id is required")
	}
	
	if r.ConversationID != nil && r.MessageID != nil {
		return fmt.Errorf("cannot specify both conversation_id and message_id")
	}
	
	if r.Rating < 1 || r.Rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}
	
	return nil
}

// Validate checks if the tag model is valid
func (t *Tag) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("name is required")
	}
	
	if len(t.Name) > 50 {
		return fmt.Errorf("name cannot exceed 50 characters")
	}
	
	if t.Color != nil && len(*t.Color) > 0 {
		// Hex color validation
		color := *t.Color
		if len(color) != 7 || color[0] != '#' {
			return fmt.Errorf("color must be a valid hex color code (e.g., #FF0000)")
		}
		
		// Validate hex characters
		for i := 1; i < 7; i++ {
			c := color[i]
			if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')) {
				return fmt.Errorf("color must be a valid hex color code (e.g., #FF0000)")
			}
		}
	}
	
	return nil
}

// Utility methods

// AddMessage adds a message to the conversation and updates counts
func (c *Conversation) AddMessage(message Message) {
	c.Messages = append(c.Messages, message)
	if message.MessageType == MessageTypePrompt {
		c.PromptCount++
	}
	c.TotalCharacters += message.CharacterCount
	c.UpdatedAt = time.Now()
}

// AddRating adds a rating to the conversation
func (c *Conversation) AddRating(rating Rating) {
	c.Ratings = append(c.Ratings, rating)
}

// AddTag adds a tag to the conversation
func (c *Conversation) AddTag(tag Tag) {
	// Check if tag already exists
	for _, existingTag := range c.Tags {
		if existingTag.ID == tag.ID {
			return // Tag already exists
		}
	}
	c.Tags = append(c.Tags, tag)
}

// GetAverageRating calculates the average rating for the conversation
func (c *Conversation) GetAverageRating() *float64 {
	if len(c.Ratings) == 0 {
		return nil
	}
	
	var total int
	for _, rating := range c.Ratings {
		total += rating.Rating
	}
	
	avg := float64(total) / float64(len(c.Ratings))
	return &avg
}

// ToSummary converts a conversation to a summary
func (c *Conversation) ToSummary() ConversationSummary {
	responseCount := 0
	for _, msg := range c.Messages {
		if msg.MessageType == MessageTypeResponse {
			responseCount++
		}
	}
	
	return ConversationSummary{
		ID:              c.ID,
		SessionID:       c.SessionID,
		Title:           c.Title,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
		PromptCount:     c.PromptCount,
		ResponseCount:   responseCount,
		TotalCharacters: c.TotalCharacters,
		AvgRating:       c.GetAverageRating(),
		TagCount:        len(c.Tags),
		Tags:            c.Tags,
	}
}

// JSON serialization helpers

// MarshalToolCalls converts tool calls to JSON string for database storage
func MarshalToolCalls(toolCalls []ToolCall) (*string, error) {
	if len(toolCalls) == 0 {
		return nil, nil
	}
	
	data, err := json.Marshal(toolCalls)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tool calls: %w", err)
	}
	
	result := string(data)
	return &result, nil
}

// UnmarshalToolCalls parses JSON string from database into tool calls
func UnmarshalToolCalls(jsonStr *string) ([]ToolCall, error) {
	if jsonStr == nil || *jsonStr == "" {
		return nil, nil
	}
	
	var toolCalls []ToolCall
	err := json.Unmarshal([]byte(*jsonStr), &toolCalls)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool calls: %w", err)
	}
	
	return toolCalls, nil
}