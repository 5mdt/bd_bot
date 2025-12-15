// Package handlers provides HTTP request handlers for the birthday management web interface.
// It manages rendering the birthday list, handling form submissions for add/edit/delete operations,
// and displaying bot status information.
package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"5mdt/bd_bot/internal/logger"
	"5mdt/bd_bot/internal/models"
	"5mdt/bd_bot/internal/storage"
)

// PageData contains the data passed to page templates.
type PageData struct {
	// Birthdays is the list of birthday records to display.
	Birthdays []models.Birthday
	// BotInfo contains Telegram bot status and statistics.
	BotInfo BotInfo
}

// BotInfo represents the Telegram bot's current status and configuration.
type BotInfo struct {
	// Status is the current bot status (e.g., "running", "stopped", "not configured").
	Status string
	// Username is the Telegram bot's username (without @).
	Username string
	// FirstName is the Telegram bot's display name.
	FirstName string
	// Uptime is the human-readable uptime duration.
	Uptime string
	// NotificationsSent is the total number of birthday notifications sent.
	NotificationsSent int64
	// NotificationHours is the configured notification time window (e.g., "08:00 - 20:00 UTC").
	NotificationHours string
	// NextCheckTime is the next scheduled birthday check time.
	NextCheckTime string
	// CurrentHourInWindow indicates whether the current hour falls within the notification window.
	CurrentHourInWindow bool
	// Configured indicates whether the bot is properly configured with a valid token.
	Configured bool
}

// BotStatusProvider defines the interface for querying bot status and metrics.
type BotStatusProvider interface {
	// GetStatus returns the current bot status.
	GetStatus() string
	// GetUsername returns the bot's username.
	GetUsername() string
	// GetFirstName returns the bot's display name.
	GetFirstName() string
	// GetUptime returns the duration since bot startup.
	GetUptime() time.Duration
	// GetNotificationsSent returns the total notifications sent.
	GetNotificationsSent() int64
	// GetNotificationHours returns start and end hours for notifications.
	GetNotificationHours() (int, int)
}

func formatUptime(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}

func formatNotificationHours(startHour, endHour int) string {
	if startHour <= endHour {
		return fmt.Sprintf("%02d:00 - %02d:00 UTC", startHour, endHour)
	} else {
		// Crosses midnight
		return fmt.Sprintf("%02d:00 - %02d:00 UTC (next day)", startHour, endHour)
	}
}

func isCurrentlyInNotificationWindow(startHour, endHour int) bool {
	currentHour := time.Now().UTC().Hour()
	if startHour <= endHour {
		return currentHour >= startHour && currentHour <= endHour
	} else {
		// Crosses midnight
		return currentHour >= startHour || currentHour <= endHour
	}
}

func calculateNextCheckTime() string {
	now := time.Now().UTC()
	nextMinute := now.Truncate(time.Minute).Add(time.Minute)
	return nextMinute.Format("15:04:05 UTC")
}

func parseIdx(r *http.Request) (int, error) {
	if err := r.ParseForm(); err != nil {
		return 0, err
	}
	return strconv.Atoi(r.FormValue("idx"))
}

func updateBirthdayFromForm(b *models.Birthday, r *http.Request) {
	originalBirthDate := b.BirthDate
	b.Name = r.FormValue("name")
	b.BirthDate = normalizeDateWithOriginal(r.FormValue("birth_date"), originalBirthDate)

	// Parse timestamp from form
	if timestampStr := r.FormValue("last_notification"); timestampStr != "" {
		if timestamp, err := time.Parse(time.RFC3339, timestampStr); err == nil {
			b.LastNotification = timestamp.UTC()
		}
	}

	if id, err := strconv.ParseInt(r.FormValue("chat_id"), 10, 64); err == nil {
		b.ChatID = id
	}
}

func normalizeDate(s string) string {
	if s == "" {
		return ""
	}

	// Handle "MM-DD" format - convert to "0000-MM-DD"
	if len(s) == 5 && strings.Count(s, "-") == 1 {
		return "0000-" + s
	}

	// Handle "YYYY-MM-DD" format
	parsedDate, err := time.Parse("2006-01-02", s)
	if err != nil {
		return ""
	}

	// For current year dates, store as "0000-MM-DD" (year unknown)
	currentYear := time.Now().Year()
	if parsedDate.Year() == currentYear {
		month := parsedDate.Format("01")
		day := parsedDate.Format("02")
		return "0000-" + month + "-" + day
	}

	// For other years, keep the full date
	return s
}

