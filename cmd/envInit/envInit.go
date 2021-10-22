package envInit

import (
	"github.com/rewardenv/reward/internal/commands"
	"github.com/rewardenv/reward/internal/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Cmd = &cobra.Command{
	Use:                   "env-init <name>",
	Short:                 "Create the .env file",
	Long:                  `Create the .env file`,
	DisableFlagsInUseLine: false,
	Args:                  cobra.RangeArgs(0, 2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return commands.EnvInitCmd(cmd, args)
	},
}

func init() {
	flagEnvName := "environment-name"
	Cmd.Flags().String(flagEnvName, " ", "name for the new environment")
	// _ = Cmd.MarkFlagRequired(flagEnvName)
	_ = viper.BindPFlag(core.AppName+"_env_name", Cmd.Flags().Lookup(flagEnvName))

	flagEnvType := "environment-type"
	Cmd.Flags().String(flagEnvType, "magento2", "type of the new environment")
	_ = Cmd.RegisterFlagCompletionFunc(
		flagEnvType, func(envInitCmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return commands.GetValidEnvTypes(), cobra.ShellCompDirectiveDefault
		})
	_ = viper.BindPFlag(core.AppName+"_env_type", Cmd.Flags().Lookup(flagEnvType))
}
