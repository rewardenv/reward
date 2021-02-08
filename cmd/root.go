/*
Package cmd represents the commands of the application.

Copyright © 2021 JANOS MIKO <janos.miko@itg.cloud>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"os"
	"path/filepath"
	"strings"

	. "reward/internal"

	dockerClient "github.com/docker/docker/client"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	logLevel   string
	appHomeDir string
	cfgFile    string
)

var rootCmd = &cobra.Command{
	Use:   AppName + " [command]",
	Short: AppName + " is a cli tool which helps you to run local dev environments",
	Long: ` ██▀███  ▓█████  █     █░ ▄▄▄       ██▀███  ▓█████▄
▓██ ▒ ██▒▓█   ▀ ▓█░ █ ░█░▒████▄    ▓██ ▒ ██▒▒██▀ ██▌
▓██ ░▄█ ▒▒███   ▒█░ █ ░█ ▒██  ▀█▄  ▓██ ░▄█ ▒░██   █▌
▒██▀▀█▄  ▒▓█  ▄ ░█░ █ ░█ ░██▄▄▄▄██ ▒██▀▀█▄  ░▓█▄   ▌
░██▓ ▒██▒░▒████▒░░██▒██▓  ▓█   ▓██▒░██▓ ▒██▒░▒████▓
░ ▒▓ ░▒▓░░░ ▒░ ░░ ▓░▒ ▒   ▒▒   ▓▒█░░ ▒▓ ░▒▓░ ▒▒▓  ▒
  ░▒ ░ ▒░ ░ ░  ░  ▒ ░ ░    ▒   ▒▒ ░  ░▒ ░ ▒░ ░ ▒  ▒
  ░░   ░    ░     ░   ░    ░   ▒     ░░   ░  ░ ░  ░
   ░        ░  ░    ░          ░  ░   ░        ░
                                             ░      `,
	Version: GetAppVersion().String(),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return CheckInvokerUser(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return RootCmd(cmd)
	},
}

// Execute runs the rootCmd itself.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(setLogLevel)
	cobra.OnInitialize(configureHiddenCommands)

	home, err := homedir.Dir()
	if err != nil {
		log.Fatalln(err)
	}

	// --app-dir
	rootCmd.PersistentFlags().StringVar(
		&appHomeDir, "app-dir", filepath.Join(home, "."+AppName), "app home directory")

	_ = viper.BindPFlag(AppName+"_home_dir", rootCmd.PersistentFlags().Lookup("app-dir"))

	// --log-level
	rootCmd.PersistentFlags().String(
		"log-level", "info", "logging level (options: trace, debug, info, warning, error)")

	_ = viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level"))

	// --debug
	rootCmd.PersistentFlags().Bool(
		"debug", false, "enable debug mode (same as --log-level=debug)")

	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	// --disable-colors
	rootCmd.PersistentFlags().Bool(
		"disable-colors", false, "disable colors in output")

	_ = viper.BindPFlag("disable_colors", rootCmd.PersistentFlags().Lookup("disable-colors"))

	// --config
	rootCmd.PersistentFlags().StringVarP(
		&cfgFile, "config", "c", filepath.Join(GetHomeDir(), "."+AppName+".yml"), "config file")

	_ = viper.BindPFlag(AppName+"_config_file", rootCmd.PersistentFlags().Lookup("config"))

	// --docker-host
	rootCmd.PersistentFlags().String(
		"docker-host", dockerClient.DefaultDockerHost, "docker host")

	_ = viper.BindPFlag("docker_host", rootCmd.PersistentFlags().Lookup("docker-host"))

	// --driver
	// rootCmd.PersistentFlags().String(
	// 	"driver", "docker-compose", "orchestration driver")
	// _ = viper.BindPFlag(AppName+"_driver", rootCmd.PersistentFlags().Lookup("driver"))

	// --service-domain
	rootCmd.PersistentFlags().String(
		"service-domain", AppName+".test", "service domain for global services")

	rootCmd.PersistentFlags().Lookup("service-domain").Hidden = true

	_ = viper.BindPFlag(AppName+"_service_domain", rootCmd.PersistentFlags().Lookup("service-domain"))

	// --print-environment
	rootCmd.Flags().Bool(
		"print-environment", false, "environment vars")

	_ = viper.BindPFlag(AppName+"_print_environment", rootCmd.Flags().Lookup("print-environment"))
}

func initConfig() {
	viper.AddConfigPath(".")
	viper.AddConfigPath(filepath.Dir(cfgFile))
	viper.SetConfigName(filepath.Base(cfgFile))
	viper.SetConfigType("yaml")

	// Read config files in default locations
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Debugf("%v", err)
	}

	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("dotenv")

	if err := viper.MergeInConfig(); err != nil {
		log.Debugf("%v", err)
	}

	log.Debugln("Using config file:", viper.ConfigFileUsed())

	appHomeDir = GetAppHomeDir()

	// app_ssl_dir and app_composer_dir have to be configured for templating
	if !viper.IsSet(AppName + "_ssl_dir") {
		viper.Set(AppName+"_ssl_dir", filepath.Join(appHomeDir, "ssl"))
	}

	if !viper.IsSet(AppName + "_composer_dir") {
		viper.Set(AppName+"_composer_dir", filepath.Join(GetHomeDir(), ".composer"))
	}

	if !viper.IsSet(AppName + "_ssh_dir") {
		viper.Set(AppName+"_ssh_dir", filepath.Join(GetHomeDir(), ".ssh"))
	}
}

func setLogLevel() {
	if IsDebug() {
		log.SetLevel(log.DebugLevel)
	} else {
		switch logLevel = viper.GetString("log_level"); logLevel {
		case "trace":
			log.SetLevel(log.TraceLevel)
		case "debug":
			log.SetLevel(log.DebugLevel)
		case "info":
			log.SetLevel(log.InfoLevel)
		case "warning":
			log.SetLevel(log.WarnLevel)
		default:
			log.SetLevel(log.ErrorLevel)
		}
	}

	log.SetFormatter(&log.TextFormatter{
		DisableColors:          viper.GetBool("disable_colors"),
		ForceColors:            true,
		DisableLevelTruncation: true,
		FullTimestamp:          true,
		QuoteEmptyFields:       true,
	})
}

func configureHiddenCommands() {
	if !IsBlackfireEnabled() {
		blackfireCmd.Hidden = true
	}
}

func RootCmd(cmd *cobra.Command) error {
	if viper.GetBool(AppName + "_print_environment") {
		for i, v := range viper.AllSettings() {
			log.Printf("%v=%v", strings.ToUpper(i), v)
		}

		os.Exit(0)
	}

	_ = cmd.Help()

	return nil
}

// CheckInvokerUser returns an error if the invoker user is root.
func CheckInvokerUser(cmd *cobra.Command) error {
	// If the REWARD_ALLOW_SUPERUSER=1 is set or the Distro is Windows then we can skip this.
	if IsAllowedSuperuser() || GetOSDistro() == "windows" {
		return nil
	}

	// Most of the commands should run by normal users except `self-update`.
	if cmd.Name() != "self-update" && IsAdmin() {
		return ErrInvokedAsRootUser
	}

	return nil
}
