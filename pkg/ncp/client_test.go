package ncp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestLoadConfig_FromEnv(t *testing.T) {
	origAK := os.Getenv("NCLOUD_ACCESS_KEY")
	origSK := os.Getenv("NCLOUD_SECRET_KEY")
	origGW := os.Getenv("NCLOUD_API_GW")
	defer func() {
		restoreEnv("NCLOUD_ACCESS_KEY", origAK)
		restoreEnv("NCLOUD_SECRET_KEY", origSK)
		restoreEnv("NCLOUD_API_GW", origGW)
	}()

	t.Run("env takes precedence", func(t *testing.T) {
		os.Setenv("NCLOUD_ACCESS_KEY", "env-ak")
		os.Setenv("NCLOUD_SECRET_KEY", "env-sk")
		os.Setenv("NCLOUD_API_GW", "https://custom.example.com")

		cfg, err := LoadConfig("")
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}
		if cfg.AccessKey != "env-ak" {
			t.Errorf("AccessKey = %v, want env-ak", cfg.AccessKey)
		}
		if cfg.SecretKey != "env-sk" {
			t.Errorf("SecretKey = %v, want env-sk", cfg.SecretKey)
		}
		if cfg.ApiURL != "https://custom.example.com" {
			t.Errorf("ApiURL = %v, want https://custom.example.com", cfg.ApiURL)
		}
	})

	t.Run("default api url when env omits it", func(t *testing.T) {
		os.Setenv("NCLOUD_ACCESS_KEY", "env-ak")
		os.Setenv("NCLOUD_SECRET_KEY", "env-sk")
		os.Unsetenv("NCLOUD_API_GW")

		cfg, err := LoadConfig("")
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}
		if cfg.ApiURL != "https://ncloud.apigw.ntruss.com" {
			t.Errorf("ApiURL = %v, want default", cfg.ApiURL)
		}
	})

	t.Run("falls back to file when env incomplete", func(t *testing.T) {
		os.Setenv("NCLOUD_ACCESS_KEY", "only-ak")
		os.Unsetenv("NCLOUD_SECRET_KEY")
		os.Unsetenv("NCLOUD_API_GW")

		cfg, err := LoadConfig("")
		// If ~/.ncloud/configure exists with valid DEFAULT, fallback succeeds
		if err != nil {
			t.Skipf("skipping: no config file available for fallback (%v)", err)
		}
		if cfg.AccessKey == "" || cfg.SecretKey == "" {
			t.Errorf("expected valid credentials from file fallback")
		}
	})
}

func TestNewClientFromConfig(t *testing.T) {
	cfg := &Config{
		AccessKey: "ak",
		SecretKey: "sk",
		ApiURL:    "https://ncloud.apigw.ntruss.com",
	}
	client := NewClientFromConfig(cfg)
	if client.accessKey != "ak" {
		t.Errorf("accessKey = %v, want ak", client.accessKey)
	}
	if client.secretKey != "sk" {
		t.Errorf("secretKey = %v, want sk", client.secretKey)
	}
	if client.apiGw != "https://ncloud.apigw.ntruss.com" {
		t.Errorf("apiGw = %v, want https://ncloud.apigw.ntruss.com", client.apiGw)
	}
	if len(client.nksBaseURLs) != 3 {
		t.Errorf("nksBaseURLs count = %d, want 3", len(client.nksBaseURLs))
	}
}

func TestResolveNKSBaseURLs(t *testing.T) {
	tests := []struct {
		apiURL    string
		wantCount int
		wantFirst string
	}{
		{"https://ncloud.apigw.ntruss.com", 3, "https://nks.apigw.ntruss.com/vnks/v2"},
		{"https://fin-ncloud.apigw.fin-ntruss.com", 1, "https://nks.apigw.fin-ntruss.com/nks/v2"},
		{"https://ncloud.apigw.gov-ntruss.com", 2, "https://nks.apigw.gov-ntruss.com/vnks/v2"},
	}
	for _, tt := range tests {
		got := resolveNKSBaseURLs(tt.apiURL)
		if len(got) != tt.wantCount {
			t.Errorf("resolveNKSBaseURLs(%q) count = %d, want %d", tt.apiURL, len(got), tt.wantCount)
		}
		if got[0] != tt.wantFirst {
			t.Errorf("resolveNKSBaseURLs(%q)[0] = %q, want %q", tt.apiURL, got[0], tt.wantFirst)
		}
	}
}

func TestClient_ListClusters(t *testing.T) {
	mockResponse := clusterListResponse{
		Clusters: []clusterInfo{
			{
				UUID:       "test-uuid-1",
				Name:       "test-cluster-1",
				RegionCode: "KR",
				Status:     "RUNNING",
			},
			{
				UUID:       "test-uuid-2",
				Name:       "test-cluster-2",
				RegionCode: "JP",
				Status:     "RUNNING",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.Header.Get("x-ncp-iam-access-key") == "" {
			t.Error("Missing x-ncp-iam-access-key header")
		}
		if r.Header.Get("x-ncp-apigw-signature-v2") == "" {
			t.Error("Missing x-ncp-apigw-signature-v2 header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := &Client{
		accessKey:   "test-access-key",
		secretKey:   "test-secret-key",
		apiGw:       server.URL,
		nksBaseURLs: []string{server.URL},
	}

	clusters, err := client.ListClusters()
	if err != nil {
		t.Fatalf("ListClusters() error = %v", err)
	}
	if len(clusters) != 2 {
		t.Errorf("ListClusters() count = %v, want 2", len(clusters))
	}
}

func TestClient_ListClusters_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := &Client{
		accessKey:   "test-access-key",
		secretKey:   "test-secret-key",
		apiGw:       server.URL,
		nksBaseURLs: []string{server.URL},
	}

	_, err := client.ListClusters()
	if err == nil {
		t.Error("ListClusters() expected error when all endpoints fail, got nil")
	}
}

func restoreEnv(key, value string) {
	if value != "" {
		os.Setenv(key, value)
	} else {
		os.Unsetenv(key)
	}
}
