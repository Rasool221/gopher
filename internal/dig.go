package internal

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
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
// an error so gopher doesn't try to fetch them.
func ResolveHref(pageURL string, hrefValue string) (string, error) {
	slog.Debug("Resolving hrefValue on page", "hrefValue", hrefValue, "pageURL", pageURL)

	// First we need to check if the hrefValue is empty.
	if hrefValue == "" {
		return "", fmt.Errorf("empty hrefValue for pageUrl %q", pageURL)
	}

	// Parse the page URL; this is the base every relative href resolves against.
	base, err := url.Parse(pageURL)
	if err != nil {
		return "", fmt.Errorf("invalid pageUrl %q: %w", pageURL, err)
	}

	ref, err := url.Parse(hrefValue)
	if err != nil {
		return "", fmt.Errorf("invalid hrefValue %q for pageUrl %q: %w", hrefValue, pageURL, err)
	}

	// A schemeless ref with no authority (e.g. "notes.io/post", "example.com") is, per RFC 3986,
	// a relative path so it would resolve against the current host. But the author may have meant
	// an external site. We disambiguate on the first path segment: if its trailing label is a real
	// ICANN-registered TLD (.com, .io, ...) we treat it as an external host by re-parsing it as a
	// protocol-relative URL so it inherits the page's scheme. If the label is a file extension
	// (.html, .png, .txt) it's not a public suffix, so we leave it as a relative path.
	if ref.Scheme == "" && ref.Host == "" && refLooksLikeExternalHost(ref.Path) {
		slog.Debug("Schemless, authority-less href looks like an external host; treating as protocol-relative URL", "hrefValue", hrefValue)
		ref, err = url.Parse("//" + hrefValue)
		if err != nil {
			return "", fmt.Errorf("invalid hrefValue %q for pageUrl %q: %w", hrefValue, pageURL, err)
		}
	}

	// Now if the hrefValue is a relative URL, we can resolve it against the pageURL.
	// If it's an absolute URL to the same host we can just return it.
	// Otherwise if it's an absolute URL to a different host, we should validate it and return it if it's valid.
	// ResolveReference handles all of these shapes per RFC 3986, inheriting the base's
	// scheme/host for relative and protocol-relative refs and returning absolute refs unchanged.
	resolved := base.ResolveReference(ref)

	slog.Debug("Resolved hrefValue to URL", "hrefValue", hrefValue, "resolvedURL", resolved.String(), "scheme", resolved.Scheme, "host", resolved.Host)

	// Reject anything that isn't http(s) (e.g. mailto:, tel:, javascript:) so gopher doesn't fetch it.
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return "", fmt.Errorf("unsupported scheme %q in hrefValue %q for pageUrl %q", resolved.Scheme, hrefValue, pageURL)
	}

	return resolved.String(), nil
}

// refLooksLikeExternalHost reports whether a schemeless, authority-less ref path looks like it was
// meant to be an external host rather than a relative path. It inspects only the first path segment
// (so "notes.io/post" is judged on "notes.io") and treats it as a host when its trailing label is an
// ICANN-registered public suffix e.g. "example.com" / "notes.io" yes, "about.html" / "logo.txt" no.
// Labels in fileExtensionTLDs (e.g. .md, .sh) are kept relative even though they're valid TLDs.
func refLooksLikeExternalHost(path string) bool {
	if path == "" || strings.HasPrefix(path, "/") {
		return false
	}

	slog.Debug("Checking if ref path looks like external host", "path", path)

	// Only the first segment can be the host; everything after the first "/" is a path on it.
	segment := path
	if i := strings.IndexByte(segment, '/'); i >= 0 {
		slog.Debug("Ref path contains slash; treating everything after first slash as path", "segment", segment, "path", path)
		segment = segment[:i]
	}

	// Guard against relative markers like "." / ".." that contain dots but aren't hosts.
	if segment == "" || strings.HasPrefix(segment, ".") || strings.Contains(segment, "..") {
		slog.Debug("Ref path segment is empty or starts with dot or contains dots, treating as relative path", "segment", segment)
		return false
	}
	if !strings.Contains(segment, ".") {
		slog.Debug("Ref path segment contains no dots, treating as relative path", "segment", segment)
		return false
	}

	// icann==true means the suffix is a real registered TLD (not a file extension); suffix != segment
	// ensures there's an actual host label in front of the TLD (rules out a bare "com").
	suffix, icann := publicsuffix.PublicSuffix(segment)
	slog.Debug("Extracted public suffix from ref path segment", "segment", segment, "suffix", suffix, "icann", icann)
	if !icann || suffix == segment {
		slog.Debug("Ref path segment does not have an ICANN-registered public suffix or is just a suffix with no host label, treating as relative path", "segment", segment, "suffix", suffix, "icann", icann)
		return false
	}

	// This part is tricky, our fileExtensionTLDs contains both common file extensions that can also be TLDs
	// In that case, we just treat them as file extensions.
	isSuffixTld := fileExtensionTLDs[suffix]
	if isSuffixTld {
		slog.Debug("Ref path segment has a public suffix that is likely a file extension, treating as relative path", "segment", segment, "suffix", suffix)
		return false
	} else {
		slog.Debug("Ref path segment has a public suffix that is a real TLD, treating as external host", "segment", segment, "suffix", suffix)
		return true
	}
}

