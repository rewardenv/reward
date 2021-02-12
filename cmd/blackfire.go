package cmd

import (
	"fmt"
	"strings"

	. "reward/internal"

	"github.com/spf13/cobra"
)

var blackfireCmd = &cobra.Command{
	Use: "blackfire [command]",
	Short: fmt.Sprintf(
		"Interacts with the blackfire service on an environment (disabled if %v_BLACKFIRE is not 1)",
		strings.ToUpper(AppName)),
	Long: fmt.Sprintf(
		`Interacts with the blackfire service on an environment (disabled if %v_BLACKFIRE is not 1)`,
		strings.ToUpper(AppName)),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := CheckDocker(); err != nil {
			return err
		}

		if err := EnvCheck(); err != nil {
			return err
		}

		if !IsBlackfireEnabled() || !IsContainerRunning(GetBlackfireContainer()) {
			return CannotFindContainerError(GetBlackfireContainer())
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return BlackfireCmd(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(blackfireCmd)
}
