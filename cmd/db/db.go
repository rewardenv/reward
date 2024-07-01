package db

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/internal/docker"
	"github.com/rewardenv/reward/internal/logic"
)

func NewCmdDB(conf *config.Config) *cmdpkg.Command {
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
				if !conf.IsSvcEnabled("db") || !conf.Docker.ContainerRunning("db") {
					return docker.ErrCannotFindContainer("db", nil)
				}

				return nil
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := cmd.Help()
				if err != nil {
					return errors.Wrap(err, "running db command")
				}

				return nil
			},
		},
		Config: conf,
	}

	cmd.AddCommands(
		newCmdDBConnect(conf),
		newCmdDBImport(conf),
		newCmdDBDump(conf),
	)

	return cmd
}

func newCmdDBConnect(conf *config.Config) *cmdpkg.Command {
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
				err := logic.New(conf).RunCmdDBConnect(cmd, args)
				if err != nil {
					return errors.Wrap(err, "running db connect command")
				}

				return nil
			},
		},
		Config: conf,
	}

	cmd.Flags().Bool("root", false, "connect as mysql root user")

	return cmd
}

func newCmdDBImport(conf *config.Config) *cmdpkg.Command {
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
				err := logic.New(conf).RunCmdDBImport(cmd, args)
				if err != nil {
					return errors.Wrap(err, "running db import command")
				}

				return nil
			},
		},
		Config: conf,
	}

	cmd.Flags().Bool("root", false, "import as mysql root user")
	cmd.Flags().Int("line-buffer-size", 10, "line buffer size in mb for database import")
	_ = conf.BindPFlag("db_import_line_buffer_size", cmd.Flags().Lookup("line-buffer-size"))

	return cmd
}

func newCmdDBDump(conf *config.Config) *cmdpkg.Command {
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
				err := logic.New(conf).RunCmdDBDump(cmd, args)
				if err != nil {
					return errors.Wrap(err, "running db dump command")
				}

				return nil
			},
		},
		Config: conf,
	}

	cmd.Flags().Bool("root", false, "dump database as mysql root user")

	return cmd
}
