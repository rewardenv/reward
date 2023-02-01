package sync

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/internal/logic"
)

func NewCmdSync(c *config.Config) *cmdpkg.Command {
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
				// TODO
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
		Config: c,
	}

	cmd.AddCommands(
		newCmdSyncStart(c),
		newCmdSyncStop(c),
		newCmdSyncList(c),
		newCmdSyncMonitor(c),
		newCmdSyncFlush(c),
		newCmdSyncPause(c),
		newCmdSyncResume(c),
		newCmdSyncReset(c),
		newCmdSyncTerminate(c),
		// newCmdSyncDaemon(c),
	)

	return cmd
}

func newCmdSyncStart(c *config.Config) *cmdpkg.Command {
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
				err := logic.New(c).RunCmdSyncStart()
				if err != nil {
					return fmt.Errorf("error starting mutagen sync: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}
}

func newCmdSyncStop(c *config.Config) *cmdpkg.Command {
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
				err := logic.New(c).RunCmdSyncStop()
				if err != nil {
					return fmt.Errorf("error stopping mutagen sync: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}
}

func newCmdSyncList(c *config.Config) *cmdpkg.Command {
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
				_, err := logic.New(c).RunCmdSyncList()
				if err != nil {
					return fmt.Errorf("error listing mutagen sync: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}
}

func newCmdSyncMonitor(c *config.Config) *cmdpkg.Command {
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
				err := logic.New(c).RunCmdSyncMonitor()
				if err != nil {
					return fmt.Errorf("error monitoring mutagen sync: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}
}

func newCmdSyncFlush(c *config.Config) *cmdpkg.Command {
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
				err := logic.New(c).RunCmdSyncFlush()
				if err != nil {
					return fmt.Errorf("error flushing mutagen sync: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}
}

func newCmdSyncPause(c *config.Config) *cmdpkg.Command {
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
				err := logic.New(c).RunCmdSyncPause()
				if err != nil {
					return fmt.Errorf("error pausing mutagen sync: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}
}

func newCmdSyncResume(c *config.Config) *cmdpkg.Command {
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
				err := logic.New(c).RunCmdSyncResume()
				if err != nil {
					return fmt.Errorf("error resuming mutagen sync: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}
}

func newCmdSyncReset(c *config.Config) *cmdpkg.Command {
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
				err := logic.New(c).RunCmdSyncReset()
				if err != nil {
					return fmt.Errorf("error resetting mutagen sync: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}
}

func newCmdSyncTerminate(c *config.Config) *cmdpkg.Command {
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
				err := logic.New(c).RunCmdSyncTerminate()
				if err != nil {
					return fmt.Errorf("error terminating mutagen sync: %w", err)
				}

				return nil
			},
		},
		Config: c,
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
