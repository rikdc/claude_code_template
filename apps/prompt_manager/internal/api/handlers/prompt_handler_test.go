package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/claude-code-template/prompt-manager/internal/database"
)

func setupTestDB(t *testing.T) *database.DB {
	// Create temporary database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	
	config := &database.Config{
		DatabasePath:  dbPath,
		MigrationsDir: "../../../database/migrations",
	}
	
	db, err := database.New(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	
	// Run migrations
	if err := db.RunMigrations(config.MigrationsDir); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}
	
	return db
}

func TestNewPromptHandler(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewPromptHandler(db)
	
	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}
	
	if handler.db != db {
		t.Error("Expected handler to store database reference")
	}
}

func TestPromptHandler_HandlePromptSubmit_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewPromptHandler(db)
	
	// Create test payload
	hookData := HookData{
		Event:     "UserPromptSubmit",
		Timestamp: time.Now().Format(time.RFC3339),
		SessionID: "test-session-123",
		Filename:  "activity-monitor",
		Data: map[string]interface{}{
			"prompt": "Test prompt content",
			"cwd":    "/test/directory",
		},
	}
	
	payload, err := json.Marshal(hookData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}
	
	// Create request
	req := httptest.NewRequest(http.MethodPost, "/messages/prompt", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	
	// Create response recorder
	w := httptest.NewRecorder()
	
	// Execute request
	handler.HandlePromptSubmit(w, req)
	
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
	
	if data["type"] != "prompt" {
		t.Errorf("Expected type 'prompt', got %v", data["type"])
	}
}

func TestPromptHandler_HandlePromptSubmit_MethodNotAllowed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewPromptHandler(db)
	
	// Test GET request (not allowed)
	req := httptest.NewRequest(http.MethodGet, "/messages/prompt", nil)
	w := httptest.NewRecorder()
	
	handler.HandlePromptSubmit(w, req)
	
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
	
	var response APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if response.Success {
		t.Error("Expected response.Success to be false")
	}
	
	if response.Error == nil || *response.Error != "Method not allowed" {
		t.Error("Expected 'Method not allowed' error")
	}
}

func TestPromptHandler_HandlePromptSubmit_InvalidJSON(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewPromptHandler(db)
	
	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/messages/prompt", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	handler.HandlePromptSubmit(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	var response APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if response.Success {
		t.Error("Expected response.Success to be false")
	}
	
	if response.Error == nil || *response.Error != "Invalid JSON request body" {
		t.Error("Expected 'Invalid JSON request body' error")
	}
}

func TestPromptHandler_HandlePromptSubmit_MissingSessionID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewPromptHandler(db)
	
	// Create payload without session_id
	hookData := HookData{
		Event:     "UserPromptSubmit",
		Timestamp: time.Now().Format(time.RFC3339),
		// SessionID missing
		Filename: "activity-monitor",
		Data: map[string]interface{}{
			"prompt": "Test prompt content",
		},
	}
	
	payload, _ := json.Marshal(hookData)
	
	req := httptest.NewRequest(http.MethodPost, "/messages/prompt", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	handler.HandlePromptSubmit(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	var response APIResponse
	json.NewDecoder(w.Body).Decode(&response)
	
	if response.Success {
		t.Error("Expected response.Success to be false")
	}
	
	if response.Error == nil || *response.Error != "session_id is required" {
		t.Error("Expected 'session_id is required' error")
	}
}

func TestPromptHandler_HandlePromptSubmit_MissingPromptData(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewPromptHandler(db)
	
	// Create payload without prompt data
	hookData := HookData{
		Event:     "UserPromptSubmit",
		Timestamp: time.Now().Format(time.RFC3339),
		SessionID: "test-session-123",
		Filename:  "activity-monitor",
		Data:      map[string]interface{}{
			// prompt missing
			"cwd": "/test/directory",
		},
	}
	
	payload, _ := json.Marshal(hookData)
	
	req := httptest.NewRequest(http.MethodPost, "/messages/prompt", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	handler.HandlePromptSubmit(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	var response APIResponse
	json.NewDecoder(w.Body).Decode(&response)
	
	if response.Success {
		t.Error("Expected response.Success to be false")
	}
	
	if response.Error == nil || *response.Error != "no prompt data in request" {
		t.Error("Expected 'no prompt data in request' error")
	}
}

func TestPromptHandler_HandlePromptSubmit_InvalidPromptDataType(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewPromptHandler(db)
	
	// Create payload with non-string prompt data
	hookData := HookData{
		Event:     "UserPromptSubmit",
		Timestamp: time.Now().Format(time.RFC3339),
		SessionID: "test-session-123",
		Filename:  "activity-monitor",
		Data: map[string]interface{}{
			"prompt": 123, // Should be string, not number
		},
	}
	
	payload, _ := json.Marshal(hookData)
	
	req := httptest.NewRequest(http.MethodPost, "/messages/prompt", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	handler.HandlePromptSubmit(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	var response APIResponse
	json.NewDecoder(w.Body).Decode(&response)
	
	if response.Success {
		t.Error("Expected response.Success to be false")
	}
	
	if response.Error == nil || *response.Error != "prompt data must be a string" {
		t.Error("Expected 'prompt data must be a string' error")
	}
}

func TestPromptHandler_CreateConversationAndMessage(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewPromptHandler(db)
	
	// Submit first prompt for new session
	hookData1 := HookData{
		Event:     "UserPromptSubmit",
		Timestamp: time.Now().Format(time.RFC3339),
		SessionID: "test-session-456",
		Filename:  "activity-monitor",
		Data: map[string]interface{}{
			"prompt":          "First prompt",
			"cwd":             "/test/directory",
			"transcript_path": "/test/transcript.md",
		},
	}
	
	payload1, _ := json.Marshal(hookData1)
	req1 := httptest.NewRequest(http.MethodPost, "/messages/prompt", bytes.NewBuffer(payload1))
	req1.Header.Set("Content-Type", "application/json")
	
	w1 := httptest.NewRecorder()
	handler.HandlePromptSubmit(w1, req1)
	
	if w1.Code != http.StatusCreated {
		t.Fatalf("First request failed with status %d", w1.Code)
	}
	
	var response1 APIResponse
	json.NewDecoder(w1.Body).Decode(&response1)
	
	data1 := response1.Data.(map[string]interface{})
	conversationID1 := data1["conversation_id"]
	
	// Submit second prompt for same session
	hookData2 := HookData{
		Event:     "UserPromptSubmit",
		Timestamp: time.Now().Format(time.RFC3339),
		SessionID: "test-session-456", // Same session
		Filename:  "activity-monitor",
		Data: map[string]interface{}{
			"prompt": "Second prompt",
		},
	}
	
	payload2, _ := json.Marshal(hookData2)
	req2 := httptest.NewRequest(http.MethodPost, "/messages/prompt", bytes.NewBuffer(payload2))
	req2.Header.Set("Content-Type", "application/json")
	
	w2 := httptest.NewRecorder()
	handler.HandlePromptSubmit(w2, req2)
	
	if w2.Code != http.StatusCreated {
		t.Fatalf("Second request failed with status %d", w2.Code)
	}
	
	var response2 APIResponse
	json.NewDecoder(w2.Body).Decode(&response2)
	
	data2 := response2.Data.(map[string]interface{})
	conversationID2 := data2["conversation_id"]
	
	// Should use same conversation for same session
	if conversationID1 != conversationID2 {
		t.Errorf("Expected same conversation ID for same session, got %v and %v", conversationID1, conversationID2)
	}
}