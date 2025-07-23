package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewResponseHandler(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewResponseHandler(db)
	
	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}
	
	if handler.db != db {
		t.Error("Expected handler to store database reference")
	}
}

func TestResponseHandler_HandleResponseSubmit_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewResponseHandler(db)
	
	// Create test payload
	hookData := HookData{
		Event:     "PostToolUse",
		Timestamp: time.Now().Format(time.RFC3339),
		SessionID: "test-session-456",
		Filename:  "activity-monitor",
		Data: map[string]interface{}{
			"response": "This is an assistant response",
			"tool_calls": []map[string]interface{}{
				{
					"name":      "Read",
					"arguments": map[string]interface{}{"file_path": "/test/file.txt"},
				},
			},
			"execution_time": 1500, // milliseconds
		},
	}
	
	payload, err := json.Marshal(hookData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}
	
	// Create request
	req := httptest.NewRequest(http.MethodPost, "/messages/response", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	
	// Create response recorder
	w := httptest.NewRecorder()
	
	// Execute request
	handler.HandleResponseSubmit(w, req)
	
	// Check response
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
	
	// Parse response
	var response APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if !response.Success {
		t.Error("Expected response.Success to be true")
	}
	
	if response.Error != nil {
		t.Errorf("Expected no error, got %s", *response.Error)
	}
	
	// Verify data structure
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected response.Data to be a map")
	}
	
	if data["session_id"] != hookData.SessionID {
		t.Errorf("Expected session_id %s, got %v", hookData.SessionID, data["session_id"])
	}
	
	if data["type"] != "response" {
		t.Errorf("Expected type 'response', got %v", data["type"])
	}
}

func TestResponseHandler_HandleResponseSubmit_MethodNotAllowed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewResponseHandler(db)
	
	// Test GET request (not allowed)
	req := httptest.NewRequest(http.MethodGet, "/messages/response", nil)
	w := httptest.NewRecorder()
	
	handler.HandleResponseSubmit(w, req)
	
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
	
	var response APIResponse
	json.NewDecoder(w.Body).Decode(&response)
	
	if response.Success {
		t.Error("Expected response.Success to be false")
	}
	
	if response.Error == nil || *response.Error != "Method not allowed" {
		t.Error("Expected 'Method not allowed' error")
	}
}

func TestResponseHandler_HandleResponseSubmit_InvalidJSON(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewResponseHandler(db)
	
	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/messages/response", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	handler.HandleResponseSubmit(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	var response APIResponse
	json.NewDecoder(w.Body).Decode(&response)
	
	if response.Success {
		t.Error("Expected response.Success to be false")
	}
}

func TestResponseHandler_HandleResponseSubmit_MissingSessionID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewResponseHandler(db)
	
	// Create payload without session_id
	hookData := HookData{
		Event:     "PostToolUse",
		Timestamp: time.Now().Format(time.RFC3339),
		// SessionID missing
		Filename: "activity-monitor",
		Data: map[string]interface{}{
			"response": "Test response content",
		},
	}
	
	payload, _ := json.Marshal(hookData)
	
	req := httptest.NewRequest(http.MethodPost, "/messages/response", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	handler.HandleResponseSubmit(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	var response APIResponse
	json.NewDecoder(w.Body).Decode(&response)
	
	if response.Success {
		t.Error("Expected response.Success to be false")
	}
}

func TestResponseHandler_HandleResponseSubmit_MissingResponseData(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewResponseHandler(db)
	
	// Create payload without response data
	hookData := HookData{
		Event:     "PostToolUse",
		Timestamp: time.Now().Format(time.RFC3339),
		SessionID: "test-session-123",
		Filename:  "activity-monitor",
		Data: map[string]interface{}{
			// response missing
			"tool_calls": []interface{}{},
		},
	}
	
	payload, _ := json.Marshal(hookData)
	
	req := httptest.NewRequest(http.MethodPost, "/messages/response", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	handler.HandleResponseSubmit(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	var response APIResponse
	json.NewDecoder(w.Body).Decode(&response)
	
	if response.Success {
		t.Error("Expected response.Success to be false")
	}
}

func TestResponseHandler_HandleResponseSubmit_InvalidResponseDataType(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewResponseHandler(db)
	
	// Create payload with non-string response data
	hookData := HookData{
		Event:     "PostToolUse",
		Timestamp: time.Now().Format(time.RFC3339),
		SessionID: "test-session-123",
		Filename:  "activity-monitor",
		Data: map[string]interface{}{
			"response": 456, // Should be string, not number
		},
	}
	
	payload, _ := json.Marshal(hookData)
	
	req := httptest.NewRequest(http.MethodPost, "/messages/response", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	handler.HandleResponseSubmit(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	var response APIResponse
	json.NewDecoder(w.Body).Decode(&response)
	
	if response.Success {
		t.Error("Expected response.Success to be false")
	}
}

func TestResponseHandler_HandleResponseSubmit_WithToolCalls(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewResponseHandler(db)
	
	// Create test payload with tool calls
	hookData := HookData{
		Event:     "PostToolUse",
		Timestamp: time.Now().Format(time.RFC3339),
		SessionID: "test-session-789",
		Filename:  "activity-monitor",
		Data: map[string]interface{}{
			"response": "Used Read tool to check file contents",
			"tool_calls": []map[string]interface{}{
				{
					"name": "Read",
					"arguments": map[string]interface{}{
						"file_path": "/test/example.txt",
						"limit":     100,
					},
				},
				{
					"name": "Write",
					"arguments": map[string]interface{}{
						"file_path": "/test/output.txt",
						"content":   "test content",
					},
				},
			},
			"execution_time": 2300,
		},
	}
	
	payload, _ := json.Marshal(hookData)
	
	req := httptest.NewRequest(http.MethodPost, "/messages/response", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	handler.HandleResponseSubmit(w, req)
	
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
	
	var response APIResponse
	json.NewDecoder(w.Body).Decode(&response)
	
	if !response.Success {
		t.Error("Expected response.Success to be true")
	}
	
	// Verify that tool calls data is preserved
	data := response.Data.(map[string]interface{})
	if data["type"] != "response" {
		t.Errorf("Expected type 'response', got %v", data["type"])
	}
}

func TestResponseHandler_ConversationCreation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewResponseHandler(db)
	
	// Submit response for new session (should create conversation)
	hookData := HookData{
		Event:     "PostToolUse",
		Timestamp: time.Now().Format(time.RFC3339),
		SessionID: "new-session-999",
		Filename:  "activity-monitor",
		Data: map[string]interface{}{
			"response":       "Assistant response without prior prompt",
			"cwd":            "/new/directory",
			"transcript_path": "/new/transcript.md",
		},
	}
	
	payload, _ := json.Marshal(hookData)
	
	req := httptest.NewRequest(http.MethodPost, "/messages/response", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	handler.HandleResponseSubmit(w, req)
	
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
	
	var response APIResponse
	json.NewDecoder(w.Body).Decode(&response)
	
	if !response.Success {
		t.Error("Expected response.Success to be true")
	}
	
	// Verify conversation was created
	data := response.Data.(map[string]interface{})
	if data["conversation_id"] == nil {
		t.Error("Expected conversation_id to be set")
	}
	
	if data["session_id"] != hookData.SessionID {
		t.Errorf("Expected session_id %s, got %v", hookData.SessionID, data["session_id"])
	}
}