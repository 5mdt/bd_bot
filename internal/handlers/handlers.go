package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"5mdt/bd_bot/internal/models"
	"5mdt/bd_bot/internal/storage"
)

type PageData struct {
	Birthdays []models.Birthday
	BotInfo   BotInfo
}

type BotInfo struct {
	Status            string
	Username          string
	FirstName         string
	Uptime            string
	NotificationsSent int64
	Configured        bool
}

type BotStatusProvider interface {
	GetStatus() string
	GetUsername() string
	GetFirstName() string
	GetUptime() time.Duration
	GetNotificationsSent() int64
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

func parseIdx(r *http.Request) (int, error) {
	if err := r.ParseForm(); err != nil {
		return 0, err
	}
	return strconv.Atoi(r.FormValue("idx"))
}

func updateBirthdayFromForm(b *models.Birthday, r *http.Request) {
	b.Name = r.FormValue("name")
	b.BirthDate = normalizeDate(r.FormValue("birth_date"))

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

func loadBirthdaysOrError(w http.ResponseWriter) ([]models.Birthday, bool) {
	bs, err := storage.LoadBirthdays()
	if err != nil {
		log.Println("LoadBirthdays error:", err)
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
			botInfo = BotInfo{
				Status:            botProvider.GetStatus(),
				Username:          botProvider.GetUsername(),
				FirstName:         botProvider.GetFirstName(),
				Uptime:            formatUptime(botProvider.GetUptime()),
				NotificationsSent: botProvider.GetNotificationsSent(),
				Configured:        true,
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
			log.Println("Template execute error:", err)
			http.Error(w, "Render error", 500)
		}
	}
}

func SaveRowHandler(tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idx, err := parseIdx(r)
		if err != nil {
			log.Println("parseIdx error:", err)
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
				log.Println("SaveRowHandler invalid idx:", idx)
				http.Error(w, "Invalid idx", 400)
				return
			}
			updateBirthdayFromForm(&bs[idx], r)
		}

		if err := storage.SaveBirthdays(bs); err != nil {
			log.Println("SaveBirthdays error:", err)
			http.Error(w, "Save error", 500)
			return
		}
		if err := tpl.ExecuteTemplate(w, "table", bs); err != nil {
			log.Println("Template execute error:", err)
			http.Error(w, "Render error", 500)
		}
	}
}

func DeleteRowHandler(tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idx, err := parseIdx(r)
		if err != nil {
			log.Println("parseIdx error:", err)
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
				log.Println("SaveBirthdays error:", err)
				http.Error(w, "Save error", 500)
				return
			}
		} else {
			log.Println("DeleteRowHandler invalid idx:", idx)
		}
		if err := tpl.ExecuteTemplate(w, "table", bs); err != nil {
			log.Println("Template execute error:", err)
			http.Error(w, "Render error", 500)
		}
	}
}
