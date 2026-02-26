package kubeconfig

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")

	config := api.NewConfig()
	config.Clusters["test-cluster"] = &api.Cluster{
		Server: "https://test.example.com",
	}
	config.AuthInfos["test-user"] = &api.AuthInfo{}
	config.Contexts["test-context"] = &api.Context{
		Cluster:  "test-cluster",
		AuthInfo: "test-user",
	}
	config.CurrentContext = "test-context"

	if err := clientcmd.WriteToFile(*config, kubeconfigPath); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	origKubeconfig := os.Getenv("KUBECONFIG")
	os.Setenv("KUBECONFIG", kubeconfigPath)
	defer restoreEnv("KUBECONFIG", origKubeconfig)

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	if manager.path != kubeconfigPath {
		t.Errorf("path = %v, want %v", manager.path, kubeconfigPath)
	}
}

func TestNewManager_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "nonexistent", "config")

	origKubeconfig := os.Getenv("KUBECONFIG")
	os.Setenv("KUBECONFIG", kubeconfigPath)
	defer restoreEnv("KUBECONFIG", origKubeconfig)

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}
}

func TestManager_FindContext(t *testing.T) {
	manager := managerWithContexts(t, map[string]string{
		"my-cluster-user@my-cluster": "my-cluster",
	})

	tests := []struct {
		name        string
		clusterName string
		want        string
		wantErr     bool
	}{
		{
			name:        "exact context match",
			clusterName: "my-cluster-user@my-cluster",
			want:        "my-cluster-user@my-cluster",
		},
		{
			name:        "partial match",
			clusterName: "my-cluster",
			want:        "my-cluster-user@my-cluster",
		},
		{
			name:        "not found",
			clusterName: "nonexistent",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := manager.FindContext(tt.clusterName)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("FindContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_SwitchContext(t *testing.T) {
	manager := managerWithContexts(t, map[string]string{
		"ctx-a": "cluster-a",
		"ctx-b": "cluster-b",
	})

	if err := manager.SwitchContext("ctx-b"); err != nil {
		t.Fatalf("SwitchContext() error = %v", err)
	}

	if manager.GetCurrentContext() != "ctx-b" {
		t.Errorf("current context = %v, want ctx-b", manager.GetCurrentContext())
	}

	// Reload from disk to verify persistence
	reloaded, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() after switch error = %v", err)
	}
	if reloaded.GetCurrentContext() != "ctx-b" {
		t.Errorf("reloaded current context = %v, want ctx-b", reloaded.GetCurrentContext())
	}
}

func TestManager_SwitchContext_NotFound(t *testing.T) {
	manager := managerWithContexts(t, map[string]string{
		"ctx-a": "cluster-a",
	})

	if err := manager.SwitchContext("nonexistent"); err == nil {
		t.Error("SwitchContext() expected error for nonexistent context")
	}
}

func TestManager_ListContextNames(t *testing.T) {
	manager := managerWithContexts(t, map[string]string{
		"ctx-1": "cluster-1",
		"ctx-2": "cluster-2",
		"ctx-3": "cluster-3",
	})

	names := manager.ListContextNames()
	if len(names) != 3 {
		t.Errorf("ListContextNames() count = %v, want 3", len(names))
	}
}

func TestManager_FindContextByCluster(t *testing.T) {
	manager := managerWithContexts(t, map[string]string{
		"my-ctx": "my-nks-cluster",
	})

	got := manager.FindContextByCluster("my-nks-cluster")
	if got != "my-ctx" {
		t.Errorf("FindContextByCluster() = %v, want my-ctx", got)
	}

	got = manager.FindContextByCluster("nonexistent")
	if got != "" {
		t.Errorf("FindContextByCluster(nonexistent) = %v, want empty", got)
	}
}

// managerWithContexts creates a Manager backed by a temporary kubeconfig file.
func managerWithContexts(t *testing.T, contexts map[string]string) *Manager {
	t.Helper()

	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")

	config := api.NewConfig()
	for ctxName, clusterName := range contexts {
		config.Clusters[clusterName] = &api.Cluster{Server: "https://" + clusterName + ".example.com"}
		config.AuthInfos[ctxName+"-user"] = &api.AuthInfo{}
		config.Contexts[ctxName] = &api.Context{
			Cluster:  clusterName,
			AuthInfo: ctxName + "-user",
		}
	}

	if err := clientcmd.WriteToFile(*config, kubeconfigPath); err != nil {
		t.Fatalf("write config: %v", err)
	}

	origKubeconfig := os.Getenv("KUBECONFIG")
	os.Setenv("KUBECONFIG", kubeconfigPath)
	t.Cleanup(func() { restoreEnv("KUBECONFIG", origKubeconfig) })

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager(): %v", err)
	}
	return manager
}

func restoreEnv(key, value string) {
	if value != "" {
		os.Setenv(key, value)
	} else {
		os.Unsetenv(key)
	}
}
