package database

import (
	"database/sql"
	"fmt"
	"log"
)

// RunMigrations runs all database migrations
func RunMigrations() error {
	migrations := []string{
		createUsersTable,
		createLeavesTable,
		createStandupsTable,
		createStandupMembersTable,
	}

	for i, migration := range migrations {
		if _, err := DB.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i, err)
		}
	}

	log.Println("All migrations completed successfully")
	return nil
}

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    google_chat_user_id TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    email TEXT,
    is_active BOOLEAN DEFAULT 1,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_google_chat_id ON users(google_chat_user_id);
`

const createLeavesTable = `
CREATE TABLE IF NOT EXISTS leaves (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    leave_type TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    reason TEXT,
    status TEXT DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_leaves_user ON leaves(user_id);
CREATE INDEX IF NOT EXISTS idx_leaves_dates ON leaves(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_leaves_status ON leaves(status);
`

const createStandupsTable = `
CREATE TABLE IF NOT EXISTS standups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    message TEXT NOT NULL,
    run_at TEXT NOT NULL,
    is_active BOOLEAN DEFAULT 1,
    last_facilitator_id INTEGER,
    created_by TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (last_facilitator_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_standups_active ON standups(is_active);
CREATE INDEX IF NOT EXISTS idx_standups_run_at ON standups(run_at);
`

const createStandupMembersTable = `
CREATE TABLE IF NOT EXISTS standup_members (
    standup_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (standup_id, user_id),
    FOREIGN KEY (standup_id) REFERENCES standups(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_standup_members_standup ON standup_members(standup_id);
CREATE INDEX IF NOT EXISTS idx_standup_members_user ON standup_members(user_id);
CREATE INDEX IF NOT EXISTS idx_standup_members_order ON standup_members(standup_id, display_order);
`

// GetEligibleUsersForStandup returns users assigned to a standup who are active and not on leave
func GetEligibleUsersForStandup(standupID int) ([]User, error) {
	query := `
		SELECT DISTINCT u.id, u.google_chat_user_id, u.display_name, u.email, u.is_active,
		       u.joined_at, u.left_at, u.created_at, u.updated_at
		FROM users u
		INNER JOIN standup_members sm ON u.id = sm.user_id
		WHERE sm.standup_id = ?
		AND u.is_active = 1
		AND u.id NOT IN (
			SELECT user_id FROM leaves
			WHERE status = 'active'
			AND start_date <= date('now')
			AND end_date >= date('now')
		)
	`

	rows, err := DB.Query(query, standupID)
	if err != nil {
		return nil, fmt.Errorf("failed to query eligible users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
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

// ExpireOldLeaves marks leaves as completed if their end_date has passed
func ExpireOldLeaves() error {
	query := `
		UPDATE leaves
		SET status = 'completed', updated_at = CURRENT_TIMESTAMP
		WHERE status = 'active'
		AND end_date < date('now')
	`

	result, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to expire old leaves: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("Expired %d old leave(s)", rowsAffected)
	}

	return nil
}

// LeaveWithUser represents a leave record with user information
type LeaveWithUser struct {
	Leave
	User User
}

// GetActiveLeavesForStandup returns active leaves for standup members on a specific date
func GetActiveLeavesForStandup(standupID int) ([]LeaveWithUser, error) {
	query := `
		SELECT l.id, l.user_id, l.leave_type, l.start_date, l.end_date, l.reason, l.status,
		       l.created_at, l.updated_at,
		       u.id, u.google_chat_user_id, u.display_name, u.email, u.is_active,
		       u.joined_at, u.left_at, u.created_at, u.updated_at
		FROM leaves l
		INNER JOIN users u ON l.user_id = u.id
		INNER JOIN standup_members sm ON u.id = sm.user_id
		WHERE sm.standup_id = ?
		AND l.status = 'active'
		AND l.start_date <= date('now')
		AND l.end_date >= date('now')
		ORDER BY u.display_name
	`

	rows, err := DB.Query(query, standupID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active leaves: %w", err)
	}
	defer rows.Close()

	var leaves []LeaveWithUser
	for rows.Next() {
		var lwu LeaveWithUser
		var userLeftAt sql.NullTime
		err := rows.Scan(
			&lwu.Leave.ID,
			&lwu.Leave.UserID,
			&lwu.Leave.LeaveType,
			&lwu.Leave.StartDate,
			&lwu.Leave.EndDate,
			&lwu.Leave.Reason,
			&lwu.Leave.Status,
			&lwu.Leave.CreatedAt,
			&lwu.Leave.UpdatedAt,
			&lwu.User.ID,
			&lwu.User.GoogleChatUserID,
			&lwu.User.DisplayName,
			&lwu.User.Email,
			&lwu.User.IsActive,
			&lwu.User.JoinedAt,
			&userLeftAt,
			&lwu.User.CreatedAt,
			&lwu.User.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leave: %w", err)
		}

		if userLeftAt.Valid {
			lwu.User.LeftAt = &userLeftAt.Time
		}

		leaves = append(leaves, lwu)
	}

	return leaves, nil
}
