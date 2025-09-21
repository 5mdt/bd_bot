package storage

import (
	"os"

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

func LoadBirthdays() ([]models.Birthday, error) {
	filePath := getPath()
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
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

func SaveBirthdays(bs []models.Birthday) error {
	filePath := getPath()
	data, err := yaml.Marshal(bs)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, filePerm)
}