func normalizeDateWithOriginal(s string, originalBirthDate string) string {
	if s == "" {
		return ""
	}

	// Handle "MM-DD" format - convert to "0000-MM-DD"
	if len(s) == 5 && strings.Count(s, "-") == 1 {
		return "0000-" + s
	}

	// Handle "YYYY-MM-DD" format
	parsedDate, err := time.Parse("2006-01-02", s)
	if err != nil {
		return ""
	}

	currentYear := time.Now().Year()

	// If original was a 0000 date and user enters current year, keep it as 0000
	if strings.HasPrefix(originalBirthDate, "0000-") && parsedDate.Year() == currentYear {
		month := parsedDate.Format("01")
		day := parsedDate.Format("02")
		return "0000-" + month + "-" + day
	}

	// For new birthdays (empty originalBirthDate) with current year, normalize to 0000
	if originalBirthDate == "" && parsedDate.Year() == currentYear {
		month := parsedDate.Format("01")
		day := parsedDate.Format("02")
		return "0000-" + month + "-" + day
	}

	// For any other case, keep the entered date as-is
	return s
}

func loadBirthdaysOrError(w http.ResponseWriter) ([]models.Birthday, bool) {
	bs, err := storage.LoadBirthdays()
	if err != nil {
		logger.Error("HANDLERS", "LoadBirthdays error: %v", err)
		http.Error(w, "Load error", 500)
		return nil, false
	}
	return bs, true
}

// IndexHandler returns an HTTP handler that renders the main birthday list page with bot status.
func IndexHandler(tpl *template.Template, botProvider BotStatusProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bs, ok := loadBirthdaysOrError(w)
		if !ok {
			return
		}

		var botInfo BotInfo
		if botProvider != nil {
			startHour, endHour := botProvider.GetNotificationHours()
			botInfo = BotInfo{
				Status:              botProvider.GetStatus(),
				Username:            botProvider.GetUsername(),
				FirstName:           botProvider.GetFirstName(),
				Uptime:              formatUptime(botProvider.GetUptime()),
				NotificationsSent:   botProvider.GetNotificationsSent(),
				NotificationHours:   formatNotificationHours(startHour, endHour),
				NextCheckTime:       calculateNextCheckTime(),
				CurrentHourInWindow: isCurrentlyInNotificationWindow(startHour, endHour),
				Configured:          true,
			}
		} else {
			botInfo = BotInfo{
				Status:     "not configured",
				Configured: false,
			}
		}

		data := PageData{
			Birthdays: bs,
			BotInfo:   botInfo,
		}

		if err := tpl.ExecuteTemplate(w, "page", data); err != nil {
			logger.Error("HANDLERS", "Template execute error: %v", err)
			http.Error(w, "Render error", 500)
		}
	}
}

// SaveRowHandler returns an HTTP handler that processes form submissions to add or update birthday records.
// For idx==-1, it adds a new record; otherwise, it updates the record at the given index.
func SaveRowHandler(tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idx, err := parseIdx(r)
		if err != nil {
			logger.Error("HANDLERS", "parseIdx error: %v", err)
			http.Error(w, "Invalid idx", 400)
			return
		}

		bs, ok := loadBirthdaysOrError(w)
		if !ok {
			return
		}

		if idx == -1 {
			b := models.Birthday{}
			updateBirthdayFromForm(&b, r)
			bs = append(bs, b)
		} else {
			if idx < 0 || idx >= len(bs) {
				logger.Error("HANDLERS", "SaveRowHandler invalid idx: %d", idx)
				http.Error(w, "Invalid idx", 400)
				return
			}
			updateBirthdayFromForm(&bs[idx], r)
		}

		if err := storage.SaveBirthdays(bs); err != nil {
			logger.Error("HANDLERS", "SaveBirthdays error: %v", err)
			http.Error(w, "Save error", 500)
			return
		}
		if err := tpl.ExecuteTemplate(w, "table", bs); err != nil {
			logger.Error("HANDLERS", "Template execute error: %v", err)
			http.Error(w, "Render error", 500)
		}
	}
}

// DeleteRowHandler returns an HTTP handler that processes requests to delete birthday records by index.
func DeleteRowHandler(tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idx, err := parseIdx(r)
		if err != nil {
			logger.Error("HANDLERS", "parseIdx error: %v", err)
			http.Error(w, "Invalid idx", 400)
			return
		}

		bs, ok := loadBirthdaysOrError(w)
		if !ok {
			return
		}

		if idx >= 0 && idx < len(bs) {
			bs = append(bs[:idx], bs[idx+1:]...)
			if err := storage.SaveBirthdays(bs); err != nil {
				logger.Error("HANDLERS", "SaveBirthdays error: %v", err)
				http.Error(w, "Save error", 500)
				return
			}
		} else {
			logger.Error("HANDLERS", "DeleteRowHandler invalid idx: %d", idx)
		}
		if err := tpl.ExecuteTemplate(w, "table", bs); err != nil {
			logger.Error("HANDLERS", "Template execute error: %v", err)
			http.Error(w, "Render error", 500)
		}
	}
}
