package logic

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rewardenv/reward/internal/shell"
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

	err := c.RunCmdEnvDockerCompose(passedArgs, shell.WithCatchOutput(false))
	if err != nil {
		return fmt.Errorf("error running docker compose command: %w", err)
	}

	return nil
}
