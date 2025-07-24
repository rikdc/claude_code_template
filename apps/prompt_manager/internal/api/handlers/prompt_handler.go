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
		ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var hookData HookData
	if err := json.NewDecoder(r.Body).Decode(&hookData); err != nil {
		ErrorResponse(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	if hookData.SessionID == "" {
		ErrorResponse(w, "session_id is required", http.StatusBadRequest)
		return
	}

	// Extract prompt content from hook data
	promptData, ok := hookData.Data["prompt"]
	if !ok {
		ErrorResponse(w, "no prompt data in request", http.StatusBadRequest)
		return
	}

	prompt, ok := promptData.(string)
	if !ok {
		ErrorResponse(w, "prompt data must be a string", http.StatusBadRequest)
		return
	}

	// Get or create conversation
	conversationID, err := GetOrCreateConversation(ph.db, hookData.SessionID, hookData.Data)
	if err != nil {
		ErrorResponse(w, fmt.Sprintf("Failed to get or create conversation: %v", err), http.StatusInternalServerError)
		return
	}

	// Create message record
	message, err := ph.db.CreateMessage(conversationID, "prompt", prompt, nil, nil)
	if err != nil {
		ErrorResponse(w, fmt.Sprintf("Failed to create message: %v", err), http.StatusInternalServerError)
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

