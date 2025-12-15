// Package bot provides the Telegram bot implementation for birthday notifications.
// It handles incoming messages, commands, birthday tracking, and scheduled notifications.
package bot

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"5mdt/bd_bot/internal/logger"
	"5mdt/bd_bot/internal/models"
	"5mdt/bd_bot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Precompiled regex patterns for date validation
var (
	mmddRegex = regexp.MustCompile(`^\d{2}-\d{2}$`)
	dateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

// Bot represents a Telegram bot instance that manages birthday notifications.
type Bot struct {
	// api is the Telegram Bot API client.
	api *tgbotapi.BotAPI
	// status is the current bot status (e.g., "connecting", "running", "stopped").
	status string
	// username is the bot's Telegram username.
	username string
	// firstName is the bot's display name.
	firstName string
	// startTime is the bot startup timestamp.
	startTime time.Time
	// notificationsSent is the counter of birthday notifications sent.
	notificationsSent int64
	// notificationStartHour is the start hour for notifications (0-23, UTC).
	notificationStartHour int
	// notificationEndHour is the end hour for notifications (0-23, UTC).
	notificationEndHour int
	// running indicates whether the bot's run loop is active.
	running bool
	// mu is the mutex for thread-safe access to bot state.
	mu sync.RWMutex
	// ctx is the context for cancellation.
	ctx context.Context
	// cancel is the function to cancel the bot's context.
	cancel context.CancelFunc
}

// New creates and initializes a new Telegram bot instance with the given token.
// It fetches bot information from Telegram and parses notification hours from environment variables.
// Returns an error if the token is invalid or Telegram API communication fails.
func New(token string) (*Bot, error) {
	if token == "" {
		return nil, fmt.Errorf("telegram bot token is required")
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	// Get bot info from Telegram
	me, err := api.GetMe()
	if err != nil {
		return nil, fmt.Errorf("failed to get bot info: %w", err)
	}

	// Parse notification hours from environment variables
	notificationStartHour := 8 // Default: 8 AM UTC
	notificationEndHour := 20  // Default: 8 PM UTC

	if startHourStr := os.Getenv("NOTIFICATION_START_HOUR"); startHourStr != "" {
		if hour, err := strconv.Atoi(startHourStr); err == nil && hour >= 0 && hour <= 23 {
			notificationStartHour = hour
		} else {
			logger.Warn("BOT", "Invalid NOTIFICATION_START_HOUR: %s, using default: %d", startHourStr, notificationStartHour)
		}
	}

	if endHourStr := os.Getenv("NOTIFICATION_END_HOUR"); endHourStr != "" {
		if hour, err := strconv.Atoi(endHourStr); err == nil && hour >= 0 && hour <= 23 {
			notificationEndHour = hour
		} else {
			logger.Warn("BOT", "Invalid NOTIFICATION_END_HOUR: %s, using default: %d", endHourStr, notificationEndHour)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	bot := &Bot{
		api:                   api,
		status:                "starting",
		username:              me.UserName,
		firstName:             me.FirstName,
		startTime:             time.Now(),
		notificationStartHour: notificationStartHour,
		notificationEndHour:   notificationEndHour,
		ctx:                   ctx,
		cancel:                cancel,
	}

	logger.Info("BOT", "Bot initialized successfully")
	logger.Info("BOT", "Username: @%s", me.UserName)
	logger.Info("BOT", "Display Name: %s", me.FirstName)
	logger.Info("BOT", "Notification hours: %02d:00 - %02d:00 UTC", notificationStartHour, notificationEndHour)
	return bot, nil
}

// Start begins the bot's message polling and birthday checking goroutines.
// This is non-blocking; the bot runs in the background.
// If the bot is already running, this method does nothing to prevent duplicate goroutines.
func (b *Bot) Start() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.running {
		logger.Warn("BOT", "Start() called but bot is already running, ignoring")
		return
	}

	b.running = true
	go b.run()
}

// Stop gracefully shuts down the bot by canceling its context and updating its status.
func (b *Bot) Stop() {
	b.cancel()
	b.setStatus("stopped")
}

// GetStatus returns the current bot status (e.g., "running", "stopped", "not configured").
// Returns "not configured" if the bot is nil.
func (b *Bot) GetStatus() string {
	if b == nil {
		return "not configured"
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status
}

// GetUsername returns the bot's Telegram username without the @ prefix.
// Returns an empty string if the bot is nil.
func (b *Bot) GetUsername() string {
	if b == nil {
		return ""
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.username
}

// GetFirstName returns the bot's display name as configured in Telegram.
// Returns an empty string if the bot is nil.
func (b *Bot) GetFirstName() string {
	if b == nil {
		return ""
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.firstName
}

// GetUptime returns the duration since the bot started.
// Returns 0 if the bot is nil.
func (b *Bot) GetUptime() time.Duration {
	if b == nil {
		return 0
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	return time.Since(b.startTime)
}

// GetNotificationsSent returns the total number of birthday notifications sent by the bot.
// Returns 0 if the bot is nil.
func (b *Bot) GetNotificationsSent() int64 {
	if b == nil {
		return 0
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.notificationsSent
}

// GetNotificationHours returns the configured start and end hours (UTC) for sending notifications.
// Returns (0, 0) if the bot is nil.
func (b *Bot) GetNotificationHours() (int, int) {
	if b == nil {
		return 0, 0
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.notificationStartHour, b.notificationEndHour
}

func (b *Bot) setStatus(status string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status = status
	logger.Info("BOT", "Status changed: %s", status)
}

func (b *Bot) run() {
	b.setStatus("connecting")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)
	b.setStatus("running")

	// Start birthday checker
	go b.checkBirthdays()

	for {
		select {
		case <-b.ctx.Done():
			b.api.StopReceivingUpdates()
			b.mu.Lock()
			b.running = false
			b.mu.Unlock()
			b.setStatus("stopped")
			return
		case update := <-updates:
			if update.Message != nil {
				// Check if bot was added to a group
				if update.Message.NewChatMembers != nil {
					for _, member := range update.Message.NewChatMembers {
						if member.ID == b.api.Self.ID {
							// Bot was added to this chat, send welcome message
							b.handleStartCommand(update.Message)
							break
						}
					}
				} else if update.Message.LeftChatMember != nil {
					// Someone left the chat - no action needed
				} else if update.Message.NewChatTitle != "" {
					// Chat title was changed, update existing birthday entry if it exists
					b.handleChatTitleChange(update.Message)
				} else {
					// Regular message handling
					b.handleMessage(update.Message)
				}
			}
		}
	}
}

func (b *Bot) handleMessage(message *tgbotapi.Message) {
	logger.LogBotMessage(message.Chat.ID, message.From.UserName, message.Text)

	// Handle commands
	if message.IsCommand() {
		b.handleCommand(message)
		return
	}

	// For group chats, ignore non-command messages
	if message.Chat.Type == "group" || message.Chat.Type == "supergroup" {
		return
	}

	// For private chats, send help prompt
	msg := tgbotapi.NewMessage(message.Chat.ID, "Hello! Send /help to see available commands.")
	if _, err := b.api.Send(msg); err != nil {
		logger.Error("BOT", "Failed to send message: %v", err)
	}
}

func (b *Bot) handleCommand(message *tgbotapi.Message) {
	command := message.Command()
	args := message.CommandArguments()

	switch command {
	case "start":
		b.handleStartCommand(message)
	case "help":
		b.handleHelpCommand(message)
	case "update_birth_date":
		b.handleUpdateBirthDateCommand(message, args)
	case "my_info":
		b.handleMyInfoCommand(message)
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown command. Send /help for available commands.")
		if _, err := b.api.Send(msg); err != nil {
			logger.Error("BOT", "Failed to send message: %v", err)
		}
	}
}

func (b *Bot) handleStartCommand(message *tgbotapi.Message) {
	welcomeText := `Hi, I am Jeeves bot. I can send you notifications about birthdays. Send me a message like:

/update_birth_date 1999-12-31

to configure your birthdate.

Note: Only one birth date can be configured per chat.

Use /help to see all available commands.`

	msg := tgbotapi.NewMessage(message.Chat.ID, welcomeText)
	if _, err := b.api.Send(msg); err != nil {
		logger.Error("BOT", "Failed to send welcome message: %v", err)
	}
}

func (b *Bot) handleHelpCommand(message *tgbotapi.Message) {
	helpText := `Available commands:

/start - Welcome message and getting started
/help - Show this help message
/update_birth_date - Set your birth date
  • YYYY-MM-DD format (e.g., /update_birth_date 1999-12-31)
  • MM-DD format (e.g., /update_birth_date 12-31) - year unknown
/my_info - Show your current information

Note: Commands work with or without the bot username (e.g., both /help and /help@bot_name work)

The bot will send you birthday greetings on your special day! 🎉`

	msg := tgbotapi.NewMessage(message.Chat.ID, helpText)
	if _, err := b.api.Send(msg); err != nil {
		logger.Error("BOT", "Failed to send help message: %v", err)
	}
}

// resolveChatName determines the appropriate display name for a chat.
// For group chats, it returns the group title. For private chats, it returns
// the user's full name (first + last) or username. Falls back to "Unknown" if
// no name information is available.
func resolveChatName(message *tgbotapi.Message) string {
	chatName := "Unknown"
	if message.Chat.Type == "group" || message.Chat.Type == "supergroup" {
		// For group chats, use the group name
		if message.Chat.Title != "" {
			chatName = message.Chat.Title
		}
	} else {
		// For private chats, use user's name
		if message.From.FirstName != "" {
			chatName = message.From.FirstName
			if message.From.LastName != "" {
				chatName += " " + message.From.LastName
			}
		} else if message.From.UserName != "" {
			chatName = message.From.UserName
		}
	}

	// Final fallback to chat title if still unknown
	if chatName == "Unknown" && message.Chat.Title != "" {
		chatName = message.Chat.Title
	}

	return chatName
}

func (b *Bot) handleUpdateBirthDateCommand(message *tgbotapi.Message, args string) {
	if args == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Please provide a birth date. Example: /update_birth_date 1999-12-31")
		if _, err := b.api.Send(msg); err != nil {
			logger.Error("BOT", "Failed to send message: %v", err)
		}
		return
	}

	// Handle MM-DD format by converting to 0000-MM-DD (year unknown)
	if mmddRegex.MatchString(args) {
		// Validate the MM-DD date
		_, err := time.Parse("01-02", args)
		if err != nil {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Invalid date. Please use a valid MM-DD format (e.g., 12-31)")
			if _, err := b.api.Send(msg); err != nil {
				logger.Error("BOT", "Failed to send message: %v", err)
			}
			return
		}
		// Convert MM-DD to 0000-MM-DD format
		args = "0000-" + args
	}

	// Validate date format (YYYY-MM-DD)
	if !dateRegex.MatchString(args) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Invalid date format. Please use YYYY-MM-DD format (e.g., 1999-12-31)")
		if _, err := b.api.Send(msg); err != nil {
			logger.Error("BOT", "Failed to send message: %v", err)
		}
		return
	}

	// Parse and validate the date
	_, err := time.Parse("2006-01-02", args)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Invalid date. Please use a valid date in YYYY-MM-DD format.")
		if _, err := b.api.Send(msg); err != nil {
			logger.Error("BOT", "Failed to send message: %v", err)
		}
		return
	}

	// Get chat name using helper function
	chatName := resolveChatName(message)

	// Load existing birthdays
	birthdays, err := storage.LoadBirthdays()
	if err != nil {
		logger.Error("STORAGE", "Failed to load birthdays: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, there was an error accessing the database.")
		if _, err := b.api.Send(msg); err != nil {
			logger.Error("BOT", "Failed to send error message: %v", err)
		}
		return
	}

	// Find existing birthday entry by chat ID
	found := false
	for i := range birthdays {
		if birthdays[i].ChatID == message.Chat.ID {
			// Update existing entry
			oldDate := birthdays[i].BirthDate
			birthdays[i].BirthDate = args
			birthdays[i].Name = chatName
			birthdays[i].LastNotification = time.Time{} // Reset notification
			found = true

			logger.Info("BOT", "Updated birthday for %s (Chat ID: %d): %s -> %s", chatName, message.Chat.ID, oldDate, args)
			break
		}
	}

	if !found {
		// Add new birthday entry
		newBirthday := models.Birthday{
			Name:             chatName,
			BirthDate:        args,
			LastNotification: time.Time{}, // Zero value (null)
			ChatID:           message.Chat.ID,
		}
		birthdays = append(birthdays, newBirthday)
		logger.Info("BOT", "Added new birthday for %s (Chat ID: %d): %s", chatName, message.Chat.ID, args)
	}

	// Save updated birthdays
	if err := storage.SaveBirthdays(birthdays); err != nil {
		logger.Error("STORAGE", "Failed to save birthdays: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, there was an error saving your information.")
		if _, err := b.api.Send(msg); err != nil {
			logger.Error("BOT", "Failed to send error message: %v", err)
		}
		return
	}

	// Send confirmation
	var responseText string
	if strings.HasPrefix(args, "0000-") {
		// MM-DD format was converted to 0000-MM-DD
		mmdd := strings.TrimPrefix(args, "0000-")
		responseText = fmt.Sprintf("✅ Your birth date has been set to %s (year unknown)!\n\nI'll send you birthday greetings every %s! 🎉", mmdd, mmdd)
	} else {
		responseText = fmt.Sprintf("✅ Your birth date has been set to %s!\n\nI'll send you birthday greetings on your special day! 🎉", args)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, responseText)
	if _, err := b.api.Send(msg); err != nil {
		logger.Error("BOT", "Failed to send confirmation message: %v", err)
	}
}

func (b *Bot) handleMyInfoCommand(message *tgbotapi.Message) {
	// Load birthdays to find user's info
	birthdays, err := storage.LoadBirthdays()
	if err != nil {
		logger.Error("STORAGE", "Failed to load birthdays: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, there was an error accessing the database.")
		if _, err := b.api.Send(msg); err != nil {
			logger.Error("BOT", "Failed to send error message: %v", err)
		}
		return
	}

	// Find user's birthday entry
	for _, birthday := range birthdays {
		if birthday.ChatID == message.Chat.ID {
			responseText := fmt.Sprintf("📋 Your Information:\n\nName: %s\nBirth Date: %s\nChat ID: %d",
				birthday.Name, birthday.BirthDate, birthday.ChatID)

			if !birthday.LastNotification.IsZero() {
				responseText += fmt.Sprintf("\nLast Notification: %s", birthday.LastNotification.Format("2006-01-02 15:04:05"))
			}

			msg := tgbotapi.NewMessage(message.Chat.ID, responseText)
			if _, err := b.api.Send(msg); err != nil {
				logger.Error("BOT", "Failed to send info message: %v", err)
			}
			return
		}
	}

	// User not found
	msg := tgbotapi.NewMessage(message.Chat.ID, "You don't have any information stored yet. Use /update_birth_date to set your birth date.")
	if _, err := b.api.Send(msg); err != nil {
		logger.Error("BOT", "Failed to send message: %v", err)
	}
}

func (b *Bot) handleChatTitleChange(message *tgbotapi.Message) {
	newTitle := message.NewChatTitle
	chatID := message.Chat.ID

	logger.Info("BOT", "Chat title changed to '%s' for chat ID: %d", newTitle, chatID)

	// Load existing birthdays
	birthdays, err := storage.LoadBirthdays()
	if err != nil {
		logger.Error("STORAGE", "Failed to load birthdays during title change: %v", err)
		return
	}

	// Find and update the birthday entry for this chat
	updated := false
	for i := range birthdays {
		if birthdays[i].ChatID == chatID {
			oldName := birthdays[i].Name
			birthdays[i].Name = newTitle
			updated = true
			logger.Info("BOT", "Updated chat name from '%s' to '%s' for chat ID: %d", oldName, newTitle, chatID)
			break
		}
	}

	if updated {
		// Save the updated birthdays
		if err := storage.SaveBirthdays(birthdays); err != nil {
			logger.Error("STORAGE", "Failed to save birthdays after title change: %v", err)
		}
	} else {
		logger.Debug("BOT", "No existing birthday entry found for chat ID: %d", chatID)
	}

	// Don't send any message to the chat for title changes
}

func (b *Bot) isWithinNotificationHours(currentHour int) bool {
	b.mu.RLock()
	startHour := b.notificationStartHour
	endHour := b.notificationEndHour
	b.mu.RUnlock()

	// Handle cases where the time window crosses midnight
	if startHour <= endHour {
		// Normal case: 10:00 - 22:00
		return currentHour >= startHour && currentHour <= endHour
	} else {
		// Crosses midnight: 22:00 - 06:00
		return currentHour >= startHour || currentHour <= endHour
	}
}

func (b *Bot) checkBirthdays() {
	// Start with minute-level checking
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Check immediately on start
	b.processBirthdays()

	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			now := time.Now().UTC()
			currentHour := now.Hour()

			// Check if we're in notification window
			if b.isWithinNotificationHours(currentHour) {
				// During notification hours: check every minute
				b.processBirthdays()
			} else {
				// Outside notification hours: only check once per hour
				if now.Minute() == 0 {
					b.processBirthdays()
				}
			}
		}
	}
}

func (b *Bot) shouldSendBirthdayNotification(birthday models.Birthday, notificationType string) bool {
	// Always send birthday today notification
	if notificationType == "BIRTHDAY_TODAY" {
		// Check if last notification was today
		now := time.Now().UTC()
		lastNotificationDate := ""
		if !birthday.LastNotification.IsZero() {
			lastNotificationDate = birthday.LastNotification.Format("2006-01-02")
		}

		todayDate := now.Format("2006-01-02")
		return lastNotificationDate != todayDate
	}

	// For 2 and 4 weeks reminders, previous checks in the function will handle skipping
	return true
}

func (b *Bot) processBirthdays() {
	now := time.Now().UTC()
	currentHour := now.Hour()

	// Check if current time is within notification hours
	if !b.isWithinNotificationHours(currentHour) {
		return // Skip logging during frequent checks
	}

	logger.LogNotification("INFO", "Starting birthday check at %s UTC (hour: %02d)", now.Format("2006-01-02 15:04:05"), currentHour)

	birthdays, err := storage.LoadBirthdays()
	if err != nil {
		logger.LogNotification("ERROR", "Failed to load birthdays: %v", err)
		return
	}

	logger.LogNotification("INFO", "Loaded %d birthday entries from storage", len(birthdays))

	today := now.Format("2006-01-02")

	logger.LogNotification("INFO", "Checking for birthdays today, in 2 weeks (+14 days), and in 4 weeks (+28 days)")

	notificationsSent := false
	entriesProcessed := 0
	entriesSkipped := 0

	for i, birthday := range birthdays {
		entriesProcessed++

		logger.LogNotification("DEBUG", "Processing entry %d: Name='%s', BirthDate='%s', ChatID=%d",
			i+1, birthday.Name, birthday.BirthDate, birthday.ChatID)

		// Skip if no chat ID configured
		if birthday.ChatID == 0 {
			logger.LogNotification("WARN", "SKIP: No chat ID configured for '%s'", birthday.Name)
			entriesSkipped++
			continue
		}

		// Extract MM-DD from birth date
		var birthdayMMDD string
		if len(birthday.BirthDate) >= 7 { // At least "0000-MM" or "YYYY-MM"
			parts := birthday.BirthDate[5:]           // Skip "0000-" or "YYYY-"
			if len(parts) >= 5 && parts[2:3] == "-" { // MM-DD
				birthdayMMDD = parts
			}
		}

		if birthdayMMDD == "" {
			logger.LogNotification("WARN", "SKIP: Invalid birth date format for '%s': '%s'", birthday.Name, birthday.BirthDate)
			entriesSkipped++
			continue
		}

		logger.LogNotification("DEBUG", "Extracted birthday MM-DD: %s for '%s'", birthdayMMDD, birthday.Name)

		// Check if we already sent notification today
		lastNotificationDate := ""
		if !birthday.LastNotification.IsZero() {
			lastNotificationDate = birthday.LastNotification.Format("2006-01-02")
		}

		logger.LogNotification("DEBUG", "Last notification for '%s': %s (today: %s)",
			birthday.Name, lastNotificationDate, today)

		if lastNotificationDate == today {
			logger.LogNotification("DEBUG", "SKIP: Already sent notification today for '%s'", birthday.Name)
			entriesSkipped++
			continue // Already sent notification today
		}

		var message string
		var notificationType string

		// Parse the birthday MM-DD to determine this year's birthday date
		thisYearBirthday, err := time.Parse("2006-01-02", fmt.Sprintf("%d-%s", now.Year(), birthdayMMDD))
		if err != nil {
			logger.LogNotification("ERROR", "SKIP: Failed to parse birthday date for '%s': %v", birthday.Name, err)
			entriesSkipped++
			continue
		}

		// Normalize current time to start of day (midnight) for accurate date comparison
		nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

		// Calculate days difference using date-only comparison
		daysDiff := int(thisYearBirthday.Sub(nowDate).Hours() / 24)

		logger.LogNotification("DEBUG", "Birthday analysis for '%s': ThisYear=%s, DaysDiff=%d",
			birthday.Name, thisYearBirthday.Format("2006-01-02"), daysDiff)

		// Check for different notification scenarios
		if daysDiff == 0 {
			// Birthday is today
			message = fmt.Sprintf("🎉 Happy Birthday, %s! 🎂", birthday.Name)
			notificationType = "BIRTHDAY_TODAY"
		} else if daysDiff == 14 {
			// Birthday is in exactly 2 weeks
			message = fmt.Sprintf("📅 Reminder: %s's birthday is in 2 weeks (%s)! 🎈", birthday.Name, birthdayMMDD)
			notificationType = "REMINDER_2_WEEKS"
		} else if daysDiff == 28 {
			// Birthday is in exactly 4 weeks
			message = fmt.Sprintf("📅 Early reminder: %s's birthday is in 4 weeks (%s)! 🗓️", birthday.Name, birthdayMMDD)
			notificationType = "REMINDER_4_WEEKS"
		} else if daysDiff < 0 {
			// Birthday has passed this year - check next year
			nextYearBirthday := thisYearBirthday.AddDate(1, 0, 0)
			nextYearDaysDiff := int(nextYearBirthday.Sub(nowDate).Hours() / 24)

			logger.LogNotification("DEBUG", "Birthday passed this year for '%s': NextYear=%s, NextYearDaysDiff=%d",
				birthday.Name, nextYearBirthday.Format("2006-01-02"), nextYearDaysDiff)

			if nextYearDaysDiff == 14 {
				// Birthday is in 2 weeks next year
				message = fmt.Sprintf("📅 Reminder: %s's birthday is in 2 weeks (%s)! 🎈", birthday.Name, birthdayMMDD)
				notificationType = "REMINDER_2_WEEKS_NEXT_YEAR"
			} else if nextYearDaysDiff == 28 {
				// Birthday is in 4 weeks next year
				message = fmt.Sprintf("📅 Early reminder: %s's birthday is in 4 weeks (%s)! 🗓️", birthday.Name, birthdayMMDD)
				notificationType = "REMINDER_4_WEEKS_NEXT_YEAR"
			} else {
				continue // No notification matches
			}
		} else {
			continue // No notification matches
		}

		// Check if this notification should be sent
		if b.shouldSendBirthdayNotification(birthday, notificationType) {
			logger.LogNotification("INFO", "SENDING: Type=%s, Name='%s', ChatID=%d, Message='%s'",
				notificationType, birthday.Name, birthday.ChatID, message)

			msg := tgbotapi.NewMessage(birthday.ChatID, message)

			if _, err := b.api.Send(msg); err != nil {
				logger.LogNotification("ERROR", "Failed to send %s notification for '%s' to ChatID %d: %v",
					notificationType, birthday.Name, birthday.ChatID, err)
				entriesSkipped++
				continue
			}

			// Increment notification counter
			b.mu.Lock()
			b.notificationsSent++
			totalSent := b.notificationsSent
			b.mu.Unlock()

			// Update last notification time
			birthdays[i].LastNotification = now
			notificationsSent = true

			logger.LogNotification("INFO", "SUCCESS: %s notification sent for '%s' (ChatID: %d, Total sent: %d)",
				notificationType, birthday.Name, birthday.ChatID, totalSent)
		} else {
			if daysDiff < 0 {
				// Birthday has passed, show next year info
				nextYearBirthday := thisYearBirthday.AddDate(1, 0, 0)
				nextYearDaysDiff := int(nextYearBirthday.Sub(nowDate).Hours() / 24)
				logger.LogNotification("DEBUG", "NO_MATCH: Birthday '%s' (%s) passed this year (%d days ago), next occurrence in %d days",
					birthday.Name, birthdayMMDD, -daysDiff, nextYearDaysDiff)
			} else {
				logger.LogNotification("DEBUG", "NO_MATCH: Birthday '%s' (%s) is in %d days (not 0, 14, or 28)",
					birthday.Name, birthdayMMDD, daysDiff)
			}
			entriesSkipped++
		}
	}

	// Save updated birthdays if any notifications were sent
	if notificationsSent {
		logger.LogNotification("INFO", "SAVING: Updating YAML file with new last_notification timestamps")
		if err := storage.SaveBirthdays(birthdays); err != nil {
			logger.LogNotification("ERROR", "Failed to save birthdays after notifications: %v", err)
		} else {
			logger.LogNotification("INFO", "SAVED: Successfully updated YAML file")
		}
	} else {
		logger.LogNotification("DEBUG", "NO_SAVE: No notifications sent, YAML file unchanged")
	}

	notificationsSentCount := entriesProcessed - entriesSkipped
	logger.LogNotification("INFO", "SUMMARY: Processed=%d, Sent=%d, Skipped=%d, Duration=%v",
		entriesProcessed, notificationsSentCount, entriesSkipped, time.Since(now).Truncate(time.Millisecond))
}
