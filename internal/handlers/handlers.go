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

type PageData struct {
	Birthdays []models.Birthday
	BotInfo   BotInfo
}

type BotInfo struct {
	Status              string
	Username            string
	FirstName           string
	Uptime              string
	NotificationsSent   int64
	NotificationHours   string
	NextCheckTime       string
	CurrentHourInWindow bool
	Configured          bool
}

type BotStatusProvider interface {
	GetStatus() string
	GetUsername() string
	GetFirstName() string
	GetUptime() time.Duration
	GetNotificationsSent() int64
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
