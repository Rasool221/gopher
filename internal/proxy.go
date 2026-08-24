package internal

import (
	"bufio"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"os"
)

type ProxyQueue struct {
	proxies      []string            // Slice to hold the proxies in the queue
	proxiesSet   map[string]struct{} // To ensure uniqueness of proxies in the queue
	currentIndex int                 // Index of the next proxy to be dequeued, used for round-robin selection
}

// Enqueue adds a proxy to the queue which is
// a simple addition to the slice of proxies.
func (pq *ProxyQueue) enqueue(proxy string) {
	if pq.proxiesSet == nil {
		pq.proxiesSet = make(map[string]struct{})
	}

	if _, exists := pq.proxiesSet[proxy]; exists {
		return // Proxy already exists in the queue, do not add it again
	}

	pq.proxies = append(pq.proxies, proxy)
	pq.proxiesSet[proxy] = struct{}{}
}

// IsProxyAlive checks if the provided proxy is alive and reachable.
// This function attempts to make a simple HTTP GET request through the proxy to a known URL.
func IsProxyAlive(proxy string) (bool, error) {
	if !IsProxyValid(proxy) {
		return false, nil
	}

	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return false, err
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	client := &http.Client{
		Transport: transport,
	}

	var amtSuccessfulChecks int

	for _, healthcheckURL := range proxyHealthcheckURLs {
		resp, err := client.Get(healthcheckURL)
		if err != nil {
			return false, err
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
			amtSuccessfulChecks++
		}
	}

	if amtSuccessfulChecks == 0 {
		return false, errors.New("proxy is not alive, failed all health checks")
	}

	// If any of the health checks passed, we consider the proxy alive.
	return true, nil
}

// IsProxyValid checks if the provided proxy string is a valid URL and supported by the HTTP client.
// We check if the proxy is http, https, or socks5 by parsing scheme.
func IsProxyValid(proxy string) bool {
	if len(proxy) == 0 {
		return false
	}

	parsedURL, err := url.Parse(proxy)
	if err != nil {
		return false
	}

	_, supported := supportedProxySchemes[parsedURL.Scheme]
	return supported
}

// CreateProxyQueue initializes the proxy queue if a file path is provided in the config.
// First, it reads the proxies file if one is provided, then parse and ensure
// the proxies are valid and supported.
// Then, it creates a new ProxyQueue and populates it with the valid proxies before returning it.
func CreateProxyQueue(cfg Config) (*ProxyQueue, error) {
	var pq ProxyQueue

	slog.Debug("creating proxy queue", "config", cfg)

	// Opening and reading the pfoxy file
	file, err := os.Open(cfg.ProxyFile)
	if err != nil {
		return nil, errors.New("failed to open proxy file: " + err.Error())
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	slog.Debug("reading proxies from file", "file", cfg.ProxyFile)

	// Itearting line by line
	for scanner.Scan() {
		line := scanner.Text()

		// Check if the proxy is valid and supported before adding it to the queue
		if ok, err := IsProxyAlive(line); !ok || err != nil {
			pq.enqueue(line)
		}
	}

	return &pq, nil
}

// GetNextProxy gets the next proxy from the queue and moves the iterator
// left to right while wrapping around to beggining, to simulate round-robin selection of proxies.
func (pq *ProxyQueue) GetNextProxy() (string, error) {
	if len(pq.proxies) == 0 {
		return "", errors.New("no proxies available in the queue")
	}

	proxy := pq.proxies[pq.currentIndex]

	// Move the index to the next proxy, wrapping around if necessary
	pq.currentIndex += 1
	if pq.currentIndex >= len(pq.proxies) {
		pq.currentIndex = 0
	}

	return proxy, nil
}
