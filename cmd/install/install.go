package install

import (
	"github.com/rewardenv/reward/internal/commands"
	"github.com/rewardenv/reward/internal/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Cmd = &cobra.Command{
	Use:   "install",
	Short: "Install required configurations",
	Long:  `Install required configurations for reward. CA Certificate, SSH Key, etc.`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return commands.InstallCmd()
	},
}

func init() {
	Cmd.Flags().Bool("reinstall", false, "reinstall configurations")
	Cmd.Flags().Bool("uninstall", false, "uninstall configurations")
	Cmd.Flags().Bool("ca-cert", false, "install ca-certificate only")
	Cmd.Flags().Bool("dns", false, "install dns settings only")
	Cmd.Flags().Bool("ssh-key", false, "install ssh key only")
	Cmd.Flags().Bool("ssh-config", false, "install ssh config only")
	Cmd.Flags().Bool("ignore-ca-cert", false, "ignore ca-certificate creation")
	Cmd.Flags().Bool("ignore-dns", false, "ignore dns settings installation")
	Cmd.Flags().Bool("ignore-ssh-key", false, "ignore ssh key installation")
	Cmd.Flags().Bool("ignore-ssh-config", false, "ignore ssh config installation")
	Cmd.Flags().Bool("ignore-svcs", false, "ignore initializing of the common services")
	Cmd.Flags().Int("app-home-mode", 0o755, "directory mode for app home dir")

	_ = viper.BindPFlag(core.AppName+"_install_reinstall", Cmd.Flags().Lookup("reinstall"))
	_ = viper.BindPFlag(core.AppName+"_install_uninstall", Cmd.Flags().Lookup("uninstall"))
	_ = viper.BindPFlag(core.AppName+"_install_ca_cert", Cmd.Flags().Lookup("ca-cert"))
	_ = viper.BindPFlag(core.AppName+"_install_dns", Cmd.Flags().Lookup("dns"))
	_ = viper.BindPFlag(core.AppName+"_install_ssh_key", Cmd.Flags().Lookup("ssh-key"))
	_ = viper.BindPFlag(core.AppName+"_install_ssh_config", Cmd.Flags().Lookup("ssh-config"))
	_ = viper.BindPFlag(core.AppName+"_install_app_home_mode", Cmd.Flags().Lookup("app-home-mode"))
	_ = viper.BindPFlag(core.AppName+"_install_ignore_init_svcs", Cmd.Flags().Lookup("ignore-svcs"))
}
