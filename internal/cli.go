package internal

import "github.com/alecthomas/kong"

type CLIConfig struct {
	Url      string `arg:"" help:"The url to begin digging."`
	Workers  int    `short:"w" default:"1" help:"Number of workers to use for concurrent requests."` // TODO
	LogLevel int    `short:"l" default:"1" help:"Log level (0=error, 1=info, 2=debug)."`
	External bool   `short:"e" default:"false" help:"Whether to traverse include external links (links to different domains)."`
	Output   int    `short:"o" default:"0" help:"Output file to write the URL map to (0=stdout, 1=sqlite)."`                                           // TODO: SQLite & HTML page output of a graph-based URL map
	Proxies  string `short:"p" default:"" help:"Comma-separated list of proxy URLs to use for requests (e.g. http://proxy1:port,http://proxy2:port)."` // TODO
}

func ParseClI() CLIConfig {
	var cfg CLIConfig
	kong.Parse(&cfg)
	return cfg
}
