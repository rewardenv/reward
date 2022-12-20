package blackfire

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"reward/cmd"
	"reward/internal/app"
	"reward/internal/docker"
	"reward/internal/util"
)

func NewBlackfireCmd(app *app.App) *cmd.Command {
	return &cmd.Command{
		Command: &cobra.Command{
			Use: "blackfire [command]",
			Short: fmt.Sprintf(
				"Interacts with the blackfire service on an environment (disabled if %s_BLACKFIRE is not 1)",
				strings.ToUpper(app.Config.AppName()),
			),
			Long: fmt.Sprintf(
				`Interacts with the blackfire service on an environment (disabled if %s_BLACKFIRE is not 1)`,
				strings.ToUpper(app.Config.AppName()),
			),
			ValidArgsFunction: func(
				cmd *cobra.Command,
				args []string,
				toComplete string,
			) ([]string, cobra.ShellCompDirective) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			PreRunE: func(cmd *cobra.Command, args []string) error {
				if err := app.Docker.Check(); err != nil {
					return err
				}

				if err := app.Config.EnvCheck(); err != nil {
					return err
				}

				ContainerRunning, err := app.Docker.ContainerRunning(app.Config.BlackfireContainer())
				if err != nil {
					return err
				}

				if !app.Config.IsDBEnabled() || !ContainerRunning {
					return docker.ErrCannotFindContainer(app.Config.BlackfireContainer(),
						fmt.Errorf("blackfire container not found"))
				}

				return nil
			},
			RunE: func(c *cobra.Command, args []string) error {
				return BlackfireCmd(&cmd.Command{Command: c, App: app}, args)
			},
		},
		App: app,
	}
}

// BlackfireCmd represents the blackfire command.
func BlackfireCmd(cmd *cmd.Command, args []string) error {
	composeArgs := []string{
		"exec",
		cmd.App.Config.BlackfireContainer(),
		"sh",
		"-c", cmd.App.Config.BlackfireCommand(),
	}
	composeArgs = append(composeArgs, strings.Join(util.ExtractUnknownArgs(cmd.Flags(), args), " "))

	_, err := cmd.App.DockerCompose.RunCommand(composeArgs)
	if err != nil {
		return err
	}

	return nil
}
