package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"google-chat-bot/config"
	"google-chat-bot/integrations"
)

// HomeHandler serves the web UI
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/ui.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Template parsing error: %v", err)
		return
	}
	tmpl.Execute(w, nil)
}

// MessageRequest represents the incoming request from the web UI
type MessageRequest struct {
	Message      string `json:"message"`
	MessageType  string `json:"messageType"`
	CardTitle    string `json:"cardTitle,omitempty"`
	CardSubtitle string `json:"cardSubtitle,omitempty"`
}

// SendHandler handles the message sending endpoint
func SendHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Message cannot be empty"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var err error
	if req.MessageType == "card" {
		// Send as card message
		cardMsg := integrations.CardMessage{
			Cards: []integrations.Card{
				{
					Header: integrations.CardHeader{
						Title:    req.CardTitle,
						Subtitle: req.CardSubtitle,
					},
					Sections: []integrations.CardSection{
						{
							Widgets: []integrations.Widget{
								{
									TextParagraph: &integrations.TextParagraph{
										Text: req.Message,
									},
								},
							},
						},
					},
				},
			},
		}
		err = integrations.SendCardMessage(config.Config.WebhookURL, cardMsg)
	} else {
		// Send as simple message
		err = integrations.SendSimpleMessage(config.Config.WebhookURL, req.Message)
	}

	if err != nil {
		log.Printf("Failed to send message: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to send message: %v", err)})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Message sent successfully!"})
}

// SendReminderHandler is deprecated - use per-standup reminder endpoints instead
func SendReminderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "This endpoint is deprecated. Use /api/standups/{id}/send to trigger specific standup reminders.",
	})
}

// HealthHandler returns the health status of the service
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"service": "standup-bot",
	})
}
