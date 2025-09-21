package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"5mdt/bd_bot/internal/models"
)

func TestLoadSaveBirthdays(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("YAML_PATH", filepath.Join(tmp, "test.yaml"))
	defer os.Unsetenv("YAML_PATH")

	want := []models.Birthday{
		{Name: "Alice", BirthDate: "2000-01-01", LastNotification: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), ChatID: 123},
		{Name: "Bob", BirthDate: "0000-12-31", LastNotification: time.Date(2024, 2, 2, 15, 30, 0, 0, time.UTC), ChatID: 456},
	}

	if err := SaveBirthdays(want); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	got, err := LoadBirthdays()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if len(got) != len(want) {
		t.Fatalf("length mismatch: got %d, want %d", len(got), len(want))
	}

	for i := range got {
		if got[i].Name != want[i].Name ||
			got[i].BirthDate != want[i].BirthDate ||
			!got[i].LastNotification.Equal(want[i].LastNotification) ||
			got[i].ChatID != want[i].ChatID {
			t.Errorf("mismatch at %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}
