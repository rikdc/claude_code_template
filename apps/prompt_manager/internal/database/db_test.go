package database

import (
	"os"
	"testing"
)

func setupTestDB(t *testing.T) *DB {
	// Create temp database file
	tmpfile, err := os.CreateTemp("", "test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()

	config := &Config{
		DatabasePath:  tmpfile.Name(),
		MigrationsDir: "../../database/migrations",
	}

	db, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Run migrations
	err = db.RunMigrations(config.MigrationsDir)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Cleanup function
	t.Cleanup(func() {
		db.Close()
		os.Remove(tmpfile.Name())
	})

	return db
}

func TestDatabaseConnection(t *testing.T) {
	db := setupTestDB(t)

	// Test health check
	err := db.Health()
	if err != nil {
		t.Errorf("Database health check failed: %v", err)
	}

	// Test stats
	stats, err := db.Stats()
	if err != nil {
		t.Errorf("Failed to get database stats: %v", err)
	}

	if stats["conversations"] != 0 {
		t.Errorf("Expected 0 conversations, got %v", stats["conversations"])
	}
}

func TestConversationCRUD(t *testing.T) {
	db := setupTestDB(t)

	sessionID := "test-session-123"
	title := "Test Conversation"
	workingDir := "/tmp/test"

	// Create conversation
	conv, err := db.CreateConversation(sessionID, &title, &workingDir, nil)
	if err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	if conv.SessionID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, conv.SessionID)
	}

	if conv.Title == nil || *conv.Title != title {
		t.Errorf("Expected title %s, got %v", title, conv.Title)
	}

	// Get conversation
	retrieved, err := db.GetConversation(conv.ID)
	if err != nil {
		t.Fatalf("Failed to get conversation: %v", err)
	}

	if retrieved.ID != conv.ID {
		t.Errorf("Expected ID %d, got %d", conv.ID, retrieved.ID)
	}

	// Update conversation title
	newTitle := "Updated Test Conversation"
	err = db.UpdateConversationTitle(conv.ID, newTitle)
	if err != nil {
		t.Fatalf("Failed to update conversation title: %v", err)
	}

	// Verify update
	updated, err := db.GetConversation(conv.ID)
	if err != nil {
		t.Fatalf("Failed to get updated conversation: %v", err)
	}

	if updated.Title == nil || *updated.Title != newTitle {
		t.Errorf("Expected updated title %s, got %v", newTitle, updated.Title)
	}

	// List conversations
	conversations, err := db.ListConversations(10, 0)
	if err != nil {
		t.Fatalf("Failed to list conversations: %v", err)
	}

	if len(conversations) != 1 {
		t.Errorf("Expected 1 conversation, got %d", len(conversations))
	}

	// Delete conversation
	err = db.DeleteConversation(conv.ID)
	if err != nil {
		t.Fatalf("Failed to delete conversation: %v", err)
	}

	// Verify deletion
	_, err = db.GetConversation(conv.ID)
	if err == nil {
		t.Error("Expected error when getting deleted conversation")
	}
}

