package shell

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdpkg "reward/cmd"
	"reward/internal/config"
	"reward/internal/logic"
)

func NewCmdShell(c *config.Config) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "shell [command]",
			Short: "Launches into a shell within the current project environment",
			Long:  `Launches into a shell within the current project environment`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(c).RunCmdShell(cmd, args)
				if err != nil {
					return fmt.Errorf("error running shell command: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}

	cmd.Flags().StringVar(&c.ShellContainer, "container", "php-fpm", "the container you want to get in")
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_shell_container", c.AppName()), cmd.Flags().Lookup("container"))

	cmd.Flags().StringVar(&c.DefaultShellCommand, "command", "", "the container you want to get in")
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_shell_command", c.AppName()), cmd.Flags().Lookup("command"))

	cmd.Flags().StringVar(&c.ShellUser, "user", "", "the user inside the container")
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_shell_user", c.AppName()), cmd.Flags().Lookup("user"))

	return cmd
}
