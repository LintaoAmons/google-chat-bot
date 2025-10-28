package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"google-chat-bot/services"
)

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	GoogleChatUserID string `json:"google_chat_user_id"`
	DisplayName      string `json:"display_name"`
	Email            string `json:"email"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

// GetRosterHandler retrieves all users or active users only
func GetRosterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Check if we should filter for active users only
	activeOnly := r.URL.Query().Get("active") == "true"

	var users interface{}
	var err error

	if activeOnly {
		users, err = services.GetActiveUsers()
	} else {
		users, err = services.GetAllUsers()
	}

	if err != nil {
		log.Printf("Failed to get roster: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get roster"})
		return
	}

	json.NewEncoder(w).Encode(users)
}

// GetUserHandler retrieves a single user by ID
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from URL path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/roster/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid user ID"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	user, err := services.GetUserByID(id)
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
		return
	}

	json.NewEncoder(w).Encode(user)
}

// CreateUserHandler creates a new user
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if req.GoogleChatUserID == "" || req.DisplayName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "google_chat_user_id and display_name are required"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	user, err := services.CreateUser(req.GoogleChatUserID, req.DisplayName, req.Email)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create user"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// UpdateUserHandler updates an existing user
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from URL path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/roster/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid user ID"})
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = services.UpdateUser(id, req.DisplayName, req.Email)
	if err != nil {
		log.Printf("Failed to update user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update user"})
		return
	}

	// Return updated user
	user, _ := services.GetUserByID(id)
	json.NewEncoder(w).Encode(user)
}

// DeactivateUserHandler deactivates a user
func DeactivateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from URL path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/roster/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid user ID"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = services.DeactivateUser(id)
	if err != nil {
		log.Printf("Failed to deactivate user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to deactivate user"})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "User deactivated successfully"})
}

// ReactivateUserHandler reactivates a user
func ReactivateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid URL"})
		return
	}

	id, err := strconv.Atoi(pathParts[2])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid user ID"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = services.ReactivateUser(id)
	if err != nil {
		log.Printf("Failed to reactivate user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to reactivate user"})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "User reactivated successfully"})
}
