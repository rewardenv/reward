package shortcuts

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/config"
)

func NewCmdShortcut(conf *config.Config, name, target string) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:                   name,
			Short:                 fmt.Sprintf(`Shortcut target: "%s"`, target),
			DisableFlagsInUseLine: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := executeShortcuts(target); err != nil {
					return errors.Wrap(err, "executing shortcut")
				}

				return nil
			},
		},
		Config: conf,
	}

	return cmd
}

func executeShortcuts(remainingCommands string) error {
out:
	for {
		remainingCommands = strings.TrimSpace(remainingCommands)

		// no more chains, run the last part
		if !strings.Contains(remainingCommands, "&&") && !strings.Contains(remainingCommands, ";") {
			if err := exec(strings.Split(strings.TrimSpace(remainingCommands), " ")); err != nil {
				return errors.Wrap(err, "executing last command")
			}

			return nil
		}

		switch searchFirst(remainingCommands) {
		case -1:
			break out
		case 0:
			// && chain
			parts := strings.SplitN(remainingCommands, "&&", 2)
			thisCommand := strings.TrimSpace(parts[0])
			remainingCommands = strings.TrimSpace(parts[1])

			if err := exec(strings.Split(thisCommand, " ")); err != nil {
				log.Errorf(
					"Error executing `%s`. Stopping shortcut execution.",
					thisCommand,
				)

				return errors.Wrap(err, "executing command")
			}

		case 1:
			// ; chain
			parts := strings.SplitN(remainingCommands, ";", 2)
			thisCommand := strings.TrimSpace(parts[0])
			remainingCommands = strings.TrimSpace(parts[1])

			if err := exec(strings.Split(thisCommand, " ")); err != nil {
				log.Warnf("Error executing `%s`. Executing next part of the shortcut: `%s`.",
					thisCommand,
					remainingCommands,
				)
			}
		}
	}

	return nil
}

func searchFirst(s string) int {
	andIdx := strings.Index(s, "&&")
	semiIdx := strings.Index(s, ";")

	switch {
	case andIdx == -1 && semiIdx == -1:
		return -1

	// ; is first
	case andIdx == -1:
		return 1

	// && is first
	case semiIdx == -1:
		return 0

	// both are present
	// && is first
	case andIdx < semiIdx:
		return 0

	// both are present
	// ; is first
	default:
		return 1
	}
}

func exec(args []string) error {
	currentCommand, _ := os.Executable()

	if err := cmdpkg.Run(currentCommand, args, os.Environ()); err != nil {
		return err
	}

	return nil
}
