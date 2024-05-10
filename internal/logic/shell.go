package logic

import (
	"github.com/spf13/cobra"

	"github.com/rewardenv/reward/internal/shell"
	"github.com/rewardenv/reward/pkg/util"
)

// RunCmdShell opens a shell in the environment's default application container.
func (c *Client) RunCmdShell(cmd *cobra.Command, args []string) error {
	c.SetShellContainer(c.EnvType())
	c.SetDefaultShellCommand(c.ShellContainer)
	c.SetShellUser(c.ShellContainer)

	var shellCommand []string
	if len(args) > 0 {
		shellCommand = util.ExtractUnknownArgs(cmd.Flags(), args)
	} else {
		shellCommand = util.ExtractUnknownArgs(cmd.Flags(), []string{c.DefaultShellCommand})
	}

	passedArgs := append([]string{
		"exec",
		"--user",
		c.ShellUser,
		c.ShellContainer,
	}, shellCommand...)

	// Don't catch stdout
	err := c.RunCmdEnvDockerCompose(passedArgs,
		shell.WithSuppressOutput(),
	)
	if err != nil {
		return err
	}

	return nil
}
