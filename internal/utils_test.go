package internal

import "testing"

func TestValidateURLFormat(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"Valid HTTP URL", "http://example.com", false},
		{"Valid HTTPS URL", "https://example.com", false},
		{"Valid URL with port", "http://example.com:8080", false},
		{"Validate URL without scheme", "www.example.com", false},
		{"Valid URL without scheme", "example.com", false},
		{"Valid localhost URL", "http://localhost:8080", false},
		{"Valid localhost URL without port", "http://localhost", false},
		{"Valid localhost URL without scheme", "localhost:8080", false},
		{"Valid localhost URL without scheme and port", "localhost", false},
		{"Valid IP address URL", "192.168.1.1", false},
		{"Valid IP address URL with scheme", "http://192.168.1.1", false},
		{"Valid IP address URL with scheme and port", "http://192.168.1.1:8080", false},
		{"Valid IP address URL with port", "192.168.1.1:8080", false},
		{"Empty URL", "", true},
		{"Invalid URL format", "http://example", true},
		{"Invalid URL format with spaces", "example", true},
		{"Invalid URL with special characters", "http://exa&*$.com", true},
		{"Invalid URL with unsupported scheme", "ftp://example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAndValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURLFormat() test=%q url=%q error = %v, wantErr %v", tt.name, tt.url, err, tt.wantErr)
				return
			}
			if got != tt.url && !tt.wantErr {
				t.Errorf("ValidateURLFormat() got = %v, want %v", got, tt.url)
			}
		})
	}
}
