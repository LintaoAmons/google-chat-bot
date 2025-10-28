# Standup Bot for Google Chat

An automated standup reminder bot for Google Chat that manages team rosters, tracks leaves, and sends daily reminders with fun roast messages to keep your team engaged.

## 🌟 Features

### Core Functionality
- **📅 Automated Daily Reminders** - Scheduled standup reminders at configurable times
- **👥 Roster Management** - Track active team members and their status
- **🏖️ Leave Tracking** - Complete leave history (sick, vacation, PTO, personal)
- **🎭 Roast Messages** - Fun motivational messages to encourage participation
- **🎯 Smart Filtering** - Only sends reminders to active users not on leave
- **⏰ Cron Scheduler** - Automated jobs for reminders and leave expiration
- **🌐 Web Management Interface** - Full CRUD operations via modern web UI

### Management Features
- **Roster**: Add/deactivate/reactivate users
- **Leaves**: Track absences with start/end dates and reasons
- **Roasts**: Create and manage fun reminder messages
- **Status Tracking**: Visual badges for active/inactive states
- **Leave Expiration**: Automatic completion of past leaves

## 📋 Prerequisites

- Go 1.25.2 or higher
- Docker (optional, for containerized deployment)
- A Google Chat webhook URL

## 🚀 Quick Start

### 1. Get Your Webhook URL

1. Open Google Chat and go to your team space
2. Click the space name → **Manage webhooks**
3. Create a new webhook
4. Copy the webhook URL

### 2. Setup

```bash
# Clone the repository
cd google-chat-bot

# Copy environment file
cp .env.example .env

# Edit .env and add your webhook URL
GOOGLE_CHAT_WEBHOOK_URL=https://chat.googleapis.com/v1/spaces/...
DATABASE_PATH=./standup_bot.db
REMINDER_TIME=09:00
TIMEZONE=UTC
SKIP_WEEKENDS=true
PORT=8080
```

### 3. Run Locally

```bash
# Install dependencies
go mod download

# Build and run
go build -o standup-bot
./standup-bot
```

The bot will start on http://localhost:8080

### 4. Run with Docker

```bash
# Build image
docker build -t standup-bot .

# Run container
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/data:/data \
  -e DATABASE_PATH=/data/standup_bot.db \
  -e GOOGLE_CHAT_WEBHOOK_URL="your_webhook_url" \
  -e REMINDER_TIME=09:00 \
  --name standup-bot \
  standup-bot
```

## 🎨 Web Interface

Access the management interface at `http://localhost:8080`

### Tabs

1. **Send Message** - Send custom messages and test reminders
2. **Roster** - Manage team members
3. **Leaves** - Track absences and leave history
4. **Roasts** - Create fun reminder messages

### Roster Management

**Add New User:**
- Google Chat User ID (email or username)
- Display Name
- Email (optional)

**Actions:**
- View all users or active only
- Deactivate/Reactivate users
- See join dates and status

### Leave Management

**Record Leave:**
- Select user from roster
- Choose leave type (sick, vacation, pto, personal)
- Set start and end dates
- Add optional reason

**Features:**
- View all or active leaves only
- Cancel active leaves
- Automatic expiration when end date passes
- Status tracking (active/completed/cancelled)

### Roast Management

**Add Roasts:**
- Write fun motivational messages
- Attribute to creator (optional)
- Activate/deactivate messages

**Example Roasts:**
- "Rise and shine! ☀️ Time to show what you're made of!"
- "Your standup update is like coffee - the team needs it to function!"
- "Don't make me send another reminder... 😏"

## ⚙️ Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GOOGLE_CHAT_WEBHOOK_URL` | *required* | Google Chat webhook URL |
| `DATABASE_PATH` | `./standup_bot.db` | SQLite database file path |
| `PORT` | `8080` | HTTP server port |
| `REMINDER_TIME` | `09:00` | Daily reminder time (HH:MM) |
| `TIMEZONE` | `UTC` | Timezone for scheduling |
| `SKIP_WEEKENDS` | `true` | Skip reminders on weekends |
| `LOG_LEVEL` | `info` | Logging level |

### Database Configuration

Runtime configuration stored in database (modifiable via API):
- `reminder_time` - When to send daily reminders
- `timezone` - Timezone for scheduling
- `skip_weekends` - Whether to skip weekend reminders

## 🔌 API Reference

### Roster Endpoints

```bash
# Get all users
GET /api/roster

# Get active users only
GET /api/roster?active=true

# Get single user
GET /api/roster/:id

# Create user
POST /api/roster
Content-Type: application/json
{
  "google_chat_user_id": "user@example.com",
  "display_name": "John Doe",
  "email": "john.doe@example.com"
}

# Update user
PUT /api/roster/:id
Content-Type: application/json
{
  "display_name": "Jane Doe",
  "email": "jane.doe@example.com"
}

# Deactivate user
DELETE /api/roster/:id

# Reactivate user
POST /api/roster/:id/reactivate
```

### Leaves Endpoints

```bash
# Get all leaves
GET /api/leaves

# Get active leaves only
GET /api/leaves?active=true

# Get leaves for specific user
GET /api/leaves?user_id=1

# Create leave
POST /api/leaves
Content-Type: application/json
{
  "user_id": 1,
  "leave_type": "vacation",
  "start_date": "2025-01-15",
  "end_date": "2025-01-20",
  "reason": "Family vacation"
}

# Update leave
PUT /api/leaves/:id
Content-Type: application/json
{
  "leave_type": "sick",
  "start_date": "2025-01-15",
  "end_date": "2025-01-17",
  "reason": "Flu"
}

# Cancel leave
DELETE /api/leaves/:id
```

### Roasts Endpoints

```bash
# Get all roasts
GET /api/roasts

# Get active roasts only
GET /api/roasts?active=true

# Get random roast
GET /api/roasts?random=true

# Create roast
POST /api/roasts
Content-Type: application/json
{
  "message": "Time to shine! ✨",
  "created_by": "Admin"
}

# Update roast
PUT /api/roasts/:id
Content-Type: application/json
{
  "message": "Updated roast message!"
}

# Delete roast
DELETE /api/roasts/:id
```

### Admin Endpoints

```bash
# Health check
GET /health

# Manual reminder trigger (for testing)
POST /api/send-reminder

# Send custom message
POST /send
Content-Type: application/json
{
  "message": "Custom message to team",
  "messageType": "simple"
}
```

## 🏗️ Architecture

```
standup-bot/
├── main.go                     # Application entry point with routing
├── config/
│   └── config.go              # Configuration management
├── database/
│   ├── db.go                  # Database connection
│   ├── migrations.go          # Schema migrations
│   └── models.go              # Data models
├── services/
│   ├── roster.go              # Roster business logic
│   ├── leaves.go              # Leave management
│   ├── roasts.go              # Roast management
│   └── scheduler.go           # Cron jobs
├── handlers/
│   ├── web.go                 # Web UI handlers
│   ├── roster.go              # Roster API
│   ├── leaves.go              # Leaves API
│   └── roasts.go              # Roasts API
├── integrations/
│   └── webhook.go             # Google Chat webhook client
├── templates/
│   └── ui.html                # Web management interface
├── go.mod                     # Go dependencies
├── Dockerfile                 # Container definition
└── README.md                  # This file
```

## 🗄️ Database Schema

### Tables

- **users** - Team roster with active status
- **leaves** - Leave history with type, dates, and status
- **roasts** - Roast message library
- **config** - Runtime configuration

### Data Models

**User:**
- ID, Google Chat User ID, Display Name, Email
- Active status, Join/Leave dates
- Timestamps

**Leave:**
- ID, User ID, Leave Type (sick/vacation/pto/personal)
- Start/End dates, Reason
- Status (active/completed/cancelled)
- Timestamps

**Roast:**
- ID, Message, Created By
- Active status
- Timestamps

## ⏰ Scheduled Jobs

### Daily Reminder (Configurable Time)

**Runs:** Every day at configured time (default: 09:00)

**Actions:**
1. Check if weekend (skip if `SKIP_WEEKENDS=true`)
2. Query eligible users (active, not on leave)
3. Select random roast message
4. Send reminder to Google Chat with user list

**Message Format:**
```
🌅 Good morning, team!

It's time for your daily standup update!

📝 Reminder:
[Random Roast Message]

Team members to submit:
• John Doe
• Jane Smith
• Bob Johnson

Have a great day! ☀️
```

### Leave Expiration (Daily at Midnight)

**Runs:** Every day at 00:00

**Actions:**
1. Find leaves with `status = 'active'` and `end_date < today`
2. Update status to 'completed'
3. Log completion

## 🚢 Deployment

### Docker Compose (Recommended)

```yaml
version: '3.8'
services:
  standup-bot:
    image: standup-bot
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
    environment:
      - DATABASE_PATH=/data/standup_bot.db
      - GOOGLE_CHAT_WEBHOOK_URL=${GOOGLE_CHAT_WEBHOOK_URL}
      - REMINDER_TIME=09:00
      - TIMEZONE=UTC
      - SKIP_WEEKENDS=true
    restart: unless-stopped
```

Run:
```bash
docker-compose up -d
```

### Cloud Run / GCP

See [GCP_CLOUD_RUN_SETUP.md](./GCP_CLOUD_RUN_SETUP.md) for detailed deployment instructions.

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: standup-bot
spec:
  replicas: 1
  selector:
    matchLabels:
      app: standup-bot
  template:
    metadata:
      labels:
        app: standup-bot
    spec:
      containers:
      - name: standup-bot
        image: standup-bot:latest
        ports:
        - containerPort: 8080
        env:
        - name: GOOGLE_CHAT_WEBHOOK_URL
          valueFrom:
            secretKeyRef:
              name: standup-bot-secrets
              key: webhook-url
        - name: DATABASE_PATH
          value: /data/standup_bot.db
        volumeMounts:
        - name: data
          mountPath: /data
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: standup-bot-pvc
```

## 🧪 Testing

### Test Reminder

Use the web UI "Test Reminder" button or:

```bash
curl -X POST http://localhost:8080/api/send-reminder
```

### Manual Operations

```bash
# Add a user
curl -X POST http://localhost:8080/api/roster \
  -H "Content-Type: application/json" \
  -d '{
    "google_chat_user_id": "test@example.com",
    "display_name": "Test User",
    "email": "test@example.com"
  }'

# Add a roast
curl -X POST http://localhost:8080/api/roasts \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Time for standup! ⏰",
    "created_by": "Admin"
  }'

# Add a leave
curl -X POST http://localhost:8080/api/leaves \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "leave_type": "vacation",
    "start_date": "2025-01-20",
    "end_date": "2025-01-25",
    "reason": "Holiday"
  }'
```

## 🛠️ Development

### Prerequisites

- Go 1.25.2+
- SQLite3

### Setup

```bash
# Install dependencies
go mod download

# Run locally
go run main.go

# Build
go build -o standup-bot

# Run tests
go test ./...
```

### Project Structure

The project follows a clean architecture pattern:

- **config/** - Configuration management
- **database/** - Data layer (models, migrations, queries)
- **services/** - Business logic layer
- **handlers/** - HTTP handlers and API endpoints
- **integrations/** - External service clients (Google Chat)
- **templates/** - HTML templates

## 🔒 Security

- **No authentication** - Web UI is open (suitable for internal networks)
- **Environment variables** - Secrets stored securely
- **Input validation** - All user inputs validated
- **SQL injection** - Protected via parameterized queries
- **CORS** - Not enabled (single-origin app)

**Production Recommendations:**
- Deploy behind VPN or firewall
- Add authentication middleware
- Use HTTPS/TLS
- Implement rate limiting
- Enable access logs

## 📝 License

MIT

## 🤝 Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## 🐛 Troubleshooting

### Bot not sending reminders

- Check `GOOGLE_CHAT_WEBHOOK_URL` is correct
- Verify scheduler is running (check logs)
- Test manually: `POST /api/send-reminder`
- Check if users are active and not on leave

### Database errors

- Ensure `DATABASE_PATH` directory exists and is writable
- Check disk space
- Verify SQLite version compatibility

### Web UI not loading

- Check port is available: `netstat -an | grep 8080`
- Verify `templates/ui.html` exists
- Check browser console for errors

### Logs

```bash
# View Docker logs
docker logs -f standup-bot

# Local run logs
./standup-bot 2>&1 | tee standup-bot.log
```

## 📚 Additional Resources

- [Google Chat Webhooks Documentation](https://developers.google.com/chat/how-tos/webhooks)
- [Go Documentation](https://golang.org/doc/)
- [SQLite Documentation](https://www.sqlite.org/docs.html)
- [Cron Expression Guide](https://crontab.guru/)

---

**Version:** 2.0.0
**Last Updated:** 2025-10-27
**Maintainer:** Generated with [Claude Code](https://claude.com/claude-code)
