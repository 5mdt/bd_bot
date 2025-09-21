package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"5mdt/bd_bot/internal/storage"
	"5mdt/bd_bot/internal/templates"
)

func TestDatetimePickerIntegration(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("YAML_PATH", filepath.Join(tmp, "test.yaml"))
	defer os.Unsetenv("YAML_PATH")

	tpl := templates.LoadTemplates()

	// Test saving a birthday with a specific timestamp
	testTime := "2024-12-25T15:30:00Z" // UTC timestamp from datetime picker

	form := url.Values{}
	form.Set("idx", "-1")
	form.Set("name", "TestUser")
	form.Set("birth_date", "0000-12-25")
	form.Set("last_notification", testTime)
	form.Set("chat_id", "123")

	req := httptest.NewRequest("POST", "/save-row", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	SaveRowHandler(tpl)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	// Verify the data was saved correctly
	birthdays, err := storage.LoadBirthdays()
	if err != nil {
		t.Fatalf("Failed to load birthdays: %v", err)
	}

	if len(birthdays) != 1 {
		t.Fatalf("Expected 1 birthday, got %d", len(birthdays))
	}

	birthday := birthdays[0]
	if birthday.Name != "TestUser" {
		t.Errorf("Expected name TestUser, got %s", birthday.Name)
	}

	if birthday.BirthDate != "0000-12-25" {
		t.Errorf("Expected birth date 0000-12-25, got %s", birthday.BirthDate)
	}

	if birthday.ChatID != 123 {
		t.Errorf("Expected chat ID 123, got %d", birthday.ChatID)
	}

	// Check that the timestamp was parsed correctly
	expectedTime, _ := time.Parse(time.RFC3339, testTime)
	if !birthday.LastNotification.Equal(expectedTime) {
		t.Errorf("Expected timestamp %v, got %v", expectedTime, birthday.LastNotification)
	}

	// Test that the template renders with proper datetime picker attributes
	response := w.Body.String()
	if !strings.Contains(response, `type="datetime-local"`) {
		t.Error("Response should contain datetime-local input")
	}

	if !strings.Contains(response, `type="date"`) {
		t.Error("Response should contain date input for birth date")
	}
}
