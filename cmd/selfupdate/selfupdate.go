package selfupdate

import (
	"github.com/spf13/cobra"

	reward "github.com/rewardenv/reward/internal/commands"
)

var Cmd = &cobra.Command{
	Use:   "self-update",
	Short: "Checks if new version exists and updates itself",
	Long:  `"Checks if new version exists and updates itself"`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
		[]string, cobra.ShellCompDirective,
	) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Aliases: []string{"selfpudate"},
	Args:    cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return reward.SelfUpdateCmd(cmd)
	},
}

func init() {
	Cmd.Flags().BoolP("dry-run", "n", false, "only prints if there's new version available")
	Cmd.Flags().BoolP("assume-yes", "y", false, "automatically update without asking")
	Cmd.Flags().BoolP("force", "f", false, "download and install the remote version even if its not newer")
}
