package commands

import (
	"github.com/rewardenv/reward/internal/core"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// ShellCommand is the command which is called in the ShellContainer.
	ShellCommand []string
	// ShellContainer is the container used for shell command.
	ShellContainer          string
	DefaultShellCommand     = "bash"
	defaultShellCommandsMap = map[string]string{
		"php-fpm": "bash",
		"default": "sh",
	}
	defaultShellContainersMap = map[string]string{
		"pwa-studio": "node",
		"default":    "php-fpm",
	}

	ShellUser    = ""
	shellUserMap = map[string]string{
		"php-fpm":    "www-data",
		"pwa-studio": "node",
		"default":    "root",
	}
)

// ShellCmd opens a shell in the environment's default application container.
func ShellCmd(cmd *cobra.Command, args []string) error {
	// For PWA Studio the default container should be "node" instead of "php-fpm"
	// and the shell should be "sh".
	if core.CheckRegexInString("^pwa-studio", core.GetEnvType()) {
		if ShellContainer == defaultShellContainersMap["default"] {
			SetShellContainer(defaultShellContainersMap[core.GetEnvType()])
		}
	}

	SetDefaultShellCommand(ShellContainer)
	SetShellUser(ShellContainer)

	if len(args) > 0 {
		ShellCommand = core.ExtractUnknownArgs(cmd.Flags(), args)
	} else {
		ShellCommand = core.ExtractUnknownArgs(cmd.Flags(), []string{DefaultShellCommand})
	}

	log.Debugln("command:", ShellCommand)
	log.Debugln("container:", ShellContainer)

	var passedArgs []string
	passedArgs = append(passedArgs, "exec")
	passedArgs = append(passedArgs, "--user", ShellUser)
	passedArgs = append(passedArgs, ShellContainer)
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
	if DefaultShellCommand == "" {
		if keyExists(defaultShellCommandsMap, s) {
			DefaultShellCommand = defaultShellCommandsMap[s]
		} else {
			DefaultShellCommand = defaultShellCommandsMap["default"]
		}
	}
}

// SetShellUser changes the user of the reward shell command.
func SetShellUser(s string) {
	if ShellUser == "" {
		if keyExists(shellUserMap, s) {
			ShellUser = shellUserMap[s]
		} else {
			ShellUser = shellUserMap["default"]
		}
	}
}

func keyExists(decoded map[string]string, key string) bool {
	val, ok := decoded[key]
	return ok && val != ""
}
