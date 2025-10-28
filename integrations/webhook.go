package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
