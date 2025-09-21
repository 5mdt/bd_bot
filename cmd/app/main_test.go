package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"5mdt/bd_bot/internal/handlers"
	"5mdt/bd_bot/internal/templates"
)

func doRequest(t *testing.T, method, path string, form url.Values, handler http.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	body := strings.NewReader("")
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req := httptest.NewRequest(method, path, body)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func TestMainHandlers(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "birthdays.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Write([]byte("[]"))
	tmpfile.Close()
	os.Setenv("YAML_PATH", tmpfile.Name())
	defer os.Unsetenv("YAML_PATH")

	tpl := templates.LoadTemplates()

	w := doRequest(t, "GET", "/", nil, handlers.IndexHandler(tpl, nil))
	if w.Code != http.StatusOK {
		t.Errorf("GET / returned %d", w.Code)
	}

	form := url.Values{
		"idx":              {"-1"},
		"name":             {"X"},
		"birth_date":       {"01-01"},
		"last_notification":{"2025-01-01T12:00:00Z"},
		"chat_id":          {"1"},
	}
	w = doRequest(t, "POST", "/save-row", form, handlers.SaveRowHandler(tpl))
	if w.Code != http.StatusOK {
		t.Errorf("POST /save-row returned %d", w.Code)
	}

	del := url.Values{"idx": {"0"}}
	w = doRequest(t, "POST", "/delete-row", del, handlers.DeleteRowHandler(tpl))
	if w.Code != http.StatusOK {
		t.Errorf("POST /delete-row returned %d", w.Code)
	}
}
