// internal/handlers/save_row_integration_test.go
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

func TestIntegration_SaveRowHandler(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("YAML_PATH", filepath.Join(tmp, "test.yaml"))
	defer os.Unsetenv("YAML_PATH")

	tpl := templates.LoadTemplates()

	form := url.Values{}
	form.Set("idx", "-1")
	form.Set("name", "TestName")
	form.Set("birth_date", "12-31")
	form.Set("last_notification", "2025-01-01T12:00:00Z")
	form.Set("chat_id", "789")

	req := httptest.NewRequest("POST", "/save-row", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	SaveRowHandler(tpl)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "TestName") {
		t.Fatal("response missing saved name")
	}
}
