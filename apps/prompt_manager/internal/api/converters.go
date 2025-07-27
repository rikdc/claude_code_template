package api

import (
	"fmt"

	"github.com/claude-code-template/prompt-manager/internal/database"
	"github.com/claude-code-template/prompt-manager/internal/models"
)

// ConvertConversation converts a database conversation to an API conversation model
func ConvertConversation(dbConv *database.Conversation) models.Conversation {
	return models.Conversation{
		ID:               dbConv.ID,
		SessionID:        dbConv.SessionID,
		Title:            dbConv.Title,
		CreatedAt:        dbConv.CreatedAt,
		UpdatedAt:        dbConv.UpdatedAt,
		PromptCount:      dbConv.PromptCount,
		TotalCharacters:  dbConv.TotalCharacters,
		WorkingDirectory: dbConv.WorkingDirectory,
		TranscriptPath:   dbConv.TranscriptPath,
	}
}

// ConvertConversationWithMessages converts a database conversation with messages to an API model
func ConvertConversationWithMessages(dbConv *database.ConversationWithMessages) (models.Conversation, error) {
	apiConv := ConvertConversation(&dbConv.Conversation)

	// Convert messages
	apiMessages := make([]models.Message, len(dbConv.Messages))
	for i := range dbConv.Messages {
		msg, err := ConvertMessage(&dbConv.Messages[i])
		if err != nil {
			return models.Conversation{}, fmt.Errorf("failed to convert message in conversation %d: %w", dbConv.ID, err)
		}
		apiMessages[i] = msg
	}
	apiConv.Messages = apiMessages

	return apiConv, nil
}

// ConvertMessage converts a database message to an API message model
func ConvertMessage(dbMsg *database.Message) (models.Message, error) {
	toolCalls, err := models.UnmarshalToolCalls(dbMsg.ToolCalls)
	if err != nil {
		return models.Message{}, fmt.Errorf("failed to unmarshal tool calls for message %d: %w", dbMsg.ID, err)
	}

	return models.Message{
		ID:             dbMsg.ID,
		ConversationID: dbMsg.ConversationID,
		MessageType:    models.MessageType(dbMsg.MessageType),
		Content:        dbMsg.Content,
		CharacterCount: dbMsg.CharacterCount,
		Timestamp:      dbMsg.Timestamp,
		ToolCalls:      toolCalls,
		ExecutionTime:  dbMsg.ExecutionTime,
	}, nil
}

// ConvertRating converts a database rating to an API rating model
func ConvertRating(dbRating *database.Rating) models.Rating {
	return models.Rating{
		ID:             dbRating.ID,
		ConversationID: dbRating.ConversationID,
		MessageID:      dbRating.MessageID,
		Rating:         dbRating.Rating,
		Comment:        dbRating.Comment,
		CreatedAt:      dbRating.CreatedAt,
		UpdatedAt:      dbRating.UpdatedAt,
	}
}

// ConvertConversationsToSummaries converts multiple database conversations to API conversation summaries
func ConvertConversationsToSummaries(dbConversations []database.Conversation) []models.ConversationSummary {
	summaries := make([]models.ConversationSummary, len(dbConversations))
	for i := range dbConversations {
		// Convert to API model first to use the ToSummary method
		apiConv := ConvertConversation(&dbConversations[i])
		summaries[i] = apiConv.ToSummary()
	}
	return summaries
}

// ConvertRatings converts multiple database ratings to API rating models
func ConvertRatings(dbRatings []database.Rating) []models.Rating {
	apiRatings := make([]models.Rating, len(dbRatings))
	for i := range dbRatings {
		apiRatings[i] = ConvertRating(&dbRatings[i])
	}
	return apiRatings
}

