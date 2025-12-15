// Package models defines the data structures for the birthday notification application.
package models

import "time"

// Birthday represents a person's birthday information stored for notifications.
type Birthday struct {
	// Name is the person's name or chat title.
	Name string `yaml:"name"`
	// BirthDate is the birth date in YYYY-MM-DD or 0000-MM-DD (year unknown) format.
	BirthDate string `yaml:"birth_date"`
	// LastNotification is the timestamp of the last birthday notification sent.
	LastNotification time.Time `yaml:"last_notification"`
	// ChatID is the Telegram chat ID for sending notifications.
	ChatID int64 `yaml:"chat_id"`
}
