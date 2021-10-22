package debug

import (
	"github.com/rewardenv/reward/internal/commands"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "debug [command]",
	Short: "Launches debug enabled shell within current project environment",
	Long:  `Launches debug enabled shell within current project environment`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	// DisableFlagParsing: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := commands.EnvCheck()
		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return commands.DebugCmd(cmd, args)
	},
}

func init() {
}
