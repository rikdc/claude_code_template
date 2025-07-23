package main

import (
	"fmt"
	"log"

	"github.com/claude-code-template/prompt-manager/internal/database"
)

func main() {
	config := database.DefaultConfig()
	db, err := database.New(config)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.RunMigrations(config.MigrationsDir)
	if err != nil {
		log.Fatal(err)
	}

	convs, err := db.ListConversations(10, 0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Conversations: %d\n", len(convs))
	if len(convs) > 0 {
		fmt.Printf("Session: %s\n", convs[0].SessionID)
		msgs, err := db.GetMessagesByConversation(convs[0].ID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Messages: %d\n", len(msgs))
		if len(msgs) > 0 {
			fmt.Printf("Content: %s\n", msgs[0].Content)
		}
	}
}