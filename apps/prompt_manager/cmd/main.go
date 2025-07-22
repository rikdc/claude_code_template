package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/claude-code-template/prompt-manager/internal/api"
	"github.com/claude-code-template/prompt-manager/internal/database"
)

const (
	DefaultPort = "8080"
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

	// Setup routes
	router := mux.NewRouter()
	
	// Health check endpoint
	router.HandleFunc("/health", server.HealthHandler).Methods("GET")
	
	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	
	// Conversation endpoints
	api.HandleFunc("/conversations", server.ListConversationsHandler).Methods("GET")
	api.HandleFunc("/conversations", server.CreateConversationHandler).Methods("POST")
	api.HandleFunc("/conversations/{id}", server.GetConversationHandler).Methods("GET")
	api.HandleFunc("/conversations/{id}", server.UpdateConversationHandler).Methods("PUT")
	api.HandleFunc("/conversations/{id}", server.DeleteConversationHandler).Methods("DELETE")
	
	// Rating endpoints
	api.HandleFunc("/conversations/{id}/ratings", server.CreateConversationRatingHandler).Methods("POST")
	api.HandleFunc("/conversations/{id}/ratings", server.GetConversationRatingsHandler).Methods("GET")
	api.HandleFunc("/ratings/{id}", server.UpdateRatingHandler).Methods("PUT")
	api.HandleFunc("/ratings/{id}", server.DeleteRatingHandler).Methods("DELETE")
	api.HandleFunc("/ratings/stats", server.GetRatingStatsHandler).Methods("GET")
	
	fmt.Printf("Starting Prompt Manager server on port %s\n", port)
	fmt.Printf("Database: %s\n", config.DatabasePath)
	log.Fatal(http.ListenAndServe(":"+port, router))
}