package env

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/internal/dockercompose"
	"github.com/rewardenv/reward/internal/logic"
)

func NewCmdEnv(conf *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:                "env",
			Short:              "Controls an environment from any point within the root project directory",
			Long:               `Controls an environment from any point within the root project directory`,
			ValidArgsFunction:  dockercompose.Completer(),
			DisableFlagParsing: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(conf).RunCmdEnv(args)
				if err != nil {
					return fmt.Errorf("error running env command: %w", err)
				}

				return nil
			},
		},
		Config: conf,
	}
}
