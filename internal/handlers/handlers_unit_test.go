// internal/handlers/handlers_unit_test.go
package handlers

import (
	"5mdt/bd_bot/internal/models"
	"net/http"
	"net/url"
	"strings"
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
	err := updateBirthdayFromForm(b, req)
	if err != nil {
		t.Fatalf("updateBirthdayFromForm returned unexpected error: %v", err)
	}

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

func TestUpdateBirthdayFromForm_InvalidTimestamp(t *testing.T) {
	form := url.Values{
		"name":              {"Alice"},
		"birth_date":        {"12-31"},
		"last_notification": {"invalid-timestamp"},
		"chat_id":           {"123"},
	}
	req, _ := http.NewRequest("POST", "/", nil)
	req.Form = form

	b := &models.Birthday{}
	err := updateBirthdayFromForm(b, req)
	if err == nil {
		t.Fatal("Expected error for invalid timestamp, got nil")
	}
	if !strings.Contains(err.Error(), "invalid last_notification format") {
		t.Errorf("Expected error to contain 'invalid last_notification format', got: %v", err)
	}
}

func TestUpdateBirthdayFromForm_InvalidChatID(t *testing.T) {
	form := url.Values{
		"name":              {"Alice"},
		"birth_date":        {"12-31"},
		"last_notification": {"2024-01-01T15:30:00Z"},
		"chat_id":           {"not-a-number"},
	}
	req, _ := http.NewRequest("POST", "/", nil)
	req.Form = form

	b := &models.Birthday{}
	err := updateBirthdayFromForm(b, req)
	if err == nil {
		t.Fatal("Expected error for invalid chat_id, got nil")
	}
	if !strings.Contains(err.Error(), "invalid chat_id format") {
		t.Errorf("Expected error to contain 'invalid chat_id format', got: %v", err)
	}
}

func TestUpdateBirthdayFromForm_EmptyOptionalFields(t *testing.T) {
	form := url.Values{
		"name":              {"Alice"},
		"birth_date":        {"12-31"},
		"last_notification": {""},
		"chat_id":           {""},
	}
	req, _ := http.NewRequest("POST", "/", nil)
	req.Form = form

	b := &models.Birthday{}
	err := updateBirthdayFromForm(b, req)
	if err != nil {
		t.Fatalf("updateBirthdayFromForm returned unexpected error for empty optional fields: %v", err)
	}

	if b.Name != "Alice" {
		t.Errorf("Name = %q; want Alice", b.Name)
	}
	if b.BirthDate != "0000-12-31" {
		t.Errorf("BirthDate = %q; want 0000-12-31", b.BirthDate)
	}
	if !b.LastNotification.IsZero() {
		t.Errorf("LastNotification should be zero for empty input, got %v", b.LastNotification)
	}
	if b.ChatID != 0 {
		t.Errorf("ChatID should be 0 for empty input, got %d", b.ChatID)
	}
}
