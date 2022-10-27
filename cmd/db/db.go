package db

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/rewardenv/reward/internal/commands"
	"github.com/rewardenv/reward/internal/core"
)

var Cmd = &cobra.Command{
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
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var dbConnectCmd = &cobra.Command{
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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := core.CheckDocker(); err != nil {
			return err
		}

		if err := commands.EnvCheck(); err != nil {
			return err
		}

		ContainerRunning, err := core.ContainerRunning("db")
		if err != nil {
			return err
		}
		if !core.IsDBEnabled() || !ContainerRunning {
			return core.CannotFindContainerError("db")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return commands.DBConnectCmd(cmd, args)
	},
}

var dbImportCmd = &cobra.Command{
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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := core.CheckDocker(); err != nil {
			return err
		}

		if err := commands.EnvCheck(); err != nil {
			return err
		}

		ContainerRunning, err := core.ContainerRunning("db")
		if err != nil {
			return err
		}
		if !core.IsDBEnabled() || !ContainerRunning {
			return core.CannotFindContainerError("db")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return commands.DBImportCmd(cmd, args)
	},
}

var dbDumpCmd = &cobra.Command{
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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := core.CheckDocker(); err != nil {
			return err
		}

		if err := commands.EnvCheck(); err != nil {
			return err
		}

		ContainerRunning, err := core.ContainerRunning("db")
		if err != nil {
			return err
		}
		if !core.IsDBEnabled() || !ContainerRunning {
			return core.CannotFindContainerError("db")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return commands.DBDumpCmd(cmd, args)
	},
}

func init() {
	Cmd.AddCommand(dbConnectCmd)
	dbConnectCmd.Flags().Bool("root", false, "connect as mysql root user")

	Cmd.AddCommand(dbImportCmd)
	dbImportCmd.Flags().Bool("root", false, "import as mysql root user")
	dbImportCmd.Flags().Int("line-buffer-size", 10, "line buffer size in mb for database import")

	Cmd.AddCommand(dbDumpCmd)
	dbDumpCmd.Flags().Bool("root", false, "dump database as mysql root user")

	_ = viper.BindPFlag("db_import_line_buffer_size", dbImportCmd.Flags().Lookup("line-buffer-size"))
}
