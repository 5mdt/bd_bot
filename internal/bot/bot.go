package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"5mdt/bd_bot/internal/models"
	"5mdt/bd_bot/internal/storage"
)

type Bot struct {
	api                 *tgbotapi.BotAPI
	status              string
	username            string
	firstName           string
	startTime           time.Time
	notificationsSent   int64
	notificationStartHour int // Start hour for notifications (0-23, UTC)
	notificationEndHour   int // End hour for notifications (0-23, UTC)
	mu                  sync.RWMutex
	ctx                 context.Context
	cancel              context.CancelFunc
}

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
	notificationStartHour := 10 // Default: 10 AM UTC
	notificationEndHour := 22   // Default: 10 PM UTC

	if startHourStr := os.Getenv("NOTIFICATION_START_HOUR"); startHourStr != "" {
		if hour, err := strconv.Atoi(startHourStr); err == nil && hour >= 0 && hour <= 23 {
			notificationStartHour = hour
		} else {
			log.Printf("Invalid NOTIFICATION_START_HOUR: %s, using default: %d", startHourStr, notificationStartHour)
		}
	}

	if endHourStr := os.Getenv("NOTIFICATION_END_HOUR"); endHourStr != "" {
		if hour, err := strconv.Atoi(endHourStr); err == nil && hour >= 0 && hour <= 23 {
			notificationEndHour = hour
		} else {
			log.Printf("Invalid NOTIFICATION_END_HOUR: %s, using default: %d", endHourStr, notificationEndHour)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	bot := &Bot{
		api:                 api,
		status:              "starting",
		username:            me.UserName,
		firstName:           me.FirstName,
		startTime:           time.Now(),
		notificationStartHour: notificationStartHour,
		notificationEndHour:   notificationEndHour,
		ctx:                 ctx,
		cancel:              cancel,
	}

	log.Printf("Bot initialized with notification hours: %02d:00 - %02d:00 UTC", notificationStartHour, notificationEndHour)
	return bot, nil
}

func (b *Bot) Start() {
	go b.run()
}

func (b *Bot) Stop() {
	b.cancel()
	b.setStatus("stopped")
}

func (b *Bot) GetStatus() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status
}

func (b *Bot) GetUsername() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.username
}

func (b *Bot) GetFirstName() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.firstName
}

func (b *Bot) GetUptime() time.Duration {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return time.Since(b.startTime)
}

func (b *Bot) GetNotificationsSent() int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.notificationsSent
}

func (b *Bot) GetNotificationHours() (int, int) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.notificationStartHour, b.notificationEndHour
}

func (b *Bot) setStatus(status string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status = status
	log.Printf("Bot status: %s", status)
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
	log.Printf("Received message from %s (Chat ID: %d): %s", message.From.UserName, message.Chat.ID, message.Text)

	// Handle commands
	if message.IsCommand() {
		b.handleCommand(message)
		return
	}

	// Default response for non-commands
	msg := tgbotapi.NewMessage(message.Chat.ID, "Hello! Send /help to see available commands.")
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Failed to send message: %v", err)
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
			log.Printf("Failed to send message: %v", err)
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
		log.Printf("Failed to send welcome message: %v", err)
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
		log.Printf("Failed to send help message: %v", err)
	}
}

