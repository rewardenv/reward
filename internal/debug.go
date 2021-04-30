package internal

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// DebugCmd opens a shell in the debug container.
func DebugCmd(cmd *cobra.Command, args []string) error {
	var command []string
	if len(args) > 0 {
		command = ExtractUnknownArgs(cmd.Flags(), args)
	} else {
		command = ExtractUnknownArgs(cmd.Flags(), []string{"bash"})
	}

	log.Debugln("command:", command)
	log.Debugln("container:", "php-debug")

	debugHost, err := LookupContainerGatewayInNetwork("php-debug", GetEnvNetworkName())
	if err != nil {
		return err
	}

	envVarDebug := "XDEBUG_REMOTE_HOST=" + debugHost

	var passedArgs []string

	passedArgs = append(passedArgs, "exec")
	passedArgs = append(passedArgs, "-e", envVarDebug)
	passedArgs = append(passedArgs, "php-debug")
	passedArgs = append(passedArgs, command...)

	err = EnvRunDockerCompose(passedArgs, false)
	if err != nil {
		return err
	}

	return nil
}
