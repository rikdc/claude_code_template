package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/claude-code-template/prompt-manager/internal/api"
	"github.com/claude-code-template/prompt-manager/internal/api/handlers"
	"github.com/claude-code-template/prompt-manager/internal/database"
)

const (
	DefaultPort = "8082"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	// Initialize database
	config := database.DefaultConfig()
	db, err := database.New(config)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(config.MigrationsDir); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize API server
	server := api.NewServer(db)

	// Initialize message handlers
	promptHandler := handlers.NewPromptHandler(db)
	responseHandler := handlers.NewResponseHandler(db)
	sessionHandler := handlers.NewSessionHandler(db)

	// Setup routes
	router := mux.NewRouter()
	
	// Health check endpoint
	router.HandleFunc("/health", server.HealthHandler).Methods("GET")
	
	// Message endpoints for hook processing
	router.HandleFunc("/messages/prompt", promptHandler.HandlePromptSubmit).Methods("POST")
	router.HandleFunc("/messages/response", responseHandler.HandleResponseSubmit).Methods("POST")
	router.HandleFunc("/messages/session", sessionHandler.HandleSessionEvent).Methods("POST")
	
	// Conversation endpoints (at root level for activity monitor compatibility)
	router.HandleFunc("/conversations", server.ListConversationsHandler).Methods("GET")
	router.HandleFunc("/conversations", server.CreateConversationHandler).Methods("POST")
	router.HandleFunc("/conversations/{id}", server.GetConversationHandler).Methods("GET")
	router.HandleFunc("/conversations/{id}", server.UpdateConversationHandler).Methods("PUT")
	router.HandleFunc("/conversations/{id}", server.DeleteConversationHandler).Methods("DELETE")
	
	// Rating endpoints
	router.HandleFunc("/conversations/{id}/ratings", server.CreateConversationRatingHandler).Methods("POST")
	router.HandleFunc("/conversations/{id}/ratings", server.GetConversationRatingsHandler).Methods("GET")
	router.HandleFunc("/ratings/{id}", server.UpdateRatingHandler).Methods("PUT")
	router.HandleFunc("/ratings/{id}", server.DeleteRatingHandler).Methods("DELETE")
	router.HandleFunc("/ratings/stats", server.GetRatingStatsHandler).Methods("GET")
	
	fmt.Printf("Starting Prompt Manager server on port %s\n", port)
	fmt.Printf("Database: %s\n", config.DatabasePath)
	log.Fatal(http.ListenAndServe(":"+port, router))
}