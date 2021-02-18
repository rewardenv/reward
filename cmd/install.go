package cmd

import (
	"github.com/spf13/viper"

	. "reward/internal"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install required configurations",
	Long:  `Install required configurations for reward. CA Certificate, SSH Key, etc.`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return InstallCmd()
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().Bool("reinstall", false, "reinstall configurations")
	installCmd.Flags().Bool("uninstall", false, "uninstall configurations")
	installCmd.Flags().Bool("ca-cert", false, "install ca-certificate only")
	installCmd.Flags().Bool("dns", false, "install dns settings only")
	installCmd.Flags().Bool("ssh-key", false, "install ssh key only")
	installCmd.Flags().Bool("ssh-config", false, "install ssh config only")
	installCmd.Flags().Bool("ignore-ca-cert", false, "ignore ca-certificate creation")
	installCmd.Flags().Bool("ignore-dns", false, "ignore dns settings installation")
	installCmd.Flags().Bool("ignore-ssh-key", false, "ignore ssh key installation")
	installCmd.Flags().Bool("ignore-ssh-config", false, "ignore ssh config installation")
	installCmd.Flags().Bool("ignore-svcs", false, "ignore initializing of the common services")
	installCmd.Flags().Int("app-home-mode", 0o755, "directory mode for app home dir")

	_ = viper.BindPFlag(AppName+"_install_reinstall", installCmd.Flags().Lookup("reinstall"))
	_ = viper.BindPFlag(AppName+"_install_uninstall", installCmd.Flags().Lookup("uninstall"))
	_ = viper.BindPFlag(AppName+"_install_ca_cert", installCmd.Flags().Lookup("ca-cert"))
	_ = viper.BindPFlag(AppName+"_install_dns", installCmd.Flags().Lookup("dns"))
	_ = viper.BindPFlag(AppName+"_install_ssh_key", installCmd.Flags().Lookup("ssh-key"))
	_ = viper.BindPFlag(AppName+"_install_ssh_config", installCmd.Flags().Lookup("ssh-config"))
	_ = viper.BindPFlag(AppName+"_install_app_home_mode", installCmd.Flags().Lookup("app-home-mode"))
	_ = viper.BindPFlag(AppName+"_install_ignore_init_svcs", installCmd.Flags().Lookup("ignore-svcs"))
}
