package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	commit    = "dev"
	buildDate = ""
)

func init() {
	versionCmd := &cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Version:", version)
			fmt.Println("Commit:", commit)
			fmt.Println("Build date:", buildDate)
		},
		Short: "prints program version",
	}

	rootCmd.AddCommand(versionCmd)
}
