package cmd

import (
	"github.com/spf13/cobra"
)

var (
	healthcheckCmd = &cobra.Command{
		Use:   "healthcheck",
		RunE:  healthcheck,
		Short: "performs a sanity check of the environment and configuration",
	}
)

func init() {
	rootCmd.AddCommand(healthcheckCmd)
}

func healthcheck(cmd *cobra.Command, _ []string) error {
	info, err := env.App.Healthcheck(cmd.Context())
	printJson(info)
	return err
}
