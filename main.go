package main

import "gopher/internal"
import "log/slog"
import "os"

import "fmt"

func main() {
	cfg := internal.ParseClI()

	err := internal.ValidateCLI(cfg)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	loggerOptions := &slog.HandlerOptions{}

	// The heirarchy is:
	// 0: error
	// 1: info (default)
	// 2: debug
	// Each number will log that level and all levels above it.
	switch cfg.LogLevel {
	case 0: // Only log errors
		loggerOptions.Level = slog.LevelError
	case 1: // Only log info and above (default)
		loggerOptions.Level = slog.LevelInfo
	case 2: // Log debug and above
		loggerOptions.Level = slog.LevelDebug
	default: // If nothing is provided, default to info level
		loggerOptions.Level = slog.LevelInfo
	}

	// Initialize logging.
	handler := slog.NewTextHandler(os.Stdout, loggerOptions) // Log to stdout in text format.
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Build the URL map for the given URL.
	urlMap := internal.BuildUrlMap(cfg.Url, nil)

	// Print the URL map in a readable format.
	internal.PrintURLMap(urlMap, 0)
}
