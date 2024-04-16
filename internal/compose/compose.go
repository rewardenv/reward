package compose

import (
	"github.com/hashicorp/go-version"

	"github.com/rewardenv/reward/internal/shell"
)

type Client interface {
	Check() error
	Version() (*version.Version, error)
	IsMinimumVersionInstalled() bool
	RunCommand(args []string, opts ...shell.Opt) (output []byte, err error)
	RunWithConfig(args []string, details ConfigDetails, opts ...shell.Opt) (string, error)
}

// ConfigFile is a filename and the contents of the file as a Dict.
type ConfigFile struct {
	Filename string
	Config   map[string]any
}

// ConfigDetails are the details about a group of ConfigFiles.
type ConfigDetails struct {
	Version     string
	WorkingDir  string
	ConfigFiles []ConfigFile
	Environment map[string]string
}
