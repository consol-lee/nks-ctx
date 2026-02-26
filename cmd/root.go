package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/consol-lee/nks-ctx/pkg/kubeconfig"
	"github.com/consol-lee/nks-ctx/pkg/ncp"
)

var profileFlag string

var rootCmd = &cobra.Command{
	Use:   "kubectl-nks-ctx [cluster-name]",
	Short: "Manage NKS (Ncloud Kubernetes Service) cluster contexts",
	Long: `kubectl plugin for managing NKS (Ncloud Kubernetes Service) cluster contexts.

Run without arguments to sync all NKS clusters to kubeconfig and display the list.
Run with a cluster name to switch to that cluster's context.

Examples:
  # Sync clusters and show list
  kubectl nks-ctx

  # Switch to a specific cluster
  kubectl nks-ctx my-cluster

  # Use a specific NCP profile
  kubectl nks-ctx --profile finance`,
	Args:          cobra.MaximumNArgs(1),
	RunE:          run,
	SilenceUsage:  true,
	SilenceErrors: true,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		manager, err := kubeconfig.NewManager()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var matches []string
		for _, name := range manager.ListClusterNames() {
			if strings.HasPrefix(name, toComplete) {
				matches = append(matches, name)
			}
		}
		return matches, cobra.ShellCompDirectiveNoFileComp
	},
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}
	return nil
}

func init() {
	rootCmd.Flags().StringVarP(&profileFlag, "profile", "p", "", "NCP profile name (default: DEFAULT)")
}

func run(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return runSync()
	}
	return runSwitch(args[0])
}

// runSync fetches all NKS clusters, generates kubeconfig entries via
// ncp-iam-authenticator, and displays the cluster list.
func runSync() error {
	cfg, err := ncp.LoadConfig(profileFlag)
	if err != nil {
		return err
	}

	// Verify ncp-iam-authenticator is available
	authenticator := ncp.NewAuthenticator(profileFlag)
	if !authenticator.IsInstalled() {
		return fmt.Errorf(
			"ncp-iam-authenticator not found.\n" +
				"Install it from: https://guide.ncloud-docs.com/docs/nks-nkstoken",
		)
	}

	client := ncp.NewClientFromConfig(cfg)

	clusters, err := client.ListClusters()
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	if len(clusters) == 0 {
		fmt.Println("No clusters found.")
		return nil
	}

	// Load kubeconfig to check existing entries
	kubeconfigPath := kubeconfig.DefaultPath()
	manager, err := kubeconfig.NewManager()
	if err != nil {
		return fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	// Sync each cluster to kubeconfig (skip if already exists)
	syncCount := 0
	skipCount := 0
	for _, cluster := range clusters {
		if ctxName := manager.FindContextByCluster(cluster.Name); ctxName != "" {
			skipCount++
			continue
		}
		if err := authenticator.UpdateKubeconfig(cluster, kubeconfigPath, false); err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: failed to sync %s: %v\n", cluster.Name, err)
			continue
		}
		syncCount++
	}

	if syncCount > 0 || skipCount > 0 {
		fmt.Printf("Synced %d cluster(s), skipped %d already configured. (%d total)\n\n", syncCount, skipCount, len(clusters))
	}

	// Reload kubeconfig if new clusters were synced
	if syncCount > 0 {
		manager, err = kubeconfig.NewManager()
		if err != nil {
			return fmt.Errorf("failed to read kubeconfig: %w", err)
		}
	}

	current := manager.GetCurrentContext()
	for _, cluster := range clusters {
		ctxName := manager.FindContextByCluster(cluster.Name)
		marker := "  "
		if ctxName != "" && ctxName == current {
			marker = "* "
		}
		fmt.Printf("%s%s\n", marker, cluster.Name)
	}

	return nil
}

// runSwitch changes the current kubeconfig context to the specified cluster.
func runSwitch(clusterName string) error {
	manager, err := kubeconfig.NewManager()
	if err != nil {
		return fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	contextName, err := manager.FindContext(clusterName)
	if err != nil {
		return fmt.Errorf(
			"context not found for '%s'.\nRun 'kubectl nks-ctx' first to sync clusters.",
			clusterName,
		)
	}

	if err := manager.SwitchContext(contextName); err != nil {
		return fmt.Errorf("failed to switch context: %w", err)
	}

	fmt.Printf("Switched to context \"%s\"\n", contextName)
	return nil
}
