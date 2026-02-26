package ncp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_FromFile(t *testing.T) {
	// Clear env vars so file loading is triggered
	origAK := os.Getenv("NCLOUD_ACCESS_KEY")
	origSK := os.Getenv("NCLOUD_SECRET_KEY")
	os.Unsetenv("NCLOUD_ACCESS_KEY")
	os.Unsetenv("NCLOUD_SECRET_KEY")
	defer func() {
		restoreEnv("NCLOUD_ACCESS_KEY", origAK)
		restoreEnv("NCLOUD_SECRET_KEY", origSK)
	}()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "configure")

	content := `[DEFAULT]
ncloud_access_key_id=file-access-key
ncloud_secret_access_key=file-secret-key
ncloud_api_url=https://custom.api.example.com
ncloud_region=KR

[finance]
ncloud_access_key_id=fin-access-key
ncloud_secret_access_key=fin-secret-key
ncloud_api_url=https://fin-ncloud.apigw.fin-ntruss.com
ncloud_region=KR
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		profile   string
		wantAK    string
		wantURL   string
		wantErr   bool
	}{
		{
			name:    "default profile",
			profile: "",
			wantAK:  "file-access-key",
			wantURL: "https://custom.api.example.com",
		},
		{
			name:    "finance profile",
			profile: "finance",
			wantAK:  "fin-access-key",
			wantURL: "https://fin-ncloud.apigw.fin-ntruss.com",
		},
		{
			name:    "nonexistent profile",
			profile: "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := loadFromFile(configPath, tt.profile)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if cfg.AccessKey != tt.wantAK {
				t.Errorf("AccessKey = %v, want %v", cfg.AccessKey, tt.wantAK)
			}
			if cfg.ApiURL != tt.wantURL {
				t.Errorf("ApiURL = %v, want %v", cfg.ApiURL, tt.wantURL)
			}
		})
	}
}

func TestLoadFromFile_DefaultApiURL(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "configure")

	content := `[DEFAULT]
ncloud_access_key_id=key
ncloud_secret_access_key=secret
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadFromFile(configPath, "")
	if err != nil {
		t.Fatalf("loadFromFile() error = %v", err)
	}

	if cfg.ApiURL != "https://ncloud.apigw.ntruss.com" {
		t.Errorf("ApiURL = %v, want default URL", cfg.ApiURL)
	}
}

func TestLoadFromFile_NoSectionHeader(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "configure")

	content := `ncloud_access_key_id = my-access-key
ncloud_secret_access_key = my-secret-key
ncloud_api_url = https://ncloud.apigw.ntruss.com
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadFromFile(configPath, "")
	if err != nil {
		t.Fatalf("loadFromFile() error = %v", err)
	}
	if cfg.AccessKey != "my-access-key" {
		t.Errorf("AccessKey = %v, want my-access-key", cfg.AccessKey)
	}
	if cfg.SecretKey != "my-secret-key" {
		t.Errorf("SecretKey = %v, want my-secret-key", cfg.SecretKey)
	}
	if cfg.ApiURL != "https://ncloud.apigw.ntruss.com" {
		t.Errorf("ApiURL = %v, want https://ncloud.apigw.ntruss.com", cfg.ApiURL)
	}
}

func TestLoadFromFile_IncompleteCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "configure")

	content := `[DEFAULT]
ncloud_access_key_id=key-only
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := loadFromFile(configPath, "")
	if err == nil {
		t.Error("expected error for incomplete credentials")
	}
}
