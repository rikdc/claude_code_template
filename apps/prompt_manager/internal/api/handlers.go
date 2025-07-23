package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/claude-code-template/prompt-manager/internal/database"
	"github.com/claude-code-template/prompt-manager/internal/models"
)

// Server holds the database connection and provides HTTP handlers
type Server struct {
	db *database.DB
}

// NewServer creates a new API server
func NewServer(db *database.DB) *Server {
	return &Server{db: db}
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *string     `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta provides pagination and additional response metadata
type Meta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// Error response helpers
func errorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := APIResponse{
		Success: false,
		Error:   &message,
	}
	
	json.NewEncoder(w).Encode(response)
}

func successResponse(w http.ResponseWriter, data interface{}, meta *Meta) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
	
	json.NewEncoder(w).Encode(response)
}

// Health check handler
func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	// Check database health
	if err := s.db.Health(); err != nil {
		errorResponse(w, fmt.Sprintf("Database unhealthy: %v", err), http.StatusServiceUnavailable)
		return
	}
	
	// Get database stats
	stats, err := s.db.Stats()
	if err != nil {
		errorResponse(w, fmt.Sprintf("Failed to get stats: %v", err), http.StatusInternalServerError)
		return
	}
	
	healthData := map[string]interface{}{
		"status":    "healthy",
		"service":   "prompt-manager",
		"timestamp": time.Now().UTC(),
		"database":  stats,
	}
	
	successResponse(w, healthData, nil)
}

// Conversation handlers

// ListConversationsHandler returns a paginated list of conversations
func (s *Server) ListConversationsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page := 1
	perPage := 20
	
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}
	
	offset := (page - 1) * perPage
	
	conversations, err := s.db.ListConversations(perPage, offset)
	if err != nil {
		errorResponse(w, fmt.Sprintf("Failed to list conversations: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Convert to summaries for list view
	summaries := make([]models.ConversationSummary, len(conversations))
	for i, conv := range conversations {
		// Create a model conversation to use ToSummary method
		modelConv := models.Conversation{
			ID:               conv.ID,
			SessionID:        conv.SessionID,
			Title:            conv.Title,
			CreatedAt:        conv.CreatedAt,
			UpdatedAt:        conv.UpdatedAt,
			PromptCount:      conv.PromptCount,
			TotalCharacters:  conv.TotalCharacters,
			WorkingDirectory: conv.WorkingDirectory,
			TranscriptPath:   conv.TranscriptPath,
		}
		summaries[i] = modelConv.ToSummary()
	}
	
	meta := &Meta{
		Page:    page,
		PerPage: perPage,
	}
	
	successResponse(w, summaries, meta)
}

// GetConversationHandler returns a specific conversation with messages
func (s *Server) GetConversationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, exists := vars["id"]
	if !exists {
		errorResponse(w, "Conversation ID is required", http.StatusBadRequest)
		return
	}
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorResponse(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}
	
	conv, err := s.db.GetConversationWithMessages(id)
	if err != nil {
		if err.Error() == "conversation not found" {
			errorResponse(w, "Conversation not found", http.StatusNotFound)
			return
		}
		errorResponse(w, fmt.Sprintf("Failed to get conversation: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Convert database models to API models
	apiConv := models.Conversation{
		ID:               conv.ID,
		SessionID:        conv.SessionID,
		Title:            conv.Title,
		CreatedAt:        conv.CreatedAt,
		UpdatedAt:        conv.UpdatedAt,
		PromptCount:      conv.PromptCount,
		TotalCharacters:  conv.TotalCharacters,
		WorkingDirectory: conv.WorkingDirectory,
		TranscriptPath:   conv.TranscriptPath,
	}
	
	// Convert messages
	apiMessages := make([]models.Message, len(conv.Messages))
	for i, msg := range conv.Messages {
		toolCalls, _ := models.UnmarshalToolCalls(msg.ToolCalls)
		
		apiMessages[i] = models.Message{
			ID:             msg.ID,
			ConversationID: msg.ConversationID,
			MessageType:    models.MessageType(msg.MessageType),
			Content:        msg.Content,
			CharacterCount: msg.CharacterCount,
			Timestamp:      msg.Timestamp,
			ToolCalls:      toolCalls,
			ExecutionTime:  msg.ExecutionTime,
		}
	}
	apiConv.Messages = apiMessages
	
	successResponse(w, apiConv, nil)
}

// CreateConversationHandler creates a new conversation
func (s *Server) CreateConversationHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID        string  `json:"session_id"`
		Title            *string `json:"title"`
		WorkingDirectory *string `json:"working_directory"`
		TranscriptPath   *string `json:"transcript_path"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}
	
	if req.SessionID == "" {
		errorResponse(w, "session_id is required", http.StatusBadRequest)
		return
	}
	
	conv, err := s.db.CreateConversation(req.SessionID, req.Title, req.WorkingDirectory, req.TranscriptPath)
	if err != nil {
		errorResponse(w, fmt.Sprintf("Failed to create conversation: %v", err), http.StatusInternalServerError)
		return
	}
	
	apiConv := models.Conversation{
		ID:               conv.ID,
		SessionID:        conv.SessionID,
		Title:            conv.Title,
		CreatedAt:        conv.CreatedAt,
		UpdatedAt:        conv.UpdatedAt,
		PromptCount:      conv.PromptCount,
		TotalCharacters:  conv.TotalCharacters,
		WorkingDirectory: conv.WorkingDirectory,
		TranscriptPath:   conv.TranscriptPath,
	}
	
	w.WriteHeader(http.StatusCreated)
	successResponse(w, apiConv, nil)
}

