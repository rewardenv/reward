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
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/rewardenv/reward/cmd/blackfire"
	"github.com/rewardenv/reward/cmd/bootstrap"
	"github.com/rewardenv/reward/cmd/completion"
	"github.com/rewardenv/reward/cmd/db"
	"github.com/rewardenv/reward/cmd/debug"
	"github.com/rewardenv/reward/cmd/env"
	"github.com/rewardenv/reward/cmd/envInit"
	"github.com/rewardenv/reward/cmd/install"
	"github.com/rewardenv/reward/cmd/selfUpdate"
	"github.com/rewardenv/reward/cmd/shell"
	"github.com/rewardenv/reward/cmd/signCertificate"
	"github.com/rewardenv/reward/cmd/svc"
	"github.com/rewardenv/reward/cmd/sync"
	"github.com/rewardenv/reward/cmd/version"
	"github.com/rewardenv/reward/internal/core"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	dockerClient "github.com/docker/docker/client"
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
	Use:   core.AppName + " [command]",
	Short: core.AppName + " is a cli tool which helps you to run local dev environments",
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
	Version: core.GetAppVersion().String(),
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

	addCommands()
	addFlags()

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

	appHomeDir = core.GetAppHomeDir()

	// app_ssl_dir and app_composer_dir have to be configured for templating
	if !viper.IsSet(core.AppName + "_ssl_dir") {
		viper.Set(core.AppName+"_ssl_dir", filepath.Join(appHomeDir, "ssl"))
	}

	if !viper.IsSet(core.AppName + "_composer_dir") {
		viper.Set(core.AppName+"_composer_dir", filepath.Join(core.GetHomeDir(), ".composer"))
	}

	if !viper.IsSet(core.AppName + "_ssh_dir") {
		viper.Set(core.AppName+"_ssh_dir", filepath.Join(core.GetHomeDir(), ".ssh"))
	}
}

func setLogLevel() {
	if core.IsDebug() {
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
	} else {
		switch logLevel = viper.GetString("log_level"); logLevel {
		case "trace":
			log.SetLevel(log.TraceLevel)
			log.SetReportCaller(true)
		case "debug":
			log.SetLevel(log.DebugLevel)
			log.SetReportCaller(true)
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
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := strings.Replace(path.Base(f.File), "github.com/rewardenv/reward", "", 1)
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
		},
	})

}

func configureHiddenCommands() {
	if !core.IsBlackfireEnabled() {
		blackfire.Cmd.Hidden = true
	}
}

// RootCmd represents the root command.
func RootCmd(cmd *cobra.Command) error {
	if viper.GetBool(core.AppName + "_print_environment") {
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
	if core.IsAllowedSuperuser() || core.GetOSDistro() == "windows" {
		return nil
	}

	// Most of the commands should run by normal users except `self-update`.
	if cmd.Name() != "self-update" && core.IsAdmin() {
		return core.ErrInvokedAsRootUser
	}

	return nil
}

func addCommands() {
	rootCmd.AddCommand(blackfire.Cmd)
	rootCmd.AddCommand(bootstrap.Cmd)
	rootCmd.AddCommand(completion.Cmd)
	rootCmd.AddCommand(db.Cmd)
	rootCmd.AddCommand(debug.Cmd)
	rootCmd.AddCommand(env.Cmd)
	rootCmd.AddCommand(envInit.Cmd)
	rootCmd.AddCommand(install.Cmd)
	rootCmd.AddCommand(selfUpdate.Cmd)
	rootCmd.AddCommand(shell.Cmd)
	rootCmd.AddCommand(signCertificate.Cmd)
	rootCmd.AddCommand(svc.Cmd)
	rootCmd.AddCommand(sync.Cmd)
	rootCmd.AddCommand(version.Cmd)
}

func addFlags() {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatalln(err)
	}

	// --app-dir
	rootCmd.PersistentFlags().StringVar(
		&appHomeDir, "app-dir", filepath.Join(home, "."+core.AppName), "app home directory")

	_ = viper.BindPFlag(core.AppName+"_home_dir", rootCmd.PersistentFlags().Lookup("app-dir"))

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
		&cfgFile, "config", "c", filepath.Join(core.GetHomeDir(), "."+core.AppName+".yml"), "config file")

	_ = viper.BindPFlag(core.AppName+"_config_file", rootCmd.PersistentFlags().Lookup("config"))

	// --docker-host
	rootCmd.PersistentFlags().String(
		"docker-host", dockerClient.DefaultDockerHost, "docker host")

	_ = viper.BindPFlag("docker_host", rootCmd.PersistentFlags().Lookup("docker-host"))

	if core.GetOSDistro() == "windows" {
		// --docker-host
		rootCmd.PersistentFlags().Bool(
			"wsl2-direct-mount", false, "use direct mount in WSL2 instead of syncing")

		_ = viper.BindPFlag(core.AppName+"_wsl2_direct_mount", rootCmd.PersistentFlags().Lookup("wsl2-direct-mount"))
	}

	// --driver
	// rootCmd.PersistentFlags().String(
	// 	"driver", "docker-compose", "orchestration driver")
	// _ = viper.BindPFlag(AppName+"_driver", rootCmd.PersistentFlags().Lookup("driver"))

	// --service-domain
	rootCmd.PersistentFlags().String(
		"service-domain", core.AppName+".test", "service domain for global services")

	rootCmd.PersistentFlags().Lookup("service-domain").Hidden = true

	_ = viper.BindPFlag(core.AppName+"_service_domain", rootCmd.PersistentFlags().Lookup("service-domain"))

	// --print-environment
	rootCmd.Flags().Bool(
		"print-environment", false, "environment vars")

	_ = viper.BindPFlag(core.AppName+"_print_environment", rootCmd.Flags().Lookup("print-environment"))
}
