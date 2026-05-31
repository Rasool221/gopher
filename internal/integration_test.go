//go:build integration

package internal

import (
	"strings"
	"testing"
)

// testSiteURL points at the dockerized test site
var testSiteURL = "http://localhost:8080"

// collectReachedURLs walks a URLMap tree and records every URL it visits into
// the reached set. Used by traversal tests to assert which pages gopher hit.
func collectReachedURLs(node URLMap, reached map[string]struct{}) {
	reached[node.URL] = struct{}{}
	for _, child := range node.links {
		collectReachedURLs(child, reached)
	}
}

func TestIntegration_GetPageContent_Home(t *testing.T) {
	body, err := GetPageContent(testSiteURL + "/")
	if err != nil {
		t.Fatalf("GetPageContent: %v", err)
	}
	if !strings.Contains(body, "Gopher Test Site") {
		t.Errorf("expected home page to contain site title, got: %.200s...", body)
	}
}

func TestIntegration_GetPageContent_404(t *testing.T) {
	if _, err := GetPageContent(testSiteURL + "/does-not-exist.html"); err == nil {
		t.Fatal("expected error on 404, got nil")
	}
}

func TestIntegration_GetPageContent_410(t *testing.T) {
	if _, err := GetPageContent(testSiteURL + "/gone"); err == nil {
		t.Fatal("expected error on 410, got nil")
	}
}

// Go's default http.Client follows redirects, so a 302 should resolve to the
// destination body transparently.
func TestIntegration_GetPageContent_FollowsRedirect(t *testing.T) {
	body, err := GetPageContent(testSiteURL + "/redirect-once")
	if err != nil {
		t.Fatalf("GetPageContent: %v", err)
	}
	if !strings.Contains(body, "About") {
		t.Errorf("expected redirect target (about page) body, got: %.200s...", body)
	}
}

// TestIntegration_BuildUrlMap_TraversesSite walks the test site from the root
// and asserts the crawler reached a sampling of pages from each section.
// It also implicitly tests that cycles (e.g. blog post-3 linking to itself)
// don't hang gopher — if they did, this test would time out instead of
// failing with an assertion.
func TestIntegration_BuildUrlMap_TraversesSite(t *testing.T) {
	root := testSiteURL + "/"
	got := BuildUrlMap(root, nil)
	if got.URL != root {
		t.Fatalf("expected root URL %q, got %q", root, got.URL)
	}

	// Flatten the URLMap tree into a flat set of every URL we reached.
	reached := map[string]struct{}{}
	collectReachedURLs(got, reached)

	// A handful of URLs we expect gopher to reach by following links
	// from the home page (directly or transitively).
	wantURLs := []string{
		testSiteURL + "/about.html",
		testSiteURL + "/products/",
		testSiteURL + "/products/widget.html",
		testSiteURL + "/products/gadget.html",
		testSiteURL + "/blog/",
		testSiteURL + "/blog/post-1.html",
		testSiteURL + "/blog/post-2.html",
		testSiteURL + "/blog/post-3.html",
		testSiteURL + "/contact.html",
		testSiteURL + "/deep/level-1.html",
		testSiteURL + "/deep/level-3.html",
	}

	for _, want := range wantURLs {
		if _, ok := reached[want]; !ok {
			t.Errorf("gopher did not reach %q", want)
		}
	}
}
