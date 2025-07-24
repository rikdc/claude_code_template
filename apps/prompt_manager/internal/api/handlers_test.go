package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/claude-code-template/prompt-manager/internal/database"
)

func setupTestServer(t *testing.T) *Server {
	// Create temp database file
	tmpfile, err := os.CreateTemp("", "test_api_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()

	config := &database.Config{
		DatabasePath:  tmpfile.Name(),
		MigrationsDir: "../../database/migrations",
	}

	db, err := database.New(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Run migrations
	err = db.RunMigrations(config.MigrationsDir)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	server := NewServer(db)

	// Cleanup function
	t.Cleanup(func() {
		db.Close()
		os.Remove(tmpfile.Name())
	})

	return server
}

func TestHealthHandler(t *testing.T) {
	server := setupTestServer(t)

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HealthHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success=true in response")
	}

	// Check that database stats are included
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected response.Data to be a map")
	}

	if data["status"] != "healthy" {
		t.Errorf("Expected status=healthy, got %v", data["status"])
	}

	if _, exists := data["database"]; !exists {
		t.Error("Expected database stats in response")
	}
}

func TestCreateConversation(t *testing.T) {
	server := setupTestServer(t)

	reqBody := map[string]interface{}{
		"session_id": "test-session-123",
		"title":      "Test Conversation",
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/api/v1/conversations", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.CreateConversationHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var response APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success=true in response")
	}

	// Verify conversation data
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected response.Data to be a map")
	}

	if data["session_id"] != "test-session-123" {
		t.Errorf("Expected session_id=test-session-123, got %v", data["session_id"])
	}

	if data["title"] != "Test Conversation" {
		t.Errorf("Expected title=Test Conversation, got %v", data["title"])
	}
}

func TestCreateConversationInvalidRequest(t *testing.T) {
	server := setupTestServer(t)

	// Test missing session_id
	reqBody := map[string]interface{}{
		"title": "Test Conversation",
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/api/v1/conversations", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.CreateConversationHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	var response APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Success {
		t.Error("Expected success=false in response")
	}

	if response.Error == nil {
		t.Error("Expected error message in response")
	}
}

func TestGetConversation(t *testing.T) {
	server := setupTestServer(t)

	// Create a conversation first
	conv, err := server.db.CreateConversation("test-session", stringPtr("Test Conversation"), nil, nil)
	if err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	// Add a message
	_, err = server.db.CreateMessage(conv.ID, "prompt", "Hello world", nil, nil)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	req, err := http.NewRequest("GET", "/api/v1/conversations/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	
	// Use mux router to handle path variables
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/conversations/{id}", server.GetConversationHandler)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success=true in response")
	}

	// Verify conversation data
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected response.Data to be a map")
	}

	if data["session_id"] != "test-session" {
		t.Errorf("Expected session_id=test-session, got %v", data["session_id"])
	}

	// Check messages array
	messages, ok := data["messages"].([]interface{})
	if !ok {
		t.Fatal("Expected messages to be an array")
	}

	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

func TestGetConversationNotFound(t *testing.T) {
	server := setupTestServer(t)

	req, err := http.NewRequest("GET", "/api/v1/conversations/999", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/conversations/{id}", server.GetConversationHandler)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	var response APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Success {
		t.Error("Expected success=false in response")
	}
}

func TestListConversations(t *testing.T) {
	server := setupTestServer(t)

	// Create test conversations
	conv1, err := server.db.CreateConversation("session-1", stringPtr("Conversation 1"), nil, nil)
	if err != nil {
		t.Fatalf("Failed to create conversation 1: %v", err)
	}

	conv2, err := server.db.CreateConversation("session-2", stringPtr("Conversation 2"), nil, nil)
	if err != nil {
		t.Fatalf("Failed to create conversation 2: %v", err)
	}

	req, err := http.NewRequest("GET", "/api/v1/conversations", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.ListConversationsHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success=true in response")
	}

	// Verify conversations data
	conversations, ok := response.Data.([]interface{})
	if !ok {
		t.Fatal("Expected response.Data to be an array")
	}

	if len(conversations) != 2 {
		t.Errorf("Expected 2 conversations, got %d", len(conversations))
	}

	// Check pagination meta
	if response.Meta == nil {
		t.Error("Expected meta information in response")
	}

	_ = conv1 // Suppress unused variable warning
	_ = conv2
}

func TestCreateConversationRating(t *testing.T) {
	server := setupTestServer(t)

	// Create a conversation first
	conv, err := server.db.CreateConversation("test-session", nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	reqBody := map[string]interface{}{
		"rating":  5,
		"comment": "Great conversation!",
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/api/v1/conversations/1/ratings", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/conversations/{id}/ratings", server.CreateConversationRatingHandler)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var response APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success=true in response")
	}

	// Verify rating data
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected response.Data to be a map")
	}

	if data["rating"] != float64(5) { // JSON unmarshaling converts numbers to float64
		t.Errorf("Expected rating=5, got %v", data["rating"])
	}

	if data["comment"] != "Great conversation!" {
		t.Errorf("Expected comment='Great conversation!', got %v", data["comment"])
	}

	_ = conv // Suppress unused variable warning
}

func TestCreateRatingInvalidRange(t *testing.T) {
	server := setupTestServer(t)

	// Create a conversation first
	_, err := server.db.CreateConversation("test-session", nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	reqBody := map[string]interface{}{
		"rating": 10, // Invalid rating > 5
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/api/v1/conversations/1/ratings", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/conversations/{id}/ratings", server.CreateConversationRatingHandler)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	var response APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Success {
		t.Error("Expected success=false in response")
	}

	if response.Error == nil {
		t.Error("Expected error message in response")
	}
}

func TestGetRatingStats(t *testing.T) {
	server := setupTestServer(t)

	// Create conversations and ratings
	conv, err := server.db.CreateConversation("test-session", nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	_, err = server.db.CreateConversationRating(conv.ID, 5, nil)
	if err != nil {
		t.Fatalf("Failed to create rating: %v", err)
	}

	_, err = server.db.CreateConversationRating(conv.ID, 3, nil)
	if err != nil {
		t.Fatalf("Failed to create rating: %v", err)
	}

	req, err := http.NewRequest("GET", "/api/v1/ratings/stats", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.GetRatingStatsHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success=true in response")
	}

	// Verify stats data
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected response.Data to be a map")
	}

	if data["total_ratings"] != float64(2) {
		t.Errorf("Expected total_ratings=2, got %v", data["total_ratings"])
	}

	if data["average_rating"] != float64(4) { // (5+3)/2 = 4
		t.Errorf("Expected average_rating=4, got %v", data["average_rating"])
	}

	// Check distribution
	if _, exists := data["distribution"]; !exists {
		t.Error("Expected distribution in stats")
	}
}

