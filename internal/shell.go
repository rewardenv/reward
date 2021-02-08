package internal

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	defaultShellCommand = "bash"
	ShellCommand        []string
	ShellContainer      string
)

func ShellCmd(cmd *cobra.Command, args []string) error {
	if CheckRegexInString("^pwa-studio", GetEnvType()) {
		SetShellContainer("node")
		SetDefaultShellCommand("sh")
	}

	if len(args) > 0 {
		ShellCommand = ExtractUnknownArgs(cmd.Flags(), args)
	} else {
		ShellCommand = ExtractUnknownArgs(cmd.Flags(), []string{defaultShellCommand})
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

func SetShellContainer(s string) {
	ShellContainer = s
}
func SetDefaultShellCommand(s string) {
	defaultShellCommand = s
}
