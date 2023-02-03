package shell

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/internal/logic"
)

func NewCmdShell(conf *config.Config) *cmdpkg.Command {
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
				err := logic.New(conf).RunCmdShell(cmd, args)
				if err != nil {
					return fmt.Errorf("error running shell command: %w", err)
				}

				return nil
			},
		},
		Config: conf,
	}

	cmd.Flags().StringVar(&conf.ShellContainer, "container", "php-fpm", "the container you want to get in")
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_shell_container", conf.AppName()), cmd.Flags().Lookup("container"))

	cmd.Flags().StringVar(&conf.DefaultShellCommand, "command", "", "the container you want to get in")
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_shell_command", conf.AppName()), cmd.Flags().Lookup("command"))

	cmd.Flags().StringVar(&conf.ShellUser, "user", "", "the user inside the container")
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_shell_user", conf.AppName()), cmd.Flags().Lookup("user"))

	return cmd
}
