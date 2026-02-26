package ncp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Client communicates with the NCP NKS API.
type Client struct {
	accessKey   string
	secretKey   string
	apiGw       string   // ncloud API URL (for auth/signing)
	nksBaseURLs []string // NKS API base URLs per region
}

// Cluster represents an NKS cluster.
type Cluster struct {
	UUID   string
	Name   string
	Region string
	Status string
}

type clusterListResponse struct {
	Clusters []clusterInfo `json:"clusters"`
}

type clusterInfo struct {
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	RegionCode string `json:"regionCode"`
	Status     string `json:"status"`
}

// resolveNKSBaseURLs derives all regional NKS API base URLs from the ncloud API URL.
//
// Mapping:
//
//	ncloud.apigw.ntruss.com         → nks.apigw.ntruss.com/vnks/{v2, sgn-v2, jpn-v2}
//	fin-ncloud.apigw.fin-ntruss.com → nks.apigw.fin-ntruss.com/nks/v2
//	ncloud.apigw.gov-ntruss.com     → nks.apigw.gov-ntruss.com/vnks/{v2, krs-v2}
func resolveNKSBaseURLs(apiURL string) []string {
	if strings.Contains(apiURL, "fin-ntruss.com") {
		return []string{
			"https://nks.apigw.fin-ntruss.com/nks/v2",
		}
	}
	if strings.Contains(apiURL, "gov-ntruss.com") {
		return []string{
			"https://nks.apigw.gov-ntruss.com/vnks/v2",
			"https://nks.apigw.gov-ntruss.com/vnks/krs-v2",
		}
	}
	return []string{
		"https://nks.apigw.ntruss.com/vnks/v2",
		"https://nks.apigw.ntruss.com/vnks/sgn-v2",
		"https://nks.apigw.ntruss.com/vnks/jpn-v2",
	}
}

// NewClientFromConfig creates an NCP client from a Config.
func NewClientFromConfig(cfg *Config) *Client {
	return &Client{
		accessKey:   cfg.AccessKey,
		secretKey:   cfg.SecretKey,
		apiGw:       cfg.ApiURL,
		nksBaseURLs: resolveNKSBaseURLs(cfg.ApiURL),
	}
}

// ListClusters retrieves clusters from all regional NKS API endpoints.
// If some endpoints fail but others succeed, warnings are printed and partial results are returned.
// If all endpoints fail, an error is returned.
func (c *Client) ListClusters() ([]Cluster, error) {
	var allClusters []Cluster
	var errors []string
	successCount := 0

	for _, baseURL := range c.nksBaseURLs {
		clusters, err := c.listClustersFromEndpoint(baseURL)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", baseURL, err))
			continue
		}
		successCount++
		allClusters = append(allClusters, clusters...)
	}

	if successCount == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("all API endpoints failed:\n  %s", strings.Join(errors, "\n  "))
	}

	for _, e := range errors {
		fmt.Fprintf(os.Stderr, "  Warning: %s\n", e)
	}

	return allClusters, nil
}

func (c *Client) listClustersFromEndpoint(baseURL string) ([]Cluster, error) {
	url := fmt.Sprintf("%s/clusters", baseURL)
	uri := ExtractURI(url)
	method := "GET"

	headers, err := c.PrepareAuthHeaders(method, uri, "")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare auth headers: %w", err)
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var listResp clusterListResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w (body: %s)", err, string(body))
	}

	clusters := make([]Cluster, 0, len(listResp.Clusters))
	for _, info := range listResp.Clusters {
		clusters = append(clusters, Cluster{
			UUID:   info.UUID,
			Name:   info.Name,
			Region: info.RegionCode,
			Status: info.Status,
		})
	}

	return clusters, nil
}
