package logic

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/rewardenv/reward/pkg/util"
)

// RunCmdSPX opens a shell in the spx container.
func (c *Client) RunCmdSPX(cmd *cobra.Command, args []string) error {
	var command []string
	if len(args) > 0 {
		command = util.ExtractUnknownArgs(cmd.Flags(), args)
	} else {
		command = util.ExtractUnknownArgs(cmd.Flags(), []string{"bash"})
	}

	passedArgs := append(
		[]string{
			"exec",
			"php-spx",
		}, command...,
	)

	// Don't catch stdout
	err := c.RunCmdEnvDockerCompose(passedArgs)
	if err != nil {
		return errors.Wrap(err, "running docker compose command")
	}

	return nil
}
