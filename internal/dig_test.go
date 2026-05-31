package internal

import (
	"sort"
	"testing"
)

// How do i add tests for this function? Ill come back to this later.
// func testGetPageContent(t *testing.T) {
// }

func TestExtractLinksFromHTML(t *testing.T) {
	tests := []struct {
		pageURL  string // The URL the HTML was "fetched" from; relative hrefs resolve against this.
		content  string
		expected []string // Expected resolved URLs, order-independent.
	}{
		// Absolute http URL: passes through unchanged.
		{"http://test.local/", `<html><body><a href="http://example.com">Example</a></body></html>`, []string{"http://example.com"}},

		// Multiple links. Note: "google.com" and "104.20.23.154" are RELATIVE paths per RFC 3986
		// (no scheme), so they resolve against the page URL — not treated as bare hosts.
		{
			"http://test.local/",
			`<html><body><a href="http://example.com">Example</a><div href="google.com">Google</div><img href="104.20.23.154">Example by IP address</img></body></html>`,
			[]string{"http://example.com", "http://test.local/google.com", "http://test.local/104.20.23.154"},
		},

		// Root-relative href resolves against the page's scheme+host.
		{"http://test.local/products/", `<html><a href="/about.html">About</a></html>`, []string{"http://test.local/about.html"}},

		// Document-relative href resolves against the page's directory.
		{"http://test.local/products/", `<html><a href="widget.html">Widget</a></html>`, []string{"http://test.local/products/widget.html"}},

		// Parent-directory href.
		{"http://test.local/products/", `<html><a href="../about.html">Up one</a></html>`, []string{"http://test.local/about.html"}},

		// Non-HTTP(S) schemes get rejected into the errors slice, not the links slice.
		{"http://test.local/", `<html><a href="mailto:hello@example.com">mail</a><a href="javascript:void(0)">js</a></html>`, []string{}},

		// Empty result on empty HTML.
		{"http://test.local/", `<html></html>`, []string{}},

		// Malformed HTML still tokenizes; no hrefs means no links.
		{"http://test.local/", `<html><not-a-real-tag>`, []string{}},
	}

	for _, test := range tests {
		result, _ := ExtractLinksFromHTML(test.pageURL, test.content)
		if len(result) != len(test.expected) {
			t.Errorf("Expected %d links: %v, got %d links: %v", len(test.expected), test.expected, len(result), result)
			continue
		}

		// ExtractLinksFromHTML returns links in non-deterministic order (map-based dedupe),
		// so sort both sides before comparing element-wise.
		got := append([]string(nil), result...)
		want := append([]string(nil), test.expected...)
		sort.Strings(got)
		sort.Strings(want)

		for i, link := range got {
			if link != want[i] {
				t.Errorf("Expected link '%s', got '%s'", want[i], link)
			}
		}
	}
}
