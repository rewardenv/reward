package sync

import (
	"github.com/pkg/errors"
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
				if err := logic.New(conf).RunCmdSyncStart(); err != nil {
					return errors.Wrap(err, "starting mutagen sync")
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
				if err := logic.New(conf).RunCmdSyncStop(); err != nil {
					return errors.Wrap(err, "stopping mutagen sync")
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
				if _, err := logic.New(conf).RunCmdSyncList(); err != nil {
					return errors.Wrap(err, "error listing mutagen sync")
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
				if err := logic.New(conf).RunCmdSyncMonitor(); err != nil {
					return errors.Wrap(err, "monitoring mutagen sync")
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
				if err := logic.New(conf).RunCmdSyncFlush(); err != nil {
					return errors.Wrap(err, "flushing mutagen sync")
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
				if err := logic.New(conf).RunCmdSyncPause(); err != nil {
					return errors.Wrap(err, "pausing mutagen sync")
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
				if err := logic.New(conf).RunCmdSyncResume(); err != nil {
					return errors.Wrap(err, "resuming mutagen sync")
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
				if err := logic.New(conf).RunCmdSyncReset(); err != nil {
					return errors.Wrap(err, "resetting mutagen sync")
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
				if err := logic.New(conf).RunCmdSyncTerminate(); err != nil {
					return errors.Wrap(err, "terminating mutagen sync")
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
// 					errors.Wrap(err, "mutagen sync daemon failed")
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
