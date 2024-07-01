package logic

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

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
	// 	return errors.Wrap(err, "error getting debug container gateway")
	// }

	passedArgs := append(
		[]string{
			"exec",
			// "-e",
			// fmt.Sprintf("XDEBUG_REMOTE_HOST=%s", debugHost),
			"php-debug",
		}, command...,
	)

	// Don't catch stdout
	err := c.RunCmdEnvDockerCompose(passedArgs)
	if err != nil {
		return errors.Wrap(err, "running docker compose command")
	}

	return nil
}