// UpdateConversationHandler updates a conversation's title
func (s *Server) UpdateConversationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, exists := vars["id"]
	if !exists {
		errorResponse(w, "Conversation ID is required", http.StatusBadRequest)
		return
	}
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorResponse(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		Title string `json:"title"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}
	
	if req.Title == "" {
		errorResponse(w, "title is required", http.StatusBadRequest)
		return
	}
	
	if err := s.db.UpdateConversationTitle(id, req.Title); err != nil {
		if err.Error() == "conversation not found" {
			errorResponse(w, "Conversation not found", http.StatusNotFound)
			return
		}
		errorResponse(w, fmt.Sprintf("Failed to update conversation: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Return updated conversation
	conv, err := s.db.GetConversation(id)
	if err != nil {
		errorResponse(w, fmt.Sprintf("Failed to get updated conversation: %v", err), http.StatusInternalServerError)
		return
	}
	
	apiConv := models.Conversation{
		ID:               conv.ID,
		SessionID:        conv.SessionID,
		Title:            conv.Title,
		CreatedAt:        conv.CreatedAt,
		UpdatedAt:        conv.UpdatedAt,
		PromptCount:      conv.PromptCount,
		TotalCharacters:  conv.TotalCharacters,
		WorkingDirectory: conv.WorkingDirectory,
		TranscriptPath:   conv.TranscriptPath,
	}
	
	successResponse(w, apiConv, nil)
}

// DeleteConversationHandler deletes a conversation
func (s *Server) DeleteConversationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, exists := vars["id"]
	if !exists {
		errorResponse(w, "Conversation ID is required", http.StatusBadRequest)
		return
	}
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorResponse(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}
	
	if err := s.db.DeleteConversation(id); err != nil {
		if err.Error() == "conversation not found" {
			errorResponse(w, "Conversation not found", http.StatusNotFound)
			return
		}
		errorResponse(w, fmt.Sprintf("Failed to delete conversation: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// Rating handlers

// CreateConversationRatingHandler creates a rating for a conversation
func (s *Server) CreateConversationRatingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, exists := vars["id"]
	if !exists {
		errorResponse(w, "Conversation ID is required", http.StatusBadRequest)
		return
	}
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorResponse(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		Rating  int     `json:"rating"`
		Comment *string `json:"comment"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}
	
	if req.Rating < 1 || req.Rating > 5 {
		errorResponse(w, "rating must be between 1 and 5", http.StatusBadRequest)
		return
	}
	
	rating, err := s.db.CreateConversationRating(id, req.Rating, req.Comment)
	if err != nil {
		errorResponse(w, fmt.Sprintf("Failed to create rating: %v", err), http.StatusInternalServerError)
		return
	}
	
	apiRating := models.Rating{
		ID:             rating.ID,
		ConversationID: rating.ConversationID,
		MessageID:      rating.MessageID,
		Rating:         rating.Rating,
		Comment:        rating.Comment,
		CreatedAt:      rating.CreatedAt,
		UpdatedAt:      rating.UpdatedAt,
	}
	
	w.WriteHeader(http.StatusCreated)
	successResponse(w, apiRating, nil)
}

