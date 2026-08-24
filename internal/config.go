package internal

import "github.com/google/uuid"

// Config is the internal, domain-level representation of a crawl's runtime settings.
type Config struct {
	Output    string // Where to send the result (0=stdout, 1=sqlite).
	External  bool   // Whether to include external links (links to different domains).
	URL       string // The URL entry point of the crawl
	DBPath    string // The path to the SQLite db for output (if option is selected).
	RunID     string // Unique identifier for this run of Gopher.
	ProxyFile string // Path to the file containing proxy URLs (if provided).
}

func NewConfig(cli CLIConfig) Config {
	return Config{
		Output:    cli.Output,
		External:  cli.External,
		URL:       cli.URL,
		DBPath:    cli.DBPath,
		ProxyFile: cli.ProxyFile,
		RunID:     uuid.New().String(),
	}
}
