package templates

import (
	"errors"
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
