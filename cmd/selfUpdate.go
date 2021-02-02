package cmd

import (
	. "reward/internal"

	"github.com/spf13/cobra"
)

var selfUpdateCmd = &cobra.Command{
	Use:   "self-update",
	Short: "Checks if new version exists and updates itself",
	Long:  `"Checks if new version exists and updates itself"`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Aliases: []string{"selfpudate"},
	Args:    cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return SelfUpdateCmd(cmd)
	},
}

func init() {
	rootCmd.AddCommand(selfUpdateCmd)

	selfUpdateCmd.Flags().BoolP("dry-run", "n", false, "only prints if there's new version available")
	selfUpdateCmd.Flags().BoolP("assume-yes", "y", false, "automatically update without asking")
}
