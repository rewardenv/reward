package logic

import (
	"strings"

	"github.com/pkg/errors"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/pkg/util"
)

// RunCmdBlackfire represents the blackfire command.
func (c *Client) RunCmdBlackfire(cmd *cmdpkg.Command, args []string) error {
	composeArgs := []string{
		"exec",
		cmd.Config.BlackfireContainer(),
		"sh",
		"-c", cmd.Config.BlackfireCommand(),
	}
	composeArgs = append(composeArgs, strings.Join(util.ExtractUnknownArgs(cmd.Flags(), args), " "))

	_, err := cmd.Config.Compose.RunCommand(composeArgs)
	if err != nil {
		return errors.Wrap(err, "running blackfire command")
	}

	return nil
}
