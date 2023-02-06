package sync

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/internal/logic"
)

func NewCmdSync(conf *config.Config) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "sync",
			Short: "Manipulate syncing",
			Long:  `Manipulate syncing`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			PreRun: func(syncCheckCmd *cobra.Command, args []string) {},
			Args:   cobra.ExactArgs(0),
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				// TODO; Check if this is needed
				// err := reward.SyncCheck()
				// if err != nil {
				// 	return err
				// }
				//
				// reward.SetSyncSettings()

				return nil
			},
			Run: func(cmd *cobra.Command, args []string) {
				_ = cmd.Help()
			},
		},
		Config: conf,
	}

	cmd.AddCommands(
		newCmdSyncStart(conf),
		newCmdSyncStop(conf),
		newCmdSyncList(conf),
		newCmdSyncMonitor(conf),
		newCmdSyncFlush(conf),
		newCmdSyncPause(conf),
		newCmdSyncResume(conf),
		newCmdSyncReset(conf),
		newCmdSyncTerminate(conf),
		// newCmdSyncDaemon(conf),
	)

	return cmd
}

func newCmdSyncStart(conf *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "start",
			Short: "Starts mutagen sync for the current project environment",
			Long:  `Starts mutagen sync for the current project environment`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			Args: cobra.ExactArgs(0),
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(conf).RunCmdSyncStart()
				if err != nil {
					return fmt.Errorf("error starting mutagen sync: %w", err)
				}

				return nil
			},
		},
		Config: conf,
	}
}

func newCmdSyncStop(conf *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "stop",
			Short: "Stops the mutagen sync for the current project environment",
			Long:  `Stops the mutagen sync for the current project environment`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			Args: cobra.ExactArgs(0),
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(conf).RunCmdSyncStop()
				if err != nil {
					return fmt.Errorf("error stopping mutagen sync: %w", err)
				}

				return nil
			},
		},
		Config: conf,
	}
}

func newCmdSyncList(conf *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use: "list",
			Short: "Lists mutagen session status for current project environment and optionally (with -l) " +
				"the full configuration",
			Long: `Lists mutagen session status for current project environment and optionally (with -l) ` +
				`the full configuration`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			Args: cobra.ExactArgs(0),
			RunE: func(cmd *cobra.Command, args []string) error {
				_, err := logic.New(conf).RunCmdSyncList()
				if err != nil {
					return fmt.Errorf("error listing mutagen sync: %w", err)
				}

				return nil
			},
		},
		Config: conf,
	}
}

func newCmdSyncMonitor(conf *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "monitor",
			Short: "Continuously lists mutagen session status for current project",
			Long:  `Continuously lists mutagen session status for current project`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			Args: cobra.ExactArgs(0),
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(conf).RunCmdSyncMonitor()
				if err != nil {
					return fmt.Errorf("error monitoring mutagen sync: %w", err)
				}

				return nil
			},
		},
		Config: conf,
	}
}

func newCmdSyncFlush(conf *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "flush",
			Short: "Force a synchronization cycle on sync session for current project",
			Long:  `Force a synchronization cycle on sync session for current project`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			Args: cobra.ExactArgs(0),
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(conf).RunCmdSyncFlush()
				if err != nil {
					return fmt.Errorf("error flushing mutagen sync: %w", err)
				}

				return nil
			},
		},
		Config: conf,
	}
}

func newCmdSyncPause(conf *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "pause",
			Short: "Pauses the mutagen sync for the current project environment",
			Long:  `Pauses the mutagen sync for the current project environment`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			Args: cobra.ExactArgs(0),
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(conf).RunCmdSyncPause()
				if err != nil {
					return fmt.Errorf("error pausing mutagen sync: %w", err)
				}

				return nil
			},
		},
		Config: conf,
	}
}

func newCmdSyncResume(conf *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "resume",
			Short: "Resumes the mutagen sync for the current project environment",
			Long:  `Resumes the mutagen sync for the current project environment`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			Args: cobra.ExactArgs(0),
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(conf).RunCmdSyncResume()
				if err != nil {
					return fmt.Errorf("error resuming mutagen sync: %w", err)
				}

				return nil
			},
		},
		Config: conf,
	}
}

func newCmdSyncReset(conf *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "reset",
			Short: "Reset synchronization session history for current project environment",
			Long:  `Reset synchronization session history for current project environment`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			Args: cobra.ExactArgs(0),
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(conf).RunCmdSyncReset()
				if err != nil {
					return fmt.Errorf("error resetting mutagen sync: %w", err)
				}

				return nil
			},
		},
		Config: conf,
	}
}

func newCmdSyncTerminate(conf *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "terminate",
			Short: "Permanently terminate a synchronization session",
			Long:  `Permanently terminate a synchronization session`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			Args: cobra.ExactArgs(0),
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(conf).RunCmdSyncTerminate()
				if err != nil {
					return fmt.Errorf("error terminating mutagen sync: %w", err)
				}

				return nil
			},
		},
		Config: conf,
	}
}

// TODO
// func newCmdSyncDaemon(c *config.Config) *cmdpkg.Command {
// 	cmd := &cmdpkg.Command{
// 		Command: &cobra.Command{
// 			Use:   "daemon",
// 			Short: "Control the lifecycle of the Mutagen daemon",
// 			Long:  `Control the lifecycle of the Mutagen daemon`,
// 			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
// 				[]string, cobra.ShellCompDirective,
// 			) {
// 				return nil, cobra.ShellCompDirectiveNoFileComp
// 			},
// 			Args: cobra.ExactArgs(0),
// 			Run: func(cmd *cobra.Command, args []string) {
// 				err := logic.New(c).RunCmdSyncDaemon()
// 				if err != nil {
// 					fmt.Errorf("mutagen sync daemon failed: %w", err)
// 				}
// 			},
// 		},
// 		Config: c,
// 	}
//
// 	cmd.AddCommands(
// 		newSyncDaemon(c),
// 	)
//
// 	return cmd
// }
