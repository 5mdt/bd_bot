package templates

import (
	"strings"
	"testing"
	"time"

	"5mdt/bd_bot/internal/models"
)

func TestDatetimePickerRendering(t *testing.T) {
	tpl := LoadTemplates()

	// Test with a birthday that has a timestamp
	testTime := time.Date(2024, 12, 25, 15, 30, 0, 0, time.UTC)
	birthday := models.Birthday{
		Name:             "Test User",
		BirthDate:        "12-25",
		LastNotification: testTime,
		ChatID:           123,
	}

	var buf strings.Builder
	data := map[string]interface{}{
		"Idx": 0,
		"B":   birthday,
	}

	err := tpl.ExecuteTemplate(&buf, "card", data)
	if err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	output := buf.String()

	// Check that datetime picker input is present
	if !strings.Contains(output, `type="datetime-local"`) {
		t.Error("Expected datetime-local input to be present")
	}

	// Check that datetime picker has the correct class
	if !strings.Contains(output, `datetime-picker`) {
		t.Error("Expected datetime-picker class to be present")
	}

	// Check that "Now" button is present
	if !strings.Contains(output, `class="set-now-btn"`) {
		t.Error("Expected set-now-btn class to be present")
	}

	// Check that hidden input for last_notification is present
	if !strings.Contains(output, `name="last_notification"`) {
		t.Error("Expected last_notification hidden input to be present")
	}

	// Check that data-utc attribute contains the formatted time
	expectedTime := "2024-12-25T15:30:00Z"
	if !strings.Contains(output, expectedTime) {
		t.Errorf("Expected UTC time %s to be present in output", expectedTime)
	}
}

func TestEmptyDatetimePickerRendering(t *testing.T) {
	tpl := LoadTemplates()

	// Test with a birthday that has no timestamp (zero time)
	birthday := models.Birthday{
		Name:      "Test User",
		BirthDate: "01-01",
		ChatID:    456,
	}

	var buf strings.Builder
	data := map[string]interface{}{
		"Idx": 0,
		"B":   birthday,
	}

	err := tpl.ExecuteTemplate(&buf, "card", data)
	if err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	output := buf.String()

	// Check that datetime picker input is present even with empty timestamp
	if !strings.Contains(output, `type="datetime-local"`) {
		t.Error("Expected datetime-local input to be present even with empty timestamp")
	}

	// Check that data-utc attribute is empty for zero time
	if !strings.Contains(output, `data-utc=""`) {
		t.Error("Expected empty data-utc attribute for zero time")
	}
}

func TestBirthDatePickerRendering(t *testing.T) {
	tpl := LoadTemplates()

	// Test with a birthday that has a birth date
	birthday := models.Birthday{
		Name:      "Test User",
		BirthDate: "0000-12-25", // Year unknown format
		ChatID:    123,
	}

	var buf strings.Builder
	data := map[string]interface{}{
		"Idx": 0,
		"B":   birthday,
	}

	err := tpl.ExecuteTemplate(&buf, "card", data)
	if err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	output := buf.String()

	// Check that date picker input is present
	if !strings.Contains(output, `type="date"`) {
		t.Error("Expected date input to be present")
	}

	// Check that date input is present
	if !strings.Contains(output, `type="date"`) {
		t.Error("Expected date input to be present")
	}

	// Check that birth date value is present
	if !strings.Contains(output, `value="0000-12-25"`) {
		t.Error("Expected birth date value to be present")
	}

	// Check that birth date input name is present
	if !strings.Contains(output, `name="birth_date"`) {
		t.Error("Expected birth_date name attribute to be present")
	}
}
