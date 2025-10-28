package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"google-chat-bot/database"
	"google-chat-bot/services"
)

// CreateStandupRequest represents the request to create a standup
type CreateStandupRequest struct {
	Name      string `json:"name"`
	Message   string `json:"message"`
	RunAt     string `json:"run_at"` // HH:MM format
	CreatedBy string `json:"created_by"`
	Members   []int  `json:"members"` // User IDs
}

// UpdateStandupRequest represents the request to update a standup
type UpdateStandupRequest struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	RunAt   string `json:"run_at"` // HH:MM format
	Members []int  `json:"members"` // User IDs (optional, for updating members)
}

// GetStandupsHandler retrieves all standups or active standups only
func GetStandupsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Check if we should filter for active standups only
	activeOnly := r.URL.Query().Get("active") == "true"

	var standups interface{}
	var err error

	if activeOnly {
		standups, err = services.GetActiveStandups()
	} else {
		standups, err = services.GetAllStandups()
	}

	if err != nil {
		log.Printf("Failed to get standups: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get standups"})
		return
	}

	json.NewEncoder(w).Encode(standups)
}

// GetStandupHandler retrieves a single standup by ID with its members
func GetStandupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract standup ID from URL path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/standups/")
	// Remove /members suffix if present
	idStr = strings.TrimSuffix(idStr, "/members")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid standup ID"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	standup, err := services.GetStandupWithMembers(id)
	if err != nil {
		log.Printf("Failed to get standup: %v", err)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Standup not found"})
		return
	}

	json.NewEncoder(w).Encode(standup)
}

// CreateStandupHandler creates a new standup
func CreateStandupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateStandupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if req.Name == "" || req.Message == "" || req.RunAt == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "name, message, and run_at are required"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	standup, err := services.CreateStandup(req.Name, req.Message, req.RunAt, req.CreatedBy)
	if err != nil {
		log.Printf("Failed to create standup: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create standup"})
		return
	}

	// Add members if provided
	if len(req.Members) > 0 {
		err = services.SetStandupMembers(standup.ID, req.Members)
		if err != nil {
			log.Printf("Failed to add members to standup: %v", err)
			// Continue anyway, standup is created
		}
	}

	// Refresh scheduler to include new standup
	if err := services.RefreshScheduler(); err != nil {
		log.Printf("Failed to refresh scheduler: %v", err)
	}

	// Return standup with members
	standupWithMembers, _ := services.GetStandupWithMembers(standup.ID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(standupWithMembers)
}

// UpdateStandupHandler updates an existing standup
func UpdateStandupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract standup ID from URL path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/standups/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid standup ID"})
		return
	}

	var req UpdateStandupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = services.UpdateStandup(id, req.Name, req.Message, req.RunAt)
	if err != nil {
		log.Printf("Failed to update standup: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update standup"})
		return
	}

	// Update members if provided
	if len(req.Members) > 0 {
		err = services.SetStandupMembers(id, req.Members)
		if err != nil {
			log.Printf("Failed to update members: %v", err)
		}
	}

	// Refresh scheduler to update schedule
	if err := services.RefreshScheduler(); err != nil {
		log.Printf("Failed to refresh scheduler: %v", err)
	}

	// Return updated standup with members
	standup, _ := services.GetStandupWithMembers(id)
	json.NewEncoder(w).Encode(standup)
}

// DeleteStandupHandler deactivates a standup
func DeleteStandupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract standup ID from URL path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/standups/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid standup ID"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = services.DeleteStandup(id)
	if err != nil {
		log.Printf("Failed to delete standup: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete standup"})
		return
	}

	// Refresh scheduler to remove standup
	if err := services.RefreshScheduler(); err != nil {
		log.Printf("Failed to refresh scheduler: %v", err)
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Standup deleted successfully"})
}

// GetStandupMembersHandler retrieves all members of a standup
func GetStandupMembersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract standup ID from URL: /api/standups/:id/members
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid URL"})
		return
	}

	id, err := strconv.Atoi(parts[2])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid standup ID"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	members, err := services.GetStandupMembers(id)
	if err != nil {
		log.Printf("Failed to get standup members: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get standup members"})
		return
	}

	json.NewEncoder(w).Encode(members)
}

