package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"google-chat-bot/services"
)

// CreateLeaveRequest represents the request to create a leave
type CreateLeaveRequest struct {
	UserID    int    `json:"user_id"`
	LeaveType string `json:"leave_type"`
	StartDate string `json:"start_date"` // Format: YYYY-MM-DD
	EndDate   string `json:"end_date"`   // Format: YYYY-MM-DD
	Reason    string `json:"reason"`
}

// UpdateLeaveRequest represents the request to update a leave
type UpdateLeaveRequest struct {
	LeaveType string `json:"leave_type"`
	StartDate string `json:"start_date"` // Format: YYYY-MM-DD
	EndDate   string `json:"end_date"`   // Format: YYYY-MM-DD
	Reason    string `json:"reason"`
}

// GetLeavesHandler retrieves all leaves or filters by status
func GetLeavesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Check for filters
	activeOnly := r.URL.Query().Get("active") == "true"
	userIDStr := r.URL.Query().Get("user_id")

	var leaves interface{}
	var err error

	if userIDStr != "" {
		// Get leaves for specific user
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid user_id"})
			return
		}
		leaves, err = services.GetLeavesByUserID(userID)
	} else if activeOnly {
		// Get only active leaves
		leaves, err = services.GetActiveLeaves()
	} else {
		// Get all leaves
		leaves, err = services.GetAllLeaves()
	}

	if err != nil {
		log.Printf("Failed to get leaves: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get leaves"})
		return
	}

	json.NewEncoder(w).Encode(leaves)
}

// GetLeaveHandler retrieves a single leave by ID
func GetLeaveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract leave ID from URL path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/leaves/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid leave ID"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	leave, err := services.GetLeaveByID(id)
	if err != nil {
		log.Printf("Failed to get leave: %v", err)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Leave not found"})
		return
	}

	json.NewEncoder(w).Encode(leave)
}

// CreateLeaveHandler creates a new leave
func CreateLeaveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateLeaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if req.UserID == 0 || req.LeaveType == "" || req.StartDate == "" || req.EndDate == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "user_id, leave_type, start_date, and end_date are required"})
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid start_date format (use YYYY-MM-DD)"})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid end_date format (use YYYY-MM-DD)"})
		return
	}

	// Validate dates
	if endDate.Before(startDate) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "end_date must be after start_date"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	leave, err := services.CreateLeave(req.UserID, req.LeaveType, startDate, endDate, req.Reason)
	if err != nil {
		log.Printf("Failed to create leave: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create leave"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(leave)
}

// UpdateLeaveHandler updates an existing leave
func UpdateLeaveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract leave ID from URL path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/leaves/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid leave ID"})
		return
	}

	var req UpdateLeaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid start_date format (use YYYY-MM-DD)"})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid end_date format (use YYYY-MM-DD)"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = services.UpdateLeave(id, req.LeaveType, startDate, endDate, req.Reason)
	if err != nil {
		log.Printf("Failed to update leave: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update leave"})
		return
	}

	// Return updated leave
	leave, _ := services.GetLeaveByID(id)
	json.NewEncoder(w).Encode(leave)
}

// CancelLeaveHandler cancels a leave
func CancelLeaveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract leave ID from URL path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/leaves/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid leave ID"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = services.CancelLeave(id)
	if err != nil {
		log.Printf("Failed to cancel leave: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to cancel leave"})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Leave cancelled successfully"})
}
