package main

import (
	"log"
	"net/http"

	"5mdt/bd_bot/internal/handlers"
	"5mdt/bd_bot/internal/templates"
)

func main() {
	tpl := templates.LoadTemplates()

	http.HandleFunc("/", handlers.IndexHandler(tpl))
	http.HandleFunc("/save-row", handlers.SaveRowHandler(tpl))
	http.HandleFunc("/delete-row", handlers.DeleteRowHandler(tpl))

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
