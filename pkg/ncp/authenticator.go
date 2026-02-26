package ncp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Authenticator wraps the ncp-iam-authenticator binary.
type Authenticator struct {
	binaryPath string
	profile    string
}

// NewAuthenticator creates an Authenticator, locating the ncp-iam-authenticator binary.
func NewAuthenticator(profile string) *Authenticator {
	return &Authenticator{
		binaryPath: findBinary(),
		profile:    profile,
	}
}

func findBinary() string {
	candidates := []string{
		"/usr/local/bin/ncp-iam-authenticator",
		"/opt/homebrew/bin/ncp-iam-authenticator",
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	if p, err := exec.LookPath("ncp-iam-authenticator"); err == nil {
		return p
	}

	return "ncp-iam-authenticator"
}

// IsInstalled checks if ncp-iam-authenticator is available.
func (a *Authenticator) IsInstalled() bool {
	_, err := exec.LookPath(a.binaryPath)
	if err != nil {
		_, err = os.Stat(a.binaryPath)
	}
	return err == nil
}

// UpdateKubeconfig runs ncp-iam-authenticator to add/update a cluster entry in kubeconfig.
func (a *Authenticator) UpdateKubeconfig(cluster Cluster, kubeconfigPath string, overwrite bool) error {
	dir := filepath.Dir(kubeconfigPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create kubeconfig directory: %w", err)
	}

	args := []string{
		"update-kubeconfig",
		"--region", cluster.Region,
		"--clusterUuid", cluster.UUID,
		"--clusterName", cluster.Name,
		"--kubeconfig", kubeconfigPath,
	}

	if a.profile != "" {
		args = append(args, "--profile", a.profile)
	}

	if overwrite {
		args = append(args, "--overwrite")
	}

	cmd := exec.Command(a.binaryPath, args...)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ncp-iam-authenticator failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}
