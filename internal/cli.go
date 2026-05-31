package internal

import "github.com/alecthomas/kong"

type CLIConfig struct {
	Url      string `arg:"" help:"The url to begin digging."`
	Workers  int    `short:"w" default:"1" help:"Number of workers to use for concurrent requests."`              // TODO
	LogLevel int    `short:"l" default:"1" help:"Log level (0=error, 1=info, 2=debug)."`                          // TODO: implement logging with levels
	Depth    int    `short:"d" default:"0" help:"Maximum depth to crawl links from beginning (0 for unlimited)."` // TODO
	MaxUrls  int    `short:"m" default:"0" help:"Maximum number of URLs to crawl (0 for unlimited)."`             // TODO
	Output   int    `short:"o" default:"" help:"Output file to write the URL map to (0=stdout, 1=sqlite)."`       // TODO
}

func ParseClI() CLIConfig {
	var cfg CLIConfig
	kong.Parse(&cfg)
	return cfg
}
