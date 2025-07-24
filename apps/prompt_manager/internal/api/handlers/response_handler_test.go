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

func TestResponseHandler_HandleResponseSubmit(t *testing.T) {
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
			name:   "successful response submission",
			method: http.MethodPost,
			payload: HookData{
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
					"execution_time": 1500,
				},
			},
			expectedStatus: http.StatusCreated,
			expectSuccess:  true,
			validateData: func(t *testing.T, data map[string]interface{}) {
				if data["session_id"] != "test-session-456" {
					t.Errorf("Expected session_id 'test-session-456', got %v", data["session_id"])
				}
				if data["type"] != "response" {
					t.Errorf("Expected type 'response', got %v", data["type"])
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
				Event:     "PostToolUse",
				Timestamp: time.Now().Format(time.RFC3339),
				// SessionID missing
				Filename: "activity-monitor",
				Data: map[string]interface{}{
					"response": "Test response content",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "session_id is required",
			expectSuccess:  false,
		},
		{
			name:   "missing response data",
			method: http.MethodPost,
			payload: HookData{
				Event:     "PostToolUse",
				Timestamp: time.Now().Format(time.RFC3339),
				SessionID: "test-session-123",
				Filename:  "activity-monitor",
				Data: map[string]interface{}{
					// response missing
					"tool_calls": []interface{}{},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "no response content in request",
			expectSuccess:  false,
		},
		{
			name:   "invalid response data type",
			method: http.MethodPost,
			payload: HookData{
				Event:     "PostToolUse",
				Timestamp: time.Now().Format(time.RFC3339),
				SessionID: "test-session-123",
				Filename:  "activity-monitor",
				Data: map[string]interface{}{
					"response": 456, // Should be string, not number
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "no response content in request",
			expectSuccess:  false,
		},
		{
			name:   "response with complex tool calls",
			method: http.MethodPost,
			payload: HookData{
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
			},
			expectedStatus: http.StatusCreated,
			expectSuccess:  true,
			validateData: func(t *testing.T, data map[string]interface{}) {
				if data["type"] != "response" {
					t.Errorf("Expected type 'response', got %v", data["type"])
				}
				if data["session_id"] != "test-session-789" {
					t.Errorf("Expected session_id 'test-session-789', got %v", data["session_id"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			db := setupTestDB(t)
			defer db.Close()
			
			handler := NewResponseHandler(db)
			
			// Prepare request
			var req *http.Request
			if tt.payload == nil {
				req = httptest.NewRequest(tt.method, "/messages/response", nil)
			} else if str, ok := tt.payload.(string); ok {
				// Handle string payload (invalid JSON case)
				req = httptest.NewRequest(tt.method, "/messages/response", bytes.NewBufferString(str))
			} else {
				// Handle struct payload
				payload, err := json.Marshal(tt.payload)
				if err != nil {
					t.Fatalf("Failed to marshal test payload: %v", err)
				}
				req = httptest.NewRequest(tt.method, "/messages/response", bytes.NewBuffer(payload))
			}
			
			if tt.method == http.MethodPost {
				req.Header.Set("Content-Type", "application/json")
			}
			
			// Execute request
			w := httptest.NewRecorder()
			handler.HandleResponseSubmit(w, req)
			
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
			"response":        "Assistant response without prior prompt",
			"cwd":             "/new/directory",
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