package cmd

import (
	"github.com/spf13/cobra"

	. "reward/internal"
)

var envCmd = &cobra.Command{
	Use:                "env",
	Short:              "Controls an environment from any point within the root project directory",
	Long:               `Controls an environment from any point within the root project directory`,
	ValidArgsFunction:  DockerComposeCompleter(),
	DisableFlagParsing: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := CheckDockerIsRunning(); err != nil {
			return err
		}

		if err := EnvCheck(); err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return EnvCmd(args)
	},
}

func init() {
	rootCmd.AddCommand(envCmd)
}
