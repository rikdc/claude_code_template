package api

import (
	"testing"
	"time"

	"github.com/claude-code-template/prompt-manager/internal/database"
	"github.com/claude-code-template/prompt-manager/internal/models"
)

func TestConvertConversation(t *testing.T) {
	now := time.Now()
	title := "Test Conversation"
	workingDir := "/test/dir"
	transcriptPath := "/test/transcript.md"

	dbConv := &database.Conversation{
		ID:               1,
		SessionID:        "session-123",
		Title:            &title,
		CreatedAt:        now,
		UpdatedAt:        now,
		PromptCount:      5,
		TotalCharacters:  150,
		WorkingDirectory: &workingDir,
		TranscriptPath:   &transcriptPath,
	}

	apiConv := ConvertConversation(dbConv)

	// Verify all fields are converted correctly
	if apiConv.ID != dbConv.ID {
		t.Errorf("Expected ID %d, got %d", dbConv.ID, apiConv.ID)
	}
	if apiConv.SessionID != dbConv.SessionID {
		t.Errorf("Expected SessionID %s, got %s", dbConv.SessionID, apiConv.SessionID)
	}
	if apiConv.Title == nil || *apiConv.Title != *dbConv.Title {
		t.Errorf("Expected Title %v, got %v", dbConv.Title, apiConv.Title)
	}
	if !apiConv.CreatedAt.Equal(dbConv.CreatedAt) {
		t.Errorf("Expected CreatedAt %v, got %v", dbConv.CreatedAt, apiConv.CreatedAt)
	}
	if !apiConv.UpdatedAt.Equal(dbConv.UpdatedAt) {
		t.Errorf("Expected UpdatedAt %v, got %v", dbConv.UpdatedAt, apiConv.UpdatedAt)
	}
	if apiConv.PromptCount != dbConv.PromptCount {
		t.Errorf("Expected PromptCount %d, got %d", dbConv.PromptCount, apiConv.PromptCount)
	}
	if apiConv.TotalCharacters != dbConv.TotalCharacters {
		t.Errorf("Expected TotalCharacters %d, got %d", dbConv.TotalCharacters, apiConv.TotalCharacters)
	}
	if apiConv.WorkingDirectory == nil || *apiConv.WorkingDirectory != *dbConv.WorkingDirectory {
		t.Errorf("Expected WorkingDirectory %v, got %v", dbConv.WorkingDirectory, apiConv.WorkingDirectory)
	}
	if apiConv.TranscriptPath == nil || *apiConv.TranscriptPath != *dbConv.TranscriptPath {
		t.Errorf("Expected TranscriptPath %v, got %v", dbConv.TranscriptPath, apiConv.TranscriptPath)
	}
}

func TestConvertMessage(t *testing.T) {
	tests := []struct {
		name          string
		dbMsg         *database.Message
		expectError   bool
		expectedError string
		validateMsg   func(t *testing.T, msg models.Message)
	}{
		{
			name: "successful conversion with tool calls",
			dbMsg: &database.Message{
				ID:             1,
				ConversationID: 1,
				MessageType:    "prompt",
				Content:        "Test message content",
				CharacterCount: 20,
				Timestamp:      time.Now(),
				ToolCalls:      stringPtr(`[{"name": "test_tool", "arguments": {"key": "value"}}]`),
				ExecutionTime:  intPtr(150),
			},
			expectError: false,
			validateMsg: func(t *testing.T, msg models.Message) {
				if msg.ID != 1 {
					t.Errorf("Expected ID 1, got %d", msg.ID)
				}
				if len(msg.ToolCalls) != 1 {
					t.Errorf("Expected 1 tool call, got %d", len(msg.ToolCalls))
				}
				if len(msg.ToolCalls) > 0 && msg.ToolCalls[0].Name != "test_tool" {
					t.Errorf("Expected tool call name 'test_tool', got %s", msg.ToolCalls[0].Name)
				}
			},
		},
		{
			name: "successful conversion with nil tool calls",
			dbMsg: &database.Message{
				ID:             2,
				ConversationID: 1,
				MessageType:    "response",
				Content:        "Response content",
				CharacterCount: 16,
				Timestamp:      time.Now(),
				ToolCalls:      nil,
				ExecutionTime:  nil,
			},
			expectError: false,
			validateMsg: func(t *testing.T, msg models.Message) {
				if msg.ID != 2 {
					t.Errorf("Expected ID 2, got %d", msg.ID)
				}
				if len(msg.ToolCalls) != 0 {
					t.Errorf("Expected 0 tool calls, got %d", len(msg.ToolCalls))
				}
			},
		},
		{
			name: "error on malformed tool calls JSON",
			dbMsg: &database.Message{
				ID:             3,
				ConversationID: 1,
				MessageType:    "prompt",
				Content:        "Test content",
				CharacterCount: 12,
				Timestamp:      time.Now(),
				ToolCalls:      stringPtr(`{"invalid": "json"}`), // Invalid: not an array
				ExecutionTime:  nil,
			},
			expectError:   true,
			expectedError: "failed to unmarshal tool calls for message 3",
		},
		{
			name: "error on completely invalid JSON",
			dbMsg: &database.Message{
				ID:             4,
				ConversationID: 1,
				MessageType:    "prompt",
				Content:        "Test content",
				CharacterCount: 12,
				Timestamp:      time.Now(),
				ToolCalls:      stringPtr(`{invalid json`),
				ExecutionTime:  nil,
			},
			expectError:   true,
			expectedError: "failed to unmarshal tool calls for message 4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiMsg, err := ConvertMessage(tt.dbMsg)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, but got none")
				} else if tt.expectedError != "" && !containsString(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing %q, got %q", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				} else if tt.validateMsg != nil {
					tt.validateMsg(t, apiMsg)
				}

				// Verify basic fields for successful conversions
				if apiMsg.ConversationID != tt.dbMsg.ConversationID {
					t.Errorf("Expected ConversationID %d, got %d", tt.dbMsg.ConversationID, apiMsg.ConversationID)
				}
				if string(apiMsg.MessageType) != tt.dbMsg.MessageType {
					t.Errorf("Expected MessageType %s, got %s", tt.dbMsg.MessageType, string(apiMsg.MessageType))
				}
				if apiMsg.Content != tt.dbMsg.Content {
					t.Errorf("Expected Content %s, got %s", tt.dbMsg.Content, apiMsg.Content)
				}
			}
		})
	}
}

