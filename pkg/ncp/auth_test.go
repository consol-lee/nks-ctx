package ncp

import (
	"testing"
)

func TestGenerateHMACSignature(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		uri        string
		query      string
		timestamp  string
		accessKey  string
		secretKey  string
		wantLength int
	}{
		{
			name:       "basic signature",
			method:     "GET",
			uri:        "/v1/cluster",
			query:      "",
			timestamp:  "1234567890",
			accessKey:  "test-key",
			secretKey:  "test-secret",
			wantLength: 44, // Base64 encoded SHA256 HMAC is 44 chars
		},
		{
			name:       "signature with query",
			method:     "GET",
			uri:        "/v1/cluster",
			query:      "regionCode=KR",
			timestamp:  "1234567890",
			accessKey:  "test-key",
			secretKey:  "test-secret",
			wantLength: 44,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := GenerateHMACSignature(tt.method, tt.uri, tt.query, tt.timestamp, tt.accessKey, tt.secretKey)
			if len(sig) != tt.wantLength {
				t.Errorf("GenerateHMACSignature() length = %d, want %d", len(sig), tt.wantLength)
			}
			if sig == "" {
				t.Error("GenerateHMACSignature() returned empty string")
			}
		})
	}
}

func TestExtractURI(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "full URL with path",
			url:  "https://fin-ncloud.apigw.fin-ntruss.com/v1/cluster",
			want: "/v1/cluster",
		},
		{
			name: "URL with query string",
			url:  "https://fin-ncloud.apigw.fin-ntruss.com/v1/cluster?regionCode=KR",
			want: "/v1/cluster",
		},
		{
			name: "path only",
			url:  "/v1/cluster",
			want: "/v1/cluster",
		},
		{
			name: "root path",
			url:  "https://example.com/",
			want: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractURI(tt.url)
			if got != tt.want {
				t.Errorf("ExtractURI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractQueryString(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "URL with query",
			url:  "https://example.com/path?key=value&region=KR",
			want: "key=value&region=KR",
		},
		{
			name: "URL without query",
			url:  "https://example.com/path",
			want: "",
		},
		{
			name: "empty query",
			url:  "https://example.com/path?",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractQueryString(tt.url)
			if got != tt.want {
				t.Errorf("ExtractQueryString() = %v, want %v", got, tt.want)
			}
		})
	}
}
