# Google Chat Bot

A simple Go application to send messages to Google Chat via webhooks.

## Features

- Send simple text messages to Google Chat
- Send rich card messages with headers, key-value pairs, and formatted text
- Easy configuration via environment variables
- No external dependencies (uses only Go standard library)

## Prerequisites

- Go 1.16 or higher
- A Google Chat webhook URL

## Getting Your Webhook URL

1. Open Google Chat and go to the space where you want to receive messages
2. Click on the space name at the top
3. Select "Manage webhooks"
4. Create a new webhook or use an existing one
5. Copy the webhook URL (it will look like: `https://chat.googleapis.com/v1/spaces/XXXXX/messages?key=XXXXX&token=XXXXX`)

## Setup

1. Clone or create this project:
```bash
cd google-chat-bot
```

2. Copy the example environment file:
```bash
cp .env.example .env
```

3. Edit `.env` and add your webhook URL:
```bash
GOOGLE_CHAT_WEBHOOK_URL=https://chat.googleapis.com/v1/spaces/YOUR_SPACE_ID/messages?key=YOUR_KEY&token=YOUR_TOKEN
```

## Usage

### Running the example

Export the environment variable and run:
```bash
export GOOGLE_CHAT_WEBHOOK_URL="your_webhook_url_here"
go run main.go
```

Or source the .env file (if using a tool like direnv or manually):
```bash
source .env
go run main.go
```

### Building the binary

```bash
go build -o google-chat-bot
```

Then run:
```bash
export GOOGLE_CHAT_WEBHOOK_URL="your_webhook_url_here"
./google-chat-bot
```

## Code Examples

### Simple Text Message

```go
package main

import (
    "log"
)

func main() {
    webhookURL := "your_webhook_url"

    err := SendSimpleMessage(webhookURL, "Hello from Go!")
    if err != nil {
        log.Fatal(err)
    }
}
```

### Card Message

```go
package main

import (
    "log"
)

func main() {
    webhookURL := "your_webhook_url"

    cardMsg := CardMessage{
        Cards: []Card{
            {
                Header: CardHeader{
                    Title:    "Deployment Status",
                    Subtitle: "Production Environment",
                },
                Sections: []CardSection{
                    {
                        Widgets: []Widget{
                            {
                                KeyValue: &KeyValue{
                                    TopLabel: "Status",
                                    Content:  "Deployed",
                                },
                            },
                            {
                                KeyValue: &KeyValue{
                                    TopLabel: "Version",
                                    Content:  "v1.2.3",
                                },
                            },
                            {
                                TextParagraph: &TextParagraph{
                                    Text: "Deployment completed successfully at 2024-01-15 10:30:00",
                                },
                            },
                        },
                    },
                },
            },
        },
    }

    err := SendCardMessage(webhookURL, cardMsg)
    if err != nil {
        log.Fatal(err)
    }
}
```

## Use Cases

- System monitoring alerts
- CI/CD pipeline notifications
- Application error reporting
- Status updates
- Scheduled reports
- Integration with other services

## Project Structure

```
google-chat-bot/
├── main.go           # Main application with message sending functions
├── go.mod            # Go module definition
├── .env.example      # Example environment configuration
├── .gitignore        # Git ignore rules
└── README.md         # This file
```

## API Reference

### SendSimpleMessage

Sends a simple text message to Google Chat.

```go
func SendSimpleMessage(webhookURL, message string) error
```

**Parameters:**
- `webhookURL`: Your Google Chat webhook URL
- `message`: The text message to send

**Returns:** Error if sending fails, nil on success

### SendCardMessage

Sends a formatted card message to Google Chat.

```go
func SendCardMessage(webhookURL string, cardMsg CardMessage) error
```

**Parameters:**
- `webhookURL`: Your Google Chat webhook URL
- `cardMsg`: A CardMessage struct containing the card content

**Returns:** Error if sending fails, nil on success

## License

MIT

## Contributing

Feel free to submit issues and pull requests!
