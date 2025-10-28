package services

import (
	"database/sql"
	"fmt"
	"log"

	"google-chat-bot/database"
)

// CreateStandup creates a new standup meeting
func CreateStandup(name, message, runAt, createdBy string) (*database.Standup, error) {
	query := `
		INSERT INTO standups (name, message, run_at, created_by, is_active)
		VALUES (?, ?, ?, ?, 1)
	`

	result, err := database.DB.Exec(query, name, message, runAt, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create standup: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return GetStandupByID(int(id))
}

// GetStandupByID retrieves a standup by ID
func GetStandupByID(id int) (*database.Standup, error) {
	query := `
		SELECT id, name, message, run_at, is_active, last_facilitator_id, created_by, created_at, updated_at
		FROM standups
		WHERE id = ?
	`

	var standup database.Standup
	var facilitatorID sql.NullInt64
	err := database.DB.QueryRow(query, id).Scan(
		&standup.ID,
		&standup.Name,
		&standup.Message,
		&standup.RunAt,
		&standup.IsActive,
		&facilitatorID,
		&standup.CreatedBy,
		&standup.CreatedAt,
		&standup.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get standup: %w", err)
	}

	if facilitatorID.Valid {
		id := int(facilitatorID.Int64)
		standup.LastFacilitatorID = &id
	}

	return &standup, nil
}

// GetStandupWithMembers retrieves a standup with its assigned members
func GetStandupWithMembers(id int) (*database.StandupWithMembers, error) {
	standup, err := GetStandupByID(id)
	if err != nil {
		return nil, err
	}

	members, err := GetStandupMembers(id)
	if err != nil {
		return nil, err
	}

	result := &database.StandupWithMembers{
		Standup: *standup,
		Members: members,
	}

	// Get last facilitator if set
	if standup.LastFacilitatorID != nil {
		facilitator, err := getUserByID(*standup.LastFacilitatorID)
		if err == nil {
			result.LastFacilitator = facilitator
		}
	}

	// Calculate current facilitator from eligible users
	eligibleUsers, err := database.GetEligibleUsersForStandup(id)
	if err == nil && len(eligibleUsers) > 0 {
		currentFac, err := GetCurrentFacilitator(id, eligibleUsers)
		if err == nil {
			result.CurrentFacilitator = currentFac
		}
	}

	return result, nil
}

// getUserByID retrieves a user by ID (helper function)
func getUserByID(id int) (*database.User, error) {
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
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if leftAt.Valid {
		user.LeftAt = &leftAt.Time
	}

	return &user, nil
}

// GetAllStandups retrieves all standups
func GetAllStandups() ([]database.Standup, error) {
	query := `
		SELECT id, name, message, run_at, is_active, last_facilitator_id, created_by, created_at, updated_at
		FROM standups
		ORDER BY run_at, name
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query standups: %w", err)
	}
	defer rows.Close()

	var standups []database.Standup
	for rows.Next() {
		var standup database.Standup
		var facilitatorID sql.NullInt64
		err := rows.Scan(
			&standup.ID,
			&standup.Name,
			&standup.Message,
			&standup.RunAt,
			&standup.IsActive,
			&facilitatorID,
			&standup.CreatedBy,
			&standup.CreatedAt,
			&standup.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan standup: %w", err)
		}
		if facilitatorID.Valid {
			id := int(facilitatorID.Int64)
			standup.LastFacilitatorID = &id
		}
		standups = append(standups, standup)
	}

	return standups, nil
}

// GetActiveStandups retrieves all active standups
func GetActiveStandups() ([]database.Standup, error) {
	query := `
		SELECT id, name, message, run_at, is_active, last_facilitator_id, created_by, created_at, updated_at
		FROM standups
		WHERE is_active = 1
		ORDER BY run_at, name
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active standups: %w", err)
	}
	defer rows.Close()

	var standups []database.Standup
	for rows.Next() {
		var standup database.Standup
		var facilitatorID sql.NullInt64
		err := rows.Scan(
			&standup.ID,
			&standup.Name,
			&standup.Message,
			&standup.RunAt,
			&standup.IsActive,
			&facilitatorID,
			&standup.CreatedBy,
			&standup.CreatedAt,
			&standup.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan standup: %w", err)
		}
		if facilitatorID.Valid {
			id := int(facilitatorID.Int64)
			standup.LastFacilitatorID = &id
		}
		standups = append(standups, standup)
	}

	return standups, nil
}

