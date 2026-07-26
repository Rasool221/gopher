package internal

import (
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

type ExtractionResult struct {
	Links     []string // Links found in the HTML content
	Resources []string // Resources found in the HTML content
	Errors    []error  // Errors encountered while processing the HTML content

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
func (er *ExtractionResult) addResource(resource string) {
	if er.ResourcesMap == nil {
		er.ResourcesMap = make(map[string]struct{})
	}

	if _, exists := er.ResourcesMap[resource]; !exists {
		er.Resources = append(er.Resources, resource)
		er.ResourcesMap[resource] = struct{}{}
	}
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
					// Sometimes resources will be in the hef key, so we need to check whether the file extension from url.
					resolvedFilePath := filepath.Ext(resolved)
					if resolvedFilePath != "" && !strings.HasPrefix(resolvedFilePath, ".html") {
						extractionResult.addResource(resolved)
					} else {
						extractionResult.addLink(resolved)
					}
				case "src", "poster":
					extractionResult.addResource(resolved)
				}
			}
		}
	}
	slog.Debug("Extracted links from HTML content", "pageURL", pageURL, "len(extractionResult.Links)", len(extractionResult.Links), "len(extractionResult.Errors)", len(extractionResult.Errors))

	return extractionResult
}
