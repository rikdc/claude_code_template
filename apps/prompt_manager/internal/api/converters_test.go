package api

import (
	"testing"
	"time"

	"github.com/claude-code-template/prompt-manager/internal/database"
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
	now := time.Now()
	toolCallsJSON := `[{"name": "test_tool", "arguments": {"key": "value"}}]`
	executionTime := 150

	dbMsg := &database.Message{
		ID:             1,
		ConversationID: 1,
		MessageType:    "prompt",
		Content:        "Test message content",
		CharacterCount: 20,
		Timestamp:      now,
		ToolCalls:      &toolCallsJSON,
		ExecutionTime:  &executionTime,
	}

	apiMsg := ConvertMessage(dbMsg)

	// Verify all fields are converted correctly
	if apiMsg.ID != dbMsg.ID {
		t.Errorf("Expected ID %d, got %d", dbMsg.ID, apiMsg.ID)
	}
	if apiMsg.ConversationID != dbMsg.ConversationID {
		t.Errorf("Expected ConversationID %d, got %d", dbMsg.ConversationID, apiMsg.ConversationID)
	}
	if string(apiMsg.MessageType) != dbMsg.MessageType {
		t.Errorf("Expected MessageType %s, got %s", dbMsg.MessageType, string(apiMsg.MessageType))
	}
	if apiMsg.Content != dbMsg.Content {
		t.Errorf("Expected Content %s, got %s", dbMsg.Content, apiMsg.Content)
	}
	if apiMsg.CharacterCount != dbMsg.CharacterCount {
		t.Errorf("Expected CharacterCount %d, got %d", dbMsg.CharacterCount, apiMsg.CharacterCount)
	}
	if !apiMsg.Timestamp.Equal(dbMsg.Timestamp) {
		t.Errorf("Expected Timestamp %v, got %v", dbMsg.Timestamp, apiMsg.Timestamp)
	}
	if apiMsg.ExecutionTime == nil || *apiMsg.ExecutionTime != *dbMsg.ExecutionTime {
		t.Errorf("Expected ExecutionTime %v, got %v", dbMsg.ExecutionTime, apiMsg.ExecutionTime)
	}
	// Tool calls should be unmarshaled
	if len(apiMsg.ToolCalls) != 1 {
		t.Errorf("Expected 1 tool call, got %d", len(apiMsg.ToolCalls))
	}
	if len(apiMsg.ToolCalls) > 0 && apiMsg.ToolCalls[0].Name != "test_tool" {
		t.Errorf("Expected tool call name 'test_tool', got %s", apiMsg.ToolCalls[0].Name)
	}
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
	now := time.Now()
	title := "Test Conversation"

	dbConv := &database.ConversationWithMessages{
		Conversation: database.Conversation{
			ID:               1,
			SessionID:        "session-123",
			Title:            &title,
			CreatedAt:        now,
			UpdatedAt:        now,
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
				Timestamp:      now,
				ToolCalls:      nil,
				ExecutionTime:  nil,
			},
			{
				ID:             2,
				ConversationID: 1,
				MessageType:    "response",
				Content:        "Test response",
				CharacterCount: 13,
				Timestamp:      now.Add(time.Second),
				ToolCalls:      nil,
				ExecutionTime:  nil,
			},
		},
	}

	apiConv := ConvertConversationWithMessages(dbConv)

	// Verify conversation fields
	if apiConv.ID != dbConv.ID {
		t.Errorf("Expected ID %d, got %d", dbConv.ID, apiConv.ID)
	}
	if apiConv.SessionID != dbConv.SessionID {
		t.Errorf("Expected SessionID %s, got %s", dbConv.SessionID, apiConv.SessionID)
	}

	// Verify messages are converted
	if len(apiConv.Messages) != len(dbConv.Messages) {
		t.Errorf("Expected %d messages, got %d", len(dbConv.Messages), len(apiConv.Messages))
	}

	for i, msg := range apiConv.Messages {
		expectedMsg := dbConv.Messages[i]
		if msg.ID != expectedMsg.ID {
			t.Errorf("Message %d: Expected ID %d, got %d", i, expectedMsg.ID, msg.ID)
		}
		if msg.Content != expectedMsg.Content {
			t.Errorf("Message %d: Expected Content %s, got %s", i, expectedMsg.Content, msg.Content)
		}
		if string(msg.MessageType) != expectedMsg.MessageType {
			t.Errorf("Message %d: Expected MessageType %s, got %s", i, expectedMsg.MessageType, string(msg.MessageType))
		}
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