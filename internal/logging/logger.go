// Package logging provides structured logging with debug support
package logging

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

// InitLogger initializes the global logger with debug flag
func InitLogger(debug bool) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if debug {
		opts.Level = slog.LevelDebug
	}

	logger = slog.New(slog.NewTextHandler(os.Stderr, opts))
	slog.SetDefault(logger)
}

// Debug logs a debug message with structured fields
func Debug(msg string, args ...any) {
	if logger != nil {
		logger.Debug(msg, args...)
	}
}

// Info logs an info message with structured fields
func Info(msg string, args ...any) {
	if logger != nil {
		logger.Info(msg, args...)
	}
}

// Warn logs a warning message with structured fields
func Warn(msg string, args ...any) {
	if logger != nil {
		logger.Warn(msg, args...)
	}
}

// Error logs an error message with structured fields
func Error(msg string, args ...any) {
	if logger != nil {
		logger.Error(msg, args...)
	}
}

// GetLogger returns the global logger instance
func GetLogger() *slog.Logger {
	return logger
}
