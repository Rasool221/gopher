package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type URLMap struct {
	URL       string
	links     []URLMap // Links to other URLs found on the page
	resources []string // Resources (like images, scripts) found on the page
	errors    []error  // Errors encountered while processing the page
}

// ResolveHref takes the URL of the page we're currently scanning and the href value
// of a link found on that page, and returns the fully-resolved absolute URL.
// It handles all the shapes an href can take per RFC 3986: absolute URLs
// ("http://example.com/x"), root-relative ("/about"), document-relative
// ("widget.html"), parent ("../up.html"), protocol-relative ("//host/x"),
// and query/fragment-only ("?q=1", "#anchor").
// Non-HTTP(S) schemes like mailto:, tel:, and javascript: are rejected with
// an error so the crawler doesn't try to fetch them.
func ResolveHref(pageURL string, hrefValue string) (string, error) {
	// Parse the page URL first; if it isn't a valid URL we can't resolve anything against it.
	base, err := url.Parse(pageURL)
	if err != nil {
		return "", fmt.Errorf("invalid page URL %q: %w", pageURL, err)
	}

	// Parse the href. url.Parse is permissive — it treats arbitrary strings
	// as path-only references — so the real validation happens via the scheme
	// check on the resolved URL below.
	ref, err := url.Parse(hrefValue)
	if err != nil {
		return "", fmt.Errorf("invalid href %q: %w", hrefValue, err)
	}

	// ResolveReference inherits the base's scheme/host when the href doesn't
	// supply them, so we check scheme on the result rather than on the raw href.
	// This way "mailto:..." (scheme = "mailto") gets rejected, but "/about"
	// (no scheme, inherits "http" from the base) gets through.
	resolved := base.ResolveReference(ref)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return "", fmt.Errorf("unsupported scheme %q in href %q", resolved.Scheme, hrefValue)
	}

	return resolved.String(), nil
}

// GetPageContent makes an HTTP request to the given URL and returns the HTML content as a string.
// Note that GetPageContent expects the url to be a valid URL that is reachable.
// It also handles any errors that may occur during the request, which ultimately is returned.
func GetPageContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch page content: %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

// ExtractLinksFromHTML extracts links from HTML content using the html tokenizer.
// We iterate through every HTML token and through its attributes, looking for "href" keys.
// Each href is resolved against pageURL (the URL the HTML was fetched from) so that
// relative hrefs become full absolute URLs. Successfully-resolved URLs are deduped
// via a map and returned as the first slice; any per-href resolution errors (e.g.
// unsupported scheme like mailto:) are collected and returned as the second slice.
// The two slices are NOT parallel-indexed — they're independent collections.
func ExtractLinksFromHTML(pageURL string, htmlContent string) ([]string, []error) {
	// Map of resolved URLs we've seen, used to dedupe within a single page.
	linksMap := make(map[string]struct{})

	// Errors encountered while resolving individual hrefs. Eventually surfaced
	// up to the user if gopher is executed with verbose mode.
	var parseErrors []error

	tokenizer := html.NewTokenizer(strings.NewReader(htmlContent))

	for {
		tokenType := tokenizer.Next()

		// ErrorToken represents the EOF or some error during tokenization.
		// If we encounter an EOF, we break the loop and return the links we've found so far.
		// If we encounter any other error, we return an empty list of links plus the error.
		if tokenType == html.ErrorToken {
			if tokenizer.Err() == io.EOF {
				break
			}

			return []string{}, []error{tokenizer.Err()}
		}

		// Iterate through tokens, then iterate through that token's attributes, looking for "href" keys.
		// If we find one, resolve it against the page URL and stash the result.
		token := tokenizer.Token()
		if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken {
			for _, attr := range token.Attr {
				if attr.Key != "href" {
					continue
				}

				resolved, err := ResolveHref(pageURL, attr.Val)
				if err != nil {
					parseErrors = append(parseErrors, err)
					continue
				}

				linksMap[resolved] = struct{}{}
			}
		}
	}

	// Transform the map of links into a list of links to return.
	links := make([]string, 0, len(linksMap))
	for link := range linksMap {
		links = append(links, link)
	}

	return links, parseErrors
}

// buildUrlMap takes a URL as input, fetches page content, and recursively
// scans other URLs found to build a map of the webserver.
// It returns a URLMap struct containing the URL, its links, and resources.
// We keep track of visited URLs to avoid infinite loops and redundant processing.
func BuildUrlMap(url string, visited map[string]struct{}) URLMap {
	// When the buildUrlMap is first invoked, visited will be nil, so we need to initialize it as an empty map
	// for the first invocation. Subsequent calls will recieve a populated visited map, so we won't reinitialize it.
	if visited == nil {
		visited = make(map[string]struct{})
	}

	// First, let's avoid infinite loops by checking if we've already visited this URL. If we have, we return an empty URLMap.
	if _, ok := visited[url]; ok {
		return URLMap{}
	}

	// Mark the current URL as visited.
	visited[url] = struct{}{}

	// Fetch the page content for the given URL.
	pageContent, err := GetPageContent(url)
	if err != nil {
		fmt.Printf("Error fetching page content for %s: %v\n", url, err)
		return URLMap{URL: url}
	}

	// Extract links from the page content. We pass the page URL itself (not just
	// the scheme+host) so that document-relative hrefs like "widget.html" resolve
	// against the directory the page lives in.
	links, errors := ExtractLinksFromHTML(url, pageContent)

	urlMap := URLMap{
		URL:    url,
		links:  []URLMap{},
		errors: errors,
	}

	// Create a URLMap for the current URL and recursively build URLMaps for each link found.
	for _, link := range links {
		childUrlMap := BuildUrlMap(link, visited)
		if childUrlMap.URL != "" {
			urlMap.links = append(urlMap.links, childUrlMap)
		}
	}

	return urlMap
}
