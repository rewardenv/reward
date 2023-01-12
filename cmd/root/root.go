package root

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	cmdpkg "reward/cmd"
	"reward/cmd/blackfire"
	"reward/cmd/bootstrap"
	"reward/cmd/completion"
	"reward/cmd/db"
	"reward/cmd/debug"
	"reward/cmd/env"
	"reward/cmd/envinit"
	"reward/cmd/install"
	"reward/cmd/plugin"
	"reward/cmd/selfupdate"
	"reward/cmd/shell"
	"reward/cmd/shortcuts"
	"reward/cmd/signcertificate"
	"reward/cmd/svc"
	"reward/cmd/sync"
	"reward/cmd/version"
	"reward/internal/config"
	"reward/internal/logic"
	"reward/internal/util"
)

func NewCmdRoot(c *config.Config) *cmdpkg.Command {
	cobra.EnableCommandSorting = false

	c.Init()

	var cmd = &cmdpkg.Command{
		Command: &cobra.Command{
			Use: fmt.Sprintf("%s [command]", c.AppName()),
			Short: fmt.Sprintf("%s is a cli tool which helps you to run local dev environments",
				c.AppName()),
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
			Version: c.AppVersion(),
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			SilenceErrors: c.SilenceErrors(),
			SilenceUsage:  true,
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				err := validateFlags(&cmdpkg.Command{Command: cmd, Config: c})
				if err != nil {
					return fmt.Errorf("an error occurred validating flags: %w", err)
				}

				err = c.Check(cmd, args)
				if err != nil {
					return fmt.Errorf("an error occurred checking requirements: %w", err)
				}

				return nil
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(c).RunCmdRoot(&cmdpkg.Command{Command: cmd, Config: c})
				if err != nil {
					return fmt.Errorf("an error occurred running command: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}

	configureFlags(cmd)
	c.Init()

	if c.EnvInitialized() {
		cmd.AddGroups("Environment Commands:",
			blackfire.NewBlackfireCmd(c),
			bootstrap.NewBootstrapCmd(c),
			db.NewCmdDB(c),
			debug.NewCmdDebug(c),
			env.NewCmdEnv(c),
			shell.NewCmdShell(c),
			sync.NewCmdSync(c),
		)
	}

	cmd.AddGroups("Global Commands:",
		envinit.NewCmdEnvInit(c),
		install.NewCmdInstall(c),
		selfupdate.NewCmdSelfUpdate(c),
		signcertificate.NewCmdSignCertificate(c),
		plugin.NewCmdPlugin(c),
		svc.NewCmdSvc(c),
	)

	cmd.AddCommands(
		completion.NewCompletionCmd(c),
		version.NewCmdVersion(c),
	)

	configurePlugins(cmd)
	configureShortcuts(cmd)
	configureHiddenCommands(cmd)

	return cmd
}

func configureFlags(cmd *cmdpkg.Command) {
	// --app-dir
	cmd.PersistentFlags().String(
		"app-dir",
		filepath.Join(util.HomeDir(), fmt.Sprintf(".%s", cmd.Config.AppName())),
		"app home directory",
	)
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_home_dir", cmd.Config.AppName()),
		cmd.PersistentFlags().Lookup("app-dir"))

	// --log-level
	cmd.PersistentFlags().String(
		"log-level", "info", "logging level (options: trace, debug, info, warning, error)",
	)
	_ = cmd.Config.BindPFlag("log_level", cmd.PersistentFlags().Lookup("log-level"))

	// --debug
	cmd.PersistentFlags().Bool(
		"debug", false, "enable debug mode (same as --log-level=debug)",
	)
	_ = cmd.Config.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))

	// --assume-yes
	cmd.PersistentFlags().BoolP(
		"assume-yes", "y", false, "Automatic yes to prompts.",
	)
	_ = cmd.Config.BindPFlag("assume_yes", cmd.PersistentFlags().Lookup("assume-yes"))

	// --disable-colors
	cmd.PersistentFlags().Bool(
		"disable-colors", false, "disable colors in output",
	)
	_ = cmd.Config.BindPFlag("disable_colors", cmd.PersistentFlags().Lookup("disable-colors"))

	// --config
	cmd.PersistentFlags().StringP(
		"config",
		"c",
		filepath.Join(util.HomeDir(), fmt.Sprintf(".%s.yml", cmd.Config.AppName())),
		"config file",
	)
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_config_file", cmd.Config.AppName()),
		cmd.PersistentFlags().Lookup("config"))

	// --docker-host
	cmd.PersistentFlags().String(
		"docker-host", util.DockerHost(), "docker host",
	)
	_ = cmd.Config.BindPFlag("docker_host", cmd.PersistentFlags().Lookup("docker-host"))

	// --driver
	cmd.PersistentFlags().String(
		"driver", "docker-compose", "orchestration driver")
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_driver", cmd.Config.AppName()), cmd.PersistentFlags().Lookup("driver"))

	// --service-domain
	cmd.PersistentFlags().String(
		"service-domain", fmt.Sprintf("%s.test", cmd.Config.AppName()), "service domain for global services",
	)
	cmd.PersistentFlags().Lookup("service-domain").Hidden = true
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_service_domain", cmd.Config.AppName()),
		cmd.PersistentFlags().Lookup("service-domain"))

	// --print-environment
	cmd.Flags().Bool(
		"print-environment", false, "environment vars",
	)
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_print_environment", cmd.Config.AppName()),
		cmd.Flags().Lookup("print-environment"))

	// --skip-cleanup
	cmd.Flags().Bool(
		"skip-cleanup", false, "skip cleanup of temporary files",
	)
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_skip_cleanup", cmd.Config.AppName()),
		cmd.Flags().Lookup("skip-cleanup"))
}

func configureHiddenCommands(cmd *cmdpkg.Command) {
	if !cmd.Config.BlackfireEnabled() {
		for _, v := range cmd.Commands() {
			if v.Name() == "blackfire" {
				v.Hidden = true
			}
		}
	}
}

func configurePlugins(cmd *cmdpkg.Command) {
	if len(os.Args) > 1 {
		cmdPathPieces := os.Args[1:]

		// only look for suitable extension executables if
		// the specified command does not already exist
		if _, _, err := cmd.Command.Find(cmdPathPieces); err != nil {
			// Also check the commands that will be added by Cobra.
			// These commands are only added once rootCmd.Execute() is called, so we
			// need to check them explicitly here.
			var cmdName string // first "non-flag" arguments

			for _, arg := range cmdPathPieces {
				if !strings.HasPrefix(arg, "-") {
					cmdName = arg

					break
				}
			}

			switch cmdName {
			case "help", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
				// Don't search for a plugin
			default:
				if err := cmd.HandlePluginCommand(cmdPathPieces); err != nil {
					log.Errorf("Error: %s\n", err)
					os.Exit(1)
				}
			}
		}
	}

	if len(cmd.Config.Plugins()) > 0 {
		cmd.AddPlugins()
	}
}

func configureShortcuts(cmd *cmdpkg.Command) {
	var sc []*cmdpkg.Command

	for k, v := range cmd.Config.Shortcuts() {
		sc = append(sc, shortcuts.NewCmdShortcut(cmd.Config, k, v))
	}

	cmd.AddGroups("Shortcuts:", sc...)
}

func validateFlags(cmd *cmdpkg.Command) error {
	driver := cmd.Config.GetString(fmt.Sprintf("%s_driver", cmd.Config.AppName()))
	if !regexp.MustCompile(`^docker-compose$`).MatchString(driver) {
		return fmt.Errorf("invalid value for --driver: %s", driver)
	}

	return nil
}