// GetConversationRatingsHandler returns all ratings for a conversation
func (s *Server) GetConversationRatingsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, exists := vars["id"]
	if !exists {
		errorResponse(w, "Conversation ID is required", http.StatusBadRequest)
		return
	}
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorResponse(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}
	
	ratings, err := s.db.GetConversationRatings(id)
	if err != nil {
		errorResponse(w, fmt.Sprintf("Failed to get ratings: %v", err), http.StatusInternalServerError)
		return
	}
	
	apiRatings := make([]models.Rating, len(ratings))
	for i, rating := range ratings {
		apiRatings[i] = models.Rating{
			ID:             rating.ID,
			ConversationID: rating.ConversationID,
			MessageID:      rating.MessageID,
			Rating:         rating.Rating,
			Comment:        rating.Comment,
			CreatedAt:      rating.CreatedAt,
			UpdatedAt:      rating.UpdatedAt,
		}
	}
	
	successResponse(w, apiRatings, nil)
}

// UpdateRatingHandler updates a rating
func (s *Server) UpdateRatingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, exists := vars["id"]
	if !exists {
		errorResponse(w, "Rating ID is required", http.StatusBadRequest)
		return
	}
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorResponse(w, "Invalid rating ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		Rating  int     `json:"rating"`
		Comment *string `json:"comment"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}
	
	if req.Rating < 1 || req.Rating > 5 {
		errorResponse(w, "rating must be between 1 and 5", http.StatusBadRequest)
		return
	}
	
	if err := s.db.UpdateRating(id, req.Rating, req.Comment); err != nil {
		if err.Error() == "rating not found" {
			errorResponse(w, "Rating not found", http.StatusNotFound)
			return
		}
		errorResponse(w, fmt.Sprintf("Failed to update rating: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Return updated rating
	rating, err := s.db.GetRating(id)
	if err != nil {
		errorResponse(w, fmt.Sprintf("Failed to get updated rating: %v", err), http.StatusInternalServerError)
		return
	}
	
	apiRating := models.Rating{
		ID:             rating.ID,
		ConversationID: rating.ConversationID,
		MessageID:      rating.MessageID,
		Rating:         rating.Rating,
		Comment:        rating.Comment,
		CreatedAt:      rating.CreatedAt,
		UpdatedAt:      rating.UpdatedAt,
	}
	
	successResponse(w, apiRating, nil)
}

// DeleteRatingHandler deletes a rating
func (s *Server) DeleteRatingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, exists := vars["id"]
	if !exists {
		errorResponse(w, "Rating ID is required", http.StatusBadRequest)
		return
	}
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorResponse(w, "Invalid rating ID", http.StatusBadRequest)
		return
	}
	
	if err := s.db.DeleteRating(id); err != nil {
		if err.Error() == "rating not found" {
			errorResponse(w, "Rating not found", http.StatusNotFound)
			return
		}
		errorResponse(w, fmt.Sprintf("Failed to delete rating: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// GetRatingStatsHandler returns rating statistics
func (s *Server) GetRatingStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := s.db.GetRatingStats()
	if err != nil {
		errorResponse(w, fmt.Sprintf("Failed to get rating stats: %v", err), http.StatusInternalServerError)
		return
	}
	
	successResponse(w, stats, nil)
}