// UpdateStandup updates a standup
func UpdateStandup(id int, name, message, runAt string) error {
	// Get current standup to log changes
	oldStandup, err := GetStandupByID(id)
	if err != nil {
		return fmt.Errorf("failed to get standup: %w", err)
	}

	query := `
		UPDATE standups
		SET name = ?, message = ?, run_at = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := database.DB.Exec(query, name, message, runAt, id)
	if err != nil {
		return fmt.Errorf("failed to update standup: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("standup not found")
	}

	// Log if schedule time changed
	if oldStandup.RunAt != runAt {
		log.Printf("📅 [SCHEDULE UPDATE] Standup '%s' (ID: %d) schedule changed from %s to %s", name, id, oldStandup.RunAt, runAt)
	}

	return nil
}

// DeleteStandup deactivates a standup
func DeleteStandup(id int) error {
	query := `
		UPDATE standups
		SET is_active = 0, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := database.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete standup: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("standup not found")
	}

	return nil
}

// ReactivateStandup reactivates a standup
func ReactivateStandup(id int) error {
	query := `
		UPDATE standups
		SET is_active = 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := database.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to reactivate standup: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("standup not found")
	}

	return nil
}

// AddStandupMember adds a user to a standup roster
func AddStandupMember(standupID, userID int) error {
	query := `
		INSERT OR IGNORE INTO standup_members (standup_id, user_id)
		VALUES (?, ?)
	`

	_, err := database.DB.Exec(query, standupID, userID)
	if err != nil {
		return fmt.Errorf("failed to add standup member: %w", err)
	}

	return nil
}

// RemoveStandupMember removes a user from a standup roster
func RemoveStandupMember(standupID, userID int) error {
	// Start transaction
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the display_order of the member being removed
	var removedOrder int
	err = tx.QueryRow(
		"SELECT display_order FROM standup_members WHERE standup_id = ? AND user_id = ?",
		standupID, userID,
	).Scan(&removedOrder)

	if err != nil {
		return fmt.Errorf("member not found")
	}

	// Delete the member
	_, err = tx.Exec(
		"DELETE FROM standup_members WHERE standup_id = ? AND user_id = ?",
		standupID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove standup member: %w", err)
	}

	// Reorder remaining members to fill the gap
	_, err = tx.Exec(
		"UPDATE standup_members SET display_order = display_order - 1 WHERE standup_id = ? AND display_order > ?",
		standupID, removedOrder,
	)
	if err != nil {
		return fmt.Errorf("failed to reorder members: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetStandupMembers retrieves all users assigned to a standup, ordered by display_order
func GetStandupMembers(standupID int) ([]database.User, error) {
	query := `
		SELECT u.id, u.google_chat_user_id, u.display_name, u.email, u.is_active,
		       u.joined_at, u.left_at, u.created_at, u.updated_at
		FROM users u
		INNER JOIN standup_members sm ON u.id = sm.user_id
		WHERE sm.standup_id = ?
		ORDER BY sm.display_order, u.display_name
	`

	rows, err := database.DB.Query(query, standupID)
	if err != nil {
		return nil, fmt.Errorf("failed to query standup members: %w", err)
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

// SetStandupMembers replaces all members of a standup with a new list
func SetStandupMembers(standupID int, userIDs []int) error {
	// Start transaction
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing members
	_, err = tx.Exec("DELETE FROM standup_members WHERE standup_id = ?", standupID)
	if err != nil {
		return fmt.Errorf("failed to delete existing members: %w", err)
	}

	// Insert new members with display_order
	stmt, err := tx.Prepare("INSERT INTO standup_members (standup_id, user_id, display_order) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i, userID := range userIDs {
		_, err = stmt.Exec(standupID, userID, i)
		if err != nil {
			return fmt.Errorf("failed to insert member %d: %w", userID, err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SetLastFacilitator sets the last facilitator for a standup
func SetLastFacilitator(standupID, userID int) error {
	query := `
		UPDATE standups
		SET last_facilitator_id = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := database.DB.Exec(query, userID, standupID)
	if err != nil {
		return fmt.Errorf("failed to set last facilitator: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("standup not found")
	}

	return nil
}

// GetCurrentFacilitator calculates the current facilitator from eligible users based on last facilitator
func GetCurrentFacilitator(standupID int, eligibleUsers []database.User) (*database.User, error) {
	standup, err := GetStandupByID(standupID)
	if err != nil {
		return nil, err
	}

	if len(eligibleUsers) == 0 {
		return nil, fmt.Errorf("no eligible users")
	}

	// If no last facilitator, return first eligible user
	if standup.LastFacilitatorID == nil {
		return &eligibleUsers[0], nil
	}

	// Get all members ordered by display_order
	allMembers, err := GetStandupMembers(standupID)
	if err != nil {
		return nil, err
	}

	// Find last facilitator index in the full member list
	lastFacIndex := -1
	for i, member := range allMembers {
		if member.ID == *standup.LastFacilitatorID {
			lastFacIndex = i
			break
		}
	}

	// If last facilitator not found, start from beginning
	if lastFacIndex == -1 {
		return &eligibleUsers[0], nil
	}

	// Find next eligible facilitator after last facilitator in the rotation order
	for i := lastFacIndex + 1; i < len(allMembers); i++ {
		for _, eligible := range eligibleUsers {
			if allMembers[i].ID == eligible.ID {
				return &eligible, nil
			}
		}
	}

	// Wrap around: check from start to lastFacIndex
	for i := 0; i <= lastFacIndex; i++ {
		for _, eligible := range eligibleUsers {
			if allMembers[i].ID == eligible.ID {
				return &eligible, nil
			}
		}
	}

	// Fallback: return first eligible user
	return &eligibleUsers[0], nil
}

// RotateFacilitator updates last_facilitator_id to the current facilitator
// This should be called after sending a message
func RotateFacilitator(standupID int, currentFacilitatorID int) error {
	return SetLastFacilitator(standupID, currentFacilitatorID)
}

// GetNextFacilitator returns who tomorrow's facilitator will be (calculated from eligible users)
func GetNextFacilitator(standupID int, eligibleUsers []database.User, currentFacilitatorID int) (*database.User, error) {
	if len(eligibleUsers) == 0 {
		return nil, fmt.Errorf("no eligible users")
	}

	// Get all members ordered by display_order
	allMembers, err := GetStandupMembers(standupID)
	if err != nil {
		return nil, err
	}

	// Find current facilitator index in the full member list
	currentFacIndex := -1
	for i, member := range allMembers {
		if member.ID == currentFacilitatorID {
			currentFacIndex = i
			break
		}
	}

	// If current facilitator not found, return first eligible user
	if currentFacIndex == -1 {
		return &eligibleUsers[0], nil
	}

	// Find next eligible facilitator after current facilitator in the rotation order
	for i := currentFacIndex + 1; i < len(allMembers); i++ {
		for _, eligible := range eligibleUsers {
			if allMembers[i].ID == eligible.ID {
				return &eligible, nil
			}
		}
	}

	// Wrap around: check from start to currentFacIndex
	for i := 0; i <= currentFacIndex; i++ {
		for _, eligible := range eligibleUsers {
			if allMembers[i].ID == eligible.ID {
				return &eligible, nil
			}
		}
	}

	// Fallback: return first eligible user
	return &eligibleUsers[0], nil
}

// MoveMemberUp moves a member up in the display order
func MoveMemberUp(standupID, userID int) error {
	// Get current order
	var currentOrder int
	err := database.DB.QueryRow(
		"SELECT display_order FROM standup_members WHERE standup_id = ? AND user_id = ?",
		standupID, userID,
	).Scan(&currentOrder)

	if err != nil {
		return fmt.Errorf("failed to get current order: %w", err)
	}

	if currentOrder == 0 {
		return fmt.Errorf("member is already at the top")
	}

	// Start transaction
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Swap with member above
	// Member above has order = currentOrder - 1
	_, err = tx.Exec(
		"UPDATE standup_members SET display_order = ? WHERE standup_id = ? AND display_order = ?",
		currentOrder, standupID, currentOrder-1,
	)
	if err != nil {
		return fmt.Errorf("failed to update member above: %w", err)
	}

	// Move current member up
	_, err = tx.Exec(
		"UPDATE standup_members SET display_order = ? WHERE standup_id = ? AND user_id = ?",
		currentOrder-1, standupID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to move member up: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// MoveMemberDown moves a member down in the display order
func MoveMemberDown(standupID, userID int) error {
	// Get current order and max order
	var currentOrder, maxOrder int
	err := database.DB.QueryRow(
		"SELECT display_order FROM standup_members WHERE standup_id = ? AND user_id = ?",
		standupID, userID,
	).Scan(&currentOrder)

	if err != nil {
		return fmt.Errorf("failed to get current order: %w", err)
	}

	err = database.DB.QueryRow(
		"SELECT MAX(display_order) FROM standup_members WHERE standup_id = ?",
		standupID,
	).Scan(&maxOrder)

	if err != nil {
		return fmt.Errorf("failed to get max order: %w", err)
	}

	if currentOrder >= maxOrder {
		return fmt.Errorf("member is already at the bottom")
	}

	// Start transaction
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Swap with member below
	// Member below has order = currentOrder + 1
	_, err = tx.Exec(
		"UPDATE standup_members SET display_order = ? WHERE standup_id = ? AND display_order = ?",
		currentOrder, standupID, currentOrder+1,
	)
	if err != nil {
		return fmt.Errorf("failed to update member below: %w", err)
	}

	// Move current member down
	_, err = tx.Exec(
		"UPDATE standup_members SET display_order = ? WHERE standup_id = ? AND user_id = ?",
		currentOrder+1, standupID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to move member down: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
