package internal

import (
	"reflect"
	"sort"
	"testing"
)

func TestExtractDataFromHTML(t *testing.T) {
	tests := []struct {
		name              string
		pageURL           string
		content           string
		expectedLinks     []string
		expectedResources []Resource
		expectedErrCount  int
	}{
		{
			name:              "absolute http URL",
			pageURL:           "http://example.net/",
			content:           `<html><body><a href="http://example.com">Example</a></body></html>`,
			expectedLinks:     []string{"http://example.com"},
			expectedResources: nil,
			expectedErrCount:  0,
		},
		{
			name:    "multiple links with schemeless and IP",
			pageURL: "http://example.net/",
			// "google.com" promoted to external host; "104.20.23.154" stays relative.
			content:           `<html><body><a href="http://example.com">Example</a><div href="google.com">Google</div><img href="104.20.23.154">Example by IP address</img></body></html>`,
			expectedLinks:     []string{"http://example.com", "http://google.com", "http://example.net/104.20.23.154"},
			expectedResources: nil,
			expectedErrCount:  0,
		},
		{
			name:              "root-relative href",
			pageURL:           "http://example.net/products/",
			content:           `<html><a href="/about.html">About</a></html>`,
			expectedLinks:     []string{"http://example.net/about.html"},
			expectedResources: nil,
			expectedErrCount:  0,
		},
		{
			name:              "document-relative href",
			pageURL:           "http://example.net/products/",
			content:           `<html><a href="widget.html">Widget</a></html>`,
			expectedLinks:     []string{"http://example.net/products/widget.html"},
			expectedResources: nil,
			expectedErrCount:  0,
		},
		{
			name:              "parent-directory href",
			pageURL:           "http://example.net/products/",
			content:           `<html><a href="../about.html">Up one</a></html>`,
			expectedLinks:     []string{"http://example.net/about.html"},
			expectedResources: nil,
			expectedErrCount:  0,
		},
		{
			name:              "non-HTTP schemes produce errors not links",
			pageURL:           "http://example.net/",
			content:           `<html><a href="mailto:hello@example.com">mail</a><a href="javascript:void(0)">js</a></html>`,
			expectedLinks:     nil,
			expectedResources: nil,
			expectedErrCount:  2,
		},
		{
			name:              "empty HTML",
			pageURL:           "http://example.net/",
			content:           `<html></html>`,
			expectedLinks:     nil,
			expectedResources: nil,
			expectedErrCount:  0,
		},
		{
			name:              "malformed HTML",
			pageURL:           "http://example.net/",
			content:           `<html><not-a-real-tag>`,
			expectedLinks:     nil,
			expectedResources: nil,
			expectedErrCount:  0,
		},

		// --- Resource tests ---

		{
			name:              "img src is a resource",
			pageURL:           "http://example.net/",
			content:           `<html><body><img src="/images/photo.jpg"></body></html>`,
			expectedLinks: nil,
			expectedResources: []Resource{
				{URL: "http://example.net/images/photo.jpg", Path: []string{"http://example.net/", "http://example.net/images/", "http://example.net/images/photo.jpg"}},
			},
			expectedErrCount: 0,
		},
		{
			name:              "video src and poster are resources",
			pageURL:           "http://example.net/",
			content:           `<html><body><video src="/media/clip.mp4" poster="/media/poster.jpg"></video></body></html>`,
			expectedLinks:     nil,
			expectedResources: []Resource{
				{URL: "http://example.net/media/clip.mp4", Path: []string{"http://example.net/", "http://example.net/media/", "http://example.net/media/clip.mp4"}},
				{URL: "http://example.net/media/poster.jpg", Path: []string{"http://example.net/", "http://example.net/media/", "http://example.net/media/poster.jpg"}},
			},
			expectedErrCount: 0,
		},
		{
			name:              "gif via img src",
			pageURL:           "http://example.net/",
			content:           `<html><body><img src="/images/animation.gif"></body></html>`,
			expectedLinks:     nil,
			expectedResources: []Resource{
				{URL: "http://example.net/images/animation.gif", Path: []string{"http://example.net/", "http://example.net/images/", "http://example.net/images/animation.gif"}},
			},
			expectedErrCount: 0,
		},
		{
			name:              "file download link via href",
			pageURL:           "http://example.net/",
			content:           `<html><body><a href="/files/report.pdf">Download PDF</a></body></html>`,
			expectedLinks:     nil,
			expectedResources: []Resource{
				{URL: "http://example.net/files/report.pdf", Path: []string{"http://example.net/", "http://example.net/files/", "http://example.net/files/report.pdf"}},
			},
			expectedErrCount: 0,
		},
		{
			name:              "deeply nested resource builds path up to base host",
			pageURL:           "http://example.net/",
			content:           `<html><body><img src="/a/b/c/logo.png"></body></html>`,
			expectedLinks:     nil,
			expectedResources: []Resource{
				{URL: "http://example.net/a/b/c/logo.png", Path: []string{"http://example.net/", "http://example.net/a/", "http://example.net/a/b/", "http://example.net/a/b/c/", "http://example.net/a/b/c/logo.png"}},
			},
			expectedErrCount: 0,
		},
		{
			name:              "mixed links, resources, and errors",
			pageURL:           "http://example.net/",
			content:           `<html><body><a href="/about.html">About</a><img src="/img/logo.png"><a href="mailto:hi@example.com">Mail</a><video src="/vid/intro.webm" poster="/vid/thumb.png"></video></body></html>`,
			expectedLinks: []string{"http://example.net/about.html"},
			expectedResources: []Resource{
				{URL: "http://example.net/img/logo.png", Path: []string{"http://example.net/", "http://example.net/img/", "http://example.net/img/logo.png"}},
				{URL: "http://example.net/vid/intro.webm", Path: []string{"http://example.net/", "http://example.net/vid/", "http://example.net/vid/intro.webm"}},
				{URL: "http://example.net/vid/thumb.png", Path: []string{"http://example.net/", "http://example.net/vid/", "http://example.net/vid/thumb.png"}},
			},
			expectedErrCount: 1,
		},
		{
			name:              "deduplication of resources",
			pageURL:           "http://example.net/",
			content:           `<html><body><img src="/img/logo.png"><img src="/img/logo.png"></body></html>`,
			expectedLinks: nil,
			expectedResources: []Resource{
				{URL: "http://example.net/img/logo.png", Path: []string{"http://example.net/", "http://example.net/img/", "http://example.net/img/logo.png"}},
			},
			expectedErrCount: 0,
		},
		{
			name:              "deduplication of links",
			pageURL:           "http://example.net/",
			content:           `<html><body><a href="/about.html">A</a><a href="/about.html">B</a></body></html>`,
			expectedLinks:     []string{"http://example.net/about.html"},
			expectedResources: nil,
			expectedErrCount:  0,
		},
		{
			name:              "absolute resource URL passes through",
			pageURL:           "http://example.net/",
			content:           `<html><body><img src="https://cdn.example.com/pic.jpg"></body></html>`,
			expectedLinks: nil,
			expectedResources: []Resource{
				{URL: "https://cdn.example.com/pic.jpg", Path: nil},
			},
			expectedErrCount: 0,
		},
		{
			name:              "source element inside video",
			pageURL:           "http://example.net/",
			content:           `<html><body><video><source src="/media/trailer.mkv" type="video/x-matroska"></video></body></html>`,
			expectedLinks: nil,
			expectedResources: []Resource{
				{URL: "http://example.net/media/trailer.mkv", Path: []string{"http://example.net/", "http://example.net/media/", "http://example.net/media/trailer.mkv"}},
			},
			expectedErrCount: 0,
		},
		{
			name:              "audio src is a resource",
			pageURL:           "http://example.net/",
			content:           `<html><body><audio src="/audio/song.mp3"></audio></body></html>`,
			expectedLinks:     nil,
			expectedResources: []Resource{
				{URL: "http://example.net/audio/song.mp3", Path: []string{"http://example.net/", "http://example.net/audio/", "http://example.net/audio/song.mp3"}},
			},
			expectedErrCount: 0,
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
				t.Errorf("result.Links %+v", result.Links)
			} else {
				for i, link := range gotLinks {
					if link != wantLinks[i] {
						t.Errorf("link[%d]: expected %q, got %q", i, wantLinks[i], link)
					}
				}
			}

			// Check resources (order-independent, but paths are order-sensitive per resource).
			gotResources := append([]Resource(nil), result.Resources...)
			wantResources := append([]Resource(nil), test.expectedResources...)
			sort.Slice(gotResources, func(i, j int) bool { return gotResources[i].URL < gotResources[j].URL })
			sort.Slice(wantResources, func(i, j int) bool { return wantResources[i].URL < wantResources[j].URL })

			if len(gotResources) != len(wantResources) {
				t.Errorf("expected %d resources: %v, got %d: %v", len(wantResources), wantResources, len(gotResources), gotResources)
				t.Errorf("result.Resources %+v", result.Resources)
			} else {
				for i, got := range gotResources {
					want := wantResources[i]
					if got.URL != want.URL {
						t.Errorf("resource[%d]: expected URL %q, got %q", i, want.URL, got.URL)
						continue
					}
					if !reflect.DeepEqual(got.Path, want.Path) {
						t.Errorf("resource[%d] %q: expected path %v, got %v", i, got.URL, want.Path, got.Path)
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
		{"http://example.net/", "http://example.com", "http://example.com", false},
		{"http://example.net/", "https://example.com", "https://example.com", false},
		{"http://example.net/", "http://example.com:8080", "http://example.com:8080", false},
		{"http://example.net/", "http://example.com/path?query=1", "http://example.com/path?query=1", false},

		// Root-relative href: resolves against base scheme+host.
		{"http://example.net/products/", "/about.html", "http://example.net/about.html", false},

		// Document-relative href: resolves against the page's directory.
		{"http://example.net/products/", "widget.html", "http://example.net/products/widget.html", false},

		// Parent-directory href.
		{"http://example.net/products/", "../about.html", "http://example.net/about.html", false},

		// Query-only href: keeps the page's path, replaces the query.
		{"http://example.net/search.html?q=old", "?q=new", "http://example.net/search.html?q=new", false},

		// Fragment-only href: keeps the page's path, attaches the fragment.
		{"http://example.net/contact.html", "#form", "http://example.net/contact.html#form", false},

		// Protocol-relative href: inherits the base's scheme.
		{"https://example.net/", "//cdn.example.com/lib.js", "https://cdn.example.com/lib.js", false},

		// Schemeless ref whose first label is a real TLD: promoted to an external host, inheriting the base scheme.
		{"http://example.net/products/", "notes.io", "http://notes.io", false},
		{"http://example.net/products/", "example.com/page?q=1", "http://example.com/page?q=1", false},
		{"https://example.net/", "www.google.com", "https://www.google.com", false},

		// Schemeless ref whose first label is a file extension (not a public suffix): stays a relative path.
		{"http://example.net/products/", "report.pdf", "http://example.net/products/report.pdf", false},

		// Non-HTTP(S) schemes are rejected.
		{"http://example.net/", "mailto:hello@example.com", "", true},
		{"http://example.net/", "javascript:void(0)", "", true},
		{"http://example.net/", "tel:+15551234567", "", true},
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
