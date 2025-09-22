package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"5mdt/bd_bot/internal/storage"
	"5mdt/bd_bot/internal/templates"
)

func TestDatePickerNewRowFunctionality(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("YAML_PATH", filepath.Join(tmp, "test.yaml"))
	defer os.Unsetenv("YAML_PATH")

	tpl := templates.LoadTemplates()

	// Test adding a new birthday with current year date (should normalize to 0000-MM-DD)
	form := url.Values{}
	form.Set("idx", "-1")
	form.Set("name", "NewUser")
	form.Set("birth_date", "2025-03-15") // Current year, should become 0000-03-15
	form.Set("last_notification", "2024-12-25T15:30:00Z")
	form.Set("chat_id", "456")

	req := httptest.NewRequest("POST", "/save-row", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	SaveRowHandler(tpl)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	// Verify the data was saved with normalized birth date
	birthdays, err := storage.LoadBirthdays()
	if err != nil {
		t.Fatalf("Failed to load birthdays: %v", err)
	}

	if len(birthdays) != 1 {
		t.Fatalf("Expected 1 birthday, got %d", len(birthdays))
	}

	birthday := birthdays[0]
	if birthday.Name != "NewUser" {
		t.Errorf("Expected name NewUser, got %s", birthday.Name)
	}

	// Check that current year date was normalized to year-unknown format
	if birthday.BirthDate != "0000-03-15" {
		t.Errorf("Expected birth date 0000-03-15, got %s", birthday.BirthDate)
	}

	// Test the response contains proper HTML structure for date pickers
	response := w.Body.String()

	// Should contain visible date picker
	if !strings.Contains(response, `type="date"`) {
		t.Error("Response should contain date input type")
	}

	// Should contain hidden input for birth_date
	if !strings.Contains(response, `name="birth_date"`) {
		t.Error("Response should contain birth_date name attribute")
	}

	// Should contain date input
	if !strings.Contains(response, `type="date"`) {
		t.Error("Response should contain date input type")
	}

	// Should contain birth_date name attribute
	if !strings.Contains(response, `name="birth_date"`) {
		t.Error("Response should contain birth_date name attribute")
	}
}

func TestDatePickerPastYearFunctionality(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("YAML_PATH", filepath.Join(tmp, "test.yaml"))
	defer os.Unsetenv("YAML_PATH")

	tpl := templates.LoadTemplates()

	// Test adding a birthday with past year (should keep full date)
	form := url.Values{}
	form.Set("idx", "-1")
	form.Set("name", "OldUser")
	form.Set("birth_date", "1990-07-20") // Past year, should stay 1990-07-20
	form.Set("last_notification", "")
	form.Set("chat_id", "789")

	req := httptest.NewRequest("POST", "/save-row", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	SaveRowHandler(tpl)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	// Verify the data was saved with full date
	birthdays, err := storage.LoadBirthdays()
	if err != nil {
		t.Fatalf("Failed to load birthdays: %v", err)
	}

	if len(birthdays) != 1 {
		t.Fatalf("Expected 1 birthday, got %d", len(birthdays))
	}

	birthday := birthdays[0]
	if birthday.Name != "OldUser" {
		t.Errorf("Expected name OldUser, got %s", birthday.Name)
	}

	// Check that past year date was kept as full date
	if birthday.BirthDate != "1990-07-20" {
		t.Errorf("Expected birth date 1990-07-20, got %s", birthday.BirthDate)
	}
}
