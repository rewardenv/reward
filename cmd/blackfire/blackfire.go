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

func NewBlackfireCmd(c *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use: "blackfire [command]",
			Short: fmt.Sprintf(
				"Interacts with the blackfire service on an environment (disabled if %s_BLACKFIRE is not 1)",
				strings.ToUpper(c.AppName()),
			),
			Long: fmt.Sprintf(
				`Interacts with the blackfire service on an environment (disabled if %s_BLACKFIRE is not 1)`,
				strings.ToUpper(c.AppName()),
			),
			ValidArgsFunction: func(
				cmd *cobra.Command,
				args []string,
				toComplete string,
			) ([]string, cobra.ShellCompDirective) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			PreRunE: func(cmd *cobra.Command, args []string) error {
				if !c.Docker.ContainerRunning(c.BlackfireContainer()) {
					return docker.ErrCannotFindContainer(c.BlackfireContainer(),
						fmt.Errorf("blackfire container not found"))
				}

				return nil
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(c).RunCmdBlackfire(&cmdpkg.Command{Command: cmd, Config: c}, args)
				if err != nil {
					return fmt.Errorf("error running blackfire command: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}
}
