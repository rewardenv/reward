package bootstrap

import (
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/internal/logic"
)

func NewBootstrapCmd(c *config.Config) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:          "bootstrap [command]",
			Short:        "Install and Configure the basic settings for the environment",
			Long:         `Install and Configure the basic settings for the environment`,
			SilenceUsage: true,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(c).RunCmdBootstrap()
				if err != nil {
					return fmt.Errorf("error running bootstrap command: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}

	// --no-pull
	cmd.Flags().Bool("no-pull", false,
		"when specified latest images will not be explicitly pulled "+
			"prior to environment startup to facilitate use of locally built images",
	)
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_no_pull", c.AppName()), cmd.Flags().Lookup("no-pull"))

	// --full
	cmd.Flags().Bool("full", false, "includes sample data install and reindexing")
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_full_bootstrap", c.AppName()), cmd.Flags().Lookup("full"))

	// --no-parallel
	cmd.Flags().Bool("no-parallel", false, "disable hirak/prestissimo composer module")
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_composer_no_parallel", c.AppName()), cmd.Flags().Lookup("no-parallel"))

	// --skip-composer-install
	cmd.Flags().Bool("skip-composer-install", false, "dont run composer install")
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_skip_composer_install", c.AppName()),
		cmd.Flags().Lookup("skip-composer-install"))

	// --reset-admin-url
	cmd.Flags().Bool("reset-admin-url", false,
		"set admin/url/use_custom and admin/url/use_custom_path configurations to 0")
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_reset_admin_url", c.AppName()), cmd.Flags().Lookup("reset-admin-url"))

	if c.EnvType() == "magento1" || c.EnvType() == "magento2" {
		// --magento-type
		cmd.Flags().String("magento-type", "community", "magento type to install (community or enterprise)")
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_magento_type", c.AppName()),
			cmd.Flags().Lookup("magento-type"))

		// --magento-version
		cmd.Flags().String(
			"magento-version", version.Must(c.MagentoVersion()).String(), "magento version",
		)
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_magento_version", c.AppName()),
			cmd.Flags().Lookup("magento-version"))

		// --with-sampledata
		cmd.Flags().Bool("with-sampledata", false, "starts m2demo using demo images with sampledata")
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_with_sampledata", c.AppName()),
			cmd.Flags().Lookup("with-sampledata"))

		// --disable-tfa
		cmd.Flags().Bool("disable-tfa", false, "disable magento 2 two-factor authentication")
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_magento_disable_tfa", c.AppName()),
			cmd.Flags().Lookup("disable-tfa"))

		// --magento-mode
		cmd.Flags().String("magento-mode", "developer", "mage mode (developer or production)")
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_magento_mode", c.AppName()),
			cmd.Flags().Lookup("magento-mode"))

		// --crypt-key
		cmd.Flags().String("crypt-key", "", "crypt key for magento")
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_crypt_key", c.AppName()), cmd.Flags().Lookup("crypt-key"))

		// --db-prefix
		cmd.Flags().String("db-prefix", "", "database table prefix")
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_db_prefix", c.AppName()), cmd.Flags().Lookup("db-prefix"))
	}

	if c.EnvType() == "wordpress" {
		// --db-prefix
		cmd.Flags().String("db-prefix", "", "database table prefix")
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_db_prefix", c.AppName()), cmd.Flags().Lookup("db-prefix"))
	}

	if c.EnvType() == "shopware" {
		// --shopware-version
		cmd.Flags().String(
			"shopware-version", version.Must(c.ShopwareVersion()).String(), "shopware version",
		)
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_shopware_version", c.AppName()),
			cmd.Flags().Lookup("shopware-version"))

		// --shopware-mode
		cmd.Flags().String(
			"shopware-mode", c.ShopwareMode(), "shopware mode",
		)
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_shopware_mode", c.AppName()),
			cmd.Flags().Lookup("shopware-mode"))

		// --with-sampledata
		cmd.Flags().Bool("with-sampledata", false, "install shopware demo data")
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_with_sampledata", c.AppName()),
			cmd.Flags().Lookup("with-sampledata"))
	}

	return cmd
}
