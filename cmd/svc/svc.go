package svc

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdpkg "reward/cmd"
	"reward/internal/config"
	"reward/internal/dockercompose"
	"reward/internal/logic"
)

func NewCmdSvc(c *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:                "svc",
			Short:              "Orchestrates global services such as traefik, portainer and dnsmasq via docker-compose",
			Long:               `Orchestrates global services such as traefik, portainer and dnsmasq via docker-compose`,
			ValidArgsFunction:  dockercompose.Completer(),
			DisableFlagParsing: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(c).RunCmdSvc(args)
				if err != nil {
					return fmt.Errorf("error running svc command: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}
}
