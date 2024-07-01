package selfupdate

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/internal/logic"
)

func NewCmdSelfUpdate(conf *config.Config) *cmdpkg.Command {
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
				if err := logic.New(conf).RunCmdSelfUpdate(&cmdpkg.Command{Command: cmd, Config: conf}); err != nil {
					return errors.Wrap(err, "running self-update command")
				}

				return nil
			},
		},
		Config: conf,
	}

	cmd.Flags().BoolP("dry-run", "n", false, "only prints if there's new version available")
	cmd.Flags().BoolP("force", "f", false, "download and install the remote version even if its not newer")
	cmd.Flags().Bool("prerelease", false, "allow checking prerelease versions")

	return cmd
}
