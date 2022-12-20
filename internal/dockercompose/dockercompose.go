package dockercompose

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"

	"reward/internal/shell"
)

const (
	requiredVersion = "1.25.0"
)

var (
	ErrDockerComposeVersionMismatch = func(s string) error {
		return fmt.Errorf("docker-compose version is too old: %s", s)
	}
)

type Client struct {
	shell.Shell
}

func NewClient() *Client {
	return &Client{
		Shell: &shell.LocalShell{},
	}
}

func NewMockClient(command string, output []byte, err error) *Client {
	return &Client{
		Shell: &shell.MockShell{
			LastCommand: command,
			Output:      output,
			Err:         err,
		},
	}
}

func (c *Client) Check() error {
	ver, err := c.Version()
	if err != nil {
		return err
	}

	if !c.isMinimumVersionInstalled() {
		return ErrDockerComposeVersionMismatch(
			fmt.Sprintf(
				"your docker-compose version is %v, required version: %v",
				ver.String(),
				requiredVersion,
			),
		)
	}

	return nil
}

func (c *Client) Version() (*version.Version, error) {
	log.Debugln("Checking docker-compose version...")

	data, err := c.RunCommand([]string{"version", "--short"}, shell.WithSuppressOutput(true))
	if err != nil {
		return nil, err
	}

	v, err := version.NewVersion(strings.TrimSpace(string(data)))
	if err != nil {
		return nil, err
	}

	log.Debugf("...docker-compose version is: %s.", v.String())

	return v, err
}

func (c *Client) isMinimumVersionInstalled() bool {
	log.Debugln()

	v, err := c.Version()
	if err != nil {
		return false
	}

	if v.LessThan(version.Must(version.NewVersion(requiredVersion))) {
		return false
	}

	return true
}

// RunCommand runs the passed parameters with docker-compose and returns the output.
func (c *Client) RunCommand(args []string, opts ...shell.Opt) (output []byte, err error) {
	log.Debugf("Running command: docker-compose %v", strings.Join(args, " "))

	return c.ExecuteWithOptions("docker-compose", args, opts...)
}
