package sync

import (
	reward "github.com/rewardenv/reward/internal/commands"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "sync",
	Short: "Manipulate syncing",
	Long:  `Manipulate syncing`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	PreRun: func(syncCheckCmd *cobra.Command, args []string) {},
	Args:   cobra.ExactArgs(0),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := reward.SyncCheck()
		if err != nil {
			return err
		}

		reward.SetSyncSettingsByEnvType()

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
		err := reward.SyncStartCmd()
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
		err := reward.SyncStopCmd()
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
		_, err := reward.SyncListCmd()
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
		err := reward.SyncMonitorCmd()
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
		err := reward.SyncFlushCmd()
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
		err := reward.SyncPauseCmd()
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
		err := reward.SyncResumeCmd()
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
		err := reward.SyncResetCmd()
		if err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	Cmd.AddCommand(syncStartCmd)
	Cmd.AddCommand(syncStopCmd)
	Cmd.AddCommand(syncListCmd)
	Cmd.AddCommand(syncMonitorCmd)
	Cmd.AddCommand(syncFlushCmd)
	Cmd.AddCommand(syncPauseCmd)
	Cmd.AddCommand(syncResumeCmd)
	Cmd.AddCommand(syncResetCmd)
}
