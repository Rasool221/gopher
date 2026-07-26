package internal

import (
	"sort"
	"testing"
)

func TestExtractDataFromHTML(t *testing.T) {
	tests := []struct {
		name              string
		pageURL           string
		content           string
		expectedLinks     []string
		expectedResources []string
		expectedErrCount  int
	}{
		{
			name:              "absolute http URL",
			pageURL:           "http://test.local/",
			content:           `<html><body><a href="http://example.com">Example</a></body></html>`,
			expectedLinks:     []string{"http://example.com"},
			expectedResources: nil,
			expectedErrCount:  0,
		},
		{
			name:    "multiple links with schemeless and IP",
			pageURL: "http://test.local/",
			// "google.com" promoted to external host; "104.20.23.154" stays relative.
			content:           `<html><body><a href="http://example.com">Example</a><div href="google.com">Google</div><img href="104.20.23.154">Example by IP address</img></body></html>`,
			expectedLinks:     []string{"http://example.com", "http://google.com", "http://test.local/104.20.23.154"},
			expectedResources: nil,
			expectedErrCount:  0,
		},
		{
			name:              "root-relative href",
			pageURL:           "http://test.local/products/",
			content:           `<html><a href="/about.html">About</a></html>`,
			expectedLinks:     []string{"http://test.local/about.html"},
			expectedResources: nil,
			expectedErrCount:  0,
		},
		{
			name:              "document-relative href",
			pageURL:           "http://test.local/products/",
			content:           `<html><a href="widget.html">Widget</a></html>`,
			expectedLinks:     []string{"http://test.local/products/widget.html"},
			expectedResources: nil,
			expectedErrCount:  0,
		},
		{
			name:              "parent-directory href",
			pageURL:           "http://test.local/products/",
			content:           `<html><a href="../about.html">Up one</a></html>`,
			expectedLinks:     []string{"http://test.local/about.html"},
			expectedResources: nil,
			expectedErrCount:  0,
		},
		{
			name:              "non-HTTP schemes produce errors not links",
			pageURL:           "http://test.local/",
			content:           `<html><a href="mailto:hello@example.com">mail</a><a href="javascript:void(0)">js</a></html>`,
			expectedLinks:     nil,
			expectedResources: nil,
			expectedErrCount:  2,
		},
		{
			name:              "empty HTML",
			pageURL:           "http://test.local/",
			content:           `<html></html>`,
			expectedLinks:     nil,
			expectedResources: nil,
			expectedErrCount:  0,
		},
		{
			name:              "malformed HTML",
			pageURL:           "http://test.local/",
			content:           `<html><not-a-real-tag>`,
			expectedLinks:     nil,
			expectedResources: nil,
			expectedErrCount:  0,
		},

		// --- Resource tests ---

		{
			name:              "img src is a resource",
			pageURL:           "http://test.local/",
			content:           `<html><body><img src="/images/photo.jpg"></body></html>`,
			expectedLinks:     nil,
			expectedResources: []string{"http://test.local/images/photo.jpg"},
			expectedErrCount:  0,
		},
		{
			name:              "video src and poster are resources",
			pageURL:           "http://test.local/",
			content:           `<html><body><video src="/media/clip.mp4" poster="/media/poster.jpg"></video></body></html>`,
			expectedLinks:     nil,
			expectedResources: []string{"http://test.local/media/clip.mp4", "http://test.local/media/poster.jpg"},
			expectedErrCount:  0,
		},
		{
			name:              "gif via img src",
			pageURL:           "http://test.local/",
			content:           `<html><body><img src="/images/animation.gif"></body></html>`,
			expectedLinks:     nil,
			expectedResources: []string{"http://test.local/images/animation.gif"},
			expectedErrCount:  0,
		},
		{
			name:              "file download link via href",
			pageURL:           "http://test.local/",
			content:           `<html><body><a href="/files/report.pdf">Download PDF</a></body></html>`,
			expectedLinks:     nil,
			expectedResources: []string{"http://test.local/files/report.pdf"},
			expectedErrCount:  0,
		},
		{
			name:              "mixed links, resources, and errors",
			pageURL:           "http://test.local/",
			content:           `<html><body><a href="/about.html">About</a><img src="/img/logo.png"><a href="mailto:hi@example.com">Mail</a><video src="/vid/intro.webm" poster="/vid/thumb.png"></video></body></html>`,
			expectedLinks:     []string{"http://test.local/about.html"},
			expectedResources: []string{"http://test.local/img/logo.png", "http://test.local/vid/intro.webm", "http://test.local/vid/thumb.png"},
			expectedErrCount:  1,
		},
		{
			name:              "deduplication of resources",
			pageURL:           "http://test.local/",
			content:           `<html><body><img src="/img/logo.png"><img src="/img/logo.png"></body></html>`,
			expectedLinks:     nil,
			expectedResources: []string{"http://test.local/img/logo.png"},
			expectedErrCount:  0,
		},
		{
			name:              "deduplication of links",
			pageURL:           "http://test.local/",
			content:           `<html><body><a href="/about.html">A</a><a href="/about.html">B</a></body></html>`,
			expectedLinks:     []string{"http://test.local/about.html"},
			expectedResources: nil,
			expectedErrCount:  0,
		},
		{
			name:              "absolute resource URL passes through",
			pageURL:           "http://test.local/",
			content:           `<html><body><img src="https://cdn.example.com/pic.jpg"></body></html>`,
			expectedLinks:     nil,
			expectedResources: []string{"https://cdn.example.com/pic.jpg"},
			expectedErrCount:  0,
		},
		{
			name:              "source element inside video",
			pageURL:           "http://test.local/",
			content:           `<html><body><video><source src="/media/trailer.mkv" type="video/x-matroska"></video></body></html>`,
			expectedLinks:     nil,
			expectedResources: []string{"http://test.local/media/trailer.mkv"},
			expectedErrCount:  0,
		},
		{
			name:              "audio src is a resource",
			pageURL:           "http://test.local/",
			content:           `<html><body><audio src="/audio/song.mp3"></audio></body></html>`,
			expectedLinks:     nil,
			expectedResources: []string{"http://test.local/audio/song.mp3"},
			expectedErrCount:  0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ExtractDataFromHTML(test.pageURL, test.content)

			// Check error count.
			if len(result.Errors) != test.expectedErrCount {
				t.Errorf("expected %d errors, got %d: %v", test.expectedErrCount, len(result.Errors), result.Errors)
			}

			// Check links (order-independent).
			gotLinks := append([]string(nil), result.Links...)
			wantLinks := append([]string(nil), test.expectedLinks...)
			sort.Strings(gotLinks)
			sort.Strings(wantLinks)

			if len(gotLinks) != len(wantLinks) {
				t.Errorf("expected %d links: %v, got %d: %v", len(wantLinks), wantLinks, len(gotLinks), gotLinks)
			} else {
				for i, link := range gotLinks {
					if link != wantLinks[i] {
						t.Errorf("link[%d]: expected %q, got %q", i, wantLinks[i], link)
					}
				}
			}

			// Check resources (order-independent).
			gotResources := append([]string(nil), result.Resources...)
			wantResources := append([]string(nil), test.expectedResources...)
			sort.Strings(gotResources)
			sort.Strings(wantResources)

			if len(gotResources) != len(wantResources) {
				t.Errorf("expected %d resources: %v, got %d: %v", len(wantResources), wantResources, len(gotResources), gotResources)
			} else {
				for i, res := range gotResources {
					if res != wantResources[i] {
						t.Errorf("resource[%d]: expected %q, got %q", i, wantResources[i], res)
					}
				}
			}
		})
	}
}

