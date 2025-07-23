package database

import (
	"database/sql"
	"fmt"
	"time"
)

// Rating represents a rating record
type Rating struct {
	ID             int       `json:"id"`
	ConversationID *int      `json:"conversation_id"`
	MessageID      *int      `json:"message_id"`
	Rating         int       `json:"rating"`
	Comment        *string   `json:"comment"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateConversationRating creates a rating for a conversation
func (db *DB) CreateConversationRating(conversationID int, rating int, comment *string) (*Rating, error) {
	if rating < 1 || rating > 5 {
		return nil, fmt.Errorf("rating must be between 1 and 5")
	}

	query := `
	INSERT INTO ratings (conversation_id, rating, comment)
	VALUES (?, ?, ?)
	RETURNING id, conversation_id, message_id, rating, comment, created_at, updated_at`

	var r Rating
	err := db.conn.QueryRow(query, conversationID, rating, comment).Scan(
		&r.ID, &r.ConversationID, &r.MessageID, &r.Rating, &r.Comment, &r.CreatedAt, &r.UpdatedAt,
	)
	
	if err != nil {
		// Fallback for SQLite versions that don't support RETURNING
		result, err := db.conn.Exec(
			"INSERT INTO ratings (conversation_id, rating, comment) VALUES (?, ?, ?)",
			conversationID, rating, comment,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert rating: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("failed to get last insert ID: %w", err)
		}

		return db.GetRating(int(id))
	}

	return &r, nil
}

// CreateMessageRating creates a rating for a message
func (db *DB) CreateMessageRating(messageID int, rating int, comment *string) (*Rating, error) {
	if rating < 1 || rating > 5 {
		return nil, fmt.Errorf("rating must be between 1 and 5")
	}

	query := `
	INSERT INTO ratings (message_id, rating, comment)
	VALUES (?, ?, ?)
	RETURNING id, conversation_id, message_id, rating, comment, created_at, updated_at`

	var r Rating
	err := db.conn.QueryRow(query, messageID, rating, comment).Scan(
		&r.ID, &r.ConversationID, &r.MessageID, &r.Rating, &r.Comment, &r.CreatedAt, &r.UpdatedAt,
	)
	
	if err != nil {
		// Fallback for SQLite versions that don't support RETURNING
		result, err := db.conn.Exec(
			"INSERT INTO ratings (message_id, rating, comment) VALUES (?, ?, ?)",
			messageID, rating, comment,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert rating: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("failed to get last insert ID: %w", err)
		}

		return db.GetRating(int(id))
	}

	return &r, nil
}

// GetRating retrieves a rating by ID
func (db *DB) GetRating(id int) (*Rating, error) {
	query := `
	SELECT id, conversation_id, message_id, rating, comment, created_at, updated_at
	FROM ratings WHERE id = ?`

	var r Rating
	err := db.conn.QueryRow(query, id).Scan(
		&r.ID, &r.ConversationID, &r.MessageID, &r.Rating, &r.Comment, &r.CreatedAt, &r.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("rating not found")
		}
		return nil, fmt.Errorf("failed to get rating: %w", err)
	}

	return &r, nil
}

// GetConversationRatings retrieves all ratings for a conversation
func (db *DB) GetConversationRatings(conversationID int) ([]Rating, error) {
	query := `
	SELECT id, conversation_id, message_id, rating, comment, created_at, updated_at
	FROM ratings 
	WHERE conversation_id = ?
	ORDER BY created_at DESC`

	rows, err := db.conn.Query(query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation ratings: %w", err)
	}
	defer rows.Close()

	var ratings []Rating
	for rows.Next() {
		var r Rating
		err := rows.Scan(
			&r.ID, &r.ConversationID, &r.MessageID, &r.Rating, &r.Comment, &r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rating: %w", err)
		}
		ratings = append(ratings, r)
	}

	return ratings, nil
}

// GetMessageRatings retrieves all ratings for a message
func (db *DB) GetMessageRatings(messageID int) ([]Rating, error) {
	query := `
	SELECT id, conversation_id, message_id, rating, comment, created_at, updated_at
	FROM ratings 
	WHERE message_id = ?
	ORDER BY created_at DESC`

	rows, err := db.conn.Query(query, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message ratings: %w", err)
	}
	defer rows.Close()

	var ratings []Rating
	for rows.Next() {
		var r Rating
		err := rows.Scan(
			&r.ID, &r.ConversationID, &r.MessageID, &r.Rating, &r.Comment, &r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rating: %w", err)
		}
		ratings = append(ratings, r)
	}

	return ratings, nil
}

// UpdateRating updates a rating's score and comment
func (db *DB) UpdateRating(id int, rating int, comment *string) error {
	if rating < 1 || rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}

	query := "UPDATE ratings SET rating = ?, comment = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	result, err := db.conn.Exec(query, rating, comment, id)
	if err != nil {
		return fmt.Errorf("failed to update rating: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("rating not found")
	}

	return nil
}

// DeleteRating deletes a rating
func (db *DB) DeleteRating(id int) error {
	query := "DELETE FROM ratings WHERE id = ?"
	result, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete rating: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("rating not found")
	}

	return nil
}

// GetRatingStats returns rating statistics
func (db *DB) GetRatingStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Average rating
	var avgRating float64
	err := db.conn.QueryRow("SELECT AVG(rating) FROM ratings").Scan(&avgRating)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get average rating: %w", err)
	}
	stats["average_rating"] = avgRating

	// Rating distribution
	rows, err := db.conn.Query("SELECT rating, COUNT(*) FROM ratings GROUP BY rating ORDER BY rating")
	if err != nil {
		return nil, fmt.Errorf("failed to get rating distribution: %w", err)
	}
	defer rows.Close()

	distribution := make(map[int]int)
	for rows.Next() {
		var rating, count int
		if err := rows.Scan(&rating, &count); err != nil {
			return nil, fmt.Errorf("failed to scan rating distribution: %w", err)
		}
		distribution[rating] = count
	}
	stats["distribution"] = distribution

	// Total ratings
	var totalRatings int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM ratings").Scan(&totalRatings)
	if err != nil {
		return nil, fmt.Errorf("failed to count ratings: %w", err)
	}
	stats["total_ratings"] = totalRatings

	return stats, nil
}