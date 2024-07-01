package info

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/internal/logic"
)

func NewCmdInfo(conf *config.Config) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "info",
			Short: "Print out information about the environment",
			Long:  `Print out information about the environment`,
			ValidArgsFunction: func(
				cmd *cobra.Command,
				args []string,
				toComplete string,
			) ([]string, cobra.ShellCompDirective) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			Args: cobra.ExactArgs(0),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := logic.New(conf).RunCmdInfo(&cmdpkg.Command{Command: cmd}); err != nil {
					return errors.Wrap(err, "running info command")
				}

				return nil
			},
		},
		Config: conf,
	}

	cmd.Flags().String("style",
		"default",
		"styling for the output (options: default, black, double, bright, light, dark, csv, markdown, html)")
	_ = cmd.Config.BindPFlag("style", cmd.Flags().Lookup("style"))

	return cmd
}
