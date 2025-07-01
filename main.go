package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Birthday struct {
	Name             string `yaml:"name"`
	BirthDate        string `yaml:"birth_date"`
	LastNotification string `yaml:"last_notification"`
	ChatID           int64  `yaml:"chat_id"`
}

var yamlFile = "/data/birthdays.yaml"
var tpl *template.Template

func dict(v ...interface{}) (map[string]interface{}, error) {
	if len(v)%2 != 0 {
		return nil, fmt.Errorf("dict expects even number of args")
	}
	m := make(map[string]interface{}, len(v)/2)
	for i := 0; i < len(v); i += 2 {
		m[v[i].(string)] = v[i+1]
	}
	return m, nil
}

func ensureFile(fn string) error {
	if _, err := os.Stat(fn); os.IsNotExist(err) {
		return os.WriteFile(fn, []byte("[]\n"), 0644)
	}
	return nil
}

func loadBirthdays() ([]Birthday, error) {
	if err := ensureFile(yamlFile); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(yamlFile)
	if err != nil {
		return nil, err
	}
	var bs []Birthday
	return bs, yaml.Unmarshal(data, &bs)
}

func saveBirthdays(bs []Birthday) error {
	data, err := yaml.Marshal(bs)
	if err != nil {
		return err
	}
	return os.WriteFile(yamlFile, data, 0644)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	bs, _ := loadBirthdays()
	tpl.ExecuteTemplate(w, "page", bs)
}

func saveRowHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	idx, _ := strconv.Atoi(r.FormValue("idx"))
	bs, _ := loadBirthdays()
	if idx == -1 {
		b := Birthday{
			Name:             r.FormValue("name"),
			BirthDate:        r.FormValue("birth_date"),
			LastNotification: r.FormValue("last_notification"),
		}
		if id, err := strconv.ParseInt(r.FormValue("chat_id"), 10, 64); err == nil {
			b.ChatID = id
		}
		bs = append(bs, b)
	} else {
		bs[idx].Name = r.FormValue("name")
		bs[idx].BirthDate = r.FormValue("birth_date")
		bs[idx].LastNotification = r.FormValue("last_notification")
		if id, err := strconv.ParseInt(r.FormValue("chat_id"), 10, 64); err == nil {
			bs[idx].ChatID = id
		}
	}
	saveBirthdays(bs)
	tpl.ExecuteTemplate(w, "table", bs)
}

func deleteRowHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	idx, _ := strconv.Atoi(r.FormValue("idx"))
	bs, _ := loadBirthdays()
	if idx >= 0 && idx < len(bs) {
		bs = append(bs[:idx], bs[idx+1:]...)
		saveBirthdays(bs)
	}
	tpl.ExecuteTemplate(w, "table", bs)
}

const pageTpl = `
{{define "page"}}
<html><head>
<script src="https://unpkg.com/htmx.org@1.9.3"></script>
</head><body>
<h1>Edit Birthdays</h1>
{{template "table" .}}
</body></html>
{{end}}

{{define "table"}}
<div id="table">
<table border="1" cellpadding="5">
<tr><th>Name</th><th>Birth Date</th><th>Last Notification</th><th>Chat ID</th><th>Action</th></tr>
{{range $i, $b := .}}
  {{template "row" dict "Idx" $i "B" $b}}
{{end}}
<tr>
  <form hx-post="/save-row" hx-target="#table" hx-swap="outerHTML">
    <input type="hidden" name="idx" value="-1">
    <td><input name="name"></td>
    <td><input name="birth_date"></td>
    <td><input name="last_notification"></td>
    <td><input name="chat_id"></td>
    <td><button type="submit">Add</button></td>
  </form>
</tr>
</table>
</div>
{{end}}
`

const rowTpl = `
{{define "row"}}
<tr>
  <form hx-post="/save-row" hx-target="#table" hx-swap="outerHTML" style="display:inline">
    <input type="hidden" name="idx" value="{{.Idx}}">
    <td><input name="name" value="{{.B.Name}}"></td>
    <td><input name="birth_date" value="{{.B.BirthDate}}"></td>
    <td><input name="last_notification" value="{{.B.LastNotification}}"></td>
    <td><input name="chat_id" value="{{.B.ChatID}}"></td>
    <td><button type="submit">Save</button></form>
    <form hx-post="/delete-row" hx-target="#table" hx-swap="outerHTML" style="display:inline">
      <input type="hidden" name="idx" value="{{.Idx}}">
      <button type="submit">Delete</button>
    </form>
  </td>
</tr>
{{end}}
`

func main() {
	tpl = template.Must(template.New("").Funcs(template.FuncMap{"dict": dict}).Parse(pageTpl + rowTpl))
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/save-row", saveRowHandler)
	http.HandleFunc("/delete-row", deleteRowHandler)
	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
