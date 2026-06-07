//go:build integration

package internal

import (
	"fmt"
	"strings"
	"testing"
)

// testSiteURL and externalSiteURL point at the two dockerized test sites. They're
// served as distinct base domains via Docker network aliases (see test/docker-compose.yml),
// so these tests must run inside that network — use `just integration-tests`.
var testSiteURL = "http://primary.com"
var externalSiteURL = "http://external.com"

// primaryURLs and externalURLs are every URL gopher should traverse on each site.
// Both containers serve identical content, so the two lists mirror each other under
// their respective base domains. The set is deliberately exhaustive — it includes the
// fragment variant (/contact.html#form), the query-string variants (/search.html?...),
// the redirect URLs (/redirect-once, /redirect-chain), and the error pages (/missing.html
// 404, /gone 410). gopher records all of these as visited nodes, so the traversal tests
// assert the reached set matches these lists exactly.
var primaryURLs = []string{
	testSiteURL + "/",
	testSiteURL + "/about.html",
	testSiteURL + "/blog/",
	testSiteURL + "/blog/post-1.html",
	testSiteURL + "/blog/post-2.html",
	testSiteURL + "/blog/post-3.html",
	testSiteURL + "/contact.html",
	testSiteURL + "/contact.html#form",
	testSiteURL + "/deep/level-1.html",
	testSiteURL + "/deep/level-2.html",
	testSiteURL + "/deep/level-3.html",
	testSiteURL + "/gone",
	testSiteURL + "/missing.html",
	testSiteURL + "/products/",
	testSiteURL + "/products/gadget.html",
	testSiteURL + "/products/widget.html",
	testSiteURL + "/redirect-chain",
	testSiteURL + "/redirect-once",
	testSiteURL + "/search.html?q=gopher&page=1",
	testSiteURL + "/search.html?q=gopher&page=2",
	testSiteURL + "/search.html?q=gopher&page=3",
	testSiteURL + "/search.html?q=other",
}

var externalURLs = []string{
	externalSiteURL + "/",
	externalSiteURL + "/about.html",
	externalSiteURL + "/blog/",
	externalSiteURL + "/blog/post-1.html",
	externalSiteURL + "/blog/post-2.html",
	externalSiteURL + "/blog/post-3.html",
	externalSiteURL + "/contact.html",
	externalSiteURL + "/contact.html#form",
	externalSiteURL + "/deep/level-1.html",
	externalSiteURL + "/deep/level-2.html",
	externalSiteURL + "/deep/level-3.html",
	externalSiteURL + "/gone",
	externalSiteURL + "/missing.html",
	externalSiteURL + "/products/",
	externalSiteURL + "/products/gadget.html",
	externalSiteURL + "/products/widget.html",
	externalSiteURL + "/redirect-chain",
	externalSiteURL + "/redirect-once",
	externalSiteURL + "/search.html?q=gopher&page=1",
	externalSiteURL + "/search.html?q=gopher&page=2",
	externalSiteURL + "/search.html?q=gopher&page=3",
	externalSiteURL + "/search.html?q=other",
}

// collectReachedURLs walks a URLMap tree and records every URL it visits into
// the reached set. Used by traversal tests to assert which pages gopher hit.
func collectReachedURLs(node URLMap, reached map[string]struct{}) {
	reached[node.URL] = struct{}{}
	for _, child := range node.links {
		collectReachedURLs(child, reached)
	}
}

// printCrawl writes the crawl result to stdout in the same format gopher prints by
// default (internal.PrintURLMap), so each test run shows the URL map it built.
func printCrawl(label string, m URLMap) {
	fmt.Printf("\n===== %s =====\n", label)
	PrintURLMap(m, 0)
}

// assertReachedExactly fails if the crawl reached a different set of URLs than want:
// it flags every expected route that wasn't traversed and every URL that was traversed
// but isn't in the expected list.
func assertReachedExactly(t *testing.T, m URLMap, want []string) {
	t.Helper()

	reached := map[string]struct{}{}
	collectReachedURLs(m, reached)

	wantSet := make(map[string]struct{}, len(want))
	for _, w := range want {
		wantSet[w] = struct{}{}
		if _, ok := reached[w]; !ok {
			t.Errorf("expected gopher to traverse %q, but it did not", w)
		}
	}
	for r := range reached {
		if _, ok := wantSet[r]; !ok {
			t.Errorf("gopher traversed unexpected URL %q (not in the expected route list)", r)
		}
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

// TestIntegration_BuildUrlMap_TraversesSite walks the primary site from the root and
// asserts gopher reached exactly the routes in primaryURLs (External defaults to false,
// so the external site is not crawled). It also implicitly tests that cycles (e.g. blog
// post-3 linking to itself) don't hang gopher — if they did, this test would time out
// instead of failing an assertion.
func TestIntegration_BuildUrlMap_TraversesSite(t *testing.T) {
	root := testSiteURL + "/"
	got := NewGopher(NewConfig(CLIConfig{})).BuildURLMap(root)
	printCrawl("BuildUrlMap (primary, External=false)", got)

	if got.URL != root {
		t.Fatalf("expected root URL %q, got %q", root, got.URL)
	}

	assertReachedExactly(t, got, primaryURLs)
}

// TestIntegration_External_FlagControlsCrossDomainTraversal exercises the cfg.External
// toggle. The home page links to a second site served under a different base domain
// (external.com vs primary.com). With External=true gopher should traverse both sites'
// full route sets; with External=false it should traverse only the primary site's routes
// and skip the external site entirely.
func TestIntegration_External_FlagControlsCrossDomainTraversal(t *testing.T) {
	root := testSiteURL + "/"

	// External enabled: every primary AND external route should be traversed.
	withExternal := NewGopher(NewConfig(CLIConfig{External: true})).BuildURLMap(root)
	printCrawl("External=true (primary + external)", withExternal)
	allURLs := append(append([]string{}, primaryURLs...), externalURLs...)
	assertReachedExactly(t, withExternal, allURLs)

	// External disabled: only the primary routes should be traversed.
	withoutExternal := NewGopher(NewConfig(CLIConfig{External: false})).BuildURLMap(root)
	printCrawl("External=false (primary only)", withoutExternal)
	assertReachedExactly(t, withoutExternal, primaryURLs)
}
