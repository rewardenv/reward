package compose

import (
	"bufio"
	"bytes"
	"container/list"
	"fmt"
	"os"
	"strings"

	dockercompose "github.com/docker/cli/cli/compose/types"
	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/rewardenv/reward/internal/shell"
)

const (
	requiredVersionDockerCompose   = "1.25.0"
	requiredVersionDockerComposeV2 = "2.0"
)

var ErrDockerComposeVersionMismatch = func(s string) error {
	return fmt.Errorf("docker-compose version is too old: %s", s)
}

type Client struct {
	shell.Shell
	TmpFiles     *list.List
	useComposeV2 bool
}

type Opt func(*Client)

func WithUseComposeV2() Opt {
	return func(c *Client) {
		c.useComposeV2 = true
	}
}

func NewClient(s shell.Shell, tmpFiles *list.List, opts ...Opt) *Client {
	c := &Client{
		Shell:    s,
		TmpFiles: tmpFiles,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) AppName() string {
	return viper.GetString("app_name")
}

func (c *Client) EnvName() string {
	return strings.ToLower(viper.GetString(fmt.Sprintf("%s_env_name", c.AppName())))
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
		return fmt.Errorf("failed to fetch docker-compose version: %w", err)
	}

	if !c.isMinimumVersionInstalled() {
		return ErrDockerComposeVersionMismatch(
			fmt.Sprintf(
				"your docker-compose version is %s, required version: %s",
				ver.String(),
				c.requiredVersion(),
			),
		)
	}

	return nil
}

func (c *Client) Version() (*version.Version, error) {
	log.Debugln("Checking docker-compose version...")

	data, err := c.RunCommand([]string{"version", "--short"},
		shell.WithCatchOutput(true),
		shell.WithSuppressOutput(true))
	if err != nil {
		return nil, err
	}

	v, err := version.NewVersion(strings.TrimSpace(string(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse docker-compose version: %w", err)
	}

	log.Debugf("...docker-compose version is: %s.", v.String())

	return v, nil
}

func (c *Client) isMinimumVersionInstalled() bool {
	log.Debugln("Checking docker-compose version requirements...")

	v, err := c.Version()
	if err != nil {
		log.Debugln("...docker-compose version requirements not met.")

		return false
	}

	if v.LessThan(version.Must(version.NewVersion(c.requiredVersion()))) {
		log.Debugln("...docker-compose version requirements not met.")

		return false
	}

	log.Debugln("...docker-compose version requirements are met.")

	return true
}

// RunCommand runs the passed parameters with docker-compose and returns the output.
func (c *Client) RunCommand(args []string, opts ...shell.Opt) (output []byte, err error) {
	log.Debugf("Running command: docker-compose %s", strings.Join(args, " "))

	command := "docker-compose"
	if c.useComposeV2 {
		args = append([]string{"compose"}, args...)
		command = "docker"
	}

	return c.ExecuteWithOptions(command, args, opts...)
}

// Completer returns a cobra Command completer function for docker-compose.
func Completer() func(cmd *cobra.Command, args []string, toComplete string) (
	[]string, cobra.ShellCompDirective,
) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		args = append(args, "--help")
		out, _ := NewClient(shell.NewLocalShellWithOpts(), nil).RunCommand(args,
			shell.WithCatchOutput(true),
			shell.WithSuppressOutput(true))

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

// RunWithConfig calls docker-compose with the converted configuration settings (from templates).
func (c *Client) RunWithConfig(args []string, details dockercompose.ConfigDetails, opts ...shell.Opt) (string, error) {
	tmpFiles := make([]string, 0, len(details.ConfigFiles))

	for _, conf := range details.ConfigFiles {
		log.Traceln("Reading config:")
		log.Traceln(conf.Filename)

		bs, err := yaml.Marshal(conf.Config)

		log.Traceln(string(bs))

		if err != nil {
			return "", fmt.Errorf("failed to marshal config: %w", err)
		}

		tmpFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s-", c.AppName()))
		if err != nil {
			return "", fmt.Errorf("failed to create temporary file: %w", err)
		}

		c.TmpFiles.PushBack(tmpFile.Name())

		tmpFiles = append(tmpFiles, tmpFile.Name())

		if _, err = tmpFile.Write(bs); err != nil {
			return "", fmt.Errorf("failed to write to temporary file: %w", err)
		}

		if err := tmpFile.Close(); err != nil {
			return "", fmt.Errorf("failed to close temporary file: %w", err)
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

func (c *Client) requiredVersion() string {
	if c.useComposeV2 {
		return requiredVersionDockerComposeV2
	}

	return requiredVersionDockerCompose
}
