package cmd

import (
	reward "github.com/rewardenv/reward/internal"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:                "env",
	Short:              "Controls an environment from any point within the root project directory",
	Long:               `Controls an environment from any point within the root project directory`,
	ValidArgsFunction:  reward.DockerComposeCompleter(),
	DisableFlagParsing: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// If the command is 'env config' then skip docker api check.
		if !reward.ContainsString([]string{args[0]}, "config") {
			if err := reward.CheckDocker(); err != nil {
				return err
			}
		}

		if err := reward.EnvCheck(); err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return reward.EnvCmd(args)
	},
}

func init() {
	rootCmd.AddCommand(envCmd)
}
