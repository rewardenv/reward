package envinit

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/rewardenv/reward/internal/commands"
	"github.com/rewardenv/reward/internal/core"
)

var Cmd = &cobra.Command{
	Use:                   "env-init <name>",
	Short:                 "Create the .env file",
	Long:                  `Create the .env file`,
	DisableFlagsInUseLine: false,
	Args:                  cobra.RangeArgs(0, 2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
		[]string, cobra.ShellCompDirective,
	) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return commands.EnvInitCmd(cmd, args)
	},
}

func init() {
	Cmd.Flags().String("environment-name", " ", "name for the new environment")
	_ = viper.BindPFlag(core.AppName+"_env_name", Cmd.Flags().Lookup("environment-name"))

	Cmd.Flags().String("environment-type", "magento2", "type of the new environment")
	_ = Cmd.RegisterFlagCompletionFunc(
		"environment-type",
		func(envInitCmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return commands.ValidEnvTypes(), cobra.ShellCompDirectiveDefault
		},
	)

	_ = viper.BindPFlag(core.AppName+"_env_type", Cmd.Flags().Lookup("environment-type"))
}
