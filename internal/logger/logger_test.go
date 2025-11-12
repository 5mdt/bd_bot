package logger

import (
	"os"
	"testing"
)

func TestLoggerInitialization(t *testing.T) {
	// Test that logger initializes correctly
	if formatMessage("INFO", "TEST", "test message %s", "arg") == "" {
		t.Error("formatMessage should not return empty string")
	}
}

func TestDebugEnvVariable(t *testing.T) {
	// Save original state
	originalDebug := os.Getenv("DEBUG")
	originalLogLevel := os.Getenv("LOG_LEVEL")

	// Clean up after test
	defer func() {
		os.Setenv("DEBUG", originalDebug)
		os.Setenv("LOG_LEVEL", originalLogLevel)
	}()

	// Test DEBUG=true enables debug logging
	os.Setenv("DEBUG", "true")
	os.Setenv("LOG_LEVEL", "")

	// Reinitialize (simulate package init)
	if debug := os.Getenv("DEBUG"); debug == "true" {
		debugEnabled = true
		logLevel = DEBUG
	}

	if !IsDebugEnabled() {
		t.Error("DEBUG=true should enable debug logging")
	}

	// Test LOG_LEVEL=ERROR
	os.Setenv("DEBUG", "")
	os.Setenv("LOG_LEVEL", "ERROR")

	if level := os.Getenv("LOG_LEVEL"); level == "ERROR" {
		logLevel = ERROR
		debugEnabled = false
	}

	if shouldLog(DEBUG) {
		t.Error("DEBUG level should not be logged when LOG_LEVEL=ERROR")
	}

	if !shouldLog(ERROR) {
		t.Error("ERROR level should be logged when LOG_LEVEL=ERROR")
	}
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		currentLevel LogLevel
		testLevel    LogLevel
		shouldLog    bool
	}{
		{DEBUG, DEBUG, true},
		{DEBUG, INFO, true},
		{DEBUG, WARN, true},
		{DEBUG, ERROR, true},
		{INFO, DEBUG, false},
		{INFO, INFO, true},
		{INFO, WARN, true},
		{INFO, ERROR, true},
		{ERROR, DEBUG, false},
		{ERROR, INFO, false},
		{ERROR, WARN, false},
		{ERROR, ERROR, true},
	}

	for _, test := range tests {
		// Temporarily set log level
		originalLevel := logLevel
		logLevel = test.currentLevel

		result := shouldLog(test.testLevel)
		if result != test.shouldLog {
			t.Errorf("Level %d with test level %d: expected %t, got %t",
				test.currentLevel, test.testLevel, test.shouldLog, result)
		}

		// Restore original level
		logLevel = originalLevel
	}
}
