package models

import "time"

type Birthday struct {
	Name             string    `yaml:"name"`
	BirthDate        string    `yaml:"birth_date"`
	LastNotification time.Time `yaml:"last_notification"`
	ChatID           int64     `yaml:"chat_id"`
}
