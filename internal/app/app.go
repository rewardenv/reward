package app

import (
	"container/list"
	"fmt"
	"os"

	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"reward/internal/config"
	"reward/internal/docker"
	"reward/internal/dockercompose"
)

type App struct {
	name    string
	version string
	// CfgFile       string
	Config        *config.Config
	Docker        *docker.Client
	DockerCompose *dockercompose.Client
	tmpFiles      *list.List
}

func New(name, ver string) *App {
	return &App{
		name:     name,
		version:  ver,
		Config:   config.New(name, ver).Init(),
		tmpFiles: list.New(),
	}
}

func (a *App) Name() string {
	return a.name
}

func (a *App) Version() *version.Version {
	return version.Must(version.NewVersion(a.version))
}

func (a *App) Init() {
	a.Config.Init()
	a.Docker = docker.Must(docker.NewClient(a.Config))
	a.DockerCompose = dockercompose.NewClient()
}

func (a *App) Check(cmd *cobra.Command) error {
	err := a.Config.CheckInvokerUser(cmd)
	if err != nil {
		return err
	}

	err = a.Docker.Check()
	if err != nil {
		return err
	}

	err = a.DockerCompose.Check()
	if err != nil {
		return err
	}

	return nil
}

// Cleanup removes all the temporary template files.
func (a *App) Cleanup() error {
	log.Debugln("Cleaning up temporary files...")

	if a.tmpFiles.Len() == 0 {
		log.Debugln("...no temporary files to clean up.")

		return nil
	}

	for e := a.tmpFiles.Front(); e != nil; e = e.Next() {
		log.Tracef("Cleaning up: %s", e.Value)

		err := os.Remove(fmt.Sprint(e.Value))
		if err != nil {
			return fmt.Errorf("failed to remove temporary file: %w", err)
		}
	}

	log.Debugln("...cleanup done.")

	return nil
}
