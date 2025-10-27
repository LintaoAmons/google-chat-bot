package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

// Global webhook URL
var webhookURL string

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

// HTML template for the web UI
const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Google Chat Bot - Send Message</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 20px;
        }
        .container {
            background: white;
            padding: 40px;
            border-radius: 10px;
            box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
            max-width: 600px;
            width: 100%;
        }
        h1 {
            color: #333;
            margin-bottom: 10px;
            font-size: 28px;
        }
        .subtitle {
            color: #666;
            margin-bottom: 30px;
            font-size: 14px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 8px;
            color: #444;
            font-weight: 500;
        }
        textarea, input[type="text"] {
            width: 100%;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 5px;
            font-size: 14px;
            font-family: inherit;
            transition: border-color 0.3s;
        }
        textarea {
            min-height: 120px;
            resize: vertical;
        }
        textarea:focus, input[type="text"]:focus {
            outline: none;
            border-color: #667eea;
        }
        .button-group {
            display: flex;
            gap: 10px;
        }
        button {
            flex: 1;
            padding: 12px 24px;
            border: none;
            border-radius: 5px;
            font-size: 16px;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.3s;
        }
        .btn-primary {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }
        .btn-primary:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 20px rgba(102, 126, 234, 0.4);
        }
        .btn-secondary {
            background: #f0f0f0;
            color: #333;
        }
        .btn-secondary:hover {
            background: #e0e0e0;
        }
        #response {
            margin-top: 20px;
            padding: 15px;
            border-radius: 5px;
            display: none;
        }
        .success {
            background: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        .error {
            background: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
        .message-type {
            display: flex;
            gap: 20px;
            margin-bottom: 20px;
        }
        .radio-option {
            display: flex;
            align-items: center;
            gap: 8px;
        }
        .radio-option input[type="radio"] {
            width: auto;
        }
        #cardOptions {
            display: none;
            margin-top: 20px;
            padding: 20px;
            background: #f9f9f9;
            border-radius: 5px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Google Chat Bot</h1>
        <p class="subtitle">Send custom messages to your Google Chat space</p>

        <form id="messageForm">
            <div class="form-group message-type">
                <div class="radio-option">
                    <input type="radio" id="simple" name="messageType" value="simple" checked>
                    <label for="simple" style="margin: 0;">Simple Message</label>
                </div>
                <div class="radio-option">
                    <input type="radio" id="card" name="messageType" value="card">
                    <label for="card" style="margin: 0;">Card Message</label>
                </div>
            </div>

            <div class="form-group">
                <label for="message">Message Text</label>
                <textarea id="message" name="message" placeholder="Enter your message here..." required></textarea>
            </div>

            <div id="cardOptions">
                <div class="form-group">
                    <label for="cardTitle">Card Title</label>
                    <input type="text" id="cardTitle" name="cardTitle" placeholder="e.g., Important Notification">
                </div>
                <div class="form-group">
                    <label for="cardSubtitle">Card Subtitle (Optional)</label>
                    <input type="text" id="cardSubtitle" name="cardSubtitle" placeholder="e.g., System Update">
                </div>
            </div>

            <div class="button-group">
                <button type="submit" class="btn-primary">Send Message</button>
                <button type="reset" class="btn-secondary">Clear</button>
            </div>
        </form>

        <div id="response"></div>
    </div>

    <script>
        const form = document.getElementById('messageForm');
        const responseDiv = document.getElementById('response');
        const messageTypeRadios = document.getElementsByName('messageType');
        const cardOptions = document.getElementById('cardOptions');

        // Show/hide card options based on message type
        messageTypeRadios.forEach(radio => {
            radio.addEventListener('change', function() {
                if (this.value === 'card') {
                    cardOptions.style.display = 'block';
                } else {
                    cardOptions.style.display = 'none';
                }
            });
        });

        form.addEventListener('submit', async (e) => {
            e.preventDefault();

            const messageType = document.querySelector('input[name="messageType"]:checked').value;
            const message = document.getElementById('message').value;
            const cardTitle = document.getElementById('cardTitle').value;
            const cardSubtitle = document.getElementById('cardSubtitle').value;

            const data = {
                message: message,
                messageType: messageType
            };

            if (messageType === 'card') {
                data.cardTitle = cardTitle || 'Notification';
                data.cardSubtitle = cardSubtitle;
            }

            try {
                const response = await fetch('/send', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(data)
                });

                const result = await response.json();

                responseDiv.style.display = 'block';
                if (response.ok) {
                    responseDiv.className = 'success';
                    responseDiv.textContent = result.message || 'Message sent successfully!';
                    form.reset();
                    cardOptions.style.display = 'none';
                } else {
                    responseDiv.className = 'error';
                    responseDiv.textContent = result.error || 'Failed to send message';
                }
            } catch (error) {
                responseDiv.style.display = 'block';
                responseDiv.className = 'error';
                responseDiv.textContent = 'Error: ' + error.message;
            }
        });
    </script>
</body>
</html>
`

// homeHandler serves the web UI
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("home").Parse(htmlTemplate)
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

// sendHandler handles the message sending endpoint
func sendHandler(w http.ResponseWriter, r *http.Request) {
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
		cardMsg := CardMessage{
			Cards: []Card{
				{
					Header: CardHeader{
						Title:    req.CardTitle,
						Subtitle: req.CardSubtitle,
					},
					Sections: []CardSection{
						{
							Widgets: []Widget{
								{
									TextParagraph: &TextParagraph{
										Text: req.Message,
									},
								},
							},
						},
					},
				},
			},
		}
		err = SendCardMessage(webhookURL, cardMsg)
	} else {
		// Send as simple message
		err = SendSimpleMessage(webhookURL, req.Message)
	}

	if err != nil {
		log.Printf("Failed to send message: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to send message: %v", err)})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Message sent successfully!"})
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Get webhook URL from environment variable and store in global variable
	webhookURL = os.Getenv("GOOGLE_CHAT_WEBHOOK_URL")
	if webhookURL == "" {
		log.Fatal("GOOGLE_CHAT_WEBHOOK_URL environment variable is not set")
	}

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Set up HTTP routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/send", sendHandler)

	// Start the server
	addr := ":" + port
	fmt.Printf("🚀 Google Chat Bot Web UI starting on http://localhost%s\n", addr)
	fmt.Printf("📝 Open your browser and navigate to http://localhost%s\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
