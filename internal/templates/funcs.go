package templates

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

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

// formatTime formats a time.Time as an ISO string for JavaScript consumption
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// isZeroTime checks if a time.Time is zero
func isZeroTime(t time.Time) bool {
	return t.IsZero()
}

// formatBirthDate formats birth date for display
// If year is 0000 (unknown), show only MM-DD
// Otherwise show the full YYYY-MM-DD
func formatBirthDate(birthDate string) string {
	if strings.HasPrefix(birthDate, "0000-") {
		// Return MM-DD for unknown year
		return strings.TrimPrefix(birthDate, "0000-")
	}
	return birthDate
}

// formatBirthDateForInput formats birth date for HTML date input
// If year is 0000 (unknown), substitute current year so browser can display it
// Otherwise return the full date for the input
func formatBirthDateForInput(birthDate string) string {
	if strings.HasPrefix(birthDate, "0000-") {
		// Replace 0000 with current year for browser display
		currentYear := time.Now().Year()
		return strings.Replace(birthDate, "0000", fmt.Sprintf("%d", currentYear), 1)
	}
	return birthDate
}

// isUnknownYear checks if a birth date has an unknown year (0000)
func isUnknownYear(birthDate string) bool {
	return strings.HasPrefix(birthDate, "0000-")
}
