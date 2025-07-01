package templates

import (
	"html/template"
)

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
    <td>
      <button type="submit">Save</button>
  </form>
  <form hx-post="/delete-row" hx-target="#table" hx-swap="outerHTML" style="display:inline">
      <input type="hidden" name="idx" value="{{.Idx}}">
      <button type="submit">Delete</button>
  </form>
  </td>
</tr>
{{end}}
`

func LoadTemplates() *template.Template {
	t := template.New("").Funcs(template.FuncMap{"dict": dict})
	t = template.Must(t.Parse(pageTpl + rowTpl))
	return t
}
