package internal

import (
	"sort"
	"testing"
)

func TestExtractLinksFromHTML(t *testing.T) {
	tests := []struct {
		pageURL  string // The URL the HTML was "fetched" from; relative hrefs resolve against this.
		content  string
		expected []string // Expected resolved URLs, order-independent.
	}{
		// Absolute http URL: passes through unchanged.
		{"http://test.local/", `<html><body><a href="http://example.com">Example</a></body></html>`, []string{"http://example.com"}},

		// Multiple links. "google.com" is schemeless but its trailing label is a real TLD (.com),
		// so it's promoted to the external host "http://google.com". "104.20.23.154" has no public
		// suffix (numeric), so it stays a relative path resolved against the page URL.
		{
			"http://test.local/",
			`<html><body><a href="http://example.com">Example</a><div href="google.com">Google</div><img href="104.20.23.154">Example by IP address</img></body></html>`,
			[]string{"http://example.com", "http://google.com", "http://test.local/104.20.23.154"},
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
