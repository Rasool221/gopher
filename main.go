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

	// Initialize logging. Logs go to stderr so they stay separate from program output (the URL map,
	// which PrintURLMap writes to stdout). To send logs to a JSON file instead, swap this handler for
	// slog.NewJSONHandler(file, loggerOptions) — the result output on stdout is unaffected.
	handler := slog.NewTextHandler(os.Stderr, loggerOptions) // Log to stderr in text format.
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Map the CLI config onto the internal domain config, then build the URL map for the given URL.
	gopher := internal.NewGopher(internal.NewConfig(cfg))
	urlMap := gopher.BuildURLMap(cfg.Url, 0)

	switch cfg.Output {
	case 0: // Print to stdout (default)
		internal.PrintURLMap(urlMap, 0)
	case 1: // Write to SQLite database
		slog.Error("Not implemented yet")
	default: // If nothing is provided, default to stdout
		slog.Error("Invalid output option, defaulting to stdout")
	}
}