// TestResolveHref covers the cases ResolveHref needs to get right:
// absolute URLs of various shapes, all the relative forms (root, document,
// parent, query-only, fragment-only, protocol-relative), and rejection of
// non-HTTP(S) schemes.
func TestResolveHref(t *testing.T) {
	tests := []struct {
		pageURL  string
		href     string
		expected string // empty when we expect an error
		wantErr  bool
	}{
		// Absolute URLs pass through.
		{"http://test.local/", "http://example.com", "http://example.com", false},
		{"http://test.local/", "https://example.com", "https://example.com", false},
		{"http://test.local/", "http://example.com:8080", "http://example.com:8080", false},
		{"http://test.local/", "http://example.com/path?query=1", "http://example.com/path?query=1", false},

		// Root-relative href: resolves against base scheme+host.
		{"http://test.local/products/", "/about.html", "http://test.local/about.html", false},

		// Document-relative href: resolves against the page's directory.
		{"http://test.local/products/", "widget.html", "http://test.local/products/widget.html", false},

		// Parent-directory href.
		{"http://test.local/products/", "../about.html", "http://test.local/about.html", false},

		// Query-only href: keeps the page's path, replaces the query.
		{"http://test.local/search.html?q=old", "?q=new", "http://test.local/search.html?q=new", false},

		// Fragment-only href: keeps the page's path, attaches the fragment.
		{"http://test.local/contact.html", "#form", "http://test.local/contact.html#form", false},

		// Protocol-relative href: inherits the base's scheme.
		{"https://test.local/", "//cdn.example.com/lib.js", "https://cdn.example.com/lib.js", false},

		// Schemeless ref whose first label is a real TLD: promoted to an external host, inheriting the base scheme.
		{"http://test.local/products/", "notes.io", "http://notes.io", false},
		{"http://test.local/products/", "example.com/page?q=1", "http://example.com/page?q=1", false},
		{"https://test.local/", "www.google.com", "https://www.google.com", false},

		// Schemeless ref whose first label is a file extension (not a public suffix): stays a relative path.
		{"http://test.local/products/", "report.pdf", "http://test.local/products/report.pdf", false},

		// Non-HTTP(S) schemes are rejected.
		{"http://test.local/", "mailto:hello@example.com", "", true},
		{"http://test.local/", "javascript:void(0)", "", true},
		{"http://test.local/", "tel:+15551234567", "", true},
	}

	for _, test := range tests {
		got, err := ResolveHref(test.pageURL, test.href)
		if test.wantErr {
			if err == nil {
				t.Errorf("ResolveHref(%q, %q): expected error, got %q", test.pageURL, test.href, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ResolveHref(%q, %q): unexpected error: %v", test.pageURL, test.href, err)
			continue
		}
		if got != test.expected {
			t.Errorf("ResolveHref(%q, %q): expected %q, got %q", test.pageURL, test.href, test.expected, got)
		}
	}
}
