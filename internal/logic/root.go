package logic

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	cmdpkg "github.com/rewardenv/reward/cmd"
)

// RunCmdRoot is the default command. If no additional args passed print the help.
func (c *Client) RunCmdRoot(cmd *cmdpkg.Command) error {
	if cmd.Config.GetBool(fmt.Sprintf("%s_print_environment", cmd.Name())) {
		for i, v := range viper.AllSettings() {
			log.Printf("%s=%v", strings.ToUpper(i), v)
		}

		return nil
	}

	_ = cmd.Help()

	return nil
}
