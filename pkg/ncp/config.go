package ncp

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds NCP credentials and API configuration.
type Config struct {
	AccessKey string
	SecretKey string
	ApiURL    string
	Region    string
}

// LoadConfig loads NCP configuration from environment variables or ~/.ncloud/configure.
// Environment variables take precedence over the config file.
// If profile is empty, "DEFAULT" is used.
func LoadConfig(profile string) (*Config, error) {
	cfg := &Config{
		AccessKey: os.Getenv("NCLOUD_ACCESS_KEY"),
		SecretKey: os.Getenv("NCLOUD_SECRET_KEY"),
		ApiURL:    os.Getenv("NCLOUD_API_GW"),
		Region:    os.Getenv("NCLOUD_REGION"),
	}

	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		if cfg.ApiURL == "" {
			cfg.ApiURL = defaultAPIURL()
		}
		return cfg, nil
	}

	fileCfg, err := loadFromFile(configFilePath(), profile)
	if err != nil {
		return nil, fmt.Errorf(
			"NCP credentials not found.\n\n" +
				"Set environment variables:\n" +
				"  export NCLOUD_ACCESS_KEY=\"your-access-key\"\n" +
				"  export NCLOUD_SECRET_KEY=\"your-secret-key\"\n" +
				"  export NCLOUD_API_GW=\"https://ncloud.apigw.ntruss.com\"  # optional\n\n" +
				"Or configure ~/.ncloud/configure with a profile.",
		)
	}

	return fileCfg, nil
}

func defaultAPIURL() string {
	return "https://ncloud.apigw.ntruss.com"
}

func configFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ncloud", "configure")
}

// loadFromFile parses the INI-style ~/.ncloud/configure file.
//
// Expected format:
//
//	[DEFAULT]
//	ncloud_access_key_id=YOUR_ACCESS_KEY
//	ncloud_secret_access_key=YOUR_SECRET_KEY
//	ncloud_api_url=https://ncloud.apigw.ntruss.com
//	ncloud_region=KR
func loadFromFile(path, profile string) (*Config, error) {
	if profile == "" {
		profile = "DEFAULT"
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	sections := make(map[string]map[string]string)
	currentSection := "DEFAULT"
	sections[currentSection] = make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.TrimSpace(line[1 : len(line)-1])
			if sections[currentSection] == nil {
				sections[currentSection] = make(map[string]string)
			}
			continue
		}

		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			sections[currentSection][key] = value
		}
	}

	data, ok := sections[profile]
	if !ok {
		return nil, fmt.Errorf("profile '%s' not found in %s", profile, path)
	}

	cfg := &Config{
		AccessKey: data["ncloud_access_key_id"],
		SecretKey: data["ncloud_secret_access_key"],
		ApiURL:    data["ncloud_api_url"],
		Region:    data["ncloud_region"],
	}

	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("incomplete credentials in profile '%s'", profile)
	}

	if cfg.ApiURL == "" {
		cfg.ApiURL = defaultAPIURL()
	}

	return cfg, nil
}
