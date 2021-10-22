package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	dbCmd = &cobra.Command{
		Use:   "db",
		Short: "database commands",
	}
	dbListSnapshotsCmd = &cobra.Command{
		Use:   "list-snapshots",
		Short: "prints a list of existing snapshots on database nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := env.App.Healthcheck(cmd.Context())
			if err != nil {
				return err
			}

			snapshots, err := env.App.ListSnapshots(cmd.Context())
			printJson(snapshots)

			return err
		},
	}
	dbRepairCmd = &cobra.Command{
		Use:   "repair",
		Short: "executes 'nodetool repair -pr' on database nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := env.App.Healthcheck(cmd.Context())
			if err != nil {
				env.Notifier.Error(
					"Could not perform a healthcheck before running repair.",
					"",
					err,
					nil,
				)

				return err
			}

			results := env.App.Repair(cmd.Context())
			fmt.Println(results.Report())

			return results.Error
		},
	}
)

func init() {
	dbCmd.AddCommand(dbListSnapshotsCmd)
	dbCmd.AddCommand(dbRepairCmd)
	rootCmd.AddCommand(dbCmd)
}
