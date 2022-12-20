/*
Package cmd represents the commands of the application.

Copyright © 2022 JANOS MIKO <info@janosmiko.com>

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
package root

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"reward/cmd"
	"reward/cmd/blackfire"
	"reward/cmd/version"
	"reward/internal/app"
	"reward/internal/util"
	// "reward/cmd/bootstrap"
	// "reward/cmd/completion"
	// "reward/cmd/db"
	// "reward/cmd/debug"
	// "reward/cmd/env"
	// "reward/cmd/envinit"
	// "reward/cmd/install"
	// "reward/cmd/selfupdate"
	// "reward/cmd/shell"
	// "reward/cmd/signcertificate"
	// "reward/cmd/svc"
	// "reward/cmd/sync"
)

func NewRootCmd(app *app.App) *cmd.Command {
	cobra.EnableCommandSorting = false

	var rootCmd = &cmd.Command{
		&cobra.Command{
			Use: fmt.Sprintf("%s [command]", app.Name()),
			Short: fmt.Sprintf("%s is a cli tool which helps you to run local dev environments",
				app.Name()),
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
			Version: app.Version().String(),
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			PersistentPreRunE: func(c *cobra.Command, args []string) error {
				if err := validateFlags(&cmd.Command{Command: c, App: app}); err != nil {
					return err
				}

				app.Init()

				if err := app.Check(c); err != nil {
					return err
				}

				return nil
			},
			RunE: func(c *cobra.Command, args []string) error {
				return Run(&cmd.Command{Command: c, App: app})
			},
		},
		app,
	}

	rootCmd.AddGroups("Environment Commands:",
		blackfire.NewBlackfireCmd(app),
		// bootstrap.Cmd,
		// db.Cmd,
		// debug.Cmd,
		// env.Cmd,
		// shell.Cmd,
		// sync.Cmd,
	)

	rootCmd.AddGroups("Global Commands:") // envinit.Cmd,
	// install.Cmd,
	// selfupdate.Cmd,
	// signcertificate.Cmd,
	// svc.Cmd,

	rootCmd.AddCommands(
		// completion.Cmd,
		version.NewVersionCmd(app),
	)

	addFlags(rootCmd)
	configureHiddenCommands(rootCmd)

	return rootCmd
}

func addFlags(c *cmd.Command) {
	// --app-dir
	c.PersistentFlags().String(
		"app-dir",
		filepath.Join(util.HomeDir(), fmt.Sprintf(".%s", c.App.Name())),
		"app home directory",
	)
	viper.BindPFlag(fmt.Sprintf("%s_home_dir", c.App.Name()), c.PersistentFlags().Lookup("app-dir"))

	// --log-level
	c.PersistentFlags().String(
		"log-level", "info", "logging level (options: trace, debug, info, warning, error)",
	)
	viper.BindPFlag("log_level", c.PersistentFlags().Lookup("log-level"))

	// --debug
	c.PersistentFlags().Bool(
		"debug", false, "enable debug mode (same as --log-level=debug)",
	)
	viper.BindPFlag("debug", c.PersistentFlags().Lookup("debug"))

	// --disable-colors
	c.PersistentFlags().Bool(
		"disable-colors", false, "disable colors in output",
	)
	viper.BindPFlag("disable_colors", c.PersistentFlags().Lookup("disable-colors"))

	// --config
	c.PersistentFlags().StringP(
		"config",
		"c",
		filepath.Join(util.HomeDir(), fmt.Sprintf(".%s.yml", c.App.Name())),
		"config file",
	)
	viper.BindPFlag(fmt.Sprintf("%s_config_file", c.App.Name()), c.PersistentFlags().Lookup("config"))

	// --docker-host
	c.PersistentFlags().String(
		"docker-host", util.DockerHost(), "docker host",
	)
	viper.BindPFlag("docker_host", c.PersistentFlags().Lookup("docker-host"))

	if util.OSDistro() == "windows" {
		// --wsl2-direct-mount
		c.PersistentFlags().Bool(
			"wsl2-direct-mount", false, "use direct mount in WSL2 instead of syncing",
		)
		viper.BindPFlag(
			fmt.Sprintf("%s_wsl2_direct_mount", c.App.Name()),
			c.PersistentFlags().Lookup("wsl2-direct-mount"),
		)
	}

	// TODO
	// --driver
	c.PersistentFlags().String(
		"driver", "docker-compose", "orchestration driver")
	_ = viper.BindPFlag(fmt.Sprintf("%s_driver", c.App.Name()), c.PersistentFlags().Lookup("driver"))

	// --service-domain
	c.PersistentFlags().String(
		"service-domain", fmt.Sprintf("%s.test", c.App.Name()), "service domain for global services",
	)
	c.PersistentFlags().Lookup("service-domain").Hidden = true
	viper.BindPFlag(fmt.Sprintf("%s_service_domain", c.App.Name()), c.PersistentFlags().Lookup("service-domain"))

	// --print-environment
	c.Flags().Bool(
		"print-environment", false, "environment vars",
	)
	viper.BindPFlag(fmt.Sprintf("%s_print_environment", c.App.Name()), c.Flags().Lookup("print-environment"))
}

func configureHiddenCommands(cmd *cmd.Command) {
	if !cmd.App.Config.BlackfireEnabled() {
		for _, v := range cmd.Commands() {
			if v.Name() == "blackfire" {
				v.Hidden = true
			}
		}
	}
}

// Run represents the root command.
func Run(cmd *cmd.Command) error {
	if cmd.App.Config.GetBool(fmt.Sprintf("%s_print_environment", cmd.App.Name())) {
		for i, v := range viper.AllSettings() {
			log.Printf("%s=%v", strings.ToUpper(i), v)
		}

		return nil
	}

	_ = cmd.Help()

	return nil
}

func validateFlags(cmd *cmd.Command) error {
	driver := viper.GetString(fmt.Sprintf("%s_driver", cmd.App.Name()))
	if !regexp.MustCompile(`^docker-compose$`).MatchString(driver) {
		return fmt.Errorf("invalid value for --driver: %s", driver)
	}

	return nil
}
