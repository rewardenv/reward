package cmd

import (
	"fmt"
	"strings"

	reward "github.com/rewardenv/reward/internal"
	"github.com/spf13/cobra"
)

var blackfireCmd = &cobra.Command{
	Use: "blackfire [command]",
	Short: fmt.Sprintf(
		"Interacts with the blackfire service on an environment (disabled if %v_BLACKFIRE is not 1)",
		strings.ToUpper(reward.AppName)),
	Long: fmt.Sprintf(
		`Interacts with the blackfire service on an environment (disabled if %v_BLACKFIRE is not 1)`,
		strings.ToUpper(reward.AppName)),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := reward.CheckDocker(); err != nil {
			return err
		}

		if err := reward.EnvCheck(); err != nil {
			return err
		}

		if !reward.IsBlackfireEnabled() || !reward.IsContainerRunning(reward.GetBlackfireContainer()) {
			return reward.CannotFindContainerError(reward.GetBlackfireContainer())
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return reward.BlackfireCmd(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(blackfireCmd)
}
