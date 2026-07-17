package internal

import (
	"fmt"
	"log/slog"
	"strings"
)

// Output handles the the output of Gopher's work, depending
// on the configuration of the Gopher instance.
func (g *Gopher) Output(urlMap URLMap) {
	switch g.cfg.Output {
	case "stdout": // Print to stdout (default)
		PrintURLMap(urlMap, 0)
	case "sqlite": // Write to SQLite database
		slog.Error("Not implemented yet")
	default: // If nothing is provided, default to stdout
		slog.Error("Invalid output option, defaulting to stdout")
	}
}

// PrintURLMap recursively prints the URL map to stdout in a readable format,
// with indentation to show the heirarchy of links.
func PrintURLMap(urlMap URLMap, indentLevel int) {
	indent := strings.Repeat("  ", indentLevel)
	fmt.Printf("%s- URL: %s\n", indent, urlMap.URL)

	for _, resource := range urlMap.resources {
		fmt.Printf("%s  * Resource: %s\n", indent, resource)
	}

	for _, link := range urlMap.links {
		PrintURLMap(link, indentLevel+1)
	}
}

// WriteURLMapToSQLite writes the URL map to a SQLite database at the path of the invoked
// Gopher command (the current working directory). The database will be named "gopher_output.db".
func WriteURLMapToSQLite(urlMap URLMap, dbPath string) {
}
