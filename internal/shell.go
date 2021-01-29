package internal

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	ShellCommand   []string
	ShellContainer string
)

func ShellCmd(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		ShellCommand = ExtractUnknownArgs(cmd.Flags(), args)
	} else {
		ShellCommand = ExtractUnknownArgs(cmd.Flags(), []string{"bash"})
	}

	log.Debugln("command:", ShellCommand)
	log.Debugln("container:", ShellContainer)

	var passedArgs []string
	passedArgs = append(passedArgs, "exec", ShellContainer)
	passedArgs = append(passedArgs, ShellCommand...)

	err := EnvRunDockerCompose(passedArgs, false)
	if err != nil {
		return err
	}

	return nil
}
