package cmd

import (
	"github.com/spf13/viper"

	. "reward/internal"

	"github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
	Use:   "db [command]",
	Short: "Interacts with the db service on an environment",
	Long:  `Interacts with the db service on an environment`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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

		if !IsDBEnabled() || !IsContainerRunning("db") {
			return CannotFindContainerError("db")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return DBConnectCmd(cmd, args)
	},
}

var dbImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Reads data from stdin and loads it into the current project's mysql database",
	Long:  `Reads data from stdin and loads it into the current project's mysql database`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := CheckDockerIsRunning(); err != nil {
			return err
		}

		if err := EnvCheck(); err != nil {
			return err
		}

		if !IsDBEnabled() || !IsContainerRunning("db") {
			return CannotFindContainerError("db")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return DBImportCmd(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(dbConnectCmd)
	dbConnectCmd.Flags().Bool("root", false, "connect as mysql root user")

	dbCmd.AddCommand(dbImportCmd)
	dbImportCmd.Flags().Bool("root", false, "import as mysql root user")

	dbImportCmd.Flags().Int("line-buffer-size", 10, "line buffer size in mb for database import")

	_ = viper.BindPFlag("db_import_line_buffer_size", dbImportCmd.Flags().Lookup("line-buffer-size"))
}