// GetPageContent makes an HTTP request to the given URL and returns the HTML content as a string.
// Note that GetPageContent expects the url to be a valid URL that is reachable.
// It also handles any errors that may occur during the request, which ultimately is returned.
func GetPageContent(url string) (string, error) {
	slog.Debug("Fetching page content for URL", "url", url)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	slog.Debug("Received HTTP response for URL", "url", url, "statusCode", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch page content: %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

// Gopher crawls a web server starting from a seed URL, building a tree (URLMap) of the pages and
// links it finds. It carries the run's Config plus the set of already-visited URLs, so the recursive
// crawl can honor limits and dedupe cycles without threading that state through every call.
type Gopher struct {
	cfg     Config
	visited map[string]struct{}
}

// NewGopher returns a Gopher ready to crawl with the given config. The visited set is initialized
// here, so callers never deal with a nil map.
func NewGopher(cfg Config) *Gopher {
	return &Gopher{
		cfg:     cfg,
		visited: make(map[string]struct{}),
	}
}

// Run starts a crawl from the given URL while incrementally outputting the results based
// on the Gopher configuration.
func (g *Gopher) Run(url string) URLMap {
	slog.Debug("Building URL map for URL", "url", url, "visitedCount", len(g.visited))

	// First, let's avoid infinite loops by checking if we've already visited this URL. If we have, we return an empty URLMap.
	if _, ok := g.visited[url]; ok {
		slog.Debug("Already visited URL, skipping to avoid cycle", "url", url)
		return URLMap{}
	}

	// Mark the current URL as visited.
	g.visited[url] = struct{}{}

	// Fetch the page content for the given URL.
	pageContent, err := GetPageContent(url)
	if err != nil {
		slog.Error("Error fetching page content", "url", url, "error", err)
		return URLMap{}
	}

	urlMap := URLMap{
		URL:       url,
		links:     []URLMap{},
		resources: []string{},
		errors:    []error{},
	}

	// Extract data from the page content. We pass the page URL itself (not just
	// the scheme+host) so that document-relative hrefs like "widget.html" resolve
	// against the directory the page lives in.
	// If any errors occur during extraction, we just keep moving, those errors are outputted.
	result := ExtractDataFromHTML(url, pageContent)

	currentBaseDomain, err := GetBaseDomain(url)
	if err != nil {
		slog.Error("Error extracting base domain from URL", "url", url, "error", err)
		return urlMap
	}

	// Add resources and errors to urlMap before we recurse exploring links
	urlMap.errors = append(result.Errors)
	urlMap.resources = append(result.Resources)

	// Create a URLMap for the current URL and recursively build URLMaps for each link found.
	// Here we will also honor the cfg.External setting to decide whether to include external links (links to different domains) in the crawl.
	for _, link := range result.Links {
		childBaseDomain, err := GetBaseDomain(link)
		if err != nil {
			slog.Error("Error extracting base domain from link URL", "linkURL", link, "error", err)
			continue
		}

		// Honoring the cfg.External setting.
		if !g.cfg.External && childBaseDomain != currentBaseDomain {
			slog.Debug("Skipping external link due to configuration", "linkURL", link, "childBaseDomain", childBaseDomain, "currentBaseDomain", currentBaseDomain)
			continue
		}

		// Checking if the link is valid.
		err = ValidateURL(link)
		if err != nil {
			slog.Error("Error validating link URL", "linkURL", link, "error", err)
			continue
		}

		// Recursively process children links.
		childURLMap := g.Run(link)
		if childURLMap.URL != "" {
			urlMap.links = append(urlMap.links, childURLMap)
		}
	}

	return urlMap
}
