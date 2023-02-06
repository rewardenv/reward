package signcertificate

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/internal/logic"
)

func NewCmdSignCertificate(conf *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "sign-certificate <hostname> [hostname2] [hostname3]",
			Short: "Create a self signed certificate for your dev hostname.",
			Long:  `Create a self signed certificate for your dev hostname.`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(conf).RunCmdSignCertificate(args)
				if err != nil {
					return fmt.Errorf("error running sign-certificate command: %w", err)
				}

				return nil
			},
		},
		Config: conf,
	}
}