func TestMessageCRUD(t *testing.T) {
	db := setupTestDB(t)

	// Create conversation first
	conv, err := db.CreateConversation("test-session", nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	// Create message
	content := "Hello, this is a test message"
	toolCalls := `[{"name": "test_tool", "args": {}}]`
	executionTime := 100

	msg, err := db.CreateMessage(conv.ID, "prompt", content, &toolCalls, &executionTime)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	if msg.Content != content {
		t.Errorf("Expected content %s, got %s", content, msg.Content)
	}

	if msg.CharacterCount != len(content) {
		t.Errorf("Expected character count %d, got %d", len(content), msg.CharacterCount)
	}

	// Get message
	retrieved, err := db.GetMessage(msg.ID)
	if err != nil {
		t.Fatalf("Failed to get message: %v", err)
	}

	if retrieved.ID != msg.ID {
		t.Errorf("Expected ID %d, got %d", msg.ID, retrieved.ID)
	}

	// Get messages by conversation
	messages, err := db.GetMessagesByConversation(conv.ID)
	if err != nil {
		t.Fatalf("Failed to get messages by conversation: %v", err)
	}

	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	// Get conversation with messages
	convWithMessages, err := db.GetConversationWithMessages(conv.ID)
	if err != nil {
		t.Fatalf("Failed to get conversation with messages: %v", err)
	}

	if len(convWithMessages.Messages) != 1 {
		t.Errorf("Expected 1 message in conversation, got %d", len(convWithMessages.Messages))
	}

	// Verify conversation stats were updated
	if convWithMessages.PromptCount != 1 {
		t.Errorf("Expected prompt count 1, got %d", convWithMessages.PromptCount)
	}

	if convWithMessages.TotalCharacters != len(content) {
		t.Errorf("Expected total characters %d, got %d", len(content), convWithMessages.TotalCharacters)
	}
}

func TestRatingCRUD(t *testing.T) {
	db := setupTestDB(t)

	// Create conversation and message
	conv, err := db.CreateConversation("test-session", nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	msg, err := db.CreateMessage(conv.ID, "prompt", "test message", nil, nil)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	// Create conversation rating
	comment := "Great conversation"
	convRating, err := db.CreateConversationRating(conv.ID, 5, &comment)
	if err != nil {
		t.Fatalf("Failed to create conversation rating: %v", err)
	}

	if convRating.Rating != 5 {
		t.Errorf("Expected rating 5, got %d", convRating.Rating)
	}

	// Create message rating
	msgRating, err := db.CreateMessageRating(msg.ID, 4, nil)
	if err != nil {
		t.Fatalf("Failed to create message rating: %v", err)
	}

	if msgRating.Rating != 4 {
		t.Errorf("Expected rating 4, got %d", msgRating.Rating)
	}

	// Get ratings
	convRatings, err := db.GetConversationRatings(conv.ID)
	if err != nil {
		t.Fatalf("Failed to get conversation ratings: %v", err)
	}

	if len(convRatings) != 1 {
		t.Errorf("Expected 1 conversation rating, got %d", len(convRatings))
	}

	msgRatings, err := db.GetMessageRatings(msg.ID)
	if err != nil {
		t.Fatalf("Failed to get message ratings: %v", err)
	}

	if len(msgRatings) != 1 {
		t.Errorf("Expected 1 message rating, got %d", len(msgRatings))
	}

	// Update rating
	newComment := "Updated comment"
	err = db.UpdateRating(convRating.ID, 3, &newComment)
	if err != nil {
		t.Fatalf("Failed to update rating: %v", err)
	}

	// Verify update
	updated, err := db.GetRating(convRating.ID)
	if err != nil {
		t.Fatalf("Failed to get updated rating: %v", err)
	}

	if updated.Rating != 3 {
		t.Errorf("Expected updated rating 3, got %d", updated.Rating)
	}

	if updated.Comment == nil || *updated.Comment != newComment {
		t.Errorf("Expected updated comment %s, got %v", newComment, updated.Comment)
	}

	// Test rating stats
	stats, err := db.GetRatingStats()
	if err != nil {
		t.Fatalf("Failed to get rating stats: %v", err)
	}

	if stats["total_ratings"] != 2 {
		t.Errorf("Expected 2 total ratings, got %v", stats["total_ratings"])
	}

	// Delete rating
	err = db.DeleteRating(msgRating.ID)
	if err != nil {
		t.Fatalf("Failed to delete rating: %v", err)
	}

	// Verify deletion
	_, err = db.GetRating(msgRating.ID)
	if err == nil {
		t.Error("Expected error when getting deleted rating")
	}
}

func TestInvalidRating(t *testing.T) {
	db := setupTestDB(t)

	// Create conversation
	conv, err := db.CreateConversation("test-session", nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	// Test invalid rating values
	_, err = db.CreateConversationRating(conv.ID, 0, nil)
	if err == nil {
		t.Error("Expected error for rating 0")
	}

	_, err = db.CreateConversationRating(conv.ID, 6, nil)
	if err == nil {
		t.Error("Expected error for rating 6")
	}
}