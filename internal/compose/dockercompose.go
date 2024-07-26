package compose

import (
	"bufio"
	"bytes"
	"container/list"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/rewardenv/reward/internal/globals"
	"github.com/rewardenv/reward/internal/shell"
)

const (
	requiredVersionDockerCompose = ">=2.0"
)

var ErrDockerComposeVersionMismatch = func(s string) error {
	return errors.Errorf("docker compose version is too old: %s", s)
}

type DockerComposeClient struct {
	shell.Shell
	tmpFiles *list.List
}

type DockerComposeOpt func(*DockerComposeClient)

func NewDockerComposeClient(s shell.Shell, tmpFiles *list.List) *DockerComposeClient {
	c := &DockerComposeClient{
		Shell:    s,
		tmpFiles: tmpFiles,
	}

	return c
}

func NewMockClient(command string, output []byte, err error) *DockerComposeClient {
	return &DockerComposeClient{
		Shell: &shell.MockShell{
			LastCommand: command,
			Output:      output,
			Err:         err,
		},
	}
}

func (c *DockerComposeClient) Check() error {
	ver, err := c.Version()
	if err != nil {
		return errors.Wrap(err, "failed to fetch docker compose version")
	}

	if !c.IsMinimumVersionInstalled() {
		return ErrDockerComposeVersionMismatch(
			fmt.Sprintf(
				"your docker compose version is %s, required version: %s",
				ver.String(),
				requiredVersionDockerCompose,
			),
		)
	}

	return nil
}

func (c *DockerComposeClient) Version() (*version.Version, error) {
	log.Debugln("Checking docker compose version...")

	data, err := c.RunCommand([]string{"version", "--short"},
		shell.WithCatchOutput(),
		shell.WithSuppressOutput())
	if err != nil {
		return nil, err
	}

	v, err := version.NewVersion(strings.TrimSpace(string(data)))
	if err != nil {
		return nil, errors.Wrap(err, "parsing docker compose version")
	}

	log.Debugf("...docker compose version is: %s.", v.String())

	return v, nil
}

func (c *DockerComposeClient) IsMinimumVersionInstalled() bool {
	log.Debugln("Checking docker compose version requirements...")

	v, err := c.Version()
	if err != nil {
		log.Debugln("...docker compose version requirements not met.")

		return false
	}

	constraints := version.MustConstraints(version.NewConstraint(requiredVersionDockerCompose))
	if !constraints.Check(v) {
		log.Debugln("...docker compose version requirements not met.")

		return false
	}

	log.Debugln("...docker compose version requirements are met.")

	return true
}

// RunCommand runs the passed parameters with docker compose and returns the output.
func (c *DockerComposeClient) RunCommand(args []string, opts ...shell.Opt) (output []byte, err error) {
	log.Debugf("Running command: docker compose %s", strings.Join(args, " "))

	command := "docker"

	args = append([]string{"compose"}, args...)

	return c.ExecuteWithOptions(command, args, opts...)
}

// Completer returns a cobra Command completer function for docker compose.
func Completer() func(cmd *cobra.Command, args []string, toComplete string) (
	[]string, cobra.ShellCompDirective,
) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		args = append(args, "--help")
		out, _ := NewDockerComposeClient(
			shell.NewLocalShellWithOpts(),
			nil).
			RunCommand(args,
				shell.WithCatchOutput(),
				shell.WithSuppressOutput())

		commandsMatched := false
		scanner := bufio.NewScanner(bytes.NewReader(out))

		var words []string

		for scanner.Scan() {
			if commandsMatched {
				if strings.Contains(scanner.Text(), "docker compose COMMAND --help") {
					continue
				}

				fields := strings.Fields(scanner.Text())
				if len(fields) > 0 {
					words = append(words, fields[0])
				}
			}

			if strings.Contains(strings.ToLower(scanner.Text()), "commands:") {
				commandsMatched = true
			}
		}

		return words, cobra.ShellCompDirectiveNoFileComp
	}
}

// RunWithConfig calls docker compose with the converted configuration settings (from templates).
func (c *DockerComposeClient) RunWithConfig(args []string, details ConfigDetails, opts ...shell.Opt) (string, error) {
	tmpFiles := make([]string, 0, len(details.ConfigFiles))

	for _, conf := range details.ConfigFiles {
		log.Traceln("Reading config:")
		log.Traceln(conf.Filename)

		bs, err := yaml.Marshal(conf.Config)

		log.Traceln(string(bs))

		if err != nil {
			return "", errors.Wrap(err, "marshaling config")
		}

		tmpFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s-", globals.APPNAME))
		if err != nil {
			return "", errors.Wrap(err, "creating temporary file")
		}

		c.tmpFiles.PushBack(tmpFile.Name())

		tmpFiles = append(tmpFiles, tmpFile.Name())

		if _, err = tmpFile.Write(bs); err != nil {
			return "", errors.Wrap(err, "writing to temporary file")
		}

		if err := tmpFile.Close(); err != nil {
			return "", errors.Wrap(err, "closing temporary file")
		}
	}

	composeArgs := make([]string, 0, len(tmpFiles))
	for _, file := range tmpFiles {
		composeArgs = append(composeArgs, "-f")
		composeArgs = append(composeArgs, file)
	}

	composeArgs = append(composeArgs, args...)

	out, err := c.RunCommand(composeArgs, opts...)
	if err != nil {
		return string(out), err
	}

	return string(out), nil
}
