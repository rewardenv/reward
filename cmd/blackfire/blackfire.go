package blackfire

import (
	"fmt"
	"strings"

	"github.com/rewardenv/reward/internal/commands"
	"github.com/rewardenv/reward/internal/core"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use: "blackfire [command]",
	Short: fmt.Sprintf(
		"Interacts with the blackfire service on an environment (disabled if %v_BLACKFIRE is not 1)",
		strings.ToUpper(core.AppName),
	),
	Long: fmt.Sprintf(
		`Interacts with the blackfire service on an environment (disabled if %v_BLACKFIRE is not 1)`,
		strings.ToUpper(core.AppName),
	),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := core.CheckDocker(); err != nil {
			return err
		}

		if err := commands.EnvCheck(); err != nil {
			return err
		}

		isContainerRunning, err := core.IsContainerRunning(core.GetBlackfireContainer())
		if err != nil {
			return err
		}
		if !core.IsDBEnabled() || !isContainerRunning {
			return core.CannotFindContainerError(core.GetBlackfireContainer())
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return commands.BlackfireCmd(cmd, args)
	},
}

func init() {
}
