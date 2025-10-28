package services

import (
	"database/sql"
	"fmt"
	"time"

	"google-chat-bot/database"
)

// CreateUser adds a new user to the roster
func CreateUser(googleChatUserID, displayName, email string) (*database.User, error) {
	query := `
		INSERT INTO users (google_chat_user_id, display_name, email, is_active)
		VALUES (?, ?, ?, 1)
	`

	result, err := database.DB.Exec(query, googleChatUserID, displayName, email)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return GetUserByID(int(id))
}

// GetUserByID retrieves a user by ID
func GetUserByID(id int) (*database.User, error) {
	query := `
		SELECT id, google_chat_user_id, display_name, email, is_active,
		       joined_at, left_at, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	var user database.User
	var leftAt sql.NullTime

	err := database.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.GoogleChatUserID,
		&user.DisplayName,
		&user.Email,
		&user.IsActive,
		&user.JoinedAt,
		&leftAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if leftAt.Valid {
		user.LeftAt = &leftAt.Time
	}

	return &user, nil
}

// GetAllUsers retrieves all users
func GetAllUsers() ([]database.User, error) {
	query := `
		SELECT id, google_chat_user_id, display_name, email, is_active,
		       joined_at, left_at, created_at, updated_at
		FROM users
		ORDER BY display_name
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []database.User
	for rows.Next() {
		var user database.User
		var leftAt sql.NullTime

		err := rows.Scan(
			&user.ID,
			&user.GoogleChatUserID,
			&user.DisplayName,
			&user.Email,
			&user.IsActive,
			&user.JoinedAt,
			&leftAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		if leftAt.Valid {
			user.LeftAt = &leftAt.Time
		}

		users = append(users, user)
	}

	return users, nil
}

// GetActiveUsers retrieves all active users (not permanently deactivated)
func GetActiveUsers() ([]database.User, error) {
	query := `
		SELECT id, google_chat_user_id, display_name, email, is_active,
		       joined_at, left_at, created_at, updated_at
		FROM users
		WHERE is_active = 1
		ORDER BY display_name
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active users: %w", err)
	}
	defer rows.Close()

	var users []database.User
	for rows.Next() {
		var user database.User
		var leftAt sql.NullTime

		err := rows.Scan(
			&user.ID,
			&user.GoogleChatUserID,
			&user.DisplayName,
			&user.Email,
			&user.IsActive,
			&user.JoinedAt,
			&leftAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		if leftAt.Valid {
			user.LeftAt = &leftAt.Time
		}

		users = append(users, user)
	}

	return users, nil
}

// UpdateUser updates a user's information
func UpdateUser(id int, displayName, email string) error {
	query := `
		UPDATE users
		SET display_name = ?, email = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := database.DB.Exec(query, displayName, email, id)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// DeactivateUser marks a user as inactive
func DeactivateUser(id int) error {
	now := time.Now()
	query := `
		UPDATE users
		SET is_active = 0, left_at = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := database.DB.Exec(query, now, id)
	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// ReactivateUser marks a user as active again
func ReactivateUser(id int) error {
	query := `
		UPDATE users
		SET is_active = 1, left_at = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := database.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to reactivate user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