// SetStandupMembersHandler replaces all members of a standup
func SetStandupMembersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract standup ID from URL: /api/standups/:id/members
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid URL"})
		return
	}

	id, err := strconv.Atoi(parts[2])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid standup ID"})
		return
	}

	var req struct {
		Members []int `json:"members"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = services.SetStandupMembers(id, req.Members)
	if err != nil {
		log.Printf("Failed to set standup members: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to set standup members"})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Standup members updated successfully"})
}

// SendStandupReminderHandler manually triggers a standup reminder
func SendStandupReminderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract standup ID from URL: /api/standups/:id/send
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid URL"})
		return
	}

	id, err := strconv.Atoi(parts[2])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid standup ID"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = services.SendManualStandupReminder(id)
	if err != nil {
		log.Printf("Failed to send manual reminder: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to send reminder: %v", err)})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Reminder sent successfully!"})
}

// SetFacilitatorHandler sets the current facilitator for a standup
func SetFacilitatorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract standup ID from URL: /api/standups/:id/facilitator
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid URL"})
		return
	}

	standupID, err := strconv.Atoi(parts[2])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid standup ID"})
		return
	}

	var req struct {
		UserID int `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = services.SetLastFacilitator(standupID, req.UserID)
	if err != nil {
		log.Printf("Failed to set last facilitator: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to set facilitator: %v", err)})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Facilitator updated successfully!"})
}

// RotateFacilitatorHandler rotates to the next facilitator
func RotateFacilitatorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract standup ID from URL: /api/standups/:id/facilitator/rotate
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid URL"})
		return
	}

	standupID, err := strconv.Atoi(parts[2])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid standup ID"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Get eligible users
	eligibleUsers, err := database.GetEligibleUsersForStandup(standupID)
	if err != nil || len(eligibleUsers) == 0 {
		log.Printf("Failed to get eligible users: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "No eligible users for standup"})
		return
	}

	// Calculate current facilitator
	currentFac, err := services.GetCurrentFacilitator(standupID, eligibleUsers)
	if err != nil {
		log.Printf("Failed to get current facilitator: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to get current facilitator: %v", err)})
		return
	}

	// Rotate by setting last_facilitator_id to current facilitator
	err = services.RotateFacilitator(standupID, currentFac.ID)
	if err != nil {
		log.Printf("Failed to rotate facilitator: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to rotate facilitator: %v", err)})
		return
	}

	// Return updated standup with new facilitator
	standup, _ := services.GetStandupWithMembers(standupID)
	json.NewEncoder(w).Encode(standup)
}

// MoveMemberUpHandler moves a member up in the display order
func MoveMemberUpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract standup ID from URL: /api/standups/:id/members/:user_id/up
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid URL"})
		return
	}

	standupID, err := strconv.Atoi(parts[2])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid standup ID"})
		return
	}

	userID, err := strconv.Atoi(parts[4])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid user ID"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = services.MoveMemberUp(standupID, userID)
	if err != nil {
		log.Printf("Failed to move member up: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("%v", err)})
		return
	}

	// Return updated standup with reordered members
	standup, _ := services.GetStandupWithMembers(standupID)
	json.NewEncoder(w).Encode(standup)
}

// MoveMemberDownHandler moves a member down in the display order
func MoveMemberDownHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract standup ID from URL: /api/standups/:id/members/:user_id/down
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid URL"})
		return
	}

	standupID, err := strconv.Atoi(parts[2])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid standup ID"})
		return
	}

	userID, err := strconv.Atoi(parts[4])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid user ID"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = services.MoveMemberDown(standupID, userID)
	if err != nil {
		log.Printf("Failed to move member down: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("%v", err)})
		return
	}

	// Return updated standup with reordered members
	standup, _ := services.GetStandupWithMembers(standupID)
	json.NewEncoder(w).Encode(standup)
}

// RemoveStandupMemberHandler removes a member from a standup
func RemoveStandupMemberHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract standup ID from URL: /api/standups/:id/members/:user_id
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid URL"})
		return
	}

	standupID, err := strconv.Atoi(parts[2])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid standup ID"})
		return
	}

	userID, err := strconv.Atoi(parts[4])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid user ID"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = services.RemoveStandupMember(standupID, userID)
	if err != nil {
		log.Printf("Failed to remove member: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("%v", err)})
		return
	}

	// Return updated standup with remaining members
	standup, _ := services.GetStandupWithMembers(standupID)
	json.NewEncoder(w).Encode(standup)
}
