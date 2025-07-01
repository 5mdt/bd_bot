package models

type Birthday struct {
	Name             string `yaml:"name"`
	BirthDate        string `yaml:"birth_date"`
	LastNotification string `yaml:"last_notification"`
	ChatID           int64  `yaml:"chat_id"`
}
