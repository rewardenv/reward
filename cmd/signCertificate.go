package cmd

import (
	. "reward/internal"

	"github.com/spf13/cobra"
)

// signCertificateCmd represents the signCertificate command.
var signCertificateCmd = &cobra.Command{
	Use:   "sign-certificate <hostname> [hostname2] [hostname3]",
	Short: "Create a self signed certificate for your dev hostname.",
	Long:  `Create a self signed certificate for your dev hostname.`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return SignCertificateCmd(args)
	},
}

func init() {
	rootCmd.AddCommand(signCertificateCmd)
}
