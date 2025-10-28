package services

import (
	"fmt"
	"time"

	"google-chat-bot/database"
)

// CreateLeave adds a new leave record
func CreateLeave(userID int, leaveType string, startDate, endDate time.Time, reason string) (*database.Leave, error) {
	query := `
		INSERT INTO leaves (user_id, leave_type, start_date, end_date, reason, status)
		VALUES (?, ?, ?, ?, ?, 'active')
	`

	result, err := database.DB.Exec(query, userID, leaveType, startDate, endDate, reason)
	if err != nil {
		return nil, fmt.Errorf("failed to create leave: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return GetLeaveByID(int(id))
}

// GetLeaveByID retrieves a leave by ID
func GetLeaveByID(id int) (*database.Leave, error) {
	query := `
		SELECT id, user_id, leave_type, start_date, end_date, reason, status,
		       created_at, updated_at
		FROM leaves
		WHERE id = ?
	`

	var leave database.Leave
	err := database.DB.QueryRow(query, id).Scan(
		&leave.ID,
		&leave.UserID,
		&leave.LeaveType,
		&leave.StartDate,
		&leave.EndDate,
		&leave.Reason,
		&leave.Status,
		&leave.CreatedAt,
		&leave.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get leave: %w", err)
	}

	return &leave, nil
}

// GetAllLeaves retrieves all leave records
func GetAllLeaves() ([]database.Leave, error) {
	query := `
		SELECT id, user_id, leave_type, start_date, end_date, reason, status,
		       created_at, updated_at
		FROM leaves
		ORDER BY start_date DESC
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaves: %w", err)
	}
	defer rows.Close()

	var leaves []database.Leave
	for rows.Next() {
		var leave database.Leave
		err := rows.Scan(
			&leave.ID,
			&leave.UserID,
			&leave.LeaveType,
			&leave.StartDate,
			&leave.EndDate,
			&leave.Reason,
			&leave.Status,
			&leave.CreatedAt,
			&leave.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leave: %w", err)
		}
		leaves = append(leaves, leave)
	}

	return leaves, nil
}

// GetActiveLeaves retrieves all currently active leaves
func GetActiveLeaves() ([]database.Leave, error) {
	query := `
		SELECT id, user_id, leave_type, start_date, end_date, reason, status,
		       created_at, updated_at
		FROM leaves
		WHERE status = 'active'
		AND start_date <= date('now')
		AND end_date >= date('now')
		ORDER BY start_date DESC
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active leaves: %w", err)
	}
	defer rows.Close()

	var leaves []database.Leave
	for rows.Next() {
		var leave database.Leave
		err := rows.Scan(
			&leave.ID,
			&leave.UserID,
			&leave.LeaveType,
			&leave.StartDate,
			&leave.EndDate,
			&leave.Reason,
			&leave.Status,
			&leave.CreatedAt,
			&leave.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leave: %w", err)
		}
		leaves = append(leaves, leave)
	}

	return leaves, nil
}

// GetLeavesByUserID retrieves all leaves for a specific user
func GetLeavesByUserID(userID int) ([]database.Leave, error) {
	query := `
		SELECT id, user_id, leave_type, start_date, end_date, reason, status,
		       created_at, updated_at
		FROM leaves
		WHERE user_id = ?
		ORDER BY start_date DESC
	`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user leaves: %w", err)
	}
	defer rows.Close()

	var leaves []database.Leave
	for rows.Next() {
		var leave database.Leave
		err := rows.Scan(
			&leave.ID,
			&leave.UserID,
			&leave.LeaveType,
			&leave.StartDate,
			&leave.EndDate,
			&leave.Reason,
			&leave.Status,
			&leave.CreatedAt,
			&leave.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leave: %w", err)
		}
		leaves = append(leaves, leave)
	}

	return leaves, nil
}

// UpdateLeave updates a leave record
func UpdateLeave(id int, leaveType string, startDate, endDate time.Time, reason string) error {
	query := `
		UPDATE leaves
		SET leave_type = ?, start_date = ?, end_date = ?, reason = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := database.DB.Exec(query, leaveType, startDate, endDate, reason, id)
	if err != nil {
		return fmt.Errorf("failed to update leave: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("leave not found")
	}

	return nil
}

// CancelLeave marks a leave as cancelled
func CancelLeave(id int) error {
	query := `
		UPDATE leaves
		SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := database.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to cancel leave: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("leave not found")
	}

	return nil
}

// CompleteLeave marks a leave as completed
func CompleteLeave(id int) error {
	query := `
		UPDATE leaves
		SET status = 'completed', updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := database.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to complete leave: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("leave not found")
	}

	return nil
}
