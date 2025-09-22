package main

import (
	"net/http"
	"os"

	"5mdt/bd_bot/internal/bot"
	"5mdt/bd_bot/internal/handlers"
	"5mdt/bd_bot/internal/logger"
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
		logger.Error("MAIN", "Failed to initialize Telegram bot: %v", err)
	}

	tpl := templates.LoadTemplates()


	http.HandleFunc("/", handlers.IndexHandler(tpl, telegramBot))
	http.HandleFunc("/bot-info", handlers.BotInfoHandler(tpl, telegramBot))
	http.HandleFunc("/save-row", handlers.SaveRowHandler(tpl))
	http.HandleFunc("/delete-row", handlers.DeleteRowHandler(tpl))

	addr := ":" + port
	logger.Info("MAIN", "Server starting on %s", addr)
	logger.Info("MAIN", "Debug logging enabled: %t", logger.IsDebugEnabled())
	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Error("MAIN", "Server failed to start: %v", err)
	}
}

func initBot() (*bot.Bot, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		logger.Warn("BOT", "TELEGRAM_BOT_TOKEN not set, bot will not start")
		return nil, nil
	}

	telegramBot, err := bot.New(token)
	if err != nil {
		return nil, err
	}

	telegramBot.Start()
	logger.Info("BOT", "Telegram bot started successfully")
	return telegramBot, nil
}
