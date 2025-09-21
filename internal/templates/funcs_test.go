package templates

import (
	"testing"
	"time"
)

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{"zero time", time.Time{}, ""},
		{"valid time", time.Date(2024, 1, 1, 15, 30, 0, 0, time.UTC), "2024-01-01T15:30:00Z"},
		{"non-UTC time", time.Date(2024, 1, 1, 15, 30, 0, 0, time.FixedZone("EST", -5*3600)), "2024-01-01T20:30:00Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTime(tt.time)
			if got != tt.want {
				t.Errorf("formatTime() = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestIsZeroTime(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want bool
	}{
		{"zero time", time.Time{}, true},
		{"valid time", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isZeroTime(tt.time)
			if got != tt.want {
				t.Errorf("isZeroTime() = %v; want %v", got, tt.want)
			}
		})
	}
}
