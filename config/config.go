package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// AppConfig holds all application configuration
type AppConfig struct {
	WebhookURL   string
	Port         string
	DatabasePath string
	ReminderTime string
	Timezone     string
	SkipWeekends bool
	LogLevel     string
}

var Config *AppConfig

// LoadConfig loads configuration from environment variables
func LoadConfig() error {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	Config = &AppConfig{
		WebhookURL:   getEnv("GOOGLE_CHAT_WEBHOOK_URL", ""),
		Port:         getEnv("PORT", "8080"),
		DatabasePath: getEnv("DATABASE_PATH", "./standup_bot.db"),
		ReminderTime: getEnv("REMINDER_TIME", "09:00"),
		Timezone:     getEnv("TIMEZONE", "UTC"),
		SkipWeekends: getEnv("SKIP_WEEKENDS", "true") == "true",
		LogLevel:     getEnv("LOG_LEVEL", "info"),
	}

	// Validate required config
	if Config.WebhookURL == "" {
		log.Fatal("GOOGLE_CHAT_WEBHOOK_URL environment variable is required")
	}

	log.Printf("Configuration loaded successfully")
	log.Printf("  Port: %s", Config.Port)
	log.Printf("  Database: %s", Config.DatabasePath)
	log.Printf("  Reminder Time: %s", Config.ReminderTime)
	log.Printf("  Timezone: %s", Config.Timezone)
	log.Printf("  Skip Weekends: %t", Config.SkipWeekends)

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
