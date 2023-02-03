package version

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/internal/shell"
)

func NewCmdVersion(conf *config.Config) *cmd.Command {
	versionCmd := &cmd.Command{
		Command: &cobra.Command{
			Use:   "version",
			Short: "Print the version information",
			Long:  fmt.Sprintf(`Print the version information for the %s application.`, conf.AppName()),
			ValidArgsFunction: func(
				cmd *cobra.Command,
				args []string,
				toComplete string,
			) ([]string, cobra.ShellCompDirective) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			Run: func(cmd *cobra.Command, args []string) {
				NewCmdVersionApp(conf).Run(cmd, []string{})
			},
		},
		Config: conf,
	}

	// version
	versionCmd.Flags().BoolP("short", "s", false, "Print version only")

	appVersionCmd := NewCmdVersionApp(conf)
	appVersionCmd.Flags().BoolP("short", "s", false, "Print version only")

	dockerVersionCmd := NewCmdVersionDocker(conf)
	dockerVersionCmd.Flags().Bool("short", false, "print version only for docker server")
	dockerVersionCmd.Flags().BoolP("short-docker-version", "s", false, "print version only for docker server")
	dockerVersionCmd.Flags().BoolP("short-api-version", "a", false, "print version only for docker API")
	dockerVersionCmd.Flags().BoolP("short-platform-version", "p", false, "print docker platform")

	dockerComposeVersionCmd := NewCmdVersionDockerCompose(conf)
	dockerComposeVersionCmd.Flags().BoolP("short", "s", false, "Print version only")

	versionCmd.AddCommands(
		appVersionCmd,
		dockerVersionCmd,
		dockerComposeVersionCmd,
	)

	return versionCmd
}

func NewCmdVersionApp(conf *config.Config) *cmd.Command {
	return &cmd.Command{
		Command: &cobra.Command{
			Use:   conf.AppName(),
			Short: fmt.Sprintf("Print the version information for %s", conf.AppName()),
			Long:  fmt.Sprintf(`Print the version information for the %s application.`, conf.AppName()),
			ValidArgsFunction: func(
				cmd *cobra.Command,
				args []string,
				toComplete string,
			) ([]string, cobra.ShellCompDirective) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			Run: func(cmd *cobra.Command, args []string) {
				short, _ := cmd.Flags().GetBool("short")
				if short {
					//nolint:forbidigo
					fmt.Printf("%s\n", conf.AppVersion())

					return
				}

				log.Printf("%s version: %s\n", conf.AppName(), conf.AppVersion())
				log.Printf("GOOS: %s\n", runtime.GOOS)
				log.Printf("GOARCH: %s\n", runtime.GOARCH)
			},
		},
		Config: conf,
	}
}

func NewCmdVersionDocker(conf *config.Config) *cmd.Command {
	return &cmd.Command{
		Command: &cobra.Command{
			Use:   "docker",
			Short: "Print the version information for docker",
			Long:  `Print the version information for Docker installed on your system.`,
			Run: func(cmd *cobra.Command, args []string) {
				data, err := conf.Docker.ServerVersion(context.Background())
				if err != nil {
					log.Panicln(err)
				}

				shortDockerVersion, _ := cmd.Flags().GetBool("short-docker-version")
				short, _ := cmd.Flags().GetBool("short")
				shortAPI, _ := cmd.Flags().GetBool("short-api-version")
				shortPlatform, _ := cmd.Flags().GetBool("short-platform-version")

				switch {
				case short, shortDockerVersion:
					//nolint:forbidigo
					fmt.Printf("%s\n", data.Version)
				case shortAPI:
					//nolint:forbidigo
					fmt.Printf("%s\n", data.APIVersion)
				case shortPlatform:
					//nolint:forbidigo
					fmt.Printf("%s\n", data.Platform.Name)
				default:
					log.Printf("docker version: %s\n", data.Version)
					log.Printf("docker API version: %s\n", data.APIVersion)
					log.Printf("docker platform: %s\n", data.Platform.Name)
				}
			},
		},
		Config: conf,
	}
}

func NewCmdVersionDockerCompose(conf *config.Config) *cmd.Command {
	return &cmd.Command{
		Command: &cobra.Command{
			Use:   "docker-compose",
			Short: "Print the version information for docker-compose",
			Long:  `Print the version information for docker-compose installed on your system.`,
			Run: func(cmd *cobra.Command, args []string) {
				out, err := conf.DockerCompose.RunCommand([]string{"version", "--short"},
					shell.WithCatchOutput(true),
					shell.WithSuppressOutput(true),
				)
				if err != nil {
					log.Panicln(err)
				}

				short, _ := cmd.Flags().GetBool("short")
				if short {
					//nolint:forbidigo
					fmt.Printf("%s\n", strings.TrimSpace(string(out)))

					return
				}

				log.Printf("docker-compose version: %s", strings.TrimSpace(string(out)))
			},
		},
		Config: conf,
	}
}
