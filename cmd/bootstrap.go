package cmd

import (
	reward "github.com/rewardenv/reward/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap [command]",
	Short: "Install and Configure the basic settings for the environment",
	Long:  `Install and Configure the basic settings for the environment`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := reward.CheckDocker(); err != nil {
			return err
		}

		if err := reward.EnvCheck(); err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return reward.BootstrapCmd()
	},
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)

	// --with-sampledata
	bootstrapCmd.Flags().Bool(
		"with-sampledata", false, "starts m2demo using demo images with sampledata")

	_ = viper.BindPFlag(reward.AppName+"_with_sampledata", bootstrapCmd.Flags().Lookup("with-sampledata"))

	// --no-pull
	bootstrapCmd.Flags().Bool(
		"no-pull",
		false,
		"when specified latest images will not be explicitly pulled "+
			"prior to environment startup to facilitate use of locally built images")

	_ = viper.BindPFlag(reward.AppName+"_no_pull", bootstrapCmd.Flags().Lookup("no-pull"))

	// --full
	bootstrapCmd.Flags().Bool(
		"full", false, "includes sample data install and reindexing")

	_ = viper.BindPFlag(reward.AppName+"_full_bootstrap", bootstrapCmd.Flags().Lookup("full"))

	// --no-parallel
	bootstrapCmd.Flags().Bool(
		"no-parallel", false, "disable hirak/prestissimo composer module")

	_ = viper.BindPFlag(reward.AppName+"_composer_no_parallel", bootstrapCmd.Flags().Lookup("no-parallel"))

	// --skip-composer-install
	bootstrapCmd.Flags().Bool(
		"skip-composer-install", false, "dont run composer install")

	_ = viper.BindPFlag(reward.AppName+"_skip_composer_install", bootstrapCmd.Flags().Lookup("skip-composer-install"))

	// --magento-type
	bootstrapCmd.Flags().String(
		"magento-type", "community", "magento type to install (community or enterprise)")

	_ = viper.BindPFlag(reward.AppName+"_magento_type", bootstrapCmd.Flags().Lookup("magento-type"))

	// --magento-version
	magentoVersion, err := reward.GetMagentoVersion()
	if err != nil {
		log.Fatalln(err)
	}
	bootstrapCmd.Flags().String(
		"magento-version", magentoVersion.String(), "magento version")

	_ = viper.BindPFlag(reward.AppName+"_magento_version", bootstrapCmd.Flags().Lookup("magento-version"))

	// --disable-tfa
	bootstrapCmd.Flags().Bool(
		"disable-tfa", false, "disable magento 2 two-factor authentication")

	_ = viper.BindPFlag(reward.AppName+"_magento_disable_tfa", bootstrapCmd.Flags().Lookup("disable-tfa"))

	// --magento-mode
	bootstrapCmd.Flags().String(
		"magento-mode", "developer", "mage mode (developer or production)")

	_ = viper.BindPFlag(reward.AppName+"_magento_mode", bootstrapCmd.Flags().Lookup("magento-mode"))

	// --reset-admin-url
	bootstrapCmd.Flags().Bool(
		"reset-admin-url", false, "set admin/url/use_custom and admin/url/use_custom_path configurations to 0")

	_ = viper.BindPFlag(reward.AppName+"_reset_admin_url", bootstrapCmd.Flags().Lookup("reset-admin-url"))
}
