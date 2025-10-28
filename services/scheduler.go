package services

import (
	"fmt"
	"log"
	"time"

	"google-chat-bot/config"
	"google-chat-bot/database"
	"google-chat-bot/integrations"
	"github.com/robfig/cron/v3"
)

var cronScheduler *cron.Cron

// StartScheduler initializes and starts the cron scheduler
func StartScheduler() error {
	cronScheduler = cron.New()

	// Schedule leave expiration check (runs daily at midnight)
	_, err := cronScheduler.AddFunc("0 0 * * *", ExpireLeaves)
	if err != nil {
		return fmt.Errorf("failed to schedule leave expiration: %w", err)
	}

	// Schedule all active standups
	err = ScheduleAllStandups()
	if err != nil {
		return fmt.Errorf("failed to schedule standups: %w", err)
	}

	cronScheduler.Start()
	log.Println("Scheduler started")
	return nil
}

// StopScheduler stops the cron scheduler
func StopScheduler() {
	if cronScheduler != nil {
		cronScheduler.Stop()
		log.Println("Scheduler stopped")
	}
}

// ScheduleAllStandups schedules reminder jobs for all active standups
func ScheduleAllStandups() error {
	standups, err := GetActiveStandups()
	if err != nil {
		return fmt.Errorf("failed to get active standups: %w", err)
	}

	for _, standup := range standups {
		err = ScheduleStandup(standup)
		if err != nil {
			log.Printf("Warning: Failed to schedule standup %d (%s): %v", standup.ID, standup.Name, err)
		}
	}

	log.Printf("Scheduled %d active standup(s)", len(standups))
	return nil
}

// ScheduleStandup schedules a single standup
func ScheduleStandup(standup database.Standup) error {
	// Parse run_at time (format: HH:MM)
	parsedTime, err := time.Parse("15:04", standup.RunAt)
	if err != nil {
		return fmt.Errorf("invalid run_at time format: %w", err)
	}

	hour := parsedTime.Hour()
	minute := parsedTime.Minute()

	// Build cron expression: "minute hour * * *"
	cronSpec := fmt.Sprintf("%d %d * * *", minute, hour)

	// Add the job
	_, err = cronScheduler.AddFunc(cronSpec, func() {
		SendStandupReminder(standup.ID)
	})

	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	log.Printf("Scheduled standup '%s' (ID: %d) at %s", standup.Name, standup.ID, standup.RunAt)
	return nil
}

