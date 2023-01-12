package install

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdpkg "reward/cmd"
	"reward/internal/config"
	"reward/internal/logic"
)

func NewCmdInstall(c *config.Config) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "install",
			Short: "Install required configurations",
			Long:  `Install required configurations for reward. CA Certificate, SSH Key, etc.`,
			ValidArgsFunction: func(
				cmd *cobra.Command,
				args []string,
				toComplete string,
			) ([]string, cobra.ShellCompDirective) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			Args: cobra.ExactArgs(0),
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(c).RunCmdInstall()
				if err != nil {
					return fmt.Errorf("error running install command: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}

	cmd.Flags().Bool("reinstall", false, "reinstall configurations")
	cmd.Flags().Bool("uninstall", false, "uninstall configurations")
	cmd.Flags().Bool("ca-cert", false, "install ca-certificate only")
	cmd.Flags().Bool("dns", false, "install dns settings only")
	cmd.Flags().Bool("ssh-key", false, "install ssh key only")
	cmd.Flags().Bool("ssh-config", false, "install ssh config only")
	cmd.Flags().Bool("ignore-ca-cert", false, "ignore ca-certificate creation")
	cmd.Flags().Bool("ignore-dns", false, "ignore dns settings installation")
	cmd.Flags().Bool("ignore-ssh-key", false, "ignore ssh key installation")
	cmd.Flags().Bool("ignore-ssh-config", false, "ignore ssh config installation")
	cmd.Flags().Bool("ignore-svcs", false, "ignore initializing of the common services")
	cmd.Flags().Int("app-home-mode", 0o755, "directory mode for app home dir")

	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_install_reinstall", c.AppName()), cmd.Flags().Lookup("reinstall"))
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_install_uninstall", c.AppName()), cmd.Flags().Lookup("uninstall"))
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_install_ca_cert", c.AppName()), cmd.Flags().Lookup("ca-cert"))
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_install_dns", c.AppName()), cmd.Flags().Lookup("dns"))
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_install_ssh_key", c.AppName()), cmd.Flags().Lookup("ssh-key"))
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_install_ssh_config", c.AppName()), cmd.Flags().Lookup("ssh-config"))
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_install_app_home_mode", c.AppName()),
		cmd.Flags().Lookup("app-home-mode"))
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_install_ignore_init_svcs", c.AppName()),
		cmd.Flags().Lookup("ignore-svcs"))

	return cmd
}
