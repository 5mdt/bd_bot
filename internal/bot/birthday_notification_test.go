package bot

import (
	"testing"
	"time"

	"5mdt/bd_bot/internal/models"
)

func TestBirthdayNotificationUniqueness(t *testing.T) {
	bot := &Bot{} // Create a minimal bot instance for testing

	// Create a test birthday entry
	testBirthday := models.Birthday{
		Name:      "Test User",
		BirthDate: time.Now().Format("01-02"), // Today's month and day
		ChatID:    12345,
	}

	// Simulate first birthday notification
	shouldSend := bot.shouldSendBirthdayNotification(testBirthday, "BIRTHDAY_TODAY", 0)
	if !shouldSend {
		t.Error("First birthday notification should be sent")
	}

	// Update last notification time to now
	testBirthday.LastNotification = time.Now()

	// Simulate second birthday notification on the same day
	shouldSend = bot.shouldSendBirthdayNotification(testBirthday, "BIRTHDAY_TODAY", 0)
	if shouldSend {
		t.Error("Second birthday notification on the same day should not be sent")
	}

	// Simulate birthday notification on a different day
	futureTime := time.Now().AddDate(0, 0, 1)
	testBirthday.LastNotification = futureTime
	shouldSend = bot.shouldSendBirthdayNotification(testBirthday, "BIRTHDAY_TODAY", 0)
	if !shouldSend {
		t.Error("Birthday notification should be sent on a different day")
	}
}
