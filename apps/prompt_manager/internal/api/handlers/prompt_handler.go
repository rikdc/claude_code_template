package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/claude-code-template/prompt-manager/internal/database"
)

// PromptHandler handles user prompt submissions
type PromptHandler struct {
	db *database.DB
}

// NewPromptHandler creates a new prompt handler
func NewPromptHandler(db *database.DB) *PromptHandler {
	return &PromptHandler{db: db}
}


// HandlePromptSubmit processes user prompt submissions
func (ph *PromptHandler) HandlePromptSubmit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		ph.errorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var hookData HookData
	if err := json.NewDecoder(r.Body).Decode(&hookData); err != nil {
		ph.errorResponse(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	if hookData.SessionID == "" {
		ph.errorResponse(w, "session_id is required", http.StatusBadRequest)
		return
	}

	// Extract prompt content from hook data
	promptData, ok := hookData.Data["prompt"]
	if !ok {
		ph.errorResponse(w, "no prompt data in request", http.StatusBadRequest)
		return
	}

	prompt, ok := promptData.(string)
	if !ok {
		ph.errorResponse(w, "prompt data must be a string", http.StatusBadRequest)
		return
	}

	// Get or create conversation
	conversationID, err := ph.getOrCreateConversation(hookData.SessionID, hookData.Data)
	if err != nil {
		ph.errorResponse(w, fmt.Sprintf("Failed to get or create conversation: %v", err), http.StatusInternalServerError)
		return
	}

	// Create message record
	message, err := ph.db.CreateMessage(conversationID, "prompt", prompt, nil, nil)
	if err != nil {
		ph.errorResponse(w, fmt.Sprintf("Failed to create message: %v", err), http.StatusInternalServerError)
		return
	}

	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message_id":      message.ID,
			"conversation_id": conversationID,
			"session_id":      hookData.SessionID,
			"type":            "prompt",
			"timestamp":       message.Timestamp,
		},
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// getOrCreateConversation gets existing conversation or creates a new one
func (ph *PromptHandler) getOrCreateConversation(sessionID string, data map[string]interface{}) (int, error) {
	// Try to find existing conversation for this session
	conversations, err := ph.db.ListConversations(10, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to list conversations: %w", err)
	}

	// Check if any conversation matches this session
	for _, conv := range conversations {
		if conv.SessionID == sessionID {
			return conv.ID, nil
		}
	}

	// Create new conversation
	workingDir := extractStringFromData(data, "cwd")
	transcriptPath := extractStringFromData(data, "transcript_path")

	conv, err := ph.db.CreateConversation(sessionID, nil, workingDir, transcriptPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create conversation: %w", err)
	}

	return conv.ID, nil
}

// extractStringFromData safely extracts a string value from map data
func extractStringFromData(data map[string]interface{}, key string) *string {
	if value, exists := data[key]; exists {
		if str, ok := value.(string); ok && str != "" {
			return &str
		}
	}
	return nil
}

// errorResponse sends an error response
func (ph *PromptHandler) errorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	response := APIResponse{
		Success: false,
		Error:   &message,
	}
	json.NewEncoder(w).Encode(response)
}