package main

import (
	"log"
	"net/http"
	"os"

	"5mdt/bd_bot/internal/bot"
	"5mdt/bd_bot/internal/handlers"
	"5mdt/bd_bot/internal/templates"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize Telegram bot
	telegramBot, err := initBot()
	if err != nil {
		log.Printf("Failed to initialize Telegram bot: %v", err)
	}

	tpl := templates.LoadTemplates()

	http.HandleFunc("/", handlers.IndexHandler(tpl, telegramBot))
	http.HandleFunc("/save-row", handlers.SaveRowHandler(tpl))
	http.HandleFunc("/delete-row", handlers.DeleteRowHandler(tpl))

	addr := ":" + port
	log.Println("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func initBot() (*bot.Bot, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Println("TELEGRAM_BOT_TOKEN not set, bot will not start")
		return nil, nil
	}

	telegramBot, err := bot.New(token)
	if err != nil {
		return nil, err
	}

	telegramBot.Start()
	log.Println("Telegram bot started")
	return telegramBot, nil
}
