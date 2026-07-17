package internal

// Config is the internal, domain-level representation of a crawl's runtime settings.
type Config struct {
	Workers  int    // Number of concurrent workers to use for requests.
	Output   string // Where to send the result (0=stdout, 1=sqlite).
	External bool   // Whether to include external links (links to different domains).
	Url      string // The URL entry point of the crawl
}

func NewConfig(cli CLIConfig) Config {
	return Config{
		Workers:  cli.Workers,
		Output:   cli.Output,
		External: cli.External,
		Url:      cli.Url,
	}
}
