# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies for SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled for SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o standup-bot .

# Final stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/standup-bot .

# Copy templates directory
COPY --from=builder /app/templates ./templates

# Create data directory for SQLite database
RUN mkdir -p /data

# Expose port
EXPOSE 8080

# Set default environment variables
ENV DATABASE_PATH=/data/standup_bot.db
ENV PORT=8080
ENV REMINDER_TIME=09:00
ENV TIMEZONE=UTC
ENV SKIP_WEEKENDS=true
ENV LOG_LEVEL=info

# Run the application
CMD ["./standup-bot"]
