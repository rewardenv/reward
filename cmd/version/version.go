package version

import (
	"context"
	"fmt"
	"strings"

	"github.com/rewardenv/reward/internal/core"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  fmt.Sprintf(`Print the version information for the %v application.`, core.AppName),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		versionAppCmd.Run(cmd, []string{})
	},
}

var versionAppCmd = &cobra.Command{
	Use:   core.AppName,
	Short: "Print the version information for " + core.AppName,
	Long:  `Print the version information for the ` + core.AppName + ` application.`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		short, err := cmd.Flags().GetBool("short")
		if err != nil {
			log.Fatalln(err)
		}
		if !short {
			log.Printf("%v version: %v\n", core.AppName, core.GetAppVersion().String())
		} else {
			log.Printf("%v\n", core.GetAppVersion().String())
		}
	},
}

var versionDockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Print the version information for docker",
	Long:  `Print the version information for Docker installed on your system.`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := core.NewDockerClient()
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
		out, err := core.RunDockerComposeCommand([]string{"version", "--short"}, true)
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
	Cmd.Flags().BoolP("short", "s", false, "Print version only")

	Cmd.AddCommand(versionAppCmd)
	versionAppCmd.Flags().BoolP("short", "s", false, "Print version only")

	// version docker
	Cmd.AddCommand(versionDockerCmd)
	versionDockerCmd.Flags().BoolP("short-docker-version", "s", false, "print version only for docker server")
	versionDockerCmd.Flags().BoolP("short-api-version", "a", false, "print version only for docker API")
	versionDockerCmd.Flags().BoolP("short-platform-version", "p", false, "print docker platform")

	// version docker-compose
	Cmd.AddCommand(versionDockerComposeCmd)
	versionDockerComposeCmd.Flags().BoolP("short", "s", false, "Print version only")
}
