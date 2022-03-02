package commands

import (
	"strings"

	"github.com/rewardenv/reward/internal/core"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// BlackfireCmd represents the blackfire command.
func BlackfireCmd(cmd *cobra.Command, args []string) error {
	command := []string{"sh", "-c", core.GetBlackfireCommand()}
	container := core.GetBlackfireContainer()

	log.Debugln("command:", command)
	log.Debugln("container:", container)

	var passedArgs []string

	passedArgs = append(passedArgs, "exec")
	passedArgs = append(passedArgs, container)
	passedArgs = append(passedArgs, command...)
	passedArgs = append(passedArgs, strings.Join(core.ExtractUnknownArgs(cmd.Flags(), args), " "))

	err := EnvRunDockerCompose(passedArgs, false)
	if err != nil {
		return err
	}

	return nil
}
