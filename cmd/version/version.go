package version

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"reward/cmd"
	"reward/internal/app"
	"reward/internal/shell"
)

func NewVersionCmd(app *app.App) *cmd.Command {
	versionCmd := &cmd.Command{
		Command: &cobra.Command{
			Use:   "version",
			Short: "Print the version information",
			Long:  fmt.Sprintf(`Print the version information for the %s application.`, app.Name()),
			ValidArgsFunction: func(
				cmd *cobra.Command,
				args []string,
				toComplete string,
			) ([]string, cobra.ShellCompDirective) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			Run: func(cmd *cobra.Command, args []string) {
				NewAppVersionCmd(app).Run(cmd, []string{})
			},
		},
		App: app,
	}

	// version
	versionCmd.Flags().BoolP("short", "s", false, "Print version only")

	appVersionCmd := NewAppVersionCmd(app)
	appVersionCmd.Flags().BoolP("short", "s", false, "Print version only")

	dockerVersionCmd := NewDockerVersionCmd(app)
	dockerVersionCmd.Flags().Bool("short", false, "print version only for docker server")
	dockerVersionCmd.Flags().BoolP("short-docker-version", "s", false, "print version only for docker server")
	dockerVersionCmd.Flags().BoolP("short-api-version", "a", false, "print version only for docker API")
	dockerVersionCmd.Flags().BoolP("short-platform-version", "p", false, "print docker platform")

	dockerComposeVersionCmd := NewDockerComposeVersionCmd(app)
	dockerComposeVersionCmd.Flags().BoolP("short", "s", false, "Print version only")

	versionCmd.AddCommands(appVersionCmd,
		dockerVersionCmd,
		dockerComposeVersionCmd,
	)

	return versionCmd
}

func NewAppVersionCmd(app *app.App) *cmd.Command {
	return &cmd.Command{
		Command: &cobra.Command{
			Use:   app.Name(),
			Short: fmt.Sprintf("Print the version information for %s", app.Name()),
			Long:  fmt.Sprintf(`Print the version information for the %s application.`, app.Name()),
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
					fmt.Printf("%s\n", app.Version().String())
					return
				}

				log.Printf("%s version: %s\n", app.Name(), app.Version().String())
				log.Printf("GOOS: %s\n", runtime.GOOS)
				log.Printf("GOARCH: %s\n", runtime.GOARCH)
			},
		},
		App: app,
	}
}

func NewDockerVersionCmd(app *app.App) *cmd.Command {
	return &cmd.Command{
		Command: &cobra.Command{
			Use:   "docker",
			Short: "Print the version information for docker",
			Long:  `Print the version information for Docker installed on your system.`,
			Run: func(cmd *cobra.Command, args []string) {
				data, err := app.Docker.ServerVersion(context.Background())
				if err != nil {
					log.Panicln(err)
				}

				shortDockerVersion, _ := cmd.Flags().GetBool("short-docker-version")
				short, _ := cmd.Flags().GetBool("short")
				shortAPI, _ := cmd.Flags().GetBool("short-api-version")
				shortPlatform, _ := cmd.Flags().GetBool("short-platform-version")

				switch {
				case short, shortDockerVersion:
					fmt.Printf("%s\n", data.Version)
				case shortAPI:
					fmt.Printf("%s\n", data.APIVersion)
				case shortPlatform:
					fmt.Printf("%s\n", data.Platform.Name)
				default:
					log.Printf("docker version: %s\n", data.Version)
					log.Printf("docker API version: %s\n", data.APIVersion)
					log.Printf("docker platform: %s\n", data.Platform.Name)
				}
			},
		},
		App: app,
	}
}

func NewDockerComposeVersionCmd(app *app.App) *cmd.Command {
	return &cmd.Command{
		Command: &cobra.Command{
			Use:   "docker-compose",
			Short: "Print the version information for docker-compose",
			Long:  `Print the version information for docker-compose installed on your system.`,
			Run: func(cmd *cobra.Command, args []string) {
				out, err := app.DockerCompose.RunCommand([]string{"version", "--short"},
					shell.WithSuppressOutput(true))
				if err != nil {
					log.Panicln(err)
				}

				short, _ := cmd.Flags().GetBool("short")
				if short {
					fmt.Printf("%s\n", strings.TrimSpace(string(out)))
					return
				}

				log.Printf("docker-compose version: %s", strings.TrimSpace(string(out)))
			},
		},
		App: app,
	}
}
