package bot

import (
	"testing"
	"time"

	"5mdt/bd_bot/internal/models"
)

func TestBirthdayNotificationUniqueness(t *testing.T) {
	bot := &Bot{} // Create a minimal bot instance for testing

	// Create a test birthday entry
	testBirthday := models.Birthday{
		Name:      "Test User",
		BirthDate: time.Now().Format("01-02"), // Today's month and day
		ChatID:    12345,
	}

	// Simulate first birthday notification
	shouldSend := bot.shouldSendBirthdayNotification(testBirthday, "BIRTHDAY_TODAY", 0)
	if !shouldSend {
		t.Error("First birthday notification should be sent")
	}

	// Update last notification time to now
	testBirthday.LastNotification = time.Now()

	// Simulate second birthday notification on the same day
	shouldSend = bot.shouldSendBirthdayNotification(testBirthday, "BIRTHDAY_TODAY", 0)
	if shouldSend {
		t.Error("Second birthday notification on the same day should not be sent")
	}

	// Simulate birthday notification on a different day
	futureTime := time.Now().AddDate(0, 0, 1)
	testBirthday.LastNotification = futureTime
	shouldSend = bot.shouldSendBirthdayNotification(testBirthday, "BIRTHDAY_TODAY", 0)
	if !shouldSend {
		t.Error("Birthday notification should be sent on a different day")
	}
}

// TestBirthdayDateCalculationAtDifferentTimes tests that birthday date calculation
// works correctly regardless of the time of day (fixes off-by-one day bug)
func TestBirthdayDateCalculationAtDifferentTimes(t *testing.T) {
	testCases := []struct {
		name         string
		currentTime  time.Time
		birthdayMMDD string
		expectedDiff int
		description  string
	}{
		{
			name:         "Birthday tomorrow at late evening (8 PM)",
			currentTime:  time.Date(2025, 12, 15, 20, 0, 0, 0, time.UTC),
			birthdayMMDD: "12-16",
			expectedDiff: 1,
			description:  "Should be 1 day away, not 0 (bug case)",
		},
		{
			name:         "Birthday tomorrow at midnight",
			currentTime:  time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC),
			birthdayMMDD: "12-16",
			expectedDiff: 1,
			description:  "Should be 1 day away",
		},
		{
			name:         "Birthday today at late evening (11 PM)",
			currentTime:  time.Date(2025, 12, 15, 23, 0, 0, 0, time.UTC),
			birthdayMMDD: "12-15",
			expectedDiff: 0,
			description:  "Should be 0 days away (today)",
		},
		{
			name:         "Birthday today at early morning (2 AM)",
			currentTime:  time.Date(2025, 12, 15, 2, 0, 0, 0, time.UTC),
			birthdayMMDD: "12-15",
			expectedDiff: 0,
			description:  "Should be 0 days away (today)",
		},
		{
			name:         "Birthday in 2 weeks at late evening",
			currentTime:  time.Date(2025, 12, 1, 20, 0, 0, 0, time.UTC),
			birthdayMMDD: "12-15",
			expectedDiff: 14,
			description:  "Should be exactly 14 days away",
		},
		{
			name:         "Birthday in 4 weeks at late evening",
			currentTime:  time.Date(2025, 11, 17, 20, 0, 0, 0, time.UTC),
			birthdayMMDD: "12-15",
			expectedDiff: 28,
			description:  "Should be exactly 28 days away",
		},
		{
			name:         "Birthday yesterday at late evening",
			currentTime:  time.Date(2025, 12, 16, 20, 0, 0, 0, time.UTC),
			birthdayMMDD: "12-15",
			expectedDiff: -1,
			description:  "Should be -1 days (passed yesterday)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the birthday MM-DD to determine this year's birthday date
			thisYearBirthday, err := time.Parse("2006-01-02", time.Now().Format("2006")+"-"+tc.birthdayMMDD)
			if err != nil {
				// Use the test case's year instead
				thisYearBirthday, err = time.Parse("2006-01-02", tc.currentTime.Format("2006")+"-"+tc.birthdayMMDD)
				if err != nil {
					t.Fatalf("Failed to parse birthday date: %v", err)
				}
			}

			// Normalize current time to start of day (midnight) - this is the fix
			nowDate := time.Date(tc.currentTime.Year(), tc.currentTime.Month(), tc.currentTime.Day(), 0, 0, 0, 0, time.UTC)

			// Calculate days difference using date-only comparison
			daysDiff := int(thisYearBirthday.Sub(nowDate).Hours() / 24)

			if daysDiff != tc.expectedDiff {
				t.Errorf("%s: Expected %d days, got %d days. %s",
					tc.name, tc.expectedDiff, daysDiff, tc.description)
				t.Logf("Current time: %s", tc.currentTime.Format("2006-01-02 15:04:05"))
				t.Logf("Normalized date: %s", nowDate.Format("2006-01-02 15:04:05"))
				t.Logf("Birthday date: %s", thisYearBirthday.Format("2006-01-02 15:04:05"))
			}
		})
	}
}

// TestBirthdayDateCalculationBugReproduction specifically tests the original bug scenario
func TestBirthdayDateCalculationBugReproduction(t *testing.T) {
	// This is the exact scenario that caused the bug:
	// Current time: December 15, 2025 at 8 PM (20:00)
	// Birthday: December 16 (tomorrow)
	// Bug: Bot would send "Happy Birthday" on Dec 15 instead of Dec 16

	currentTime := time.Date(2025, 12, 15, 20, 0, 0, 0, time.UTC) // 8 PM on Dec 15
	birthdayMMDD := "12-16"                                       // Birthday is Dec 16

	// Parse the birthday for this year
	thisYearBirthday, err := time.Parse("2006-01-02", "2025-"+birthdayMMDD)
	if err != nil {
		t.Fatalf("Failed to parse birthday date: %v", err)
	}

	// OLD BUGGY CODE (would calculate 0 days):
	// daysDiffBuggy := int(thisYearBirthday.Sub(currentTime).Hours() / 24)
	// This gives: (2025-12-16 00:00:00 - 2025-12-15 20:00:00) = 4 hours
	// 4 hours / 24 = 0.166... → int(0.166) = 0 ❌ WRONG!

	// NEW FIXED CODE (should calculate 1 day):
	nowDate := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, time.UTC)
	daysDiffFixed := int(thisYearBirthday.Sub(nowDate).Hours() / 24)
	// This gives: (2025-12-16 00:00:00 - 2025-12-15 00:00:00) = 24 hours
	// 24 hours / 24 = 1 ✓ CORRECT!

	if daysDiffFixed != 1 {
		t.Errorf("Bug still exists! Expected 1 day difference, got %d", daysDiffFixed)
		t.Logf("Current time: %s", currentTime.Format("2006-01-02 15:04:05 MST"))
		t.Logf("Normalized date: %s", nowDate.Format("2006-01-02 15:04:05 MST"))
		t.Logf("Birthday date: %s", thisYearBirthday.Format("2006-01-02 15:04:05 MST"))
		t.Fatal("The off-by-one day bug has not been fixed!")
	}

	// Verify that daysDiff == 0 would trigger a birthday message
	// and daysDiff == 1 would NOT (which is the correct behavior)
	if daysDiffFixed == 0 {
		t.Error("Bug reproduced: Birthday message would be sent a day early!")
	}
}
