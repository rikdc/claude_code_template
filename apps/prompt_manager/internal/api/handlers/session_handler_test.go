package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewSessionHandler(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewSessionHandler(db)
	
	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}
	
	if handler.db != db {
		t.Error("Expected handler to store database reference")
	}
}

func TestSessionHandler_HandleSessionEvent_SessionStart(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewSessionHandler(db)
	
	// Create test payload for session start
	hookData := HookData{
		Event:     "SessionStart",
		Timestamp: time.Now().Format(time.RFC3339),
		SessionID: "test-session-start-123",
		Filename:  "activity-monitor",
		Data: map[string]interface{}{
			"cwd":             "/test/start/directory",
			"transcript_path": "/test/transcript-start.md",
		},
	}
	
	payload, err := json.Marshal(hookData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}
	
	// Create request
	req := httptest.NewRequest(http.MethodPost, "/messages/session", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	
	// Create response recorder
	w := httptest.NewRecorder()
	
	// Execute request
	handler.HandleSessionEvent(w, req)
	
	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
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
	
	if data["event"] != "session_start" {
		t.Errorf("Expected event 'session_start', got %v", data["event"])
	}
}

func TestSessionHandler_HandleSessionEvent_SessionEnd(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewSessionHandler(db)
	
	// Create test payload for session end
	hookData := HookData{
		Event:     "SessionEnd",
		Timestamp: time.Now().Format(time.RFC3339),
		SessionID: "test-session-end-456",
		Filename:  "activity-monitor",
		Data: map[string]interface{}{
			"duration":        3600000, // 1 hour in milliseconds
			"total_messages":  25,
			"conversation_id": 123,
		},
	}
	
	payload, err := json.Marshal(hookData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}
	
	// Create request
	req := httptest.NewRequest(http.MethodPost, "/messages/session", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	
	// Create response recorder
	w := httptest.NewRecorder()
	
	// Execute request
	handler.HandleSessionEvent(w, req)
	
	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
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
}

func TestSessionHandler_HandleSessionEvent_MethodNotAllowed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewSessionHandler(db)
	
	// Test GET request (not allowed)
	req := httptest.NewRequest(http.MethodGet, "/messages/session", nil)
	w := httptest.NewRecorder()
	
	handler.HandleSessionEvent(w, req)
	
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

func TestSessionHandler_HandleSessionEvent_InvalidJSON(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewSessionHandler(db)
	
	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/messages/session", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	handler.HandleSessionEvent(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	var response APIResponse
	json.NewDecoder(w.Body).Decode(&response)
	
	if response.Success {
		t.Error("Expected response.Success to be false")
	}
}

func TestSessionHandler_HandleSessionEvent_MissingSessionID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewSessionHandler(db)
	
	// Create payload without session_id
	hookData := HookData{
		Event:     "SessionStart",
		Timestamp: time.Now().Format(time.RFC3339),
		// SessionID missing
		Filename: "activity-monitor",
		Data: map[string]interface{}{
			"cwd": "/test/directory",
		},
	}
	
	payload, _ := json.Marshal(hookData)
	
	req := httptest.NewRequest(http.MethodPost, "/messages/session", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	handler.HandleSessionEvent(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	var response APIResponse
	json.NewDecoder(w.Body).Decode(&response)
	
	if response.Success {
		t.Error("Expected response.Success to be false")
	}
}

func TestSessionHandler_HandleSessionEvent_UnknownEvent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewSessionHandler(db)
	
	// Create test payload for unknown event
	hookData := HookData{
		Event:     "CustomEvent",
		Timestamp: time.Now().Format(time.RFC3339),
		SessionID: "test-session-unknown-789",
		Filename:  "activity-monitor",
		Data: map[string]interface{}{
			"custom_field": "custom_value",
			"number_field": 42,
		},
	}
	
	payload, _ := json.Marshal(hookData)
	
	req := httptest.NewRequest(http.MethodPost, "/messages/session", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	handler.HandleSessionEvent(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	var response APIResponse
	json.NewDecoder(w.Body).Decode(&response)
	
	if response.Success {
		t.Error("Expected response.Success to be false")
	}
	
	if response.Error == nil {
		t.Error("Expected error message for unknown event")
	}
}

func TestSessionHandler_ConversationCreation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewSessionHandler(db)
	
	// Submit session event for new session (should create conversation)
	hookData := HookData{
		Event:     "SessionStart",
		Timestamp: time.Now().Format(time.RFC3339),
		SessionID: "new-session-conversation-test",
		Filename:  "activity-monitor",
		Data: map[string]interface{}{
			"cwd":             "/session/test/directory",
			"transcript_path": "/session/test/transcript.md",
		},
	}
	
	payload, _ := json.Marshal(hookData)
	
	req := httptest.NewRequest(http.MethodPost, "/messages/session", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	handler.HandleSessionEvent(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
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