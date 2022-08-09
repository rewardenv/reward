package bootstrap

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/rewardenv/reward/internal/commands"
	"github.com/rewardenv/reward/internal/core"
)

var Cmd = &cobra.Command{
	Use:   "bootstrap [command]",
	Short: "Install and Configure the basic settings for the environment",
	Long:  `Install and Configure the basic settings for the environment`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
		[]string, cobra.ShellCompDirective,
	) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := core.CheckDocker(); err != nil {
			return err
		}

		if err := commands.EnvCheck(); err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return commands.BootstrapCmd()
	},
}

func init() {
	addFlags()
}

func addFlags() {
	// --with-sampledata
	Cmd.Flags().Bool(
		"with-sampledata", false, "starts m2demo using demo images with sampledata",
	)

	_ = viper.BindPFlag(core.AppName+"_with_sampledata", Cmd.Flags().Lookup("with-sampledata"))

	// --no-pull
	Cmd.Flags().Bool(
		"no-pull",
		false,
		"when specified latest images will not be explicitly pulled "+
			"prior to environment startup to facilitate use of locally built images",
	)

	_ = viper.BindPFlag(core.AppName+"_no_pull", Cmd.Flags().Lookup("no-pull"))

	// --full
	Cmd.Flags().Bool(
		"full", false, "includes sample data install and reindexing",
	)

	_ = viper.BindPFlag(core.AppName+"_full_bootstrap", Cmd.Flags().Lookup("full"))

	// --no-parallel
	Cmd.Flags().Bool(
		"no-parallel", false, "disable hirak/prestissimo composer module",
	)

	_ = viper.BindPFlag(core.AppName+"_composer_no_parallel", Cmd.Flags().Lookup("no-parallel"))

	// --skip-composer-install
	Cmd.Flags().Bool(
		"skip-composer-install", false, "dont run composer install",
	)

	_ = viper.BindPFlag(core.AppName+"_skip_composer_install", Cmd.Flags().Lookup("skip-composer-install"))

	// --magento-type
	Cmd.Flags().String(
		"magento-type", "community", "magento type to install (community or enterprise)",
	)

	_ = viper.BindPFlag(core.AppName+"_magento_type", Cmd.Flags().Lookup("magento-type"))

	// --magento-version
	magentoVersion, err := core.GetMagentoVersion()
	if err != nil {
		log.Fatalln(err)
	}

	Cmd.Flags().String(
		"magento-version", magentoVersion.String(), "magento version",
	)

	_ = viper.BindPFlag(core.AppName+"_magento_version", Cmd.Flags().Lookup("magento-version"))

	// --disable-tfa
	Cmd.Flags().Bool(
		"disable-tfa", false, "disable magento 2 two-factor authentication",
	)

	_ = viper.BindPFlag(core.AppName+"_magento_disable_tfa", Cmd.Flags().Lookup("disable-tfa"))

	// --magento-mode
	Cmd.Flags().String(
		"magento-mode", "developer", "mage mode (developer or production)",
	)

	_ = viper.BindPFlag(core.AppName+"_magento_mode", Cmd.Flags().Lookup("magento-mode"))

	// --reset-admin-url
	Cmd.Flags().Bool(
		"reset-admin-url", false, "set admin/url/use_custom and admin/url/use_custom_path configurations to 0",
	)

	_ = viper.BindPFlag(core.AppName+"_reset_admin_url", Cmd.Flags().Lookup("reset-admin-url"))

	// --db-prefix
	Cmd.Flags().String(
		"db-prefix",
		"",
		"database table prefix",
	)

	_ = viper.BindPFlag(core.AppName+"_db_prefix", Cmd.Flags().Lookup("db-prefix"))

	// --crypt-key
	Cmd.Flags().String(
		"crypt-key",
		"",
		"crypt key for magento",
	)

	_ = viper.BindPFlag(core.AppName+"_crypt_key", Cmd.Flags().Lookup("crypt-key"))
}
