package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

// Message represents a Google Chat message
type Message struct {
	Text string `json:"text"`
}

// CardMessage represents a more complex Google Chat card message
type CardMessage struct {
	Cards []Card `json:"cards,omitempty"`
	Text  string `json:"text,omitempty"`
}

type Card struct {
	Header   CardHeader    `json:"header,omitempty"`
	Sections []CardSection `json:"sections,omitempty"`
}

type CardHeader struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle,omitempty"`
	ImageURL string `json:"imageUrl,omitempty"`
}

type CardSection struct {
	Widgets []Widget `json:"widgets"`
}

type Widget struct {
	TextParagraph *TextParagraph `json:"textParagraph,omitempty"`
	KeyValue      *KeyValue      `json:"keyValue,omitempty"`
}

type TextParagraph struct {
	Text string `json:"text"`
}

type KeyValue struct {
	TopLabel string `json:"topLabel,omitempty"`
	Content  string `json:"content"`
}

// SendSimpleMessage sends a simple text message to Google Chat webhook
func SendSimpleMessage(webhookURL, message string) error {
	msg := Message{
		Text: message,
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// SendCardMessage sends a card message to Google Chat webhook
func SendCardMessage(webhookURL string, cardMsg CardMessage) error {
	jsonData, err := json.Marshal(cardMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal card message: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send card message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Get webhook URL from environment variable
	webhookURL := os.Getenv("GOOGLE_CHAT_WEBHOOK_URL")
	if webhookURL == "" {
		log.Fatal("GOOGLE_CHAT_WEBHOOK_URL environment variable is not set")
	}

	// Example 1: Send a simple message
	if err := SendSimpleMessage(webhookURL, "Hello from Google Chat Bot! 🤖"); err != nil {
		log.Fatalf("Failed to send simple message: %v", err)
	}
	fmt.Println("Simple message sent successfully!")

	// Example 2: Send a card message
	cardMsg := CardMessage{
		Cards: []Card{
			{
				Header: CardHeader{
					Title:    "Bot Notification",
					Subtitle: "System Status Update",
				},
				Sections: []CardSection{
					{
						Widgets: []Widget{
							{
								KeyValue: &KeyValue{
									TopLabel: "Status",
									Content:  "✅ All systems operational",
								},
							},
							{
								KeyValue: &KeyValue{
									TopLabel: "Uptime",
									Content:  "99.9%",
								},
							},
							{
								TextParagraph: &TextParagraph{
									Text: "This is an example card message from the Google Chat bot.",
								},
							},
						},
					},
				},
			},
		},
	}

	if err := SendCardMessage(webhookURL, cardMsg); err != nil {
		log.Fatalf("Failed to send card message: %v", err)
	}
	fmt.Println("Card message sent successfully!")
}
