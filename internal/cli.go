// Package internal contains core functionality for Gopher.
package internal

import "github.com/alecthomas/kong"

type CLIConfig struct {
	URL      string `arg:"" help:"The url to begin digging."`
	Workers  int    `short:"w" default:"1" help:"Number of workers to use for concurrent requests."`
	LogLevel string `short:"l" default:"info" help:"Log level (error, 1=info, 2=debug)."`
	External bool   `short:"e" default:"false" help:"Whether to traverse include external links (links to different domains)."`
	Output   string `short:"o" default:"stdout" help:"Output file to write the URL map to (stdout, sqlite, json, html)."`
	Proxies  string `short:"p" default:"" help:"Comma-separated list of proxy URLs to use for requests (e.g. http://proxy1:port,http://proxy2:port)."`
	DBPath   string `short:"db" default:"" help:"Absolute path to where output.db will be written to (defaults to cwd)."`
}

func ParseClI() CLIConfig {
	var cfg CLIConfig
	kong.Parse(&cfg)
	return cfg
}
