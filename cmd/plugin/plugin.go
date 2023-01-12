package plugin

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdpkg "reward/cmd"
	"reward/internal/config"
	"reward/internal/logic"
)

func NewCmdPlugin(c *config.Config) *cmdpkg.Command {
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
		Config: c,
	}

	cmd.AddCommands(
		NewCmdPluginList(c),
		NewCmdPluginListAvailable(c),
		NewCmdPluginInstall(c),
		NewCmdPluginUpdate(c),
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
					return fmt.Errorf("error listing plugins: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}
}

// NewCmdPluginListAvailable provides a way to list available remote plugins.
func NewCmdPluginListAvailable(c *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:     "list-available",
			Short:   "List all available online plugins",
			Long:    `List all available online plugins`,
			Aliases: []string{"available"},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(c).RunCmdPluginListAvailable()
				if err != nil {
					return fmt.Errorf("error listing available plugins: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}
}

// NewCmdPluginInstall provides a way to install available remote plugins.
func NewCmdPluginInstall(c *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "install pluginname",
			Short: "Install a plugin",
			Long:  `Install a plugin`,
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(c).RunCmdPluginInstall(args)
				if err != nil {
					return fmt.Errorf("error installing plugin: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}
}

// NewCmdPluginUpdate provides a way to update an available remote plugin.
func NewCmdPluginUpdate(c *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:     "update pluginname",
			Aliases: []string{"upgrade"},
			Short:   "Update a plugin",
			Long:    `Update a plugin`,
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(c).RunCmdPluginUpdate(args)
				if err != nil {
					return fmt.Errorf("error updating plugin: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}
}
