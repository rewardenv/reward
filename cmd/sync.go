package cmd

import (
	log "github.com/sirupsen/logrus"

	. "reward/internal"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Manipulate syncing",
	Long:  `Manipulate syncing`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	PreRun: func(syncCheckCmd *cobra.Command, args []string) {},
	Args:   cobra.ExactArgs(0),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := SyncCheck()
		if err != nil {
			return err
		}

		SetSyncVarsByEnvType()

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var syncStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts mutagen sync for the current project environment",
	Long:  `Starts mutagen sync for the current project environment`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := SyncStartCmd()
		if err != nil {
			log.Fatalln(err)
		}
	},
}

var syncStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops the mutagen sync for the current project environment",
	Long:  `Stops the mutagen sync for the current project environment`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := SyncStopCmd()
		if err != nil {
			log.Fatalln(err)
		}
	},
}

var syncListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists mutagen session status for current project environment and optionally (with -l) the full configuration",
	Long:  `Lists mutagen session status for current project environment and optionally (with -l) the full configuration`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		_, err := SyncListCmd()
		if err != nil {
			log.Fatalln(err)
		}
	},
}

var syncMonitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Continuously lists mutagen session status for current project",
	Long:  `Continuously lists mutagen session status for current project`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := SyncMonitorCmd()
		if err != nil {
			log.Fatalln(err)
		}
	},
}

var syncFlushCmd = &cobra.Command{
	Use:   "flush",
	Short: "Force a synchronization cycle on sync session for current project",
	Long:  `Force a synchronization cycle on sync session for current project`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := SyncFlushCmd()
		if err != nil {
			log.Fatalln(err)
		}
	},
}

var syncPauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pauses the mutagen sync for the current project environment",
	Long:  `Pauses the mutagen sync for the current project environment`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := SyncPauseCmd()
		if err != nil {
			log.Fatalln(err)
		}
	},
}

var syncResumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resumes the mutagen sync for the current project environment",
	Long:  `Resumes the mutagen sync for the current project environment`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := SyncResumeCmd()
		if err != nil {
			log.Fatalln(err)
		}
	},
}

var syncResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset synchronization session history for current project environment",
	Long:  `Reset synchronization session history for current project environment`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := SyncResetCmd()
		if err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.AddCommand(syncStartCmd)
	syncCmd.AddCommand(syncStopCmd)
	syncCmd.AddCommand(syncListCmd)
	syncCmd.AddCommand(syncMonitorCmd)
	syncCmd.AddCommand(syncFlushCmd)
	syncCmd.AddCommand(syncPauseCmd)
	syncCmd.AddCommand(syncResumeCmd)
	syncCmd.AddCommand(syncResetCmd)
}
