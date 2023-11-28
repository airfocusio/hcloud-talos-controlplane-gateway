package cmd

import (
	"github.com/spf13/cobra"
)

var (
	verbose bool
	dir     string
	rootCmd = &cobra.Command{
		Use:          "hcloud-talos-controlplane-gateway",
		SilenceUsage: true,
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(startCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
