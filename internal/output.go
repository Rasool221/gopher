package internal

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// Output handles the the output of Gopher's work, depending
// on the configuration of the Gopher instance.
func (g *Gopher) Output(urlMap URLMap) {
	switch g.cfg.Output {
	case "stdout": // Print to stdout (default)
		PrintURLMap(urlMap, 0)
	case "sqlite": // Write to SQLite database
		err := writeURLMapToSQLite(urlMap, g.cfg)
		if err != nil {
			slog.Error("Failed to write output to SQLite", "error", err)
		}
	default: // If nothing is provided, default to stdout
		slog.Error("Invalid output option, defaulting to stdout")
	}

	fmt.Printf("\nGopher run complete. Run ID: %s\n", g.cfg.RunID)
}

// PrintURLMap recursively prints the URL map to stdout in a readable format,
// with indentation to show the heirarchy of links.
func PrintURLMap(urlMap URLMap, indentLevel int) {
	indent := strings.Repeat("  ", indentLevel)
	fmt.Printf("%s- URL: %s\n", indent, urlMap.URL)

	for _, resource := range urlMap.resources {
		fmt.Printf("%s  * Resource: %s\n", indent, resource)
	}

	for _, error := range urlMap.errors {
		fmt.Printf("%s  * Error: %s\n", indent, error)
	}

	for _, link := range urlMap.links {
		PrintURLMap(link, indentLevel+1)
	}
}

// writeURLMapToSQLite writes the URL map to a SQLite database at the path of the invoked
// Gopher command (the current working directory). The database will be named "gopher_output.db".
func writeURLMapToSQLite(urlMap URLMap, cfg Config) error {
	// Determine the path for the SQLite database.
	dbPath := cfg.DBPath
	dbFullPath := filepath.Join(dbPath, gopherSqliteDBName)

	// Creating the SQLite database binary.
	slog.Debug("Writing output to SQLite", "dbFullPath", dbFullPath)
	db, err := sql.Open("sqlite", dbFullPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Creating the schema of the DB.
	slog.Debug("Writing SQLite DB schema")
	_, err = db.Exec(gopherSqliteDBCreateSQL)
	if err != nil {
		return err
	}

	// Inserting output data to the SQLite database within a single transaction.
	slog.Debug("Inserting output data to SQLite database")
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	err = insertURLMapToSQLite(tx, urlMap, "", cfg)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// insertURLMapToSQLite inserts a URLMap node and its resources and errors, then recurses
// into its links. parentURL is empty for the root of the crawl.
func insertURLMapToSQLite(tx *sql.Tx, node URLMap, parentURL string, cfg Config) error {
	if _, err := tx.Exec(
		gopherSqliteInsertLinkSQL,
		cfg.RunID, node.URL, nullableString(parentURL),
	); err != nil {
		return err
	}

	for _, resource := range node.resources {
		if _, err := tx.Exec(
			gopherSqliteInsertResourceSQL,
			cfg.RunID, resource, resourceFileType(resource), node.URL,
		); err != nil {
			return err
		}
	}

	for _, e := range node.errors {
		if _, err := tx.Exec(
			gopherSqliteInsertErrorSQL,
			cfg.RunID, e.Error(), node.URL,
		); err != nil {
			return err
		}
	}

	for _, link := range node.links {
		if err := insertURLMapToSQLite(tx, link, node.URL, cfg); err != nil {
			return err
		}
	}

	return nil
}

// resourceFileType derives a file type from a resource URL's extension,
// falling back to "unknown" when there is none.
func resourceFileType(resource string) string {
	path := resource
	if u, err := url.Parse(resource); err == nil {
		path = u.Path
	}

	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	if ext == "" {
		return "unknown"
	}
	return strings.ToLower(ext)
}

// nullableString returns a NULL-able value for empty strings so optional
// columns store NULL rather than an empty string.
func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
