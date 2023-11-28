package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/airfocusio/hcloud-talos-controlplane-gateway/internal"
	"github.com/spf13/cobra"
)

var (
	startCmdClusterName  string
	startCmdFirewallName string
	startCmd             = &cobra.Command{
		Use:  "start",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := internal.NewLogger(verbose)
			opts := internal.ServiceOpts{
				HcloudToken:  os.Getenv("HCLOUD_TOKEN"),
				ClusterName:  startCmdClusterName,
				FirewallName: startCmdFirewallName,
			}
			ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			service, err := internal.NewService(ctx, logger, opts)
			if err != nil {
				return err
			}
			return service.Run()
		},
	}
)

func init() {
	startCmd.Flags().StringVarP(&startCmdClusterName, "cluster-name", "c", "", "")
	startCmd.Flags().StringVarP(&startCmdFirewallName, "firewall-name", "f", "", "")
}
