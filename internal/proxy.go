package internal

import (
	"bufio"
	"errors"
	"io"
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
		if IsProxyValid(line) {
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

// MakeRequest makes an HTTP GET request to the specified pageUrl
// and returns the content. The next proxy in the queue will be used, and if in un-healthy,
// will retry up to proxyUnhealthyRetryCount times.
func (g *Gopher) MakeRequest(pageUrl string) (string, error) {
	if g.proxies == nil {
		return "", errors.New("no proxies are available")
	}

	if len(pageUrl) == 0 {
		return "", errors.New("pageUrl is required")
	}

	amountFailures := 0

	for amountFailures <= proxyUnhealthyRetryCount {
		nextProxy, err := g.proxies.GetNextProxy()
		if err != nil {
			return "", err
		}

		// TODO: add check to see if we should use a proxy or not
		resp, err := MakeRequestWithProxy(pageUrl, nextProxy)
		if err != nil {
			return "", err
		}

		// If we've made a successful request we can return the content
		if resp.StatusCode <= 299 && resp.StatusCode >= 200 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", err
			}

			return string(body), nil
		}

		slog.Debug("request failed with proxy", "proxy", nextProxy, "status", resp.Status, "retryCount", amountFailures)
	}

	return "", errors.New("exceeded retries for unhealthy proxies")
}

// MakeRequestWithProxy makes an HTTP GET request to the specified pageUrl
// using the provided proxy. It returns a pointer to the response and an error.
// It is a thin wrapper around the actual HTTP request for cleaner code.
func MakeRequestWithProxy(pageUrl string, proxy string) (*http.Response, error) {
	proxyUrl, err := url.Parse(proxy)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyUrl),
	}

	client := &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= redirectFollowLimit {
				return errors.New("stopped after maximum redirects")
			}
			return nil
		},
	}

	req, err := http.NewRequest("GET", pageUrl, nil)
	if err != nil {
		return nil, err
	}

	return client.Do(req)
}