// Helper functions for tests
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestConvertRating(t *testing.T) {
	now := time.Now()
	conversationID := 1
	comment := "Great conversation"

	dbRating := &database.Rating{
		ID:             1,
		ConversationID: &conversationID,
		MessageID:      nil,
		Rating:         5,
		Comment:        &comment,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	apiRating := ConvertRating(dbRating)

	// Verify all fields are converted correctly
	if apiRating.ID != dbRating.ID {
		t.Errorf("Expected ID %d, got %d", dbRating.ID, apiRating.ID)
	}
	if apiRating.ConversationID == nil || *apiRating.ConversationID != *dbRating.ConversationID {
		t.Errorf("Expected ConversationID %v, got %v", dbRating.ConversationID, apiRating.ConversationID)
	}
	if apiRating.MessageID != nil {
		t.Errorf("Expected MessageID to be nil, got %v", apiRating.MessageID)
	}
	if apiRating.Rating != dbRating.Rating {
		t.Errorf("Expected Rating %d, got %d", dbRating.Rating, apiRating.Rating)
	}
	if apiRating.Comment == nil || *apiRating.Comment != *dbRating.Comment {
		t.Errorf("Expected Comment %v, got %v", dbRating.Comment, apiRating.Comment)
	}
	if !apiRating.CreatedAt.Equal(dbRating.CreatedAt) {
		t.Errorf("Expected CreatedAt %v, got %v", dbRating.CreatedAt, apiRating.CreatedAt)
	}
	if !apiRating.UpdatedAt.Equal(dbRating.UpdatedAt) {
		t.Errorf("Expected UpdatedAt %v, got %v", dbRating.UpdatedAt, apiRating.UpdatedAt)
	}
}

func TestConvertConversationWithMessages(t *testing.T) {
	tests := []struct {
		name         string
		dbConv       *database.ConversationWithMessages
		expectError  bool
		validateConv func(t *testing.T, conv models.Conversation)
	}{
		{
			name: "successful conversion with valid messages",
			dbConv: &database.ConversationWithMessages{
				Conversation: database.Conversation{
					ID:               1,
					SessionID:        "session-123",
					Title:            stringPtr("Test Conversation"),
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
					PromptCount:      1,
					TotalCharacters:  20,
					WorkingDirectory: nil,
					TranscriptPath:   nil,
				},
				Messages: []database.Message{
					{
						ID:             1,
						ConversationID: 1,
						MessageType:    "prompt",
						Content:        "Test prompt",
						CharacterCount: 11,
						Timestamp:      time.Now(),
						ToolCalls:      nil,
						ExecutionTime:  nil,
					},
					{
						ID:             2,
						ConversationID: 1,
						MessageType:    "response",
						Content:        "Test response",
						CharacterCount: 13,
						Timestamp:      time.Now().Add(time.Second),
						ToolCalls:      nil,
						ExecutionTime:  nil,
					},
				},
			},
			expectError: false,
			validateConv: func(t *testing.T, conv models.Conversation) {
				if conv.ID != 1 {
					t.Errorf("Expected ID 1, got %d", conv.ID)
				}
				if len(conv.Messages) != 2 {
					t.Errorf("Expected 2 messages, got %d", len(conv.Messages))
				}
			},
		},
		{
			name: "error when message has malformed tool calls",
			dbConv: &database.ConversationWithMessages{
				Conversation: database.Conversation{
					ID:               2,
					SessionID:        "session-456",
					Title:            stringPtr("Error Test"),
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
					PromptCount:      1,
					TotalCharacters:  10,
					WorkingDirectory: nil,
					TranscriptPath:   nil,
				},
				Messages: []database.Message{
					{
						ID:             3,
						ConversationID: 2,
						MessageType:    "prompt",
						Content:        "Test prompt",
						CharacterCount: 11,
						Timestamp:      time.Now(),
						ToolCalls:      stringPtr(`{invalid json`), // Malformed JSON
						ExecutionTime:  nil,
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiConv, err := ConvertConversationWithMessages(tt.dbConv)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				} else {
					// Verify basic fields
					if apiConv.ID != tt.dbConv.ID {
						t.Errorf("Expected ID %d, got %d", tt.dbConv.ID, apiConv.ID)
					}
					if apiConv.SessionID != tt.dbConv.SessionID {
						t.Errorf("Expected SessionID %s, got %s", tt.dbConv.SessionID, apiConv.SessionID)
					}

					if tt.validateConv != nil {
						tt.validateConv(t, apiConv)
					}
				}
			}
		})
	}
}

func TestConvertConversationsToSummaries(t *testing.T) {
	now := time.Now()
	title1 := "First Conversation"
	title2 := "Second Conversation"

	dbConversations := []database.Conversation{
		{
			ID:               1,
			SessionID:        "session-1",
			Title:            &title1,
			CreatedAt:        now,
			UpdatedAt:        now,
			PromptCount:      3,
			TotalCharacters:  100,
			WorkingDirectory: nil,
			TranscriptPath:   nil,
		},
		{
			ID:               2,
			SessionID:        "session-2",
			Title:            &title2,
			CreatedAt:        now.Add(time.Hour),
			UpdatedAt:        now.Add(time.Hour),
			PromptCount:      5,
			TotalCharacters:  200,
			WorkingDirectory: nil,
			TranscriptPath:   nil,
		},
	}

	summaries := ConvertConversationsToSummaries(dbConversations)

	if len(summaries) != len(dbConversations) {
		t.Errorf("Expected %d summaries, got %d", len(dbConversations), len(summaries))
	}

	for i, summary := range summaries {
		expected := dbConversations[i]
		if summary.ID != expected.ID {
			t.Errorf("Summary %d: Expected ID %d, got %d", i, expected.ID, summary.ID)
		}
		if summary.SessionID != expected.SessionID {
			t.Errorf("Summary %d: Expected SessionID %s, got %s", i, expected.SessionID, summary.SessionID)
		}
		if summary.PromptCount != expected.PromptCount {
			t.Errorf("Summary %d: Expected PromptCount %d, got %d", i, expected.PromptCount, summary.PromptCount)
		}
		if summary.TotalCharacters != expected.TotalCharacters {
			t.Errorf("Summary %d: Expected TotalCharacters %d, got %d", i, expected.TotalCharacters, summary.TotalCharacters)
		}
	}
}

func TestConvertRatings(t *testing.T) {
	now := time.Now()
	conversationID := 1
	messageID := 2
	comment1 := "Good"
	comment2 := "Excellent"

	dbRatings := []database.Rating{
		{
			ID:             1,
			ConversationID: &conversationID,
			MessageID:      nil,
			Rating:         4,
			Comment:        &comment1,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             2,
			ConversationID: nil,
			MessageID:      &messageID,
			Rating:         5,
			Comment:        &comment2,
			CreatedAt:      now.Add(time.Hour),
			UpdatedAt:      now.Add(time.Hour),
		},
	}

	apiRatings := ConvertRatings(dbRatings)

	if len(apiRatings) != len(dbRatings) {
		t.Errorf("Expected %d ratings, got %d", len(dbRatings), len(apiRatings))
	}

	for i, rating := range apiRatings {
		expected := dbRatings[i]
		if rating.ID != expected.ID {
			t.Errorf("Rating %d: Expected ID %d, got %d", i, expected.ID, rating.ID)
		}
		if rating.Rating != expected.Rating {
			t.Errorf("Rating %d: Expected Rating %d, got %d", i, expected.Rating, rating.Rating)
		}

		// Check conversation ID
		if (rating.ConversationID == nil) != (expected.ConversationID == nil) {
			t.Errorf("Rating %d: ConversationID null mismatch", i)
		} else if rating.ConversationID != nil && *rating.ConversationID != *expected.ConversationID {
			t.Errorf("Rating %d: Expected ConversationID %d, got %d", i, *expected.ConversationID, *rating.ConversationID)
		}

		// Check message ID
		if (rating.MessageID == nil) != (expected.MessageID == nil) {
			t.Errorf("Rating %d: MessageID null mismatch", i)
		} else if rating.MessageID != nil && *rating.MessageID != *expected.MessageID {
			t.Errorf("Rating %d: Expected MessageID %d, got %d", i, *expected.MessageID, *rating.MessageID)
		}
	}
}

