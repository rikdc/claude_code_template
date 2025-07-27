package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/claude-code-template/prompt-manager/internal/database"
	"github.com/claude-code-template/prompt-manager/internal/validation"
	"github.com/gorilla/mux"
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
	// Parse and validate pagination parameters
	page, perPage, err := validation.ParseAndValidatePage(
		r.URL.Query().Get("page"),
		r.URL.Query().Get("per_page"),
	)
	if err != nil {
		if validation.IsValidationError(err) {
			errorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		errorResponse(w, "Invalid pagination parameters", http.StatusBadRequest)
		return
	}

	offset := (page - 1) * perPage

	conversations, err := s.db.ListConversations(perPage, offset)
	if err != nil {
		errorResponse(w, fmt.Sprintf("Failed to list conversations: %v", err), http.StatusInternalServerError)
		return
	}

	// Get total count for pagination
	totalCount, err := s.db.GetConversationCount()
	if err != nil {
		errorResponse(w, fmt.Sprintf("Failed to get conversation count: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to summaries for list view
	summaries := ConvertConversationsToSummaries(conversations)

	totalPages := (totalCount + perPage - 1) / perPage // Calculate total pages with ceiling division
	meta := &Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      totalCount,
		TotalPages: totalPages,
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

	id, err := validation.ParseAndValidateID(idStr, "conversation_id")
	if err != nil {
		if validation.IsValidationError(err) {
			errorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		errorResponse(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	conv, err := s.db.GetConversationWithMessages(id)
	if err != nil {
		if errors.Is(err, database.ErrConversationNotFound) {
			errorResponse(w, "Conversation not found", http.StatusNotFound)
			return
		}
		errorResponse(w, fmt.Sprintf("Failed to get conversation: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert database models to API models
	apiConv, err := ConvertConversationWithMessages(conv)
	if err != nil {
		errorResponse(w, fmt.Sprintf("Failed to convert conversation: %v", err), http.StatusInternalServerError)
		return
	}

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

	// Validate session ID
	if err := validation.ValidateSessionID(req.SessionID); err != nil {
		if validation.IsValidationError(err) {
			errorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		errorResponse(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Validate title
	if err := validation.ValidateTitle(req.Title); err != nil {
		if validation.IsValidationError(err) {
			errorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		errorResponse(w, "Invalid title", http.StatusBadRequest)
		return
	}

	// Validate paths
	if err := validation.ValidatePath(req.WorkingDirectory); err != nil {
		if validation.IsValidationError(err) {
			errorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		errorResponse(w, "Invalid working directory path", http.StatusBadRequest)
		return
	}

	if err := validation.ValidatePath(req.TranscriptPath); err != nil {
		if validation.IsValidationError(err) {
			errorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		errorResponse(w, "Invalid transcript path", http.StatusBadRequest)
		return
	}

	// Sanitize strings
	if req.Title != nil {
		sanitized := validation.SanitizeString(*req.Title, validation.MaxTitleLength)
		req.Title = &sanitized
	}
	if req.WorkingDirectory != nil {
		sanitized := validation.SanitizeString(*req.WorkingDirectory, validation.MaxPathLength)
		req.WorkingDirectory = &sanitized
	}
	if req.TranscriptPath != nil {
		sanitized := validation.SanitizeString(*req.TranscriptPath, validation.MaxPathLength)
		req.TranscriptPath = &sanitized
	}

	conv, err := s.db.CreateConversation(req.SessionID, req.Title, req.WorkingDirectory, req.TranscriptPath)
	if err != nil {
		errorResponse(w, fmt.Sprintf("Failed to create conversation: %v", err), http.StatusInternalServerError)
		return
	}

	apiConv := ConvertConversation(conv)

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

	id, err := validation.ParseAndValidateID(idStr, "conversation_id")
	if err != nil {
		if validation.IsValidationError(err) {
			errorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
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

	// Validate title
	if err := validation.ValidateTitle(&req.Title); err != nil {
		if validation.IsValidationError(err) {
			errorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		errorResponse(w, "Invalid title", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		errorResponse(w, "title is required", http.StatusBadRequest)
		return
	}

	// Sanitize title
	req.Title = validation.SanitizeString(req.Title, validation.MaxTitleLength)

	if err := s.db.UpdateConversationTitle(id, req.Title); err != nil {
		if errors.Is(err, database.ErrConversationNotFound) {
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

	apiConv := ConvertConversation(conv)

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

	id, err := validation.ParseAndValidateID(idStr, "conversation_id")
	if err != nil {
		if validation.IsValidationError(err) {
			errorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		errorResponse(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	if err := s.db.DeleteConversation(id); err != nil {
		if errors.Is(err, database.ErrConversationNotFound) {
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

	id, err := validation.ParseAndValidateID(idStr, "conversation_id")
	if err != nil {
		if validation.IsValidationError(err) {
			errorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
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

	// Validate rating
	if err := validation.ValidateRating(req.Rating); err != nil {
		if validation.IsValidationError(err) {
			errorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		errorResponse(w, "Invalid rating", http.StatusBadRequest)
		return
	}

	// Validate comment
	if err := validation.ValidateComment(req.Comment); err != nil {
		if validation.IsValidationError(err) {
			errorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		errorResponse(w, "Invalid comment", http.StatusBadRequest)
		return
	}

	// Sanitize comment
	if req.Comment != nil {
		sanitized := validation.SanitizeString(*req.Comment, validation.MaxCommentLength)
		req.Comment = &sanitized
	}

	rating, err := s.db.CreateConversationRating(id, req.Rating, req.Comment)
	if err != nil {
		errorResponse(w, fmt.Sprintf("Failed to create rating: %v", err), http.StatusInternalServerError)
		return
	}

	apiRating := ConvertRating(rating)

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

	id, err := validation.ParseAndValidateID(idStr, "conversation_id")
	if err != nil {
		if validation.IsValidationError(err) {
			errorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		errorResponse(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	ratings, err := s.db.GetConversationRatings(id)
	if err != nil {
		errorResponse(w, fmt.Sprintf("Failed to get ratings: %v", err), http.StatusInternalServerError)
		return
	}

	apiRatings := ConvertRatings(ratings)

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

	id, err := validation.ParseAndValidateID(idStr, "rating_id")
	if err != nil {
		if validation.IsValidationError(err) {
			errorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
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

	// Validate rating
	if err := validation.ValidateRating(req.Rating); err != nil {
		if validation.IsValidationError(err) {
			errorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		errorResponse(w, "Invalid rating", http.StatusBadRequest)
		return
	}

	// Validate comment
	if err := validation.ValidateComment(req.Comment); err != nil {
		if validation.IsValidationError(err) {
			errorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		errorResponse(w, "Invalid comment", http.StatusBadRequest)
		return
	}

	// Sanitize comment
	if req.Comment != nil {
		sanitized := validation.SanitizeString(*req.Comment, validation.MaxCommentLength)
		req.Comment = &sanitized
	}

	if err := s.db.UpdateRating(id, req.Rating, req.Comment); err != nil {
		if errors.Is(err, database.ErrRatingNotFound) {
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

	apiRating := ConvertRating(rating)

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

	id, err := validation.ParseAndValidateID(idStr, "rating_id")
	if err != nil {
		if validation.IsValidationError(err) {
			errorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		errorResponse(w, "Invalid rating ID", http.StatusBadRequest)
		return
	}

	if err := s.db.DeleteRating(id); err != nil {
		if errors.Is(err, database.ErrRatingNotFound) {
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

