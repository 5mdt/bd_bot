package handlers

import (
	"html/template"
	"net/http"

	"5mdt/bd_bot/internal/logger"
)

// BotInfoHandler returns an HTTP handler that renders the bot status information as partial HTML.
// It queries the bot provider for current status, uptime, and notification metrics.
func BotInfoHandler(tpl *template.Template, botProvider BotStatusProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		var botInfo BotInfo
		if botProvider != nil && botProvider.GetStatus() != "not configured" {
			startHour, endHour := botProvider.GetNotificationHours()
			botInfo = BotInfo{
				Status:              botProvider.GetStatus(),
				Username:            botProvider.GetUsername(),
				FirstName:           botProvider.GetFirstName(),
				Uptime:              formatUptime(botProvider.GetUptime()),
				NotificationsSent:   botProvider.GetNotificationsSent(),
				NotificationHours:   formatNotificationHours(startHour, endHour),
				NextCheckTime:       calculateNextCheckTime(),
				CurrentHourInWindow: isCurrentlyInNotificationWindow(startHour, endHour),
				Configured:          true,
			}
		} else {
			botInfo = BotInfo{
				Status:     "not configured",
				Configured: false,
			}
		}

		// Create context for the bot info template
		data := map[string]interface{}{
			"Bot": botInfo,
		}

		// Execute just the bot-info template
		if err := tpl.ExecuteTemplate(w, "bot-info", data); err != nil {
			logger.Error("HANDLERS", "Bot info template execute error: %v", err)
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}
	}
}
