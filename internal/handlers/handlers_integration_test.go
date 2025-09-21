package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"5mdt/bd_bot/internal/storage"
	"5mdt/bd_bot/internal/templates"
)

func TestIntegration_SaveAndDeleteRow(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("YAML_PATH", filepath.Join(tmp, "test.yaml"))
	defer os.Unsetenv("YAML_PATH")

	tpl := templates.LoadTemplates()

	// Add new row
	form := "idx=-1&name=TestUser&birth_date=12-31&last_notification=2024-01-01T12:00:00Z&chat_id=123"
	req := httptest.NewRequest("POST", "/save-row", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	SaveRowHandler(tpl)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Delete row
	req = httptest.NewRequest("POST", "/delete-row", strings.NewReader("idx=0"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	DeleteRowHandler(tpl)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on delete, got %d", w.Code)
	}

	bs, err := storage.LoadBirthdays()
	if err != nil {
		t.Fatal(err)
	}
	if len(bs) != 0 {
		t.Fatalf("expected 0 records, got %d", len(bs))
	}
}
