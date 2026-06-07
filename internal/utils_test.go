package internal

import "testing"

var SHOULD_ERR = true
var SHOULD_NOT_ERR = false

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
