package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	backupCmd = &cobra.Command{
		Use:   "backup",
		Short: "backup-related commands",
	}
	backupRunCmd = &cobra.Command{
		Use:   "run",
		Short: "runs a backup (exports database schema and snapshot, uploads to remote storage, cleans up)",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := env.App.Healthcheck(cmd.Context())
			if err != nil {
				env.Notifier.Error(
					"Could not perform a healthcheck before creating backups.",
					"",
					err,
					nil,
				)
				return err
			}

			result := env.App.Backup(cmd.Context())
			fmt.Println(result.Report())

			return result.Error
		},
	}
	backupCleanupExpired = &cobra.Command{
		Use:   "cleanup-expired",
		Short: "removes expired backups from remote storage",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := env.App.Healthcheck(cmd.Context())
			if err != nil {
				return err
			}

			expired, err := env.App.CleanupExpiredBackups(cmd.Context())
			printJson(expired)

			return err
		},
	}
	backupListExpired = &cobra.Command{
		Use:   "list-expired",
		Short: "prints a list of expired backups in remote storage that can be removed",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := env.App.Healthcheck(cmd.Context())
			if err != nil {
				return err
			}

			expired, err := env.App.ListExpiredBackups(cmd.Context())
			printJson(expired)

			return err
		},
	}
	backupList = &cobra.Command{
		Use:   "list",
		Short: "prints a list of existing backups in remote storage",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := env.App.Healthcheck(cmd.Context())
			if err != nil {
				return err
			}

			expired, err := env.App.ListBackups(cmd.Context())
			printJson(expired)

			return err
		},
	}
)

func init() {
	backupCmd.AddCommand(backupRunCmd)
	backupCmd.AddCommand(backupCleanupExpired)
	backupCmd.AddCommand(backupList)
	backupCmd.AddCommand(backupListExpired)
	rootCmd.AddCommand(backupCmd)
}
