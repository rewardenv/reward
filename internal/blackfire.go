package internal

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// BlackfireCmd represents the blackfire command.
func BlackfireCmd(cmd *cobra.Command, args []string) error {
	command := []string{"sh", "-c", GetBlackfireCommand()}
	container := GetBlackfireContainer()

	log.Debugln("command:", command)
	log.Debugln("container:", container)

	var passedArgs []string

	passedArgs = append(passedArgs, "exec")
	passedArgs = append(passedArgs, container)
	passedArgs = append(passedArgs, command...)
	passedArgs = append(passedArgs, strings.Join(ExtractUnknownArgs(cmd.Flags(), args), " "))

	err := EnvRunDockerCompose(passedArgs, false)
	if err != nil {
		return err
	}

	return nil
}
