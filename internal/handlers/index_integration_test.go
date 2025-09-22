// internal/handlers/index_integration_test.go
package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"5mdt/bd_bot/internal/templates"
)

func TestIntegration_IndexHandler(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("YAML_PATH", filepath.Join(tmp, "test.yaml"))
	defer os.Unsetenv("YAML_PATH")

	tpl := templates.LoadTemplates()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	IndexHandler(tpl, nil)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "birthday-container") {
		t.Fatal("response missing birthday container")
	}
}
