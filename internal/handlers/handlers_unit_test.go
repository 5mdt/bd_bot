// internal/handlers/handlers_unit_test.go
package handlers

import (
	"5mdt/bd_bot/internal/models"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestNormalizeDate(t *testing.T) {
	currentYear := time.Now().Year()
	currentYearStr := strconv.Itoa(currentYear)

	tests := []struct {
		in, want string
	}{
		{"12-31", "0000-12-31"},
		{"2000-01-01", "2000-01-01"},
		{currentYearStr + "-03-15", "0000-03-15"}, // Current year should become year-unknown
		{"1990-07-20", "1990-07-20"},              // Past year should stay as-is
		{"invalid", ""},
		{"12345", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := normalizeDate(tt.in)
		if got != tt.want {
			t.Errorf("normalizeDate(%q) = %q; want %q", tt.in, got, tt.want)
		}
	}
}

func TestUpdateBirthdayFromForm(t *testing.T) {
	form := url.Values{
		"name":              {"Alice"},
		"birth_date":        {"12-31"},
		"last_notification": {"2024-01-01T15:30:00Z"},
		"chat_id":           {"123"},
	}
	req, _ := http.NewRequest("POST", "/", nil)
	req.Form = form

	b := &models.Birthday{}
	updateBirthdayFromForm(b, req)

	if b.Name != "Alice" {
		t.Errorf("Name = %q; want Alice", b.Name)
	}
	if b.BirthDate != "0000-12-31" {
		t.Errorf("BirthDate = %q; want 0000-12-31", b.BirthDate)
	}
	expectedTime := "2024-01-01T15:30:00Z"
	if b.LastNotification.UTC().Format("2006-01-02T15:04:05Z") != expectedTime {
		t.Errorf("LastNotification = %q; want %q", b.LastNotification.UTC().Format("2006-01-02T15:04:05Z"), expectedTime)
	}
	if b.ChatID != 123 {
		t.Errorf("ChatID = %d; want 123", b.ChatID)
	}
}
