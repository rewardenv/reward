package selfupdate

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/internal/logic"
)

func NewCmdSelfUpdate(c *config.Config) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "self-update",
			Short: "Checks if new version exists and updates itself",
			Long:  `"Checks if new version exists and updates itself"`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			Aliases: []string{"selfpudate", "self-upgrade", "selfupgrade"},
			Args:    cobra.ExactArgs(0),
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(c).RunCmdSelfUpdate(&cmdpkg.Command{Command: cmd})
				if err != nil {
					return fmt.Errorf("error running self-update command: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}

	cmd.Flags().BoolP("dry-run", "n", false, "only prints if there's new version available")
	cmd.Flags().BoolP("force", "f", false, "download and install the remote version even if its not newer")
	cmd.Flags().Bool("prerelease", false, "allow checking prerelease versions")

	return cmd
}
