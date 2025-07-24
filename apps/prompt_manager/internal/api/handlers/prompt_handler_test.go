package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)


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

func TestPromptHandler_HandlePromptSubmit(t *testing.T) {
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
			name:           "successful prompt submission",
			method:         http.MethodPost,
			payload: HookData{
				Event:     "UserPromptSubmit",
				Timestamp: time.Now().Format(time.RFC3339),
				SessionID: "test-session-123",
				Filename:  "activity-monitor",
				Data: map[string]interface{}{
					"prompt": "Test prompt content",
					"cwd":    "/test/directory",
				},
			},
			expectedStatus: http.StatusCreated,
			expectSuccess:  true,
			validateData: func(t *testing.T, data map[string]interface{}) {
				if data["session_id"] != "test-session-123" {
					t.Errorf("Expected session_id 'test-session-123', got %v", data["session_id"])
				}
				if data["type"] != "prompt" {
					t.Errorf("Expected type 'prompt', got %v", data["type"])
				}
				if data["message_id"] == nil {
					t.Error("Expected message_id to be set")
				}
				if data["conversation_id"] == nil {
					t.Error("Expected conversation_id to be set")
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
				Event:     "UserPromptSubmit",
				Timestamp: time.Now().Format(time.RFC3339),
				// SessionID missing
				Filename: "activity-monitor",
				Data: map[string]interface{}{
					"prompt": "Test prompt content",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "session_id is required",
			expectSuccess:  false,
		},
		{
			name:   "missing prompt data",
			method: http.MethodPost,
			payload: HookData{
				Event:     "UserPromptSubmit",
				Timestamp: time.Now().Format(time.RFC3339),
				SessionID: "test-session-123",
				Filename:  "activity-monitor",
				Data: map[string]interface{}{
					// prompt missing
					"cwd": "/test/directory",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "no prompt data in request",
			expectSuccess:  false,
		},
		{
			name:   "invalid prompt data type",
			method: http.MethodPost,
			payload: HookData{
				Event:     "UserPromptSubmit",
				Timestamp: time.Now().Format(time.RFC3339),
				SessionID: "test-session-123",
				Filename:  "activity-monitor",
				Data: map[string]interface{}{
					"prompt": 123, // Should be string, not number
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "prompt data must be a string",
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			db := setupTestDB(t)
			defer db.Close()
			
			handler := NewPromptHandler(db)
			
			// Prepare request
			var req *http.Request
			if tt.payload == nil {
				req = httptest.NewRequest(tt.method, "/messages/prompt", nil)
			} else if str, ok := tt.payload.(string); ok {
				// Handle string payload (invalid JSON case)
				req = httptest.NewRequest(tt.method, "/messages/prompt", bytes.NewBufferString(str))
			} else {
				// Handle struct payload
				payload, err := json.Marshal(tt.payload)
				if err != nil {
					t.Fatalf("Failed to marshal test payload: %v", err)
				}
				req = httptest.NewRequest(tt.method, "/messages/prompt", bytes.NewBuffer(payload))
			}
			
			if tt.method == http.MethodPost {
				req.Header.Set("Content-Type", "application/json")
			}
			
			// Execute request
			w := httptest.NewRecorder()
			handler.HandlePromptSubmit(w, req)
			
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

func TestPromptHandler_ConversationReuse(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	handler := NewPromptHandler(db)
	sessionID := "test-session-reuse-456"
	
	// Submit first prompt
	hookData1 := HookData{
		Event:     "UserPromptSubmit",
		Timestamp: time.Now().Format(time.RFC3339),
		SessionID: sessionID,
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
		SessionID: sessionID, // Same session
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