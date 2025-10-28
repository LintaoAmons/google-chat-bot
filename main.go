package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"google-chat-bot/config"
	"google-chat-bot/database"
	"google-chat-bot/handlers"
	"google-chat-bot/services"
)

// handleRosterRoutes routes roster API requests
func handleRosterRoutes(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/roster" || r.URL.Path == "/api/roster/" {
		// Collection routes
		switch r.Method {
		case http.MethodGet:
			handlers.GetRosterHandler(w, r)
		case http.MethodPost:
			handlers.CreateUserHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else if strings.HasSuffix(r.URL.Path, "/reactivate") && r.Method == http.MethodPost {
		// Reactivate route
		handlers.ReactivateUserHandler(w, r)
	} else {
		// Single resource routes
		switch r.Method {
		case http.MethodGet:
			handlers.GetUserHandler(w, r)
		case http.MethodPut:
			handlers.UpdateUserHandler(w, r)
		case http.MethodDelete:
			handlers.DeactivateUserHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleLeavesRoutes routes leaves API requests
func handleLeavesRoutes(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/leaves" || r.URL.Path == "/api/leaves/" {
		// Collection routes
		switch r.Method {
		case http.MethodGet:
			handlers.GetLeavesHandler(w, r)
		case http.MethodPost:
			handlers.CreateLeaveHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else {
		// Single resource routes
		switch r.Method {
		case http.MethodGet:
			handlers.GetLeaveHandler(w, r)
		case http.MethodPut:
			handlers.UpdateLeaveHandler(w, r)
		case http.MethodDelete:
			handlers.CancelLeaveHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleStandupsRoutes routes standups API requests
func handleStandupsRoutes(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/standups" || r.URL.Path == "/api/standups/" {
		// Collection routes
		switch r.Method {
		case http.MethodGet:
			handlers.GetStandupsHandler(w, r)
		case http.MethodPost:
			handlers.CreateStandupHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else if strings.HasSuffix(r.URL.Path, "/up") {
		// Move member up route: /api/standups/:id/members/:user_id/up
		if r.Method == http.MethodPost {
			handlers.MoveMemberUpHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else if strings.HasSuffix(r.URL.Path, "/down") {
		// Move member down route: /api/standups/:id/members/:user_id/down
		if r.Method == http.MethodPost {
			handlers.MoveMemberDownHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else if strings.Contains(r.URL.Path, "/members/") && r.Method == http.MethodDelete {
		// Remove member route: DELETE /api/standups/:id/members/:user_id
		handlers.RemoveStandupMemberHandler(w, r)
	} else if strings.HasSuffix(r.URL.Path, "/members") {
		// Member management routes: /api/standups/:id/members
		switch r.Method {
		case http.MethodGet:
			handlers.GetStandupMembersHandler(w, r)
		case http.MethodPut:
			handlers.SetStandupMembersHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else if strings.HasSuffix(r.URL.Path, "/facilitator/rotate") {
		// Rotate facilitator route: /api/standups/:id/facilitator/rotate
		if r.Method == http.MethodPost {
			handlers.RotateFacilitatorHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else if strings.HasSuffix(r.URL.Path, "/facilitator") {
		// Set facilitator route: /api/standups/:id/facilitator
		if r.Method == http.MethodPost {
			handlers.SetFacilitatorHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else if strings.HasSuffix(r.URL.Path, "/send") {
		// Manual reminder route: /api/standups/:id/send
		if r.Method == http.MethodPost {
			handlers.SendStandupReminderHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else {
		// Single resource routes: /api/standups/:id
		switch r.Method {
		case http.MethodGet:
			handlers.GetStandupHandler(w, r)
		case http.MethodPut:
			handlers.UpdateStandupHandler(w, r)
		case http.MethodDelete:
			handlers.DeleteStandupHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func main() {
	// Load configuration
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	if err := database.InitDB(config.Config.DatabasePath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// Start scheduler
	if err := services.StartScheduler(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}
	defer services.StopScheduler()

	// Set up HTTP routes
	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/send", handlers.SendHandler)
	http.HandleFunc("/health", handlers.HealthHandler)
	http.HandleFunc("/api/send-reminder", handlers.SendReminderHandler)

	// Roster API routes
	http.HandleFunc("/api/roster", handleRosterRoutes)
	http.HandleFunc("/api/roster/", handleRosterRoutes)

	// Leaves API routes
	http.HandleFunc("/api/leaves", handleLeavesRoutes)
	http.HandleFunc("/api/leaves/", handleLeavesRoutes)

	// Standups API routes
	http.HandleFunc("/api/standups", handleStandupsRoutes)
	http.HandleFunc("/api/standups/", handleStandupsRoutes)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down gracefully...")
		services.StopScheduler()
		database.CloseDB()
		os.Exit(0)
	}()

	// Start the server
	addr := ":" + config.Config.Port
	log.Printf("🚀 Standup Bot starting on http://localhost%s", addr)
	log.Printf("📝 Web UI: http://localhost%s", addr)
	log.Printf("🔧 Health check: http://localhost%s/health", addr)
	log.Printf("⏰ Scheduler: Running with configured standups")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
