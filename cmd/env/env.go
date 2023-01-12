package env

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdpkg "reward/cmd"
	"reward/internal/config"
	"reward/internal/dockercompose"
	"reward/internal/logic"
)

func NewCmdEnv(c *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:                "env",
			Short:              "Controls an environment from any point within the root project directory",
			Long:               `Controls an environment from any point within the root project directory`,
			ValidArgsFunction:  dockercompose.Completer(),
			DisableFlagParsing: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(c).RunCmdEnv(args)
				if err != nil {
					return fmt.Errorf("error running env command: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}
}
