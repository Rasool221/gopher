package internal

import "testing"

var (
	SHOULD_ERR     = true
	SHOULD_NOT_ERR = false
)

func TestValidateURLFormat(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"Valid HTTP URL", "http://example.com", SHOULD_NOT_ERR},
		{"Valid HTTPS URL", "https://example.com", SHOULD_NOT_ERR},
		{"Valid URL with port", "http://example.com:8080", SHOULD_NOT_ERR},
		{"Validate URL without scheme", "www.example.com", SHOULD_NOT_ERR},
		{"Valid URL without scheme", "example.com", SHOULD_NOT_ERR},
		{"Valid localhost URL", "http://localhost:8080", SHOULD_NOT_ERR},
		{"Valid localhost URL without port", "http://localhost", SHOULD_NOT_ERR},
		{"Valid localhost URL without scheme", "localhost:8080", SHOULD_NOT_ERR},
		{"Valid localhost URL without scheme and port", "localhost", SHOULD_NOT_ERR},
		{"Valid IP address URL", "192.168.1.1", SHOULD_NOT_ERR},
		{"Valid IP address URL with scheme", "http://192.168.1.1", SHOULD_NOT_ERR},
		{"Valid IP address URL with scheme and port", "http://192.168.1.1:8080", SHOULD_NOT_ERR},
		{"Valid IP address URL with port", "192.168.1.1:8080", SHOULD_NOT_ERR},
		{"Validate URL with unique TLD", "http://example.sbs", SHOULD_NOT_ERR},
		{"Empty URL", "", SHOULD_ERR},
		{"Invalid URL format", "http://example", SHOULD_ERR},
		{"Invalid URL format with spaces", "example", SHOULD_ERR},
		{"Invalid URL with special characters", "http://exa&*$.com", SHOULD_ERR},
		{"Invalid URL with unsupported scheme", "ftp://example.com", SHOULD_ERR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() test=%q url=%q error = %v, wantErr %v", tt.name, tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestGetBaseDomain(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{"Valid HTTP URL", "http://example.com", "example.com", SHOULD_NOT_ERR},
		{"Valid HTTPS URL", "https://example.com", "example.com", SHOULD_NOT_ERR},
		{"Valid URL with port", "http://example.com:8080", "example.com", SHOULD_NOT_ERR},
		{"Domain with subdomain", "http://sub.example.com", "example.com", SHOULD_NOT_ERR},
		{"Domain with multiple subdomains", "http://a.b.c.example.com", "example.com", SHOULD_NOT_ERR},
		{"Validate URL without scheme", "www.example.com", "example.com", SHOULD_NOT_ERR},
		{"Valid URL without scheme", "example.com", "example.com", SHOULD_NOT_ERR},
		{"Domain without scheme and port", "example.com:8080", "example.com", SHOULD_NOT_ERR},
		{"Valid localhost URL", "http://localhost:8080", "localhost", SHOULD_NOT_ERR},
		{"Valid localhost URL without port", "http://localhost", "localhost", SHOULD_NOT_ERR},
		{"Valid localhost URL without scheme", "localhost:8080", "localhost", SHOULD_NOT_ERR},
		{"Valid localhost URL without scheme and port", "localhost", "localhost", SHOULD_NOT_ERR},
		{"Valid IP address URL", "192.168.1.1", "192.168.1.1", SHOULD_NOT_ERR},
		{"Valid IP address URL with scheme", "http://192.168.1.1", "192.168.1.1", SHOULD_NOT_ERR},
		{"Valid IP address URL with scheme and port", "http://192.168.1.1:8080", "192.168.1.1", SHOULD_NOT_ERR},
		{"Valid IP address URL with port", "192.168.1.1:8080", "192.168.1.1", SHOULD_NOT_ERR},
		{"Empty URL", "", "", SHOULD_ERR},
		{"Invalid URL format", "http://example", "", SHOULD_ERR},
		{"Invalid URL format with spaces", "example", "", SHOULD_ERR},
		{"Invalid URL with special characters", "http://exa&*$.com", "", SHOULD_ERR},
		{"Invalid URL with unsupported scheme", "ftp://example.com", "", SHOULD_ERR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBaseDomain(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBaseDomain() test=%q url=%q error = %v, wantErr %v", tt.name, tt.url, err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("GetBaseDomain() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsLikelyFile(t *testing.T) {
	tests := []struct {
		name string
		href string
		want bool
	}{
		{"pdf direct", "https://example.com/file.pdf", true},
		{"nested file", "https://example.com/a/b/c/image.png", true},
		{"uppercase ext", "https://example.com/report.PDF", true},
		{"multi-dot name", "https://example.com/very.long.name.pdf", true},
		{"composite ext", "https://example.com/archive.tar.gz", true},
		{"query after file", "https://example.com/file.pdf?download=1&token=x", true},
		{"fragment after file", "https://example.com/file.pdf#section", true},

		{"bare filename", "report.docx", true},
		{"relative path", "../docs/spec.pdf", true},
		{"relative with query", "download.zip?ref=2", true},

		{"root", "https://example.com/", false},
		{"no ext", "https://example.com/page", false},
		{"no ext with slash", "https://example.com/news/2024/", false},
		{"ext only in query", "https://example.com/search?q=file.pdf", false},
		{"query only", "https://example.com/download?id=5", false},
		{"empty", "", false},
		{"non-http scheme", "mailto:foo@bar.com", false},
		{"js scheme", "javascript:void(0)", false},
		{"hash only", "#top", false},

		{"trailing slash after file", "https://example.com/file.pdf/", true},
		{"hidden file", "https://example.com/.env", true},
		{"hidden dotfile", ".gitignore", true},
		{"trailing dot", "https://example.com/foo.", false},
		{"no-extension download", "https://example.com/download?id=5", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLikelyFile(tt.href); got != tt.want {
				t.Errorf("IsLikelyFile(%q) = %v, want %v", tt.href, got, tt.want)
			}
		})
	}
}

func TestIsValidHost(t *testing.T) {
	// IsValidHost accepts a bare hostname only: localhost, a literal IP, or a domain
	// whose trailing label is a real ICANN TLD. No scheme, no port, no path, no
	// query, no fragment. See the doc comment on IsValidHost.
	tests := []struct {
		name string
		host string
		want bool
	}{
		{"localhost", "localhost", true},

		{"basic domain", "example.com", true},
		{"subdomain", "sub.example.com", true},
		{"multi-level subdomain", "a.b.c.example.org", true},
		{"country-code tld", "example.co.uk", true},
		{"random tld", "example.sbs", true},
		{"tld that looks like a file extension", "example.mov", true},

		{"ipv4", "123.123.123.123", true},
		{"private ipv4", "192.168.1.1", true},
		{"ipv6", "2001:db8::1", true},
		{"loopback ipv6", "::1", true},

		{"empty", "", false},
		{"no tld", "example", false},
		{"bare tld", "com", false},
		{"non-icann tld", "foo.local", false},
		{"numeric-only label", "example.123", false},
		{"invalid characters", "exa&*$.com", false},
		{"hidden dotfile", ".gitignore", false},

		{"scheme", "http://example.com", false},
		{"scheme https", "https://example.com", false},
		{"scheme other", "mailto:foo@bar.com", false},
		{"scheme javascript", "javascript:void(0)", false},

		{"port", "example.com:8080", false},
		{"localhost with port", "localhost:8080", false},
		{"ipv4 with port", "123.123.123.123:8080", false},

		{"path", "example.com/archive.tar.gz", false},
		{"relative path", "../docs/spec.pdf", false},
		{"bare filename", "report.docx", false},
		{"query", "example.com/file.pdf?download=1", false},
		{"fragment", "example.com/file.pdf#section", false},
		{"hash only", "#top", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidHost(tt.host); got != tt.want {
				t.Errorf("IsValidHost(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}
