package plugin

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/internal/logic"
)

func NewCmdPlugin(conf *config.Config) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:                   "plugin [flags]",
			DisableFlagsInUseLine: true,
			Short:                 "Provides utilities for interacting with plugins",
			Long:                  `Provides utilities for interacting with plugins.`,
			Run: func(cmd *cobra.Command, args []string) {
				cmdpkg.DefaultSubCommandRun()(cmd, args)
			},
		},
		Config: conf,
	}

	cmd.AddCommands(
		NewCmdPluginList(conf),
		NewCmdPluginListAvailable(conf),
		NewCmdPluginInstall(conf),
		NewCmdPluginRemove(conf),
	)

	return cmd
}

// NewCmdPluginList provides a way to list all plugin executables visible.
func NewCmdPluginList(c *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "list",
			Short: fmt.Sprintf("List all visible plugins in %s", c.PluginsDir()),
			Long:  fmt.Sprintf(`List all visible plugins in %s`, c.PluginsDir()),
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(c).RunCmdPluginList()
				if err != nil {
					return errors.Wrap(err, "listing plugins")
				}

				return nil
			},
		},
		Config: c,
	}
}

// NewCmdPluginListAvailable provides a way to list available remote plugins.
func NewCmdPluginListAvailable(conf *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:     "list-available",
			Short:   "List all available online plugins",
			Long:    `List all available online plugins`,
			Aliases: []string{"available"},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(conf).RunCmdPluginListAvailable()
				if err != nil {
					return errors.Wrap(err, "listing available plugins")
				}

				return nil
			},
		},
		Config: conf,
	}
}

// NewCmdPluginInstall provides a way to install available remote plugins.
func NewCmdPluginInstall(conf *config.Config) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "install pluginname",
			Short: "Install a plugin",
			Long:  `Install a plugin`,
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(conf).RunCmdPluginInstall(&cmdpkg.Command{Command: cmd, Config: conf},
					args)
				if err != nil {
					return errors.Wrap(err, "installing plugin")
				}

				return nil
			},
		},
		Config: conf,
	}

	cmd.Flags().BoolP("dry-run", "n", false, "only prints if there's new version available")
	cmd.Flags().BoolP("force", "f", false, "download and install the remote version even if its not newer")
	cmd.Flags().Bool("prerelease", false, "allow checking prerelease versions")

	return cmd
}

// NewCmdPluginRemove provides a way to delete installed plugins.
func NewCmdPluginRemove(conf *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "remove pluginname",
			Short: "Remove a plugin",
			Long:  `Remove a plugin`,
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(conf).RunCmdPluginRemove(&cmdpkg.Command{Command: cmd, Config: conf},
					args)
				if err != nil {
					return errors.Wrap(err, "removing plugin")
				}

				return nil
			},
		},
		Config: conf,
	}
}
