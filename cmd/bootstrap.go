package cmd

import (
	"github.com/spf13/viper"

	. "reward/internal"

	"github.com/spf13/cobra"
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap [command]",
	Short: "Install and Configure the basic settings for the environment",
	Long:  `Install and Configure the basic settings for the environment`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := CheckDockerIsRunning(); err != nil {
			return err
		}

		if err := EnvCheck(); err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return BootstrapCmd()
	},
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)

	bootstrapCmd.Flags().Bool(
		"with-sampledata", false, "starts m2demo using demo images with sampledata")

	_ = viper.BindPFlag(AppName+"_with_sampledata", bootstrapCmd.Flags().Lookup("with-sampledata"))

	bootstrapCmd.Flags().Bool(
		"no-pull",
		false,
		"when specified latest images will not be explicitly pulled "+
			"prior to environment startup to facilitate use of locally built images")

	_ = viper.BindPFlag(AppName+"_no_pull", bootstrapCmd.Flags().Lookup("no-pull"))

	bootstrapCmd.Flags().Bool(
		"full", false, "includes sample data install and reindexing")

	_ = viper.BindPFlag(AppName+"_full_bootstrap", bootstrapCmd.Flags().Lookup("full"))

	bootstrapCmd.Flags().Bool(
		"no-parallel", false, "disable hirak/prestissimo composer module")

	_ = viper.BindPFlag(AppName+"_composer_no_parallel", bootstrapCmd.Flags().Lookup("no-parallel"))

	bootstrapCmd.Flags().Bool(
		"skip-composer-install", false, "dont run composer install")

	_ = viper.BindPFlag(AppName+"_skip_composer_install", bootstrapCmd.Flags().Lookup("skip-composer-install"))

	bootstrapCmd.Flags().String(
		"magento-type", "community", "magento type to install (community or enterprise)")

	_ = viper.BindPFlag(AppName+"_magento_type", bootstrapCmd.Flags().Lookup("magento-type"))

	bootstrapCmd.Flags().String(
		"magento-version", GetMagentoVersion().String(), "magento version")

	_ = viper.BindPFlag(AppName+"_magento_version", bootstrapCmd.Flags().Lookup("magento-version"))
}
