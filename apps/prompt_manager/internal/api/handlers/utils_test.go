package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/claude-code-template/prompt-manager/internal/database"
)

func TestGetOrCreateConversation(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		data      map[string]interface{}
		setup     func(db *database.DB) error
		expectNew bool
		wantErr   bool
	}{
		{
			name:      "creates new conversation when none exists",
			sessionID: "new-session-123",
			data: map[string]interface{}{
				"cwd":             "/test/directory",
				"transcript_path": "/test/transcript.md",
			},
			expectNew: true,
			wantErr:   false,
		},
		{
			name:      "returns existing conversation when session matches",
			sessionID: "existing-session-456",
			data: map[string]interface{}{
				"cwd": "/test/directory",
			},
			setup: func(db *database.DB) error {
				_, err := db.CreateConversation("existing-session-456", nil, nil, nil)
				return err
			},
			expectNew: false,
			wantErr:   false,
		},
		{
			name:      "creates conversation with nil context data",
			sessionID: "nil-data-session",
			data:      map[string]interface{}{},
			expectNew: true,
			wantErr:   false,
		},
		{
			name:      "creates conversation with empty string values",
			sessionID: "empty-strings-session",
			data: map[string]interface{}{
				"cwd":             "",
				"transcript_path": "",
			},
			expectNew: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer db.Close()

			// Setup existing data if needed
			if tt.setup != nil {
				if err := tt.setup(db); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Get initial conversation count
			initialConvs, err := db.ListConversations(100, 0)
			if err != nil {
				t.Fatalf("Failed to list initial conversations: %v", err)
			}
			initialCount := len(initialConvs)

			// Test the function
			conversationID, err := GetOrCreateConversation(db, tt.sessionID, tt.data)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrCreateConversation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return // Skip further checks if error was expected
			}

			// Verify conversation ID is valid
			if conversationID <= 0 {
				t.Errorf("Expected positive conversation ID, got %d", conversationID)
			}

			// Check if new conversation was created or existing one returned
			finalConvs, err := db.ListConversations(100, 0)
			if err != nil {
				t.Fatalf("Failed to list final conversations: %v", err)
			}
			finalCount := len(finalConvs)

			if tt.expectNew {
				if finalCount != initialCount+1 {
					t.Errorf("Expected new conversation to be created. Initial: %d, Final: %d", initialCount, finalCount)
				}
			} else {
				if finalCount != initialCount {
					t.Errorf("Expected no new conversation. Initial: %d, Final: %d", initialCount, finalCount)
				}
			}

			// Verify the conversation has correct session ID
			found := false
			for _, conv := range finalConvs {
				if conv.ID == conversationID {
					if conv.SessionID != tt.sessionID {
						t.Errorf("Expected session ID %s, got %s", tt.sessionID, conv.SessionID)
					}
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Could not find conversation with ID %d", conversationID)
			}
		})
	}
}

func TestExtractStringFromData(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		expected *string
	}{
		{
			name: "extracts existing string value",
			data: map[string]interface{}{
				"test_key": "test_value",
			},
			key:      "test_key",
			expected: stringPtr("test_value"),
		},
		{
			name: "returns nil for non-existent key",
			data: map[string]interface{}{
				"other_key": "other_value",
			},
			key:      "missing_key",
			expected: nil,
		},
		{
			name: "returns nil for empty string",
			data: map[string]interface{}{
				"empty_key": "",
			},
			key:      "empty_key",
			expected: nil,
		},
		{
			name: "returns nil for non-string value",
			data: map[string]interface{}{
				"number_key": 123,
			},
			key:      "number_key",
			expected: nil,
		},
		{
			name: "returns nil for boolean value",
			data: map[string]interface{}{
				"bool_key": true,
			},
			key:      "bool_key",
			expected: nil,
		},
		{
			name: "returns nil for nil value",
			data: map[string]interface{}{
				"nil_key": nil,
			},
			key:      "nil_key",
			expected: nil,
		},
		{
			name: "extracts string with spaces",
			data: map[string]interface{}{
				"space_key": "  value with spaces  ",
			},
			key:      "space_key",
			expected: stringPtr("  value with spaces  "),
		},
		{
			name:     "handles nil data map",
			data:     nil,
			key:      "any_key",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ExtractStringFromData(tt.data, tt.key)

			// Compare pointers and values
			if tt.expected == nil && result != nil {
				t.Errorf("Expected nil, got %v", *result)
			} else if tt.expected != nil && result == nil {
				t.Errorf("Expected %v, got nil", *tt.expected)
			} else if tt.expected != nil && result != nil && *tt.expected != *result {
				t.Errorf("Expected %v, got %v", *tt.expected, *result)
			}
		})
	}
}

func TestErrorResponse(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		statusCode     int
		expectedStatus int
		expectedBody   bool // Whether to check body content
	}{
		{
			name:           "sends bad request error",
			message:        "Invalid request",
			statusCode:     http.StatusBadRequest,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   true,
		},
		{
			name:           "sends internal server error",
			message:        "Database connection failed",
			statusCode:     http.StatusInternalServerError,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   true,
		},
		{
			name:           "sends not found error",
			message:        "Resource not found",
			statusCode:     http.StatusNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   true,
		},
		{
			name:           "handles empty message",
			message:        "",
			statusCode:     http.StatusBadRequest,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   true,
		},
		{
			name:           "handles long message",
			message:        "This is a very long error message that contains multiple sentences and should be handled properly by the error response function without any issues.",
			statusCode:     http.StatusUnprocessableEntity,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create response recorder
			w := httptest.NewRecorder()

			// Call ErrorResponse
			ErrorResponse(w, tt.message, tt.statusCode)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check content type
			expectedContentType := "application/json"
			if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
				t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
			}

			// Check response body if requested
			if tt.expectedBody {
				var response APIResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response body: %v", err)
				}

				// Check success flag
				if response.Success != false {
					t.Errorf("Expected Success to be false, got %v", response.Success)
				}

				// Check error message
				if response.Error == nil {
					t.Error("Expected Error field to be set, got nil")
				} else if *response.Error != tt.message {
					t.Errorf("Expected error message %q, got %q", tt.message, *response.Error)
				}

				// Check that Data field is nil for errors
				if response.Data != nil {
					t.Errorf("Expected Data field to be nil for error response, got %v", response.Data)
				}
			}
		})
	}
}

// Helper function to create string pointers for tests
func stringPtr(s string) *string {
	return &s
}