# Standup Bot Implementation Document

## Executive Summary

This document outlines the implementation plan for transforming the current Google Chat webhook bot into a standup reminder daemon. The bot will run as a background service that sends scheduled daily standup reminders to Google Chat, manages a user roster with leave tracking, and includes fun "roast" messages to encourage participation. **This is a one-way communication system** - the bot only sends messages via webhook and does not receive or process user responses.

## Current State Analysis

### Existing Implementation
- **Architecture**: Simple HTTP server with webhook-based message sending
- **Tech Stack**: Go 1.25.2, single-file monolithic design (main.go)
- **Features**: Send simple/card messages via web UI to Google Chat
- **Communication**: One-way (send only) via incoming webhooks
- **Storage**: Stateless, no database

### Limitations for Standup Reminder Daemon
1. No persistent storage for roster and roasts data
2. No scheduled task execution capability

### Design Constraints (Intentional)
1. **One-way communication only** - Bot sends messages via webhook, does not receive responses
2. **No interactive features** - No slash commands, buttons, or dialogs (Google Chat cannot send data back to the app)
3. **Web UI for management** - All roster, roast, and leave management done via web interface

## Proposed Standup Bot Features

### Core Features

#### 1. Roster Management (via Web UI)
- **Add users to roster**: Manually add team members via web interface
- **Remove users from roster**: Deactivate users anytime
- **List active roster**: View current standup participants
- **User status tracking**: Track active/inactive status

#### 2. Leave Management (via Web UI)
- **Record leave**: Track sick leave, vacation, PTO with start/end dates
- **Leave history**: Complete historical record of all absences
- **View current leaves**: See who's currently on leave
- **Auto-return**: Automatically detect expired leaves and mark users as returned
- **Leave types**: Support different leave categories (sick, vacation, PTO, personal, etc.)

#### 3. Roast Management (via Web UI)
- **Create roasts**: Add fun roast messages
- **Update roasts**: Edit existing roast messages
- **Delete roasts**: Remove roast messages
- **List roasts**: View all available roast messages
- **Random selection**: Automatically pick random roasts for reminders

#### 4. Standup Scheduling (Automated Daemon)
- **Daily reminders**: Scheduled standup reminders at configurable times
- **Roast integration**: Include random roast in reminder messages
- **Skip users on leave**: Don't send reminders to users currently on leave
- **Skip weekends/holidays**: Configurable skip days
- **Webhook delivery**: Send messages via Google Chat webhook

## Technical Architecture

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Google Chat Space                        │
│               (Receives messages only)                       │
└──────────────────────▲──────────────────────────────────────┘
                       │
                       │ HTTPS Webhook (One-way: Bot → Chat)
                       │
┌──────────────────────┴──────────────────────────────────────┐
│              Standup Bot Daemon (Go)                         │
│                                                              │
│  ┌────────────────┐  ┌─────────────────┐                   │
│  │  HTTP Server   │  │  Cron Scheduler │                   │
│  │  (Web UI)      │  │  (Reminders)    │                   │
│  └────────────────┘  └─────────────────┘                   │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │              Business Logic Layer                       ││
│  │  • Roster Manager  • Roast Manager  • Leave Manager     ││
│  └─────────────────────────────────────────────────────────┘│
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │              Data Access Layer (SQLite)                 ││
│  └─────────────────────────────────────────────────────────┘│
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                   SQLite Database                            │
│         • users  • roasts  • leaves  • config               │
└─────────────────────────────────────────────────────────────┘
```

**Key Points:**
- **One-way communication**: Bot sends messages to Google Chat via webhook only
- **Web UI**: Admin interface for managing roster, roasts, and leaves
- **Cron scheduler**: Automated daily reminders at configured times
- **SQLite**: Lightweight persistent storage, no external database needed

### Database Schema

#### Users Table
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    google_chat_user_id TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    email TEXT,
    is_active BOOLEAN DEFAULT 1,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_active ON users(is_active);
CREATE INDEX idx_users_google_chat_id ON users(google_chat_user_id);
```

#### Leaves Table
```sql
CREATE TABLE leaves (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    leave_type TEXT NOT NULL,              -- 'sick', 'vacation', 'pto', 'personal', etc.
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    reason TEXT,
    status TEXT DEFAULT 'active',          -- 'active', 'completed', 'cancelled'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_leaves_user ON leaves(user_id);
CREATE INDEX idx_leaves_dates ON leaves(start_date, end_date);
CREATE INDEX idx_leaves_status ON leaves(status);
```

**Query for users eligible for reminders:**
```sql
SELECT * FROM users
WHERE is_active = 1
AND id NOT IN (
    SELECT user_id FROM leaves
    WHERE status = 'active'
    AND start_date <= date('now')
    AND end_date >= date('now')
)
```

#### Roasts Table
```sql
CREATE TABLE roasts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message TEXT NOT NULL,
    created_by TEXT,
    is_active BOOLEAN DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_roasts_active ON roasts(is_active);
```

#### Config Table
```sql
CREATE TABLE config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Default config values
INSERT INTO config (key, value) VALUES
    ('reminder_time', '09:00'),
    ('timezone', 'UTC'),
    ('skip_weekends', 'true');
```

### API Endpoints

#### Web UI Endpoints (Existing)
- `GET /` - Web UI home page
- `POST /send` - Send message to Google Chat

#### New API Endpoints

##### Roster Management
- `POST /api/roster` - Add user to roster
- `DELETE /api/roster/:id` - Remove user from roster
- `GET /api/roster` - List all roster members
- `GET /api/roster/active` - List active members (not on leave)

##### Leave Management
- `POST /api/leaves` - Create leave record
- `PUT /api/leaves/:id` - Update leave record
- `DELETE /api/leaves/:id` - Cancel leave
- `GET /api/leaves` - List all leaves
- `GET /api/leaves/active` - Get currently active leaves
- `GET /api/leaves/user/:userId` - Get leave history for a user

##### Roast Management
- `POST /api/roasts` - Create new roast
- `PUT /api/roasts/:id` - Update existing roast
- `DELETE /api/roasts/:id` - Delete roast
- `GET /api/roasts` - List all roasts
- `GET /api/roasts/random` - Get random roast

##### Configuration
- `GET /api/config` - Get all config values
- `PUT /api/config/:key` - Update config value

##### Admin/Testing
- `POST /api/send-reminder` - Manually trigger reminder (for testing)

### Code Structure

```
google-chat-bot/
├── main.go                          # Application entry point
├── config/
│   └── config.go                    # Configuration management
├── database/
│   ├── db.go                        # Database connection
│   ├── migrations.go                # Schema migrations
│   └── models.go                    # Data models (User, Leave, Roast)
├── handlers/
│   ├── web.go                       # Web UI handlers
│   ├── roster.go                    # Roster API handlers
│   ├── leaves.go                    # Leave API handlers
│   ├── roasts.go                    # Roast API handlers
│   └── config.go                    # Config API handlers
├── services/
│   ├── roster.go                    # Roster business logic
│   ├── leaves.go                    # Leave management logic
│   ├── roasts.go                    # Roast management logic
│   └── scheduler.go                 # Cron jobs & reminder logic
├── integrations/
│   └── webhook.go                   # Google Chat webhook client
├── templates/
│   └── ui.html                      # Web UI template
├── static/
│   ├── css/
│   │   └── style.css
│   └── js/
│       └── app.js
├── go.mod
├── go.sum
├── Dockerfile
└── README.md
```

## Implementation Phases

### Phase 1: Database Foundation
**Duration**: 1 day

**Tasks**:
1. Add SQLite database integration
   - Install `github.com/mattn/go-sqlite3` driver
   - Create database connection manager
   - Implement database migrations
   - Create data models (User, Leave, Roast, Config)

2. Refactor code structure
   - Create package structure (database/, handlers/, services/)
   - Move existing code into appropriate packages
   - Set up configuration management

**Deliverables**:
- ✅ Database schema implemented
- ✅ Database connection working
- ✅ Migrations system in place
- ✅ Code reorganized into packages

### Phase 2: Core Business Logic & Web UI
**Duration**: 1-2 days

**Tasks**:
1. Implement Roster Management
   - Service layer for roster CRUD
   - API handlers for roster endpoints
   - Web UI for roster management

2. Implement Leave Management
   - Service layer for leave CRUD
   - API handlers for leave endpoints
   - Web UI for leave tracking
   - Auto-expire leaves logic

3. Implement Roast Management
   - Service layer for roast CRUD
   - API handlers for roast endpoints
   - Web UI for roast management

4. Update Web UI
   - Modernize existing UI
   - Add roster management interface
   - Add leave management interface
   - Add roast management interface

**Deliverables**:
- ✅ Full CRUD operations for roster, leaves, and roasts
- ✅ Complete web UI for all management tasks
- ✅ API endpoints functional

### Phase 3: Scheduling & Automation
**Duration**: 1 day

**Tasks**:
1. Add Scheduler (using `github.com/robfig/cron/v3`)
   - Daily reminder job
   - Daily leave expiration check
   - Configurable schedule via database

2. Implement Reminder Logic
   - Query eligible users (active, not on leave)
   - Fetch random roast
   - Format reminder message
   - Send via webhook

3. Implement Leave Expiration
   - Check for leaves past end_date
   - Update status to 'completed'
   - Optional: log or notify

**Deliverables**:
- ✅ Automated daily reminders
- ✅ Automatic leave expiration
- ✅ Configurable schedule
- ✅ Roast integration in reminders

### Phase 4: Polish & Production Readiness
**Duration**: 1 day

**Tasks**:
1. Error Handling & Logging
   - Comprehensive error handling
   - Structured logging
   - Graceful degradation

2. Testing
   - Unit tests for services
   - Integration tests for database
   - Manual testing of all features

3. Documentation
   - README with setup instructions
   - API documentation
   - Deployment guide

4. Docker & Deployment
   - Update Dockerfile for new dependencies
   - Add volume for SQLite database
   - Environment variable configuration
   - Health check endpoint

**Deliverables**:
- ✅ Production-ready code
- ✅ Comprehensive error handling
- ✅ Complete documentation
- ✅ Docker deployment ready

## Technical Dependencies

### New Go Packages Required
```go
require (
    github.com/joho/godotenv v1.5.1              // Existing - Environment variables
    github.com/mattn/go-sqlite3 v1.14.18         // Database driver
    github.com/robfig/cron/v3 v3.0.1            // Cron scheduler
    github.com/gorilla/mux v1.8.1               // HTTP routing (optional)
)
```

### External Services
- **Google Chat Webhook URL**: For sending messages (already configured)

## Configuration

### Environment Variables
```bash
# Existing
GOOGLE_CHAT_WEBHOOK_URL=          # Google Chat incoming webhook URL
PORT=8080                         # HTTP server port

# New
DATABASE_PATH=./standup_bot.db     # SQLite database file path
REMINDER_TIME=09:00                # Daily reminder time (HH:MM)
TIMEZONE=UTC                       # Timezone for scheduling
SKIP_WEEKENDS=true                 # Skip reminders on weekends
LOG_LEVEL=info                     # Logging level (debug, info, warn, error)
```

### Database Configuration
Stored in `config` table, can be modified via Web UI:
- `reminder_time` - When to send daily reminders
- `timezone` - Timezone for scheduling
- `skip_weekends` - Whether to skip weekends

## Deployment Strategy

### Development
1. Local SQLite database
2. Hot reload during development
3. Test cron jobs manually

### Production
1. **Build Docker image**
   ```bash
   docker build -t standup-bot:latest .
   ```

2. **Run with volume for database**
   ```bash
   docker run -d \
     -p 8080:8080 \
     -v /path/to/data:/data \
     -e DATABASE_PATH=/data/standup_bot.db \
     -e GOOGLE_CHAT_WEBHOOK_URL=https://... \
     -e REMINDER_TIME=09:00 \
     --name standup-bot \
     standup-bot:latest
   ```

3. **Database migrations run automatically on startup**

4. **Monitor logs**
   ```bash
   docker logs -f standup-bot
   ```

### Rollback Plan
- Keep previous Docker image tagged
- Database migrations include version tracking
- Can rollback to previous image if needed

## Security Considerations

1. **Data Protection**: SQLite database file permissions
2. **API Security**: Input validation on all endpoints
3. **Secrets Management**: Never commit webhook URLs or credentials
4. **Access Control**: Web UI should be password-protected (future enhancement)

## Monitoring & Health Checks

1. **Health Endpoint**: `GET /health` returns service status
2. **Logging**: Structured logs with timestamps
3. **Metrics**: Track reminder send success/failure rates
4. **Alerts**: Monitor cron job failures

## Feature Comparison

| Feature | Current | After Implementation |
|---------|---------|---------------------|
| Send messages | ✅ Manual via UI | ✅ Manual + Automated |
| Roster management | ❌ | ✅ Full CRUD via UI |
| Leave tracking | ❌ | ✅ Full history |
| Roast messages | ❌ | ✅ Randomized |
| Scheduled reminders | ❌ | ✅ Configurable cron |
| Database | ❌ | ✅ SQLite |
| Skip users on leave | ❌ | ✅ Automatic |

## Open Questions for Review

1. **Roster Management**:
   - Should we require email for all users?
   - Support for user avatars/profile pictures?

2. **Leave Management**:
   - Should leaves support half-days?
   - Notification when leave is about to end?
   - Approval workflow needed?

3. **Roast Strategy**:
   - Admin-only roast creation or allow all users?
   - Severity levels for roasts (gentle → harsh)?
   - Personalized roasts per user?

4. **Reminder Strategy**:
   - What time should reminders be sent?
   - Should we send multiple reminders?
   - Format: simple text or card message?

5. **Access Control**:
   - Should Web UI be password-protected?
   - API authentication needed?
   - Different permission levels (admin vs regular user)?

## Estimated Timeline

- **Phase 1** (Database): 1 day
- **Phase 2** (Business Logic & UI): 1-2 days
- **Phase 3** (Scheduling): 1 day
- **Phase 4** (Polish & Deploy): 1 day

**Total**: 4-5 days for full implementation

With focused development: ~1 week
With part-time development: ~2 weeks

## Next Steps

1. ✅ **Review this document** - Confirm simplified architecture
2. **Answer open questions** - Clarify requirements
3. **Begin Phase 1** - Set up database and migrations
4. **Iterate quickly** - Ship Phase 1, then Phase 2, etc.

---

**Document Version**: 2.0
**Last Updated**: 2025-10-27
**Author**: Claude Code
**Status**: Ready for Implementation
