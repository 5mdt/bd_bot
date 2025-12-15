// internal/handlers/handlers_unit_test.go
package handlers

import (
	"5mdt/bd_bot/internal/models"
	"net/http"
	"net/url"
	"testing"
)

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
