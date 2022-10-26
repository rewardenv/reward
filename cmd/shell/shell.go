package shell

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/rewardenv/reward/internal/commands"
	"github.com/rewardenv/reward/internal/core"
)

var Cmd = &cobra.Command{
	Use:   "shell [command]",
	Short: "Launches into a shell within the current project environment",
	Long:  `Launches into a shell within the current project environment`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
		[]string, cobra.ShellCompDirective,
	) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	// DisableFlagParsing: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := commands.EnvCheck()

		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		return commands.ShellCmd(cmd, args)
	},
}

func init() {
	Cmd.Flags().StringVar(&commands.ShellContainer, "container", "php-fpm", "the container you want to get in")
	Cmd.Flags().StringVar(&commands.DefaultShellCommand, "command", "", "the container you want to get in")
	Cmd.Flags().StringVar(&commands.ShellUser, "user", "", "the user inside the container")
	viper.BindPFlag(fmt.Sprintf("%s_shell_user", core.AppName), Cmd.Flags().Lookup("user"))
}
