package models

import (
	"testing"
	"time"
)

func TestBirthdayStruct(t *testing.T) {
	testTime := time.Date(2025, 6, 30, 12, 0, 0, 0, time.UTC)
	b := Birthday{
		Name:             "Alice",
		BirthDate:        "2000-01-01",
		LastNotification: testTime,
		ChatID:           12345,
	}
	if b.Name != "Alice" {
		t.Errorf("expected Name Alice, got %s", b.Name)
	}
	if b.BirthDate != "2000-01-01" {
		t.Errorf("expected BirthDate 2000-01-01, got %s", b.BirthDate)
	}
	if !b.LastNotification.Equal(testTime) {
		t.Errorf("expected LastNotification %v, got %v", testTime, b.LastNotification)
	}
	if b.ChatID != 12345 {
		t.Errorf("expected ChatID 12345, got %d", b.ChatID)
	}
}
