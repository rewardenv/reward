package env

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/compose"
	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/internal/logic"
)

func NewCmdEnv(conf *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:                "env",
			Short:              "Controls an environment from any point within the root project directory",
			Long:               `Controls an environment from any point within the root project directory`,
			ValidArgsFunction:  compose.Completer(),
			DisableFlagParsing: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(conf).RunCmdEnv(args)
				if err != nil {
					return errors.Wrap(err, "running env command")
				}

				return nil
			},
		},
		Config: conf,
	}
}
