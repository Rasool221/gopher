package main

import (
	"fmt"
	"log/slog"
	"os"

	"gopher/internal"
)

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
	case "error": // Only log errors
		loggerOptions.Level = slog.LevelError
	case "info": // Only log info and above (default)
		loggerOptions.Level = slog.LevelInfo
	case "debug": // Log debug and above
		loggerOptions.Level = slog.LevelDebug
	default: // If nothing is provided, default to info level
		loggerOptions.Level = slog.LevelInfo
	}

	// Initialize logging. Logs go to stderr so they stay separate from program output (the URL map,
	// which PrintURLMap writes to stdout).
	handler := slog.NewTextHandler(os.Stderr, loggerOptions) // Log to stderr in text format.
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Map the CLI config onto the internal domain config, then build the URL map for the given URL.
	gopher := internal.NewGopher(internal.NewConfig(cfg))
	gopher.Output(gopher.Run(cfg.Url))
}
