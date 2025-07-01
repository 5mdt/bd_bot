package storage

import (
	"os"

	"gopkg.in/yaml.v3"
	"5mdt/bd_bot/internal/models"
)

var yamlFile = "/data/birthdays.yaml"

func EnsureFile() error {
	if _, err := os.Stat(yamlFile); os.IsNotExist(err) {
		return os.WriteFile(yamlFile, []byte("[]\n"), 0644)
	}
	return nil
}

func LoadBirthdays() ([]models.Birthday, error) {
	if err := EnsureFile(); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(yamlFile)
	if err != nil {
		return nil, err
	}
	var bs []models.Birthday
	return bs, yaml.Unmarshal(data, &bs)
}

func SaveBirthdays(bs []models.Birthday) error {
	data, err := yaml.Marshal(bs)
	if err != nil {
		return err
	}
	return os.WriteFile(yamlFile, data, 0644)
}
