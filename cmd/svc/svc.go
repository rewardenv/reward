package svc

import (
	"github.com/rewardenv/reward/internal/commands"
	"github.com/rewardenv/reward/internal/core"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:                "svc",
	Short:              "Orchestrates global services such as traefik, portainer and dnsmasq via docker-compose",
	Long:               `Orchestrates global services such as traefik, portainer and dnsmasq via docker-compose`,
	ValidArgsFunction:  core.DockerComposeCompleter(),
	DisableFlagParsing: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := commands.CheckIfInstalled(); err != nil {
			return err
		}

		if err := core.CheckDocker(); err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return commands.SvcCmd(args)
	},
}
