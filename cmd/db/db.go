package db

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/internal/docker"
	"github.com/rewardenv/reward/internal/logic"
)

func NewCmdDB(c *config.Config) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "db [command]",
			Short: "Interacts with the db service on an environment",
			Long:  `Interacts with the db service on an environment`,
			ValidArgsFunction: func(
				cmd *cobra.Command,
				args []string,
				toComplete string,
			) ([]string, cobra.ShellCompDirective) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				if !c.IsDBEnabled() || !c.Docker.ContainerRunning("db") {
					return docker.ErrCannotFindContainer("db", nil)
				}

				return nil
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				return cmd.Help()
			},
		},
		Config: c,
	}

	cmd.AddCommands(
		newCmdDBConnect(c),
		newCmdDBImport(c),
		newCmdDBDump(c),
	)

	return cmd
}

func newCmdDBConnect(c *config.Config) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "connect",
			Short: "Launches an interactive mysql session within the current project environment",
			Long:  `Launches an interactive mysql session within the current project environment`,
			ValidArgsFunction: func(
				cmd *cobra.Command,
				args []string,
				toComplete string,
			) ([]string, cobra.ShellCompDirective) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(c).RunCmdDBConnect(cmd, args)
				if err != nil {
					return fmt.Errorf("error running db connect command: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}

	cmd.Flags().Bool("root", false, "connect as mysql root user")

	return cmd
}

func newCmdDBImport(c *config.Config) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "import",
			Short: "Reads data from stdin and loads it into the current project's mysql database",
			Long:  `Reads data from stdin and loads it into the current project's mysql database`,
			ValidArgsFunction: func(
				cmd *cobra.Command,
				args []string,
				toComplete string,
			) ([]string, cobra.ShellCompDirective) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(c).RunCmdDBImport(cmd, args)
				if err != nil {
					return fmt.Errorf("error running db import command: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}

	cmd.Flags().Bool("root", false, "import as mysql root user")
	cmd.Flags().Int("line-buffer-size", 10, "line buffer size in mb for database import")
	_ = c.BindPFlag("db_import_line_buffer_size", cmd.Flags().Lookup("line-buffer-size"))

	return cmd
}

func newCmdDBDump(c *config.Config) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "dump",
			Short: "Dump the database from the DB container",
			Long:  `Dump the database from the DB container`,
			ValidArgsFunction: func(
				cmd *cobra.Command,
				args []string,
				toComplete string,
			) ([]string, cobra.ShellCompDirective) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(c).RunCmdDBDump(cmd, args)
				if err != nil {
					return fmt.Errorf("error running db dump command: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}

	cmd.Flags().Bool("root", false, "dump database as mysql root user")

	return cmd
}
