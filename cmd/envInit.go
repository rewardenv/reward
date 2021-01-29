package cmd

import (
	. "reward/internal"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var envInitCmd = &cobra.Command{
	Use:                   "env-init <name>",
	Short:                 "Create the .env file",
	Long:                  `Create the .env file`,
	DisableFlagsInUseLine: false,
	Args:                  cobra.RangeArgs(0, 2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return EnvInitCmd(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(envInitCmd)

	flagEnvName := "environment-name"
	envInitCmd.Flags().String(flagEnvName, " ", "name for the new environment")
	// _ = envInitCmd.MarkFlagRequired(flagEnvName)
	_ = viper.BindPFlag(AppName+"_env_name", envInitCmd.Flags().Lookup(flagEnvName))

	flagEnvType := "environment-type"
	envInitCmd.Flags().String(flagEnvType, "magento2", "type of the new environment")
	_ = envInitCmd.RegisterFlagCompletionFunc(
		flagEnvType, func(envInitCmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return GetValidEnvTypes(), cobra.ShellCompDirectiveDefault
		})
	_ = viper.BindPFlag(AppName+"_env_type", envInitCmd.Flags().Lookup(flagEnvType))
}
