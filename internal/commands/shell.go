package commands

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/rewardenv/reward/internal/core"
)

var (
	// ShellCommand is the command which is called in the ShellContainer.
	ShellCommand []string

	// DefaultShellCommand is the default shell for the container (the first element in ShellCommand).
	DefaultShellCommand string

	// ShellContainer is the container used for shell command.
	ShellContainer string

	// ShellUser represents the user of the container.
	ShellUser string
)

// ShellCmd opens a shell in the environment's default application container.
func ShellCmd(cmd *cobra.Command, args []string) error {
	SetShellContainer(core.EnvType())

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
func SetShellContainer(envType string) {
	ShellContainer = defaultShellContainer(envType)
}

// SetDefaultShellCommand changes the command invoked by reward shell command.
func SetDefaultShellCommand(containerName string) {
	DefaultShellCommand = defaultShellCommand(containerName)
}

// SetShellUser changes the user of the reward shell command.
func SetShellUser(containerName string) {
	ShellUser = defaultShellUser(containerName)
}

func defaultShellContainer(envType string) string {
	conf := viper.GetString(core.AppName + "_shell_container")
	if conf != "" {
		return conf
	}

	switch envType {
	case "pwa-studio":
		return "node"
	default:
		return "php-fpm"
	}
}

func defaultShellCommand(containerName string) string {
	conf := viper.GetString(core.AppName + "_shell_command")
	if conf != "" {
		return conf
	}

	switch containerName {
	case "php-fpm":
		return "bash"
	default:
		return "sh"
	}
}

func defaultShellUser(containerName string) string {
	conf := viper.GetString(core.AppName + "_shell_user")
	if conf != "" {
		return conf
	}

	switch containerName {
	case "php-fpm":
		return "www-data"
	case "node":
		return "node"
	default:
		return "root"
	}
}
