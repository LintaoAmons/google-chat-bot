package database

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID               int       `json:"id"`
	GoogleChatUserID string    `json:"google_chat_user_id"`
	DisplayName      string    `json:"display_name"`
	Email            string    `json:"email"`
	IsActive         bool      `json:"is_active"`
	JoinedAt         time.Time `json:"joined_at"`
	LeftAt           *time.Time `json:"left_at,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Leave represents a leave record for a user
type Leave struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	LeaveType string    `json:"leave_type"` // 'sick', 'vacation', 'pto', 'personal', etc.
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Reason    string    `json:"reason"`
	Status    string    `json:"status"` // 'active', 'completed', 'cancelled'
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Standup represents a standup meeting with its own schedule and roster
type Standup struct {
	ID                 int       `json:"id"`
	Name               string    `json:"name"`
	Message            string    `json:"message"`
	RunAt              string    `json:"run_at"` // Time in HH:MM format (e.g., "09:00")
	IsActive           bool      `json:"is_active"`
	LastFacilitatorID  *int      `json:"last_facilitator_id,omitempty"`
	CreatedBy          string    `json:"created_by"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// StandupMember represents a user assigned to a standup meeting
type StandupMember struct {
	StandupID int       `json:"standup_id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// StandupWithMembers represents a standup with its assigned members
type StandupWithMembers struct {
	Standup
	Members           []User `json:"members"`
	LastFacilitator   *User  `json:"last_facilitator,omitempty"`
	CurrentFacilitator *User  `json:"current_facilitator,omitempty"` // Dynamically calculated, not from DB
}
