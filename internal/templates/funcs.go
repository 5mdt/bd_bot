// Package templates provides template rendering and helper functions for the web interface.
package templates

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// dict creates a map from alternating key-value arguments for use in templates.
// It validates that an even number of arguments are provided and all keys are strings.
func dict(v ...interface{}) (map[string]interface{}, error) {
	if len(v)%2 != 0 {
		return nil, errors.New("dict requires even number of arguments")
	}
	m := make(map[string]interface{})
	for i := 0; i < len(v); i += 2 {
		key, ok := v[i].(string)
		if !ok {
			return nil, errors.New("dict keys must be strings")
		}
		m[key] = v[i+1]
	}
	return m, nil
}

// formatTime returns a time.Time as an RFC3339 string for JavaScript consumption,
// or an empty string if the time is zero.
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// isZeroTime returns true if the given time.Time value is zero (unset).
func isZeroTime(t time.Time) bool {
	return t.IsZero()
}

// formatBirthDate formats a birth date string for display in the UI.
// If the year is 0000 (unknown), it returns only MM-DD; otherwise, it returns the full YYYY-MM-DD.
func formatBirthDate(birthDate string) string {
	if strings.HasPrefix(birthDate, "0000-") {
		// Return MM-DD for unknown year
		return strings.TrimPrefix(birthDate, "0000-")
	}
	return birthDate
}

// formatBirthDateForInput formats a birth date string for HTML date input elements.
// If the year is 0000 (unknown), it substitutes the current year for browser display.
func formatBirthDateForInput(birthDate string) string {
	if strings.HasPrefix(birthDate, "0000-") {
		// Replace 0000 with current year for browser display
		currentYear := time.Now().Year()
		return strings.Replace(birthDate, "0000", fmt.Sprintf("%d", currentYear), 1)
	}
	return birthDate
}

// isUnknownYear returns true if the birth date has an unknown year (starts with "0000-").
func isUnknownYear(birthDate string) bool {
	return strings.HasPrefix(birthDate, "0000-")
}
