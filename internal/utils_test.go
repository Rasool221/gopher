package internal

import "testing"

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
