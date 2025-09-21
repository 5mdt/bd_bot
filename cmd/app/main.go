package main

import (
	"log"
	"net/http"
	"os"

	"5mdt/bd_bot/internal/handlers"
	"5mdt/bd_bot/internal/templates"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	tpl := templates.LoadTemplates()

	http.HandleFunc("/", handlers.IndexHandler(tpl))
	http.HandleFunc("/save-row", handlers.SaveRowHandler(tpl))
	http.HandleFunc("/delete-row", handlers.DeleteRowHandler(tpl))

	addr := ":" + port
	log.Println("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