func (b *Bot) handleUpdateBirthDateCommand(message *tgbotapi.Message, args string) {
	if args == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Please provide a birth date. Example: /update_birth_date 1999-12-31")
		if _, err := b.api.Send(msg); err != nil {
			log.Printf("Failed to send message: %v", err)
		}
		return
	}

	// Handle MM-DD format by converting to 0000-MM-DD (year unknown)
	mmddRegex := regexp.MustCompile(`^\d{2}-\d{2}$`)
	if mmddRegex.MatchString(args) {
		// Validate the MM-DD date
		_, err := time.Parse("01-02", args)
		if err != nil {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Invalid date. Please use a valid MM-DD format (e.g., 12-31)")
			if _, err := b.api.Send(msg); err != nil {
				log.Printf("Failed to send message: %v", err)
			}
			return
		}
		// Convert MM-DD to 0000-MM-DD format
		args = "0000-" + args
	}

	// Validate date format (YYYY-MM-DD)
	dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !dateRegex.MatchString(args) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Invalid date format. Please use YYYY-MM-DD format (e.g., 1999-12-31)")
		if _, err := b.api.Send(msg); err != nil {
			log.Printf("Failed to send message: %v", err)
		}
		return
	}

	// Parse and validate the date
	_, err := time.Parse("2006-01-02", args)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Invalid date. Please use a valid date in YYYY-MM-DD format.")
		if _, err := b.api.Send(msg); err != nil {
			log.Printf("Failed to send message: %v", err)
		}
		return
	}

	// Get chat name - prioritize chat title for groups, fall back to user info for private chats
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

	// Load existing birthdays
	birthdays, err := storage.LoadBirthdays()
	if err != nil {
		log.Printf("Failed to load birthdays: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, there was an error accessing the database.")
		if _, err := b.api.Send(msg); err != nil {
			log.Printf("Failed to send error message: %v", err)
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

			log.Printf("Updated birthday for %s (Chat ID: %d): %s -> %s", chatName, message.Chat.ID, oldDate, args)
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
		log.Printf("Added new birthday for %s (Chat ID: %d): %s", chatName, message.Chat.ID, args)
	}

	// Save updated birthdays
	if err := storage.SaveBirthdays(birthdays); err != nil {
		log.Printf("Failed to save birthdays: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, there was an error saving your information.")
		if _, err := b.api.Send(msg); err != nil {
			log.Printf("Failed to send error message: %v", err)
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
		log.Printf("Failed to send confirmation message: %v", err)
	}
}

func (b *Bot) handleMyInfoCommand(message *tgbotapi.Message) {
	// Load birthdays to find user's info
	birthdays, err := storage.LoadBirthdays()
	if err != nil {
		log.Printf("Failed to load birthdays: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, there was an error accessing the database.")
		if _, err := b.api.Send(msg); err != nil {
			log.Printf("Failed to send error message: %v", err)
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
				log.Printf("Failed to send info message: %v", err)
			}
			return
		}
	}

	// User not found
	msg := tgbotapi.NewMessage(message.Chat.ID, "You don't have any information stored yet. Use /update_birth_date to set your birth date.")
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

func (b *Bot) handleChatTitleChange(message *tgbotapi.Message) {
	newTitle := message.NewChatTitle
	chatID := message.Chat.ID

	log.Printf("Chat title changed to '%s' for chat ID: %d", newTitle, chatID)

	// Load existing birthdays
	birthdays, err := storage.LoadBirthdays()
	if err != nil {
		log.Printf("Failed to load birthdays during title change: %v", err)
		return
	}

	// Find and update the birthday entry for this chat
	updated := false
	for i := range birthdays {
		if birthdays[i].ChatID == chatID {
			oldName := birthdays[i].Name
			birthdays[i].Name = newTitle
			updated = true
			log.Printf("Updated chat name from '%s' to '%s' for chat ID: %d", oldName, newTitle, chatID)
			break
		}
	}

	if updated {
		// Save the updated birthdays
		if err := storage.SaveBirthdays(birthdays); err != nil {
			log.Printf("Failed to save birthdays after title change: %v", err)
		}
	} else {
		log.Printf("No existing birthday entry found for chat ID: %d", chatID)
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
	ticker := time.NewTicker(1 * time.Hour) // Check hourly for notification window
	defer ticker.Stop()

	// Check immediately on start
	b.processBirthdays()

	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			b.processBirthdays()
		}
	}
}

func (b *Bot) processBirthdays() {
	now := time.Now().UTC()
	currentHour := now.Hour()

	// Check if current time is within notification hours
	if !b.isWithinNotificationHours(currentHour) {
		log.Printf("Current hour %02d:00 UTC is outside notification window (%02d:00 - %02d:00 UTC), skipping birthday notifications",
			currentHour, b.notificationStartHour, b.notificationEndHour)
		return
	}

	birthdays, err := storage.LoadBirthdays()
	if err != nil {
		log.Printf("Failed to load birthdays: %v", err)
		return
	}

	today := now.Format("01-02") // MM-DD format
	log.Printf("Processing birthdays at %02d:00 UTC for date %s", currentHour, today)

	for i, birthday := range birthdays {
		// Skip if no chat ID configured
		if birthday.ChatID == 0 {
			continue
		}

		// Extract MM-DD from birth date
		var birthdayMMDD string
		if len(birthday.BirthDate) >= 7 { // At least "0000-MM" or "YYYY-MM"
			parts := birthday.BirthDate[5:] // Skip "0000-" or "YYYY-"
			if len(parts) >= 5 && parts[2:3] == "-" { // MM-DD
				birthdayMMDD = parts
			}
		}

		if birthdayMMDD == today {
			// Check if we already sent notification today
			if !birthday.LastNotification.IsZero() &&
			   birthday.LastNotification.Format("2006-01-02") == time.Now().Format("2006-01-02") {
				continue
			}

			// Send birthday notification
			message := fmt.Sprintf("🎉 Happy Birthday, %s! 🎂", birthday.Name)
			msg := tgbotapi.NewMessage(birthday.ChatID, message)

			if _, err := b.api.Send(msg); err != nil {
				log.Printf("Failed to send birthday notification for %s: %v", birthday.Name, err)
				continue
			}

			// Increment notification counter
			b.mu.Lock()
			b.notificationsSent++
			b.mu.Unlock()

			// Update last notification time
			birthdays[i].LastNotification = time.Now().UTC()
			log.Printf("Sent birthday notification for %s", birthday.Name)
		}
	}

	// Save updated birthdays with notification timestamps
	if err := storage.SaveBirthdays(birthdays); err != nil {
		log.Printf("Failed to save birthdays after notifications: %v", err)
	}
}
