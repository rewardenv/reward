package cmd

import (
	reward "github.com/rewardenv/reward/internal"
	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell [command]",
	Short: "Launches into a shell within the current project environment",
	Long:  `Launches into a shell within the current project environment`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	// DisableFlagParsing: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := reward.EnvCheck()
		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return reward.ShellCmd(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)

	shellCmd.Flags().StringVar(&reward.ShellContainer, "container", "php-fpm", "the container you want to get in")
}
