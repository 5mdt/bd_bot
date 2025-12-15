package templates

import (
	"testing"
)

func TestFormatBirthDateForInput(t *testing.T) {
	// Save original nowYear and restore after test
	originalNowYear := nowYear
	defer func() { nowYear = originalNowYear }()

	// Set a fixed year for deterministic testing
	nowYear = func() int { return 2025 }

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Unknown year regular date",
			input:    "0000-06-15",
			expected: "2025-06-15",
		},
		{
			name:     "Unknown year leap day (Feb 29)",
			input:    "0000-02-29",
			expected: "2000-02-29", // Should use 2000, not 2025 (non-leap year)
		},
		{
			name:     "Known year",
			input:    "1990-12-25",
			expected: "1990-12-25",
		},
		{
			name:     "Known year leap day",
			input:    "2020-02-29",
			expected: "2020-02-29",
		},
		{
			name:     "Unknown year boundary start (Jan 1)",
			input:    "0000-01-01",
			expected: "2025-01-01",
		},
		{
			name:     "Unknown year boundary end (Dec 31)",
			input:    "0000-12-31",
			expected: "2025-12-31",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Invalid format",
			input:    "invalid",
			expected: "invalid",
		},
		{
			name:     "Invalid month (13)",
			input:    "2025-13-01",
			expected: "2025-13-01",
		},
		{
			name:     "Invalid day (Feb 30)",
			input:    "2025-02-30",
			expected: "2025-02-30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBirthDateForInput(tt.input)
			if result != tt.expected {
				t.Errorf("formatBirthDateForInput(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatBirthDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Unknown year",
			input:    "0000-06-15",
			expected: "06-15",
		},
		{
			name:     "Unknown year leap day",
			input:    "0000-02-29",
			expected: "02-29",
		},
		{
			name:     "Known year",
			input:    "1990-12-25",
			expected: "1990-12-25",
		},
		{
			name:     "Unknown year boundary start (Jan 1)",
			input:    "0000-01-01",
			expected: "01-01",
		},
		{
			name:     "Unknown year boundary end (Dec 31)",
			input:    "0000-12-31",
			expected: "12-31",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Invalid format",
			input:    "invalid",
			expected: "invalid",
		},
		{
			name:     "Invalid month (13)",
			input:    "2025-13-01",
			expected: "2025-13-01",
		},
		{
			name:     "Invalid day (Feb 30)",
			input:    "2025-02-30",
			expected: "2025-02-30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBirthDate(tt.input)
			if result != tt.expected {
				t.Errorf("formatBirthDate(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsUnknownYear(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Unknown year",
			input:    "0000-06-15",
			expected: true,
		},
		{
			name:     "Unknown year leap day",
			input:    "0000-02-29",
			expected: true,
		},
		{
			name:     "Known year",
			input:    "1990-12-25",
			expected: false,
		},
		{
			name:     "Unknown year boundary start (Jan 1)",
			input:    "0000-01-01",
			expected: true,
		},
		{
			name:     "Unknown year boundary end (Dec 31)",
			input:    "0000-12-31",
			expected: true,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "Invalid format",
			input:    "invalid",
			expected: false,
		},
		{
			name:     "Invalid date with 0000 prefix",
			input:    "0000-13-01",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUnknownYear(tt.input)
			if result != tt.expected {
				t.Errorf("isUnknownYear(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
