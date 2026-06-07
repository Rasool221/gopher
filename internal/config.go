package internal

// Config is the internal, domain-level representation of a crawl's runtime settings. It's
// deliberately decoupled from CLIConfig (cli.go), which is the command-line surface defined with
// kong tags. The engine (Gopher) consumes Config only, so the rest of the package never depends on
// how settings arrive — flags today, a config file or API request tomorrow. NewConfig is the single
// place that maps the CLI layer onto the domain layer.
type Config struct {
	Workers  int  // Number of concurrent workers to use for requests.
	Output   int  // Where to send the result (0=stdout, 1=sqlite).
	External bool // Whether to include external links (links to different domains).
}

// NewConfig maps the CLI-facing CLIConfig onto the internal Config the engine uses. LogLevel is
// intentionally absent: it configures logging at startup (main.go), which is a bootstrap concern,
// not part of a crawl.
func NewConfig(cli CLIConfig) Config {
	return Config{
		Workers:  cli.Workers,
		Output:   cli.Output,
		External: cli.External,
	}
}
