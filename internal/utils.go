package internal

import (
	"errors"
	"fmt"
	"golang.org/x/net/publicsuffix"
	"log/slog"
	"net"
	"net/http"
	"net/url"
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
	err := ValidateServerReachable(cfg.Url)
	if err != nil {
		return err
	}

	// Validate LogLevel
	if cfg.LogLevel < 0 || cfg.LogLevel > 2 {
		return errors.New("Invalid log level: must be 0 (error), 1 (info), or 2 (debug)")
	}

	// Validate Depth
	if cfg.Depth < 0 {
		return errors.New("Invalid depth: must be 0 (unlimited) or a positive integer")
	}

	// Validate MaxUrls
	if cfg.MaxUrls < 0 {
		return errors.New("Invalid max URLs: must be 0 (unlimited) or a positive integer")
	}

	// Validate Output
	if cfg.Output < 0 || cfg.Output > 1 {
		return errors.New("Invalid output option: must be 0 (stdout) or 1 (sqlite)")
	}

	return nil
}

// ParseAndValidateURL checks if the provided URL is non-empty
// and is a valid HTTP or HTTPS URL. It returns an error if the URL is invalid.
// This function supports domain names, IP addresses, and localhost URLs.
// This function returns the exact URL, and an error if one occurs.
func ParseAndValidateURL(target string) (string, error) {
	if target == "" {
		return "", errors.New("url is required")
	}

	// Pull out the host. If the target carries a scheme, it must be http(s); if it's schemeless
	// (e.g. "example.com", "localhost:8080") we parse it as a network-path reference so net/url
	// populates Host. Either way we never mutate the input — the original target is returned.
	var host string
	if i := strings.Index(target, "://"); i >= 0 {
		scheme := target[:i]
		if scheme != "http" && scheme != "https" {
			return "", fmt.Errorf("invalid scheme %q: must be http or https", scheme)
		}

		parsed, err := url.Parse(target)
		if err != nil {
			return "", fmt.Errorf("invalid URL format: %w", err)
		}
		host = parsed.Hostname()
	} else {
		parsed, err := url.Parse("//" + target)
		if err != nil {
			return "", fmt.Errorf("invalid URL format: %w", err)
		}
		host = parsed.Hostname()
	}

	if host == "" {
		return "", fmt.Errorf("invalid URL format: host is required")
	}

	if !isValidHost(host) {
		return "", fmt.Errorf("invalid URL format: %q is not a valid host", host)
	}

	return target, nil
}

// isValidHost reports whether host is something we'd actually try to reach: localhost, a literal IP,
// or a domain whose trailing label is a real ICANN-registered public suffix. This is what rejects
// "example" (no TLD) and "exa&*$.com" (invalid characters) while accepting "example.com".
func isValidHost(host string) bool {
	if host == "localhost" {
		return true
	}
	if net.ParseIP(host) != nil {
		return true
	}
	if !isValidHostnameChars(host) {
		return false
	}

	// suffix != host ensures there's a registrable label in front of the TLD (rules out a bare "com").
	suffix, icann := publicsuffix.PublicSuffix(host)
	return icann && suffix != host
}

// isValidHostnameChars reports whether host contains only characters legal in a hostname
// (letters, digits, hyphens, and dots).
func isValidHostnameChars(host string) bool {
	for _, r := range host {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '.':
		default:
			return false
		}
	}
	return true
}

func ValidateServerReachable(url string) error {
	slog.Debug("checking url reachability", "url", url)

	if url == "" {
		return errors.New("url is required")
	}

	// Validate URL format
	_, err := ParseAndValidateURL(url)
	if err != nil {
		return err
	}

	// Sending an HTTP GET request to the URL to check if it's reachable.
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to reach the website: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("website returned non-OK status: %s", resp.Status)
	}

	slog.Debug("url is reachable", "url", url)
	return nil
}

// This is temp for tesing, I dont think it's worth keeping.
