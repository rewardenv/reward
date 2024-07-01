package bootstrap

import (
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/internal/logic"
)

func NewBootstrapCmd(conf *config.Config) *cmdpkg.Command {
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
				if err := logic.New(conf).RunCmdBootstrap(); err != nil {
					return errors.Wrap(err, "running bootstrap command")
				}

				return nil
			},
		},
		Config: conf,
	}

	// --no-pull
	cmd.Flags().Bool("no-pull", false,
		"when specified latest images will not be explicitly pulled "+
			"prior to environment startup to facilitate use of locally built images",
	)
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_no_pull", conf.AppName()), cmd.Flags().Lookup("no-pull"))

	if conf.EnvType() == "magento1" || conf.EnvType() == "magento2" || conf.EnvType() == "shopware" {
		// --full
		cmd.Flags().Bool("full", false, "includes sample data install and reindexing")
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_full_bootstrap", conf.AppName()), cmd.Flags().Lookup("full"))
	}

	// --no-parallel
	cmd.Flags().Bool("no-parallel", false, "disable hirak/prestissimo composer module")
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_composer_no_parallel", conf.AppName()), cmd.Flags().Lookup("no-parallel"))

	// --skip-composer-install
	cmd.Flags().Bool("skip-composer-install", false, "dont run composer install")
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_skip_composer_install", conf.AppName()),
		cmd.Flags().Lookup("skip-composer-install"))

	if conf.EnvType() == "magento1" || conf.EnvType() == "magento2" {
		// --reset-admin-url
		cmd.Flags().Bool("reset-admin-url", false,
			"set admin/url/use_custom and admin/url/use_custom_path configurations to 0")
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_reset_admin_url", conf.AppName()),
			cmd.Flags().Lookup("reset-admin-url"))

		// --magento-type
		cmd.Flags().String("magento-type", "community", "magento type to install (community or enterprise)")
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_magento_type", conf.AppName()),
			cmd.Flags().Lookup("magento-type"))

		// --magento-version
		cmd.Flags().String(
			"magento-version", version.Must(conf.MagentoVersion()).String(), "magento version",
		)
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_magento_version", conf.AppName()),
			cmd.Flags().Lookup("magento-version"))

		// --with-sampledata
		cmd.Flags().Bool("with-sampledata", false, "starts m2demo using demo images with sampledata")
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_with_sampledata", conf.AppName()),
			cmd.Flags().Lookup("with-sampledata"))

		// --disable-tfa
		cmd.Flags().Bool("disable-tfa", false, "disable magento 2 two-factor authentication")
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_magento_disable_tfa", conf.AppName()),
			cmd.Flags().Lookup("disable-tfa"))

		// --magento-mode
		cmd.Flags().String("magento-mode", "developer", "mage mode (developer or production)")
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_magento_mode", conf.AppName()),
			cmd.Flags().Lookup("magento-mode"))

		// --crypt-key
		cmd.Flags().String("crypt-key", "", "crypt key for magento")
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_crypt_key", conf.AppName()), cmd.Flags().Lookup("crypt-key"))

		// --db-prefix
		cmd.Flags().String("db-prefix", "", "database table prefix")
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_db_prefix", conf.AppName()), cmd.Flags().Lookup("db-prefix"))
	}

	if conf.EnvType() == "wordpress" {
		// --db-prefix
		cmd.Flags().String("db-prefix", "", "database table prefix")
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_db_prefix", conf.AppName()), cmd.Flags().Lookup("db-prefix"))
	}

	if conf.EnvType() == "shopware" {
		// --shopware-version
		cmd.Flags().String(
			"shopware-version", version.Must(conf.ShopwareVersion()).String(), "shopware version",
		)
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_shopware_version", conf.AppName()),
			cmd.Flags().Lookup("shopware-version"))

		// --shopware-mode
		cmd.Flags().String(
			"shopware-mode", conf.ShopwareMode(), "shopware mode",
		)
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_shopware_mode", conf.AppName()),
			cmd.Flags().Lookup("shopware-mode"))

		// --with-sampledata
		cmd.Flags().Bool("with-sampledata", false, "install shopware demo data")
		_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_with_sampledata", conf.AppName()),
			cmd.Flags().Lookup("with-sampledata"))
	}

	return cmd
}
