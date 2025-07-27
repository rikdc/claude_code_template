package database

import (
	"database/sql"
	"fmt"
	"time"
)

// Conversation represents a conversation record
type Conversation struct {
	ID               int       `json:"id"`
	SessionID        string    `json:"session_id"`
	Title            *string   `json:"title"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	PromptCount      int       `json:"prompt_count"`
	TotalCharacters  int       `json:"total_characters"`
	WorkingDirectory *string   `json:"working_directory"`
	TranscriptPath   *string   `json:"transcript_path"`
}

// Message represents a message record
type Message struct {
	ID             int       `json:"id"`
	ConversationID int       `json:"conversation_id"`
	MessageType    string    `json:"message_type"` // 'prompt' or 'response'
	Content        string    `json:"content"`
	CharacterCount int       `json:"character_count"`
	Timestamp      time.Time `json:"timestamp"`
	ToolCalls      *string   `json:"tool_calls"`
	ExecutionTime  *int      `json:"execution_time"`
}

// ConversationWithMessages includes messages in the conversation
type ConversationWithMessages struct {
	Conversation
	Messages []Message `json:"messages"`
}

// CreateConversation inserts a new conversation
func (db *DB) CreateConversation(sessionID string, title *string, workingDir *string, transcriptPath *string) (*Conversation, error) {
	query := `
	INSERT INTO conversations (session_id, title, working_directory, transcript_path)
	VALUES (?, ?, ?, ?)
	RETURNING id, session_id, title, created_at, updated_at, prompt_count, total_characters, working_directory, transcript_path`

	var conv Conversation
	err := db.conn.QueryRow(query, sessionID, title, workingDir, transcriptPath).Scan(
		&conv.ID, &conv.SessionID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt,
		&conv.PromptCount, &conv.TotalCharacters, &conv.WorkingDirectory, &conv.TranscriptPath,
	)
	
	if err != nil {
		// Fallback for SQLite versions that don't support RETURNING
		result, err := db.conn.Exec(
			"INSERT INTO conversations (session_id, title, working_directory, transcript_path) VALUES (?, ?, ?, ?)",
			sessionID, title, workingDir, transcriptPath,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert conversation: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("failed to get last insert ID: %w", err)
		}

		// Fetch the created conversation
		return db.GetConversation(int(id))
	}

	return &conv, nil
}

// GetConversation retrieves a conversation by ID
func (db *DB) GetConversation(id int) (*Conversation, error) {
	query := `
	SELECT id, session_id, title, created_at, updated_at, prompt_count, total_characters, working_directory, transcript_path
	FROM conversations WHERE id = ?`

	var conv Conversation
	err := db.conn.QueryRow(query, id).Scan(
		&conv.ID, &conv.SessionID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt,
		&conv.PromptCount, &conv.TotalCharacters, &conv.WorkingDirectory, &conv.TranscriptPath,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrConversationNotFound
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return &conv, nil
}

// GetConversationBySessionID retrieves a conversation by session ID
func (db *DB) GetConversationBySessionID(sessionID string) (*Conversation, error) {
	query := `
	SELECT id, session_id, title, created_at, updated_at, prompt_count, total_characters, working_directory, transcript_path
	FROM conversations WHERE session_id = ?`

	var conv Conversation
	err := db.conn.QueryRow(query, sessionID).Scan(
		&conv.ID, &conv.SessionID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt,
		&conv.PromptCount, &conv.TotalCharacters, &conv.WorkingDirectory, &conv.TranscriptPath,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrConversationNotFound
		}
		return nil, fmt.Errorf("failed to get conversation by session ID: %w", err)
	}

	return &conv, nil
}

// GetConversationWithMessages retrieves a conversation with its messages
func (db *DB) GetConversationWithMessages(id int) (*ConversationWithMessages, error) {
	// Get conversation
	conv, err := db.GetConversation(id)
	if err != nil {
		return nil, err
	}

	// Get messages
	messages, err := db.GetMessagesByConversation(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return &ConversationWithMessages{
		Conversation: *conv,
		Messages:     messages,
	}, nil
}

// GetConversationCount returns the total number of conversations
func (db *DB) GetConversationCount() (int, error) {
	query := "SELECT COUNT(*) FROM conversations"
	
	var count int
	err := db.conn.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get conversation count: %w", err)
	}
	
	return count, nil
}

// ListConversations retrieves conversations with pagination
func (db *DB) ListConversations(limit, offset int) ([]Conversation, error) {
	query := `
	SELECT id, session_id, title, created_at, updated_at, prompt_count, total_characters, working_directory, transcript_path
	FROM conversations 
	ORDER BY updated_at DESC
	LIMIT ? OFFSET ?`

	rows, err := db.conn.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var conv Conversation
		err := rows.Scan(
			&conv.ID, &conv.SessionID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt,
			&conv.PromptCount, &conv.TotalCharacters, &conv.WorkingDirectory, &conv.TranscriptPath,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

// UpdateConversationTitle updates the title of a conversation
func (db *DB) UpdateConversationTitle(id int, title string) error {
	query := "UPDATE conversations SET title = ? WHERE id = ?"
	result, err := db.conn.Exec(query, title, id)
	if err != nil {
		return fmt.Errorf("failed to update conversation title: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return ErrConversationNotFound
	}

	return nil
}

// DeleteConversation deletes a conversation and its messages
func (db *DB) DeleteConversation(id int) error {
	// Start transaction
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete messages first (due to foreign key)
	_, err = tx.Exec("DELETE FROM messages WHERE conversation_id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}

	// Delete conversation
	result, err := tx.Exec("DELETE FROM conversations WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return ErrConversationNotFound
	}

	return tx.Commit()
}

// CreateMessage inserts a new message
func (db *DB) CreateMessage(conversationID int, messageType, content string, toolCalls *string, executionTime *int) (*Message, error) {
	characterCount := len(content)
	
	query := `
	INSERT INTO messages (conversation_id, message_type, content, character_count, tool_calls, execution_time)
	VALUES (?, ?, ?, ?, ?, ?)
	RETURNING id, conversation_id, message_type, content, character_count, timestamp, tool_calls, execution_time`

	var msg Message
	err := db.conn.QueryRow(query, conversationID, messageType, content, characterCount, toolCalls, executionTime).Scan(
		&msg.ID, &msg.ConversationID, &msg.MessageType, &msg.Content,
		&msg.CharacterCount, &msg.Timestamp, &msg.ToolCalls, &msg.ExecutionTime,
	)
	
	if err != nil {
		// Fallback for SQLite versions that don't support RETURNING
		result, err := db.conn.Exec(
			"INSERT INTO messages (conversation_id, message_type, content, character_count, tool_calls, execution_time) VALUES (?, ?, ?, ?, ?, ?)",
			conversationID, messageType, content, characterCount, toolCalls, executionTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert message: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("failed to get last insert ID: %w", err)
		}

		// Fetch the created message
		return db.GetMessage(int(id))
	}

	return &msg, nil
}

// GetMessage retrieves a message by ID
func (db *DB) GetMessage(id int) (*Message, error) {
	query := `
	SELECT id, conversation_id, message_type, content, character_count, timestamp, tool_calls, execution_time
	FROM messages WHERE id = ?`

	var msg Message
	err := db.conn.QueryRow(query, id).Scan(
		&msg.ID, &msg.ConversationID, &msg.MessageType, &msg.Content,
		&msg.CharacterCount, &msg.Timestamp, &msg.ToolCalls, &msg.ExecutionTime,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("message not found")
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return &msg, nil
}

// GetMessagesByConversation retrieves all messages for a conversation
func (db *DB) GetMessagesByConversation(conversationID int) ([]Message, error) {
	query := `
	SELECT id, conversation_id, message_type, content, character_count, timestamp, tool_calls, execution_time
	FROM messages 
	WHERE conversation_id = ?
	ORDER BY timestamp ASC`

	rows, err := db.conn.Query(query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.MessageType, &msg.Content,
			&msg.CharacterCount, &msg.Timestamp, &msg.ToolCalls, &msg.ExecutionTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}