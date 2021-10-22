package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"github.com/kolesa-team/scylla-octopus/pkg/environment"
	"os"
)

var (
	env              environment.Environment
	configPath       string
	forceVerboseMode bool
	rootCmd          = &cobra.Command{
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, _ []string) {
			_ = cmd.Help()
		},
		PersistentPreRunE: initialize,
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			_ = env.Logger.Sync()
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(
		&configPath,
		"config",
		"config/remote.yml",
		"configuration file path",
	)
	rootCmd.PersistentFlags().BoolVarP(
		&forceVerboseMode,
		"verbose",
		"v",
		false,
		"verbose logging",
	)
}

func Execute(ctx context.Context) {
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}

// reads configuration file, performs basic validation, and initializes project components before each command.
func initialize(cmd *cobra.Command, args []string) error {
	var err error

	config, err := environment.GetConfig(configPath, forceVerboseMode)
	if err != nil {
		return err
	}

	env, err = environment.GetEnvironment(config, entity.BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    buildDate,
	})
	if err != nil && env.Notifier != nil {
		env.Notifier.Error(
			"could not initialize project environment",
			"",
			err,
			map[string]interface{}{"configPath": configPath},
		)
	}

	return err
}

func printJson(data interface{}) {
	dataJson, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(dataJson))
}
