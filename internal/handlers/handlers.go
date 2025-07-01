package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"5mdt/bd_bot/internal/models"
	"5mdt/bd_bot/internal/storage"
)

func IndexHandler(tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bs, _ := storage.LoadBirthdays()
		tpl.ExecuteTemplate(w, "page", bs)
	}
}

func SaveRowHandler(tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		idx, _ := strconv.Atoi(r.FormValue("idx"))
		bs, _ := storage.LoadBirthdays()

		if idx == -1 {
			b := models.Birthday{
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
		storage.SaveBirthdays(bs)
		tpl.ExecuteTemplate(w, "table", bs)
	}
}

func DeleteRowHandler(tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		idx, _ := strconv.Atoi(r.FormValue("idx"))
		bs, _ := storage.LoadBirthdays()
		if idx >= 0 && idx < len(bs) {
			bs = append(bs[:idx], bs[idx+1:]...)
			storage.SaveBirthdays(bs)
		}
		tpl.ExecuteTemplate(w, "table", bs)
	}
}
