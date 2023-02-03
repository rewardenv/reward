package envinit

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/internal/logic"
)

func NewCmdEnvInit(conf *config.Config) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
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
				err := logic.New(conf).RunCmdEnvInit(cmd, args)
				if err != nil {
					return fmt.Errorf("error running env-init command: %w", err)
				}

				return nil
			},
		},
		Config: conf,
	}

	cmd.Flags().String("environment-name", " ", "name for the new environment")
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_env_name", conf.AppName()), cmd.Flags().Lookup("environment-name"))

	cmd.Flags().String("environment-type", "magento2", "type of the new environment")
	_ = cmd.RegisterFlagCompletionFunc(
		"environment-type",
		func(envInitCmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return conf.ValidEnvTypes(), cobra.ShellCompDirectiveDefault
		},
	)

	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_env_type", conf.AppName()), cmd.Flags().Lookup("environment-type"))

	return cmd
}
