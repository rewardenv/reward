package cmd

import (
	. "reward/internal"

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
		err := EnvCheck()
		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return ShellCmd(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
	shellCmd.Flags().StringVar(&ShellContainer, "container", "php-fpm", "the container you want to get in")
}
