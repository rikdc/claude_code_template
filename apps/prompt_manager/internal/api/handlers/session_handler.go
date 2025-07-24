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

	switch hookData.Event {
	case "SessionStart":
		sh.handleSessionStart(w, &hookData)
		return
	case "SessionEnd", "Stop":
		sh.handleSessionEnd(w, &hookData)
		return
	default:
		ErrorResponse(w, fmt.Sprintf("Unknown session event: %s", hookData.Event), http.StatusBadRequest)
		return
	}
}

// handleSessionStart processes session start events
func (sh *SessionHandler) handleSessionStart(w http.ResponseWriter, hookData *HookData) {
	// Get or create conversation
	conversationID, err := GetOrCreateConversation(sh.db, hookData.SessionID, hookData.Data)
	if err != nil {
		ErrorResponse(w, fmt.Sprintf("Failed to get or create conversation: %v", err), http.StatusInternalServerError)
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
		ErrorResponse(w, fmt.Sprintf("Failed to list conversations: %v", err), http.StatusInternalServerError)
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
