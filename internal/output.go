package internal

import (
	"fmt"
	"strings"
)

// PrintURLMap recursively prints the URL map to stdout in a readable format, with indentation to show
// the hierarchy of links. This is program output (the crawl result), not a log, so it goes straight to
// stdout and is unaffected by the configured log level. The indentLevel parameter controls the
// indentation for nested links. Printing is one of the output types specified in the CLIConfig, and is
// the default if no output type is specified.
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
