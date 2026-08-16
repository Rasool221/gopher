package internal

import "github.com/google/uuid"

// Config is the internal, domain-level representation of a crawl's runtime settings.
type Config struct {
	Workers  int    // Number of concurrent workers to use for requests.
	Output   string // Where to send the result (0=stdout, 1=sqlite).
	External bool   // Whether to include external links (links to different domains).
	URL      string // The URL entry point of the crawl
	DBPath   string // The path to the SQLite db for output (if option is selected).
	RunID    string // Unique identifier for this run of Gopher.
}

func NewConfig(cli CLIConfig) Config {
	return Config{
		Workers:  cli.Workers,
		Output:   cli.Output,
		External: cli.External,
		URL:      cli.URL,
		DBPath:   cli.DBPath,
		RunID:    uuid.New().String(),
	}
}
