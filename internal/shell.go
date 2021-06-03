package internal

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// ShellCommand is the command which is called in the ShellContainer.
	ShellCommand []string
	// ShellContainer is the container used for shell command.
	ShellContainer          string
	defaultShellCommand     = "bash"
	defaultShellCommandsMap = map[string]string{
		"default":    "bash",
		"pwa-studio": "sh",
	}
	defaultShellContainersMap = map[string]string{
		"default":    "php-fpm",
		"pwa-studio": "node",
	}
)

// ShellCmd opens a shell in the environment's default application container.
func ShellCmd(cmd *cobra.Command, args []string) error {
	// For PWA Studio the default container should be "node" instead of "php-fpm"
	// and the shell should be "sh".
	if CheckRegexInString("^pwa-studio", GetEnvType()) {
		if ShellContainer == defaultShellContainersMap["default"] {
			SetShellContainer(defaultShellContainersMap[GetEnvType()])
		}
		SetDefaultShellCommand(defaultShellCommandsMap[GetEnvType()])
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

// SetShellContainer changes the container used for the reward shell command.
func SetShellContainer(s string) {
	ShellContainer = s
}

// SetDefaultShellCommand changes the command invoked by reward shell command.
func SetDefaultShellCommand(s string) {
	defaultShellCommand = s
}
