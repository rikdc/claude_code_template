package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/claude-code-template/prompt-manager/internal/database"
)

// SessionHandler handles session events (start/stop)
type SessionHandler struct {
	db *database.DB
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(db *database.DB) *SessionHandler {
	return &SessionHandler{db: db}
}

// HandleSessionEvent processes session start/stop events
func (sh *SessionHandler) HandleSessionEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		sh.errorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var hookData HookData
	if err := json.NewDecoder(r.Body).Decode(&hookData); err != nil {
		sh.errorResponse(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	if hookData.SessionID == "" {
		sh.errorResponse(w, "session_id is required", http.StatusBadRequest)
		return
	}

	switch hookData.Event {
	case "SessionStart":
		sh.handleSessionStart(w, &hookData)
		return
	case "SessionEnd", "Stop":
		sh.handleSessionEnd(w, &hookData)
		return
	default:
		sh.errorResponse(w, fmt.Sprintf("Unknown session event: %s", hookData.Event), http.StatusBadRequest)
		return
	}
}

// handleSessionStart processes session start events
func (sh *SessionHandler) handleSessionStart(w http.ResponseWriter, hookData *HookData) {
	// Get or create conversation
	conversationID, err := sh.getOrCreateConversation(hookData.SessionID, hookData.Data)
	if err != nil {
		sh.errorResponse(w, fmt.Sprintf("Failed to get or create conversation: %v", err), http.StatusInternalServerError)
		return
	}

	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"event":           "session_start",
			"conversation_id": conversationID,
			"session_id":      hookData.SessionID,
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleSessionEnd processes session end/stop events
func (sh *SessionHandler) handleSessionEnd(w http.ResponseWriter, hookData *HookData) {
	// Try to find existing conversation for this session
	conversations, err := sh.db.ListConversations(10, 0)
	if err != nil {
		sh.errorResponse(w, fmt.Sprintf("Failed to list conversations: %v", err), http.StatusInternalServerError)
		return
	}

	var conversationID *int
	for _, conv := range conversations {
		if conv.SessionID == hookData.SessionID {
			conversationID = &conv.ID
			break
		}
	}

	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"event":           "session_end",
			"conversation_id": conversationID,
			"session_id":      hookData.SessionID,
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// getOrCreateConversation gets existing conversation or creates a new one
func (sh *SessionHandler) getOrCreateConversation(sessionID string, data map[string]interface{}) (int, error) {
	// Try to find existing conversation for this session
	conversations, err := sh.db.ListConversations(10, 0)
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

	conv, err := sh.db.CreateConversation(sessionID, nil, workingDir, transcriptPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create conversation: %w", err)
	}

	return conv.ID, nil
}

// errorResponse sends an error response
func (sh *SessionHandler) errorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	response := APIResponse{
		Success: false,
		Error:   &message,
	}
	json.NewEncoder(w).Encode(response)
}