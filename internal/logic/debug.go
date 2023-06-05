package logic

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rewardenv/reward/internal/shell"
	"github.com/rewardenv/reward/pkg/util"
)

// RunCmdDebug opens a shell in the debug container.
func (c *Client) RunCmdDebug(cmd *cobra.Command, args []string) error {
	var command []string
	if len(args) > 0 {
		command = util.ExtractUnknownArgs(cmd.Flags(), args)
	} else {
		command = util.ExtractUnknownArgs(cmd.Flags(), []string{"bash"})
	}

	// debugHost, err := c.Docker.ContainerGatewayInNetwork("php-debug", c.EnvNetworkName())
	// if err != nil {
	// 	return fmt.Errorf("error getting debug container gateway: %w", err)
	// }

	passedArgs := append(
		[]string{
			"exec",
			// "-e",
			// fmt.Sprintf("XDEBUG_REMOTE_HOST=%s", debugHost),
			"php-debug",
		}, command...,
	)

	err := c.RunCmdEnvDockerCompose(passedArgs, shell.WithCatchOutput(false))
	if err != nil {
		return fmt.Errorf("error running docker compose command: %w", err)
	}

	return nil
}
