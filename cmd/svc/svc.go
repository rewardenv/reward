package svc

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/compose"
	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/internal/logic"
)

func NewCmdSvc(conf *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:                "svc",
			Short:              "Orchestrates global services such as traefik, portainer and dnsmasq via docker compose",
			Long:               `Orchestrates global services such as traefik, portainer and dnsmasq via docker compose`,
			ValidArgsFunction:  compose.Completer(),
			DisableFlagParsing: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(conf).RunCmdSvc(args)
				if err != nil {
					return errors.Wrap(err, "running svc command")
				}

				return nil
			},
		},
		Config: conf,
	}
}
