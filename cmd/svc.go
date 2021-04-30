package cmd

import (
	reward "github.com/rewardenv/reward/internal"
	"github.com/spf13/cobra"
)

var svcCmd = &cobra.Command{
	Use:                "svc",
	Short:              "Orchestrates global services such as traefik, portainer and dnsmasq via docker-compose",
	Long:               `Orchestrates global services such as traefik, portainer and dnsmasq via docker-compose`,
	ValidArgsFunction:  reward.DockerComposeCompleter(),
	DisableFlagParsing: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := reward.CheckIfInstalled(); err != nil {
			return err
		}

		if err := reward.CheckDocker(); err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return reward.SvcCmd(args)
	},
}

func init() {
	rootCmd.AddCommand(svcCmd)
}
