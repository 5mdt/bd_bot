// Package logger provides structured logging utilities with configurable log levels.
// It supports DEBUG, INFO, WARN, and ERROR levels controlled via environment variables
// (DEBUG=true or LOG_LEVEL=<level>). Includes specialized helpers for HTTP, bot, storage,
// and notification logging.
package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// LogLevel represents the severity level of log messages.
type LogLevel int

const (
	// DEBUG logs diagnostic information (enabled via DEBUG=true or LOG_LEVEL=DEBUG).
	DEBUG LogLevel = iota
	// INFO logs general informational messages.
	INFO
	// WARN logs warning messages for potentially problematic situations.
	WARN
	// ERROR logs error messages for failures and exceptions.
	ERROR
)

var (
	debugEnabled = false
	logLevel     = INFO
)

func init() {
	// Check DEBUG environment variable
	if debug := os.Getenv("DEBUG"); debug != "" {
		debugValue := strings.ToLower(debug)
		if debugValue == "true" || debugValue == "1" || debugValue == "yes" {
			debugEnabled = true
			logLevel = DEBUG
		}
	}

	// Check LOG_LEVEL environment variable
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		switch strings.ToUpper(level) {
		case "DEBUG":
			logLevel = DEBUG
			debugEnabled = true
		case "INFO":
			logLevel = INFO
		case "WARN", "WARNING":
			logLevel = WARN
		case "ERROR":
			logLevel = ERROR
		}
	}

	// Configure log format
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

func shouldLog(level LogLevel) bool {
	return level >= logLevel
}

func formatMessage(level string, component string, format string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)

	if component != "" {
		return fmt.Sprintf("%s [%s] [%s] %s", timestamp, level, component, message)
	}
	return fmt.Sprintf("%s [%s] %s", timestamp, level, message)
}

// Debug logs debug messages with the specified component and format (only when DEBUG=true or LOG_LEVEL=DEBUG).
func Debug(component string, format string, args ...interface{}) {
	if shouldLog(DEBUG) {
		log.Print(formatMessage("DEBUG", component, format, args...))
	}
}

// Info logs informational messages with the specified component and format.
func Info(component string, format string, args ...interface{}) {
	if shouldLog(INFO) {
		log.Print(formatMessage("INFO", component, format, args...))
	}
}

// Warn logs warning messages with the specified component and format.
func Warn(component string, format string, args ...interface{}) {
	if shouldLog(WARN) {
		log.Print(formatMessage("WARN", component, format, args...))
	}
}

// Error logs error messages with the specified component and format.
func Error(component string, format string, args ...interface{}) {
	if shouldLog(ERROR) {
		log.Print(formatMessage("ERROR", component, format, args...))
	}
}

// IsDebugEnabled returns true if debug logging is enabled via environment variables.
func IsDebugEnabled() bool {
	return debugEnabled
}

// Debugf logs debug messages without a component prefix for backward compatibility.
func Debugf(format string, args ...interface{}) {
	Debug("", format, args...)
}

// Infof logs informational messages without a component prefix for backward compatibility.
func Infof(format string, args ...interface{}) {
	Info("", format, args...)
}

// Warnf logs warning messages without a component prefix for backward compatibility.
func Warnf(format string, args ...interface{}) {
	Warn("", format, args...)
}

// Errorf logs error messages without a component prefix for backward compatibility.
func Errorf(format string, args ...interface{}) {
	Error("", format, args...)
}

// LogRequest logs HTTP request information including method, path, status code, and duration.
// In debug mode, it also logs the user agent string.
func LogRequest(method, path, userAgent string, statusCode int, duration time.Duration) {
	if debugEnabled {
		Debug("HTTP", "%s %s - %d (%v) - %s", method, path, statusCode, duration, userAgent)
	} else {
		Info("HTTP", "%s %s - %d", method, path, statusCode)
	}
}

// LogBotMessage logs incoming Telegram bot messages with chat ID and username.
// In debug mode, it also logs the full message content.
func LogBotMessage(chatID int64, username, message string) {
	if debugEnabled {
		Debug("BOT", "Message from %s (Chat ID: %d): %s", username, chatID, message)
	} else {
		Info("BOT", "Message from %s (Chat ID: %d)", username, chatID)
	}
}

// LogBotAction logs Telegram bot actions with success/failure status.
func LogBotAction(action, target string, success bool) {
	if debugEnabled {
		Debug("BOT", "Action: %s -> %s (success: %t)", action, target, success)
	} else if !success {
		Error("BOT", "Failed action: %s -> %s", action, target)
	} else {
		Info("BOT", "Action: %s -> %s", action, target)
	}
}

// LogStorage logs storage operations (load, save, etc.) with optional error details.
func LogStorage(operation string, details string, err error) {
	if err != nil {
		Error("STORAGE", "%s failed: %v - %s", operation, err, details)
	} else if debugEnabled {
		Debug("STORAGE", "%s success: %s", operation, details)
	} else {
		Info("STORAGE", "%s completed", operation)
	}
}

// LogNotification logs birthday notification events with the specified severity level.
func LogNotification(level string, message string, args ...interface{}) {
	component := "NOTIFICATION"
	formattedMsg := fmt.Sprintf(message, args...)

	switch strings.ToUpper(level) {
	case "DEBUG":
		Debug(component, formattedMsg)
	case "INFO":
		Info(component, formattedMsg)
	case "WARN", "WARNING":
		Warn(component, formattedMsg)
	case "ERROR":
		Error(component, formattedMsg)
	default:
		Info(component, formattedMsg)
	}
}
