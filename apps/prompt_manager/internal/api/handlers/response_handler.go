package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/claude-code-template/prompt-manager/internal/database"
)

// ResponseHandler handles assistant response submissions
type ResponseHandler struct {
	db *database.DB
}

// NewResponseHandler creates a new response handler
func NewResponseHandler(db *database.DB) *ResponseHandler {
	return &ResponseHandler{db: db}
}

// HandleResponseSubmit processes assistant response submissions
func (rh *ResponseHandler) HandleResponseSubmit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		rh.errorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var hookData HookData
	if err := json.NewDecoder(r.Body).Decode(&hookData); err != nil {
		rh.errorResponse(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	if hookData.SessionID == "" {
		rh.errorResponse(w, "session_id is required", http.StatusBadRequest)
		return
	}

	// Extract response content from hook data
	var responseContent string
	var toolCallsJSON *string
	var executionTime *int

	// Try to extract response content from various possible fields
	if content, ok := hookData.Data["response"]; ok {
		if str, ok := content.(string); ok {
			responseContent = str
		}
	} else if content, ok := hookData.Data["content"]; ok {
		if str, ok := content.(string); ok {
			responseContent = str
		}
	}

	if responseContent == "" {
		rh.errorResponse(w, "no response content in request", http.StatusBadRequest)
		return
	}

	// Extract tool calls if present
	if toolCalls, ok := hookData.Data["tool_calls"]; ok {
		if toolCallsData, err := json.Marshal(toolCalls); err == nil {
			toolCallsStr := string(toolCallsData)
			toolCallsJSON = &toolCallsStr
		}
	}

	// Extract execution time if present
	if execTime, ok := hookData.Data["execution_time"]; ok {
		if timeMs, ok := execTime.(float64); ok {
			execTimeInt := int(timeMs)
			executionTime = &execTimeInt
		}
	}

	// Get or create conversation
	conversationID, err := rh.getOrCreateConversation(hookData.SessionID, hookData.Data)
	if err != nil {
		rh.errorResponse(w, fmt.Sprintf("Failed to get or create conversation: %v", err), http.StatusInternalServerError)
		return
	}

	// Create message record
	message, err := rh.db.CreateMessage(conversationID, "response", responseContent, toolCallsJSON, executionTime)
	if err != nil {
		rh.errorResponse(w, fmt.Sprintf("Failed to create message: %v", err), http.StatusInternalServerError)
		return
	}

	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message_id":      message.ID,
			"conversation_id": conversationID,
			"session_id":      hookData.SessionID,
			"type":            "response",
			"timestamp":       message.Timestamp,
			"has_tool_calls":  toolCallsJSON != nil,
			"execution_time":  executionTime,
		},
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// getOrCreateConversation gets existing conversation or creates a new one
func (rh *ResponseHandler) getOrCreateConversation(sessionID string, data map[string]interface{}) (int, error) {
	// Try to find existing conversation for this session
	conversations, err := rh.db.ListConversations(10, 0)
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

	conv, err := rh.db.CreateConversation(sessionID, nil, workingDir, transcriptPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create conversation: %w", err)
	}

	return conv.ID, nil
}

// errorResponse sends an error response
func (rh *ResponseHandler) errorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	response := APIResponse{
		Success: false,
		Error:   &message,
	}
	json.NewEncoder(w).Encode(response)
}