package internal

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func ValidateCLI(cfg CLIConfig) error {
	if cfg.Url == "" {
		return errors.New("URL is required")
	}

	// Validate URL format
	if !strings.HasPrefix(cfg.Url, "http://") && !strings.HasPrefix(cfg.Url, "https://") {
		return errors.New("Malformed URL: URL must start with http:// or https://")
	}

	// Validate website is reachable
	err := ValidateUrl(cfg.Url)
	if err != nil {
		return err
	}

	// TODO: Add gocurrency (lol) at some point via the workers param
	// and ensure CPU can handle the amt of workers

	return nil
}

func ValidateUrl(url string) error {
	if url == "" {
		return errors.New("URL is required")
	}

	// Validate URL format
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return errors.New("Malformed URL: URL must start with http:// or https://")
	}

	resp, err := http.Get(url)
	if err != nil {
		return errors.New("Unable to reach the website: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Website returned non-200 status code: " + resp.Status)
	}

	return nil
}

// This is temp for tesing, I dont think it's worth keeping.
func PrintURLMap(urlMap URLMap, indentLevel int) {
	indent := strings.Repeat("  ", indentLevel)
	fmt.Printf("%s- %s\n", indent, urlMap.URL)

	for _, resource := range urlMap.resources {
		fmt.Printf("%s  * Resource: %s\n", indent, resource)
	}

	for _, link := range urlMap.links {
		PrintURLMap(link, indentLevel+1)
	}
}
