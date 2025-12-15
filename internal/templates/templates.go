package templates

import (
	"embed"
	"html/template"
	"sync"
)

//go:embed tmpl/*
var tmplFS embed.FS

var (
	tpl  *template.Template
	once sync.Once
)

// LoadTemplates loads and parses all HTML templates from the embedded filesystem,
// registers custom template functions, and returns a cached template for rendering.
// Uses sync.Once to ensure templates are loaded only once.
func LoadTemplates() *template.Template {
	once.Do(func() {
		// Load template with functions for birth date handling (updated)
		tpl = template.New("").Funcs(template.FuncMap{
			"dict":                    dict,
			"formatTime":              formatTime,
			"isZeroTime":              isZeroTime,
			"formatBirthDate":         formatBirthDate,
			"formatBirthDateForInput": formatBirthDateForInput,
			"isUnknownYear":           isUnknownYear,
		})
		tpl = template.Must(tpl.ParseFS(tmplFS, "tmpl/*.gohtml"))
	})
	return tpl
}
