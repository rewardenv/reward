package blackfire

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/internal/docker"
	"github.com/rewardenv/reward/internal/logic"
)

func NewBlackfireCmd(conf *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use: "blackfire [command]",
			Short: fmt.Sprintf(
				"Interacts with the blackfire service on an environment (disabled if %s_BLACKFIRE is not 1)",
				strings.ToUpper(conf.AppName()),
			),
			Long: fmt.Sprintf(
				`Interacts with the blackfire service on an environment (disabled if %s_BLACKFIRE is not 1)`,
				strings.ToUpper(conf.AppName()),
			),
			ValidArgsFunction: func(
				cmd *cobra.Command,
				args []string,
				toComplete string,
			) ([]string, cobra.ShellCompDirective) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			PreRunE: func(cmd *cobra.Command, args []string) error {
				if !conf.Docker.ContainerRunning(conf.BlackfireContainer()) {
					return docker.ErrCannotFindContainer(conf.BlackfireContainer(),
						fmt.Errorf("blackfire container not found"))
				}

				return nil
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(conf).RunCmdBlackfire(&cmdpkg.Command{Command: cmd, Config: conf}, args)
				if err != nil {
					return fmt.Errorf("error running blackfire command: %w", err)
				}

				return nil
			},
		},
		Config: conf,
	}
}
