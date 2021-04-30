package cmd

import (
	reward "github.com/rewardenv/reward/internal"
	"github.com/spf13/cobra"
)

var signCertificateCmd = &cobra.Command{
	Use:   "sign-certificate <hostname> [hostname2] [hostname3]",
	Short: "Create a self signed certificate for your dev hostname.",
	Long:  `Create a self signed certificate for your dev hostname.`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return reward.SignCertificateCmd(args)
	},
}

func init() {
	rootCmd.AddCommand(signCertificateCmd)
}
