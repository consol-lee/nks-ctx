package kubeconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Manager handles loading, querying, and modifying kubeconfig.
type Manager struct {
	path   string
	config *api.Config
}

// DefaultPath returns the default kubeconfig path (~/.kube/config),
// or the value of KUBECONFIG if set.
func DefaultPath() string {
	if p := os.Getenv("KUBECONFIG"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kube", "config")
}

// NewManager loads the kubeconfig from the default path.
func NewManager() (*Manager, error) {
	path := DefaultPath()

	config, err := clientcmd.LoadFromFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			config = api.NewConfig()
		} else {
			return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
		}
	}

	return &Manager{path: path, config: config}, nil
}

// GetCurrentContext returns the name of the currently active context.
func (m *Manager) GetCurrentContext() string {
	return m.config.CurrentContext
}

// FindContext finds a context by exact name or partial match.
func (m *Manager) FindContext(name string) (string, error) {
	// Exact match
	if _, ok := m.config.Contexts[name]; ok {
		return name, nil
	}

	// Partial match
	for ctxName := range m.config.Contexts {
		if strings.Contains(ctxName, name) {
			return ctxName, nil
		}
	}

	return "", fmt.Errorf("context not found: %s", name)
}

// FindContextByCluster returns the context name associated with a cluster name.
func (m *Manager) FindContextByCluster(clusterName string) string {
	for ctxName, ctx := range m.config.Contexts {
		if ctx.Cluster == clusterName || strings.Contains(ctxName, clusterName) {
			return ctxName
		}
	}
	return ""
}

// SwitchContext sets the current-context and writes kubeconfig to disk.
func (m *Manager) SwitchContext(contextName string) error {
	if _, ok := m.config.Contexts[contextName]; !ok {
		return fmt.Errorf("context '%s' not found in kubeconfig", contextName)
	}

	m.config.CurrentContext = contextName

	return clientcmd.WriteToFile(*m.config, m.path)
}

// ListContextNames returns all context names in kubeconfig.
func (m *Manager) ListContextNames() []string {
	names := make([]string, 0, len(m.config.Contexts))
	for name := range m.config.Contexts {
		names = append(names, name)
	}
	return names
}

// ListClusterNames returns all cluster names referenced by contexts in kubeconfig.
func (m *Manager) ListClusterNames() []string {
	seen := make(map[string]bool)
	var names []string
	for _, ctx := range m.config.Contexts {
		if ctx.Cluster != "" && !seen[ctx.Cluster] {
			seen[ctx.Cluster] = true
			names = append(names, ctx.Cluster)
		}
	}
	return names
}
