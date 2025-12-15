// Package storage provides YAML-based persistence for birthday data.
// It manages loading and saving birthday records from/to a configurable file path.
package storage

import (
	"os"
	"path/filepath"

	"5mdt/bd_bot/internal/models"
	"gopkg.in/yaml.v3"
)

const filePerm = 0644

func getPath() string {
	if path := os.Getenv("YAML_PATH"); path != "" {
		return path
	}
	return "/data/birthdays.yaml"
}

// LoadBirthdays reads and parses birthday data from the configured YAML file.
// It creates an empty file (and parent directories) if it doesn't exist.
// Returns a nil slice and error on failure.
func LoadBirthdays() ([]models.Birthday, error) {
	filePath := getPath()
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Ensure parent directory exists
		if dir := filepath.Dir(filePath); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, err
			}
		}
		if err := os.WriteFile(filePath, []byte("[]\n"), filePerm); err != nil {
			return nil, err
		}
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var bs []models.Birthday
	return bs, yaml.Unmarshal(data, &bs)
}

// SaveBirthdays marshals birthday data to YAML and writes it to the configured file path.
// It creates parent directories if they don't exist.
func SaveBirthdays(bs []models.Birthday) error {
	filePath := getPath()
	data, err := yaml.Marshal(bs)
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	if dir := filepath.Dir(filePath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return os.WriteFile(filePath, data, filePerm)
}
