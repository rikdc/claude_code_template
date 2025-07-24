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

func TestSessionHandler_HandleSessionEvent(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		payload        interface{}
		expectedStatus int
		expectedError  string
		expectSuccess  bool
		validateData   func(t *testing.T, data map[string]interface{})
	}{
		{
			name:   "session start event",
			method: http.MethodPost,
			payload: HookData{
				Event:     "SessionStart",
				Timestamp: time.Now().Format(time.RFC3339),
				SessionID: "test-session-start-123",
				Filename:  "activity-monitor",
				Data: map[string]interface{}{
					"cwd":             "/test/start/directory",
					"transcript_path": "/test/transcript-start.md",
				},
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
			validateData: func(t *testing.T, data map[string]interface{}) {
				if data["session_id"] != "test-session-start-123" {
					t.Errorf("Expected session_id 'test-session-start-123', got %v", data["session_id"])
				}
				if data["event"] != "session_start" {
					t.Errorf("Expected event 'session_start', got %v", data["event"])
				}
				if data["conversation_id"] == nil {
					t.Error("Expected conversation_id to be set")
				}
			},
		},
		{
			name:   "session end event",
			method: http.MethodPost,
			payload: HookData{
				Event:     "SessionEnd",
				Timestamp: time.Now().Format(time.RFC3339),
				SessionID: "test-session-end-456",
				Filename:  "activity-monitor",
				Data: map[string]interface{}{
					"duration":        3600000,
					"total_messages":  25,
					"conversation_id": 123,
				},
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
			validateData: func(t *testing.T, data map[string]interface{}) {
				if data["session_id"] != "test-session-end-456" {
					t.Errorf("Expected session_id 'test-session-end-456', got %v", data["session_id"])
				}
				if data["event"] != "session_end" {
					t.Errorf("Expected event 'session_end', got %v", data["event"])
				}
			},
		},
		{
			name:   "stop event (alias for session end)",
			method: http.MethodPost,
			payload: HookData{
				Event:     "Stop",
				Timestamp: time.Now().Format(time.RFC3339),
				SessionID: "test-session-stop-789",
				Filename:  "activity-monitor",
				Data: map[string]interface{}{
					"duration": 1800000,
				},
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
			validateData: func(t *testing.T, data map[string]interface{}) {
				if data["session_id"] != "test-session-stop-789" {
					t.Errorf("Expected session_id 'test-session-stop-789', got %v", data["session_id"])
				}
				if data["event"] != "session_end" {
					t.Errorf("Expected event 'session_end', got %v", data["event"])
				}
			},
		},
		{
			name:           "method not allowed",
			method:         http.MethodGet,
			payload:        nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "Method not allowed",
			expectSuccess:  false,
		},
		{
			name:           "invalid JSON",
			method:         http.MethodPost,
			payload:        "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid JSON request body",
			expectSuccess:  false,
		},
		{
			name:   "missing session ID",
			method: http.MethodPost,
			payload: HookData{
				Event:     "SessionStart",
				Timestamp: time.Now().Format(time.RFC3339),
				// SessionID missing
				Filename: "activity-monitor",
				Data: map[string]interface{}{
					"cwd": "/test/directory",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "session_id is required",
			expectSuccess:  false,
		},
		{
			name:   "unknown event type",
			method: http.MethodPost,
			payload: HookData{
				Event:     "CustomEvent",
				Timestamp: time.Now().Format(time.RFC3339),
				SessionID: "test-session-unknown-789",
				Filename:  "activity-monitor",
				Data: map[string]interface{}{
					"custom_field": "custom_value",
					"number_field": 42,
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Unknown session event: CustomEvent",
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			db := setupTestDB(t)
			defer db.Close()
			
			handler := NewSessionHandler(db)
			
			// Prepare request
			var req *http.Request
			if tt.payload == nil {
				req = httptest.NewRequest(tt.method, "/messages/session", nil)
			} else if str, ok := tt.payload.(string); ok {
				// Handle string payload (invalid JSON case)
				req = httptest.NewRequest(tt.method, "/messages/session", bytes.NewBufferString(str))
			} else {
				// Handle struct payload
				payload, err := json.Marshal(tt.payload)
				if err != nil {
					t.Fatalf("Failed to marshal test payload: %v", err)
				}
				req = httptest.NewRequest(tt.method, "/messages/session", bytes.NewBuffer(payload))
			}
			
			if tt.method == http.MethodPost {
				req.Header.Set("Content-Type", "application/json")
			}
			
			// Execute request
			w := httptest.NewRecorder()
			handler.HandleSessionEvent(w, req)
			
			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			
			// Parse response
			var response APIResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}
			
			// Check success flag
			if response.Success != tt.expectSuccess {
				t.Errorf("Expected success %v, got %v", tt.expectSuccess, response.Success)
			}
			
			// Check error message
			if tt.expectedError != "" {
				if response.Error == nil {
					t.Errorf("Expected error '%s', got nil", tt.expectedError)
				} else if *response.Error != tt.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tt.expectedError, *response.Error)
				}
			} else if response.Error != nil {
				t.Errorf("Expected no error, got '%s'", *response.Error)
			}
			
			// Custom data validation
			if tt.validateData != nil && response.Success {
				data, ok := response.Data.(map[string]interface{})
				if !ok {
					t.Fatal("Expected response.Data to be a map")
				}
				tt.validateData(t, data)
			}
		})
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