package envinit

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdpkg "reward/cmd"
	"reward/internal/config"
	"reward/internal/logic"
)

func NewCmdEnvInit(c *config.Config) *cmdpkg.Command {
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
				err := logic.New(c).RunCmdEnvInit(cmd, args)
				if err != nil {
					return fmt.Errorf("error running env-init command: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}

	cmd.Flags().String("environment-name", " ", "name for the new environment")
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_env_name", c.AppName()), cmd.Flags().Lookup("environment-name"))

	cmd.Flags().String("environment-type", "magento2", "type of the new environment")
	_ = cmd.RegisterFlagCompletionFunc(
		"environment-type",
		func(envInitCmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return c.ValidEnvTypes(), cobra.ShellCompDirectiveDefault
		},
	)

	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_env_type", c.AppName()), cmd.Flags().Lookup("environment-type"))

	return cmd
}
