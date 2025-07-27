package models

import (
	"testing"
	"time"
)

func TestConversationValidation(t *testing.T) {
	tests := []struct {
		name      string
		conv      Conversation
		wantError bool
	}{
		{
			name: "valid conversation",
			conv: Conversation{
				SessionID:       "test-session",
				PromptCount:     1,
				TotalCharacters: 100,
			},
			wantError: false,
		},
		{
			name: "missing session ID",
			conv: Conversation{
				PromptCount:     1,
				TotalCharacters: 100,
			},
			wantError: true,
		},
		{
			name: "negative prompt count",
			conv: Conversation{
				SessionID:       "test-session",
				PromptCount:     -1,
				TotalCharacters: 100,
			},
			wantError: true,
		},
		{
			name: "negative total characters",
			conv: Conversation{
				SessionID:       "test-session",
				PromptCount:     1,
				TotalCharacters: -100,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.conv.Validate()
			if tt.wantError && err == nil {
				t.Error("expected validation error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestMessageValidation(t *testing.T) {
	tests := []struct {
		name      string
		msg       Message
		wantError bool
	}{
		{
			name: "valid prompt message",
			msg: Message{
				ConversationID: 1,
				MessageType:    MessageTypePrompt,
				Content:        "Hello world",
				CharacterCount: 11,
			},
			wantError: false,
		},
		{
			name: "valid response message",
			msg: Message{
				ConversationID: 1,
				MessageType:    MessageTypeResponse,
				Content:        "Hi there",
				CharacterCount: 8,
			},
			wantError: false,
		},
		{
			name: "missing conversation ID",
			msg: Message{
				MessageType:    MessageTypePrompt,
				Content:        "Hello",
				CharacterCount: 5,
			},
			wantError: true,
		},
		{
			name: "invalid message type",
			msg: Message{
				ConversationID: 1,
				MessageType:    MessageType("invalid"),
				Content:        "Hello",
				CharacterCount: 5,
			},
			wantError: true,
		},
		{
			name: "empty content",
			msg: Message{
				ConversationID: 1,
				MessageType:    MessageTypePrompt,
				Content:        "",
				CharacterCount: 0,
			},
			wantError: true,
		},
		{
			name: "character count mismatch",
			msg: Message{
				ConversationID: 1,
				MessageType:    MessageTypePrompt,
				Content:        "Hello",
				CharacterCount: 10, // Should be 5
			},
			wantError: true,
		},
		{
			name: "negative execution time",
			msg: Message{
				ConversationID: 1,
				MessageType:    MessageTypeResponse,
				Content:        "Hello",
				CharacterCount: 5,
				ExecutionTime:  intPtr(-100),
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.Validate()
			if tt.wantError && err == nil {
				t.Error("expected validation error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestRatingValidation(t *testing.T) {
	tests := []struct {
		name      string
		rating    Rating
		wantError bool
	}{
		{
			name: "valid conversation rating",
			rating: Rating{
				ConversationID: intPtr(1),
				Rating:         5,
			},
			wantError: false,
		},
		{
			name: "valid message rating",
			rating: Rating{
				MessageID: intPtr(1),
				Rating:    3,
			},
			wantError: false,
		},
		{
			name: "missing both IDs",
			rating: Rating{
				Rating: 5,
			},
			wantError: true,
		},
		{
			name: "both IDs specified",
			rating: Rating{
				ConversationID: intPtr(1),
				MessageID:      intPtr(1),
				Rating:         5,
			},
			wantError: true,
		},
		{
			name: "rating too low",
			rating: Rating{
				ConversationID: intPtr(1),
				Rating:         0,
			},
			wantError: true,
		},
		{
			name: "rating too high",
			rating: Rating{
				ConversationID: intPtr(1),
				Rating:         6,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rating.Validate()
			if tt.wantError && err == nil {
				t.Error("expected validation error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestTagValidation(t *testing.T) {
	tests := []struct {
		name      string
		tag       Tag
		wantError bool
	}{
		{
			name: "valid tag",
			tag: Tag{
				Name: "bug",
			},
			wantError: false,
		},
		{
			name: "valid tag with color",
			tag: Tag{
				Name:  "feature",
				Color: stringPtr("#FF0000"),
			},
			wantError: false,
		},
		{
			name: "empty name",
			tag: Tag{
				Name: "",
			},
			wantError: true,
		},
		{
			name: "name too long",
			tag: Tag{
				Name: "this_is_a_very_long_tag_name_that_exceeds_fifty_characters_limit",
			},
			wantError: true,
		},
		{
			name: "invalid color format",
			tag: Tag{
				Name:  "test",
				Color: stringPtr("red"),
			},
			wantError: true,
		},
		{
			name: "invalid hex color",
			tag: Tag{
				Name:  "test",
				Color: stringPtr("#GGGGGG"),
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tag.Validate()
			if tt.wantError && err == nil {
				t.Error("expected validation error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestConversationMethods(t *testing.T) {
	conv := &Conversation{
		ID:        1,
		SessionID: "test-session",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test AddMessage
	msg1 := Message{
		MessageType:    MessageTypePrompt,
		Content:        "Hello",
		CharacterCount: 5,
	}
	
	msg2 := Message{
		MessageType:    MessageTypeResponse,
		Content:        "Hi there",
		CharacterCount: 8,
	}

	conv.AddMessage(msg1)
	if conv.PromptCount != 1 {
		t.Errorf("Expected prompt count 1, got %d", conv.PromptCount)
	}
	if conv.TotalCharacters != 5 {
		t.Errorf("Expected total characters 5, got %d", conv.TotalCharacters)
	}

	conv.AddMessage(msg2)
	if conv.PromptCount != 1 { // Response shouldn't increment prompt count
		t.Errorf("Expected prompt count 1, got %d", conv.PromptCount)
	}
	if conv.TotalCharacters != 13 {
		t.Errorf("Expected total characters 13, got %d", conv.TotalCharacters)
	}

	// Test AddRating
	rating := Rating{
		Rating: 5,
	}
	conv.AddRating(rating)
	if len(conv.Ratings) != 1 {
		t.Errorf("Expected 1 rating, got %d", len(conv.Ratings))
	}

	// Test GetAverageRating
	avgRating := conv.GetAverageRating()
	if avgRating == nil || *avgRating != 5.0 {
		t.Errorf("Expected average rating 5.0, got %v", avgRating)
	}

	// Add another rating
	conv.AddRating(Rating{Rating: 3})
	avgRating = conv.GetAverageRating()
	if avgRating == nil || *avgRating != 4.0 {
		t.Errorf("Expected average rating 4.0, got %v", avgRating)
	}

	// Test AddTag
	tag := Tag{ID: 1, Name: "test"}
	conv.AddTag(tag)
	if len(conv.Tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(conv.Tags))
	}

	// Try adding same tag again (should not duplicate)
	conv.AddTag(tag)
	if len(conv.Tags) != 1 {
		t.Errorf("Expected 1 tag after duplicate add, got %d", len(conv.Tags))
	}

	// Test ToSummary
	summary := conv.ToSummary()
	if summary.ID != conv.ID {
		t.Errorf("Expected summary ID %d, got %d", conv.ID, summary.ID)
	}
	if summary.PromptCount != 1 {
		t.Errorf("Expected summary prompt count 1, got %d", summary.PromptCount)
	}
	if summary.ResponseCount != 1 {
		t.Errorf("Expected summary response count 1, got %d", summary.ResponseCount)
	}
	if summary.AvgRating == nil || *summary.AvgRating != 4.0 {
		t.Errorf("Expected summary avg rating 4.0, got %v", summary.AvgRating)
	}
}

func TestToolCallsSerialization(t *testing.T) {
	toolCalls := []ToolCall{
		{
			Name: "test_tool",
			Arguments: map[string]interface{}{
				"param1": "value1",
				"param2": 42,
			},
			Duration: intPtr(100),
		},
		{
			Name: "another_tool",
			Arguments: map[string]interface{}{
				"key": "value",
			},
		},
	}

	// Test marshaling
	jsonStr, err := MarshalToolCalls(toolCalls)
	if err != nil {
		t.Fatalf("Failed to marshal tool calls: %v", err)
	}

	if jsonStr == nil {
		t.Fatal("Expected non-nil JSON string")
	}

	// Test unmarshaling
	unmarshaled, err := UnmarshalToolCalls(jsonStr)
	if err != nil {
		t.Fatalf("Failed to unmarshal tool calls: %v", err)
	}

	if len(unmarshaled) != 2 {
		t.Errorf("Expected 2 tool calls, got %d", len(unmarshaled))
	}

	if unmarshaled[0].Name != "test_tool" {
		t.Errorf("Expected first tool name 'test_tool', got '%s'", unmarshaled[0].Name)
	}

	// Test empty tool calls
	emptyJson, err := MarshalToolCalls([]ToolCall{})
	if err != nil {
		t.Fatalf("Failed to marshal empty tool calls: %v", err)
	}

	if emptyJson != nil {
		t.Error("Expected nil for empty tool calls")
	}

	// Test nil unmarshaling
	emptyUnmarshaled, err := UnmarshalToolCalls(nil)
	if err != nil {
		t.Fatalf("Failed to unmarshal nil tool calls: %v", err)
	}

	if emptyUnmarshaled != nil {
		t.Error("Expected nil for unmarshaling nil string")
	}
}

// Helper functions for test pointers
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}