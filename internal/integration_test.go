//go:build integration

package internal

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

// testSiteURL and externalSiteURL point at the two dockerized test sites. They're
// served as distinct base domains via Docker network aliases (see test/docker-compose.yml),
// so these tests must run inside that network. Use `just integration-tests`.
var (
	testSiteURL     = "http://primary.com"
	externalSiteURL = "http://external.com"
)

// primaryURLs and externalURLs are every link-gopher should walk on each site.
// Both containers serve identical content, so the two lists mirror each other under
// their respective base domains. The set is deliberately exhaustive; it includes the
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

// primaryResources maps a page URL to the resource URLs expected on that page.
// Only index.html carries resources (img, video, audio, source tags, plus downloadable files).
var primaryResources = map[string][]string{
	testSiteURL + "/": {
		testSiteURL + "/assets/test.jpg",
		testSiteURL + "/assets/logo.png",
		testSiteURL + "/assets/test.gif",
		testSiteURL + "/assets/test.mp4",
		testSiteURL + "/assets/poster.png",
		testSiteURL + "/assets/test.webm",
		testSiteURL + "/assets/theme.mp3",
		testSiteURL + "/assets/report.pdf",
		testSiteURL + "/assets/data.csv",
	},
}

// externalResources mirrors primaryResources for the external site.
var externalResources = map[string][]string{
	externalSiteURL + "/": {
		externalSiteURL + "/assets/test.jpg",
		externalSiteURL + "/assets/logo.png",
		externalSiteURL + "/assets/test.gif",
		externalSiteURL + "/assets/test.mp4",
		externalSiteURL + "/assets/poster.png",
		externalSiteURL + "/assets/test.webm",
		externalSiteURL + "/assets/theme.mp3",
		externalSiteURL + "/assets/report.pdf",
		externalSiteURL + "/assets/data.csv",
	},
}

// primaryErrors maps a page URL to the number of resolution errors expected.
// index.html has mailto:, tel:, and javascript: hrefs → 3 errors.
var primaryErrors = map[string]int{
	testSiteURL + "/": 3,
}

var externalErrors = map[string]int{
	externalSiteURL + "/": 3,
}

// collectReachedURLs walks a URLMap tree and records every URL it visits into
// the reached set. Used by traversal tests to assert which pages gopher hit.
func collectReachedURLs(node URLMap, reached map[string]struct{}) {
	reached[node.URL] = struct{}{}
	for _, child := range node.links {
		collectReachedURLs(child, reached)
	}
}

// collectResourcesByURL walks a URLMap tree and builds a map from each page URL
// to its list of resources. Returns one entry per visited page.
func collectResourcesByURL(node URLMap, out map[string][]string) {
	if len(node.resources) > 0 {
		urls := make([]string, 0, len(node.resources))
		for _, r := range node.resources {
			urls = append(urls, r.URL)
		}
		out[node.URL] = urls
	}
	for _, child := range node.links {
		collectResourcesByURL(child, out)
	}
}

// collectErrorsByURL walks a URLMap tree and builds a map from each page URL
// to its error count.
func collectErrorsByURL(node URLMap, out map[string]int) {
	if len(node.errors) > 0 {
		out[node.URL] = len(node.errors)
	}
	for _, child := range node.links {
		collectErrorsByURL(child, out)
	}
}

// printCrawl writes the crawl result to stdout in the same format gopher prints by
// default (internal.PrintURLMap), so each test run shows the URL map it built.
func printCrawl(label string, m URLMap) {
	fmt.Printf("\n===== %s =====\n", label)
	// PrintURLMap(m, 0)
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

// assertResourcesExactly fails if a page's resources don't match the expected list.
func assertResourcesExactly(t *testing.T, m URLMap, want map[string][]string) {
	t.Helper()

	got := map[string][]string{}
	collectResourcesByURL(m, got)

	for pageURL, wantResources := range want {
		gotResources, ok := got[pageURL]
		if !ok {
			if len(wantResources) > 0 {
				t.Errorf("page %q: expected %d resources, but page had none", pageURL, len(wantResources))
			}
			continue
		}

		sortedGot := append([]string(nil), gotResources...)
		sortedWant := append([]string(nil), wantResources...)
		sort.Strings(sortedGot)
		sort.Strings(sortedWant)

		if len(sortedGot) != len(sortedWant) {
			t.Errorf("page %q: expected %d resources %v, got %d %v", pageURL, len(sortedWant), sortedWant, len(sortedGot), sortedGot)
			continue
		}
		for i, r := range sortedGot {
			if r != sortedWant[i] {
				t.Errorf("page %q: resource[%d]: expected %q, got %q", pageURL, i, sortedWant[i], r)
			}
		}
	}

	// Fail if a page returned resources we didn't expect.
	for pageURL, gotResources := range got {
		if _, expected := want[pageURL]; !expected && len(gotResources) > 0 {
			t.Errorf("page %q: unexpected resources %v", pageURL, gotResources)
		}
	}
}

// assertErrorsExactly fails if a page's error count doesn't match the expected count.
func assertErrorsExactly(t *testing.T, m URLMap, want map[string]int) {
	t.Helper()

	got := map[string]int{}
	collectErrorsByURL(m, got)

	for pageURL, wantCount := range want {
		gotCount := got[pageURL]
		if gotCount != wantCount {
			t.Errorf("page %q: expected %d errors, got %d", pageURL, wantCount, gotCount)
		}
	}

	for pageURL, gotCount := range got {
		if _, expected := want[pageURL]; !expected && gotCount > 0 {
			t.Errorf("page %q: unexpected %d errors", pageURL, gotCount)
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

// TestIntegration_Run_TraversesSite walks the primary site from the root and
// asserts gopher reached exactly the routes in primaryURLs (External defaults to false,
// so the external site is not crawled). It also implicitly tests that cycles (e.g. blog
// post-3 linking to itself) don't hang gopher (if they did, this test would time out
// instead of failing an assertion.
func TestIntegration_Run_TraversesSite(t *testing.T) {
	root := testSiteURL + "/"
	got := NewGopher(NewConfig(CLIConfig{})).Run(root)
	printCrawl("Run (primary, External=false)", got)

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
	withExternal := NewGopher(NewConfig(CLIConfig{External: true})).Run(root)
	printCrawl("External=true (primary + external)", withExternal)
	allURLs := append(append([]string{}, primaryURLs...), externalURLs...)
	assertReachedExactly(t, withExternal, allURLs)

	// External disabled: only the primary routes should be traversed.
	withoutExternal := NewGopher(NewConfig(CLIConfig{External: false})).Run(root)
	printCrawl("External=false (primary only)", withoutExternal)
	assertReachedExactly(t, withoutExternal, primaryURLs)
}

// TestIntegration_Run_PrimaryResources asserts that the primary home page
// carries exactly the resources defined in index.html (img, video, audio, source).
func TestIntegration_Run_PrimaryResources(t *testing.T) {
	root := testSiteURL + "/"
	got := NewGopher(NewConfig(CLIConfig{})).Run(root)
	assertResourcesExactly(t, got, primaryResources)
}

// TestIntegration_Run_PrimaryErrors asserts that the primary home page
// produces errors for mailto:, tel:, and javascript: hrefs.
func TestIntegration_Run_PrimaryErrors(t *testing.T) {
	root := testSiteURL + "/"
	got := NewGopher(NewConfig(CLIConfig{})).Run(root)
	assertErrorsExactly(t, got, primaryErrors)
}

// TestIntegration_ExternalResources asserts that when External=true, both
// the primary and external home pages carry their respective resources.
func TestIntegration_ExternalResources(t *testing.T) {
	root := testSiteURL + "/"
	got := NewGopher(NewConfig(CLIConfig{External: true})).Run(root)

	merged := map[string][]string{}
	for k, v := range primaryResources {
		merged[k] = v
	}
	for k, v := range externalResources {
		merged[k] = v
	}
	assertResourcesExactly(t, got, merged)
}

// TestIntegration_ExternalErrors asserts that when External=true, both
// the primary and external home pages produce their expected errors.
func TestIntegration_ExternalErrors(t *testing.T) {
	root := testSiteURL + "/"
	got := NewGopher(NewConfig(CLIConfig{External: true})).Run(root)

	merged := map[string]int{}
	for k, v := range primaryErrors {
		merged[k] = v
	}
	for k, v := range externalErrors {
		merged[k] = v
	}
	assertErrorsExactly(t, got, merged)
}
