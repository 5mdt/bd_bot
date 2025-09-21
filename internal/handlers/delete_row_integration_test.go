// internal/handlers/delete_row_integration_test.go
package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"5mdt/bd_bot/internal/templates"
)

func TestIntegration_DeleteRowHandler(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("YAML_PATH", filepath.Join(tmp, "test.yaml"))
	defer os.Unsetenv("YAML_PATH")

	tpl := templates.LoadTemplates()

	// first add a row
	form := url.Values{}
	form.Set("idx", "-1")
	form.Set("name", "ToDelete")
	form.Set("birth_date", "01-01")
	form.Set("last_notification", "2025-01-01T12:00:00Z")
	form.Set("chat_id", "111")
	req := httptest.NewRequest("POST", "/save-row", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	SaveRowHandler(tpl)(httptest.NewRecorder(), req)

	// now delete it
	del := url.Values{}
	del.Set("idx", "0")
	req = httptest.NewRequest("POST", "/delete-row", strings.NewReader(del.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	DeleteRowHandler(tpl)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "ToDelete") {
		t.Fatal("deleted row still present")
	}
}
