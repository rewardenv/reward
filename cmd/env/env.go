package env

import (
	"github.com/rewardenv/reward/internal/commands"
	"github.com/rewardenv/reward/internal/core"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:                "env",
	Short:              "Controls an environment from any point within the root project directory",
	Long:               `Controls an environment from any point within the root project directory`,
	ValidArgsFunction:  core.DockerComposeCompleter(),
	DisableFlagParsing: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// If the command is 'env config' then skip docker api check.
		if !core.ContainsString([]string{args[0]}, "config") {
			if err := core.CheckDocker(); err != nil {
				return err
			}
		}

		if err := commands.EnvCheck(); err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return commands.EnvCmd(args)
	},
}

func init() {
}
