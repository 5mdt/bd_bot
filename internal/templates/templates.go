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

func LoadTemplates() *template.Template {
	once.Do(func() {
		tpl = template.New("").Funcs(template.FuncMap{
			"dict":       dict,
			"formatTime": formatTime,
			"isZeroTime": isZeroTime,
		})
		tpl = template.Must(tpl.ParseFS(tmplFS, "tmpl/*.gohtml"))
	})
	return tpl
}
