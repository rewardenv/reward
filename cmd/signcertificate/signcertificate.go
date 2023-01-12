package signcertificate

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdpkg "reward/cmd"
	"reward/internal/config"
	"reward/internal/logic"
)

func NewCmdSignCertificate(c *config.Config) *cmdpkg.Command {
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
				err := logic.New(c).RunCmdSignCertificate(args)
				if err != nil {
					return fmt.Errorf("error running sign-certificate command: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}
}
