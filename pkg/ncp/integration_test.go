//go:build integration

package ncp

import (
	"testing"
)

// TestClient_ListClusters_Integration is an integration test that calls the actual NCP API.
// To run:
//
//	go test -tags=integration ./pkg/ncp
//
// Environment variables must be set:
//   - NCLOUD_ACCESS_KEY
//   - NCLOUD_SECRET_KEY
//   - NCLOUD_API_GW (optional)
func TestClient_ListClusters_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg, err := LoadConfig("")
	if err != nil {
		t.Skip("Skipping integration test: NCP credentials not configured")
	}

	client := NewClientFromConfig(cfg)

	clusters, err := client.ListClusters()
	if err != nil {
		t.Fatalf("ListClusters() error = %v", err)
	}

	t.Logf("Found %d cluster(s)", len(clusters))
	for _, cluster := range clusters {
		t.Logf("  - %s (%s) in %s - %s", cluster.Name, cluster.UUID, cluster.Region, cluster.Status)
	}
}