// SendStandupReminder sends a reminder for a specific standup
func SendStandupReminder(standupID int) {
	startTime := time.Now()
	log.Printf("⏰ [SCHEDULE TRIGGER] Standup reminder job started for ID: %d at %s", standupID, startTime.Format("2006-01-02 15:04:05"))

	// Check if we should skip today (weekends)
	if config.Config.SkipWeekends {
		today := time.Now().Weekday()
		if today == time.Saturday || today == time.Sunday {
			log.Printf("⏭️  [SKIPPED] Standup ID: %d skipped - Weekend (%s)", standupID, today.String())
			return
		}
	}

	// Get standup details
	standup, err := GetStandupByID(standupID)
	if err != nil {
		log.Printf("Error getting standup: %v", err)
		return
	}

	// Check if standup is still active
	if !standup.IsActive {
		log.Printf("Standup %d (%s) is no longer active", standupID, standup.Name)
		return
	}

	// Get eligible users (active and not on leave)
	users, err := database.GetEligibleUsersForStandup(standupID)
	if err != nil {
		log.Printf("Error getting eligible users for standup %d: %v", standupID, err)
		return
	}

	if len(users) == 0 {
		log.Printf("No eligible users for standup %d (%s)", standupID, standup.Name)
		return
	}

	// Calculate current facilitator from eligible users
	currentFacilitator, err := GetCurrentFacilitator(standupID, users)
	if err != nil {
		log.Printf("Error getting current facilitator for standup %d: %v", standupID, err)
		return
	}

	// Calculate tomorrow's facilitator
	var nextFacilitator *database.User
	if currentFacilitator != nil {
		nextFacilitator, err = GetNextFacilitator(standupID, users, currentFacilitator.ID)
		if err != nil {
			log.Printf("Warning: Could not get next facilitator for standup %d: %v", standupID, err)
		}
	}

	// Get active leaves for today
	activeLeaves, err := database.GetActiveLeavesForStandup(standupID)
	if err != nil {
		log.Printf("Warning: Could not get active leaves for standup %d: %v", standupID, err)
	}

	// Build the reminder message
	message := fmt.Sprintf("🌅 *%s*\n\n", standup.Name)

	// Add current facilitator if available
	if currentFacilitator != nil {
		message += fmt.Sprintf("👤 *Today's Facilitator:* %s\n", currentFacilitator.DisplayName)
	}

	// Add tomorrow's facilitator if available
	if nextFacilitator != nil {
		message += fmt.Sprintf("📅 *Tomorrow's Facilitator:* %s\n", nextFacilitator.DisplayName)
	}

	message += fmt.Sprintf("\n%s\n", standup.Message)

	// Add leave information if there are active leaves
	if len(activeLeaves) > 0 {
		message += fmt.Sprintf("\n🏖️ *On Leave Today:*\n")
		for _, leave := range activeLeaves {
			message += fmt.Sprintf("• %s (%s)\n", leave.User.DisplayName, leave.LeaveType)
		}
	}

	message += fmt.Sprintf("\n_Have a great day!_ ☀️")

	// Send the message via webhook
	sendTime := time.Now()
	err = integrations.SendSimpleMessage(config.Config.WebhookURL, message)
	if err != nil {
		log.Printf("❌ [SEND FAILED] Failed to send reminder for standup %d (%s): %v", standupID, standup.Name, err)
		return
	}

	// Log successful send with details
	facilitatorInfo := "none"
	if currentFacilitator != nil {
		facilitatorInfo = currentFacilitator.DisplayName
	}
	log.Printf("✅ [MESSAGE SENT] Standup: '%s' (ID: %d) | Scheduled time: %s | Sent at: %s | Facilitator: %s | Eligible users: %d | On leave: %d",
		standup.Name,
		standupID,
		standup.RunAt,
		sendTime.Format("15:04:05"),
		facilitatorInfo,
		len(users),
		len(activeLeaves),
	)

	// Update last_facilitator_id to current facilitator for next rotation
	if currentFacilitator != nil {
		err = RotateFacilitator(standupID, currentFacilitator.ID)
		if err != nil {
			log.Printf("⚠️  [WARNING] Failed to update last facilitator for standup %d: %v", standupID, err)
		} else {
			nextName := "unknown"
			if nextFacilitator != nil {
				nextName = nextFacilitator.DisplayName
			}
			log.Printf("🔄 [ROTATION] Last facilitator set to: %s | Next facilitator will be: %s", currentFacilitator.DisplayName, nextName)
		}
	}

	// Log completion time
	duration := time.Since(startTime)
	log.Printf("✨ [COMPLETED] Standup reminder job completed in %v", duration)
}

// SendManualStandupReminder manually triggers a standup reminder (for testing)
func SendManualStandupReminder(standupID int) error {
	log.Printf("🚀 [MANUAL TRIGGER] Manually triggering standup reminder for ID: %d at %s", standupID, time.Now().Format("2006-01-02 15:04:05"))
	go SendStandupReminder(standupID)
	return nil
}

// ExpireLeaves marks leaves as completed if their end date has passed
func ExpireLeaves() {
	log.Println("Running leave expiration job...")

	err := database.ExpireOldLeaves()
	if err != nil {
		log.Printf("Error expiring leaves: %v", err)
		return
	}

	log.Println("Leave expiration check completed")
}

// RefreshScheduler stops and restarts the scheduler (useful after creating/updating standups)
func RefreshScheduler() error {
	log.Println("Refreshing scheduler...")

	// Stop current scheduler
	if cronScheduler != nil {
		cronScheduler.Stop()
	}

	// Create new scheduler
	cronScheduler = cron.New()

	// Re-add leave expiration
	_, err := cronScheduler.AddFunc("0 0 * * *", ExpireLeaves)
	if err != nil {
		return fmt.Errorf("failed to schedule leave expiration: %w", err)
	}

	// Re-schedule all standups
	err = ScheduleAllStandups()
	if err != nil {
		return fmt.Errorf("failed to schedule standups: %w", err)
	}

	// Start scheduler
	cronScheduler.Start()
	log.Println("Scheduler refreshed successfully")
	return nil
}
