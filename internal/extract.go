package internal

import (
	"io"
	"log/slog"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// Resource is a file (image, video, audio, downloadable file, etc.) referenced by a page.
// URL is the fully-resolved absolute URL of the resource. Path is the chain of links from
// the base host down to the resource itself (the resource URL is the final element), which
// lets consumers render the resource as part of the site's directory tree. Resources on a
// different base domain than the page have a nil Path and are included as-is.
type Resource struct {
	URL  string
	Path []string
}

type ExtractionResult struct {
	Links     []string   // Links found in the HTML content
	Resources []Resource // Resources found in the HTML content
	Errors    []error    // Errors encountered while processing the HTML content

	LinksMap     map[string]struct{} // Map of resolved URLs we've seen, used to dedupe within a single page.
	ResourcesMap map[string]struct{} // Map of resolved resources we've seen, used to dedupe within a single page.
}

// addLink adds a link to the ExtractionResult, ensuring no duplicates are added.
func (er *ExtractionResult) addLink(link string) {
	if er.LinksMap == nil {
		er.LinksMap = make(map[string]struct{})
	}

	if _, exists := er.LinksMap[link]; !exists {
		er.Links = append(er.Links, link)
		er.LinksMap[link] = struct{}{}
	}
}

// addResource adds a resource to the ExtractionResult, ensuring no duplicates are added.
// The resource's path chain is built from its resolved URL against the page the resource
// was found on; external resources get a nil Path.
func (er *ExtractionResult) addResource(pageURL string, resource string) {
	if er.ResourcesMap == nil {
		er.ResourcesMap = make(map[string]struct{})
	}

	if _, exists := er.ResourcesMap[resource]; !exists {
		er.Resources = append(er.Resources, Resource{
			URL:  resource,
			Path: resourcePath(pageURL, resource),
		})
		er.ResourcesMap[resource] = struct{}{}
	}
}

// resourcePath returns the chain of links from the base host down to a resource, e.g. for
// http://example.net/assets/logo.png:
//
//	[http://example.net/, http://example.net/assets/, http://example.net/assets/logo.png]
//
// It returns nil when the resource lives on a different base domain than the page, since
// external resources are kept as-is rather than expanded into the page's directory tree.
func resourcePath(pageURL string, resourceURL string) []string {
	pageBase, err := GetBaseDomain(pageURL)
	if err != nil {
		return nil
	}
	resourceBase, err := GetBaseDomain(resourceURL)
	if err != nil {
		return nil
	}
	if pageBase != resourceBase {
		return nil
	}

	u, err := url.Parse(resourceURL)
	if err != nil {
		return nil
	}

	base := u.Scheme + "://" + u.Host

	// A resource sitting at the host root has no parent directories.
	if u.Path == "" || u.Path == "/" {
		return []string{base + "/"}
	}

	path := []string{base + "/"}
	dir := ""
	segments := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	for i := 0; i < len(segments)-1; i++ {
		dir += "/" + segments[i]
		path = append(path, base+dir+"/")
	}

	return append(path, resourceURL)
}

// ExtractDataFromHTML extracts links & resources from HTML content using the html tokenizer.
// We iterate through every HTML token and through its attributes, looking for the following keys:
// "href" - for links
// "src" - for resources
// "poster" - for resources (<video /> thumbnail image)
func ExtractDataFromHTML(pageURL string, htmlContent string) ExtractionResult {
	slog.Debug("Extracting links from HTML content", "pageURL", pageURL)

	// Errors encountered while resolving individual hrefs. Eventually surfaced
	// up to the user if gopher is executed with verbose mode.
	var extractionResult ExtractionResult

	tokenizer := html.NewTokenizer(strings.NewReader(htmlContent))

	for {
		tokenType := tokenizer.Next()

		// ErrorToken represents the EOF or some error during tokenization.
		// If we encounter an EOF, we break the loop and return the links we've found so far.
		// If we encounter any other error, we return an empty list of links plus the error.
		if tokenType == html.ErrorToken {
			if tokenizer.Err() == io.EOF {
				slog.Debug("Finished tokenizing HTML content for page", "pageURL", pageURL)
				break
			}

			return ExtractionResult{Errors: []error{tokenizer.Err()}}
		}

		// Iterate through tokens, then iterate through that token's attributes, looking for "href" keys.
		// If we find one, resolve it against the page URL and stash the result.
		token := tokenizer.Token()
		if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken {
			for _, attr := range token.Attr {
				if _, exists := supportedHTMLKeys[attr.Key]; !exists {
					continue
				}

				slog.Debug("Found attribute in HTML token", "hrefValue", attr.Val, "token", token.Data, "pageURL", pageURL)

				resolved, err := ResolveHref(pageURL, attr.Val)
				if err != nil {
					extractionResult.Errors = append(extractionResult.Errors, err)
					continue
				}

				switch attr.Key {
				case "href":
					// Sometimes resources will be in the href key, so we need to check whether the file extension from url.
					if err := ValidateURL(attr.Val); err == nil {
						extractionResult.addLink(resolved)
					} else if IsLikelyFile(attr.Val) {
						extractionResult.addResource(pageURL, resolved)
					} else {
						// Not a standalone URL and not a file: a relative page link
						// (e.g. "/about.html", "../about.html"). Record the resolved link.
						extractionResult.addLink(resolved)
					}
				case "src", "poster":
					extractionResult.addResource(pageURL, resolved)
				}
			}
		}
	}
	slog.Debug("Extracted links from HTML content", "pageURL", pageURL, "len(extractionResult.Links)", len(extractionResult.Links), "len(extractionResult.Errors)", len(extractionResult.Errors))

	return extractionResult
}
