package debug

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdpkg "reward/cmd"
	"reward/internal/config"
	"reward/internal/logic"
)

func NewCmdDebug(c *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "debug [command]",
			Short: "Launches debug enabled shell within current project environment",
			Long:  `Launches debug enabled shell within current project environment`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(c).RunCmdDebug(cmd, args)
				if err != nil {
					return fmt.Errorf("error running debug command: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}
}
