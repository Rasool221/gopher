// Package internal contains core functionality for Gopher.
package internal

import "github.com/alecthomas/kong"

type CLIConfig struct {
	URL       string `arg:"" help:"The url to begin digging."`
	LogLevel  string `short:"l" default:"info" help:"Log level (error=0, 1=info, 2=debug)."`
	External  bool   `short:"e" default:"false" help:"Whether to traverse include external links (links to different domains)."`
	Output    string `short:"o" default:"stdout" help:"Output file to write the URL map to (stdout, sqlite, json, html)."`
	ProxyFile string `short:"p" placeholder:"FILE" default:"" help:"File with one proxy URL per line, requests are spread across them. See docs for more details.`
	DBPath    string `short:"d" default:"" help:"Absolute path to where output.db will be written to (defaults to cwd)."`
}

func ParseClI() CLIConfig {
	var cfg CLIConfig
	kong.Parse(&cfg)
	return cfg
}
