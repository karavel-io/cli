package main

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/karavel-io/cli/pkg/action"
	"github.com/spf13/cobra"
)

func NewBootstrapCommand() *cobra.Command {
	defaultKubeConfig := ""
	if usr, err := user.Current(); err == nil {
		defaultKubeConfig = filepath.Join(usr.HomeDir, ".kube", "config")
	}

	var kubeconfig string
	cmd := &cobra.Command{
		Use:   "bootstrap [WORKDIR]",
		Short: "Applies rendered resources (only needed for bootstrapping the cluster) in a safe order",
		RunE: func(cmd *cobra.Command, args []string) error {
			var cwd string
			if len(args) > 0 {
				cwd = args[0]
			}
			if cwd == "" {
				d, err := os.Getwd()
				if err != nil {
					return err
				}
				cwd = d
			}
			cwd, err := filepath.Abs(cwd)
			if err != nil {
				return err
			}

			return action.Bootstrap(cmd.Context(), action.BootstrapParams{
				KustomizeDir: cwd,
				KubeConfig:   kubeconfig,
			})
		},
	}
	cmd.Flags().StringVarP(&kubeconfig, "kubeconfig", "k", defaultKubeConfig, "Specify .kubeconfig file")

	return cmd
}
