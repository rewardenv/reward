package cmd

import (
	"context"
	"fmt"
	"strings"

	reward "github.com/rewardenv/reward/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  fmt.Sprintf(`Print the version information for the %v application.`, reward.AppName),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		versionAppCmd.Run(cmd, []string{})
	},
}

var versionAppCmd = &cobra.Command{
	Use:   reward.AppName,
	Short: "Print the version information for " + reward.AppName,
	Long:  `Print the version information for the ` + reward.AppName + ` application.`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		short, err := cmd.Flags().GetBool("short")
		if err != nil {
			log.Fatalln(err)
		}
		if !short {
			log.Printf("%v version: %v\n", reward.AppName, reward.GetAppVersion().String())
		} else {
			log.Printf("%v\n", reward.GetAppVersion().String())
		}
	},
}

var versionDockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Print the version information for docker",
	Long:  `Print the version information for Docker installed on your system.`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := reward.NewDockerClient()
		if err != nil {
			log.Fatalln(err)
		}

		data, err := client.ServerVersion(context.Background())
		if err != nil {
			log.Fatalln(err)
		}

		short, err := cmd.Flags().GetBool("short-docker-version")
		if err != nil {
			log.Fatalln(err)
		}
		shortAPI, err := cmd.Flags().GetBool("short-api-version")
		if err != nil {
			log.Fatalln(err)
		}
		shortPlatform, err := cmd.Flags().GetBool("short-platform-version")
		if err != nil {
			log.Fatalln(err)
		}
		if short {
			log.Printf("%v\n", data.Version)
		} else if shortAPI {
			log.Printf("%v\n", data.APIVersion)
		} else if shortPlatform {
			log.Printf("%v\n", data.Platform.Name)
		} else {
			log.Printf("docker version: %v\n", data.Version)
			log.Printf("docker API version: %v\n", data.APIVersion)
			log.Printf("docker platform: %v\n", data.Platform.Name)
		}
	},
}

var versionDockerComposeCmd = &cobra.Command{
	Use:   "docker-compose",
	Short: "Print the version information for docker-compose",
	Long:  `Print the version information for docker-compose installed on your system.`,
	Run: func(cmd *cobra.Command, args []string) {
		out, err := reward.RunDockerComposeCommand([]string{"version", "--short"}, true)
		if err != nil {
			log.Fatalln(err)
		}

		short, err := cmd.Flags().GetBool("short")
		if err != nil {
			log.Fatalln(err)
		}

		if !short {
			log.Printf("docker-compose version: %v\n", strings.TrimSuffix(out, "\n"))
		} else {
			log.Printf("%v\n", strings.TrimSuffix(out, "\n"))
		}
	},
}

func init() {
	// version
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolP("short", "s", false, "Print version only")

	versionCmd.AddCommand(versionAppCmd)
	versionAppCmd.Flags().BoolP("short", "s", false, "Print version only")

	// version docker
	versionCmd.AddCommand(versionDockerCmd)
	versionDockerCmd.Flags().BoolP("short-docker-version", "s", false, "print version only for docker server")
	versionDockerCmd.Flags().BoolP("short-api-version", "a", false, "print version only for docker API")
	versionDockerCmd.Flags().BoolP("short-platform-version", "p", false, "print docker platform")

	// version docker-compose
	versionCmd.AddCommand(versionDockerComposeCmd)
	versionDockerComposeCmd.Flags().BoolP("short", "s", false, "Print version only")
}
