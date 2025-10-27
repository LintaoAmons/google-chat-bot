# Google Chat Bot

A simple Go application to send messages to Google Chat via webhooks.

## Features

- Send simple text messages to Google Chat
- Send rich card messages with headers, key-value pairs, and formatted text
- Easy configuration via environment variables
- No external dependencies (uses only Go standard library)

## Prerequisites

- Go 1.16 or higher (for local development)
- Docker (for containerized deployment)
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

### Running with Go (Local Development)

The application automatically loads the `.env` file, so just run:
```bash
go run main.go
```

Or build and run the binary:
```bash
go build -o google-chat-bot
./google-chat-bot
```

You can also export the environment variable directly:
```bash
export GOOGLE_CHAT_WEBHOOK_URL="your_webhook_url_here"
go run main.go
```

### Running with Docker

#### Build locally

```bash
docker build -t google-chat-bot .
```

#### Run the container

```bash
docker run --env-file .env google-chat-bot
```

Or pass the webhook URL directly:
```bash
docker run -e GOOGLE_CHAT_WEBHOOK_URL="your_webhook_url_here" google-chat-bot
```

#### Pull from GitHub Container Registry

Once the GitHub Actions workflow runs, you can pull the pre-built image:

```bash
# Pull the latest version
docker pull ghcr.io/YOUR_GITHUB_USERNAME/google-chat-bot:latest

# Run it
docker run -e GOOGLE_CHAT_WEBHOOK_URL="your_webhook_url_here" ghcr.io/YOUR_GITHUB_USERNAME/google-chat-bot:latest
```

Replace `YOUR_GITHUB_USERNAME` with your actual GitHub username or organization name.

## CI/CD with GitHub Actions

This project includes a GitHub Actions workflow that automatically builds and pushes Docker images to GitHub Container Registry (GHCR).

### Workflow Features

- Builds multi-platform images (linux/amd64, linux/arm64)
- Automatic tagging based on git branches and tags
- Pushes to `ghcr.io/YOUR_USERNAME/google-chat-bot`
- Caching for faster builds
- Triggers on:
  - Push to `main` branch
  - Version tags (e.g., `v1.0.0`)
  - Pull requests
  - Manual workflow dispatch

### Setting Up GitHub Container Registry

1. The workflow uses `GITHUB_TOKEN` which is automatically provided by GitHub Actions
2. Images are public by default, but you can make them private in your repository settings
3. To pull private images, you need to authenticate:

```bash
echo $GITHUB_TOKEN | docker login ghcr.io -u YOUR_USERNAME --password-stdin
```

### Tagging Strategy

The workflow automatically creates the following tags:

- `latest` - Latest build from main branch
- `main` - Latest build from main branch
- `v1.2.3` - Semantic version tags
- `v1.2` - Major.minor version
- `v1` - Major version
- `main-<sha>` - Branch name with commit SHA

### Creating a Release

To create a new release and trigger the build:

```bash
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

This will automatically build and push the image with tags `v1.0.0`, `v1.0`, `v1`, and `latest`.

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
├── .github/
│   └── workflows/
│       └── docker-build.yml  # GitHub Actions workflow for building and pushing Docker images
├── main.go                   # Main application with message sending functions
├── go.mod                    # Go module definition
├── go.sum                    # Go dependencies checksums
├── Dockerfile                # Docker container definition
├── .dockerignore             # Docker build ignore rules
├── .env                      # Environment variables (not committed)
├── .env.example              # Example environment configuration
├── .gitignore                # Git ignore rules
└── README.md                 # This file
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
