package logic

import (
	"bufio"
	"bytes"
	"container/list"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"

	compose "github.com/docker/cli/cli/compose/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/rewardenv/reward/internal/shell"
	"github.com/rewardenv/reward/internal/templates"
	"github.com/rewardenv/reward/pkg/util"
)

// RunCmdDBConnect connects to the environment's database container.
func (c *Client) RunCmdDBConnect(cmd *cobra.Command, args []string) error {
	runAsRootUser, err := cmd.Flags().GetBool("root")
	if err != nil {
		return fmt.Errorf("failed to get flag: %w", err)
	}

	mysqlDBParam := "--database=$(printenv MYSQL_DATABASE)"

	var mysqlUserParam, mysqlPasswordParam string
	if runAsRootUser {
		mysqlUserParam = "-uroot"
		mysqlPasswordParam = "-p$(printenv MYSQL_ROOT_PASSWORD)" //nolint:gosec
	} else {
		mysqlUserParam = "-u$(printenv MYSQL_USER)"
		mysqlPasswordParam = "-p$(printenv MYSQL_PASSWORD)" //nolint:gosec
	}

	passedArgs := []string{
		"exec",
		c.DBContainer(),
		"sh",
		"-c",
		fmt.Sprintf("%s %s %s %s %s",
			c.DBCommand(),
			mysqlUserParam,
			mysqlPasswordParam,
			mysqlDBParam,
			strings.Join(util.ExtractUnknownArgs(cmd.Flags(), args), " "),
		),
	}

	err = c.RunCmdEnvDockerCompose(passedArgs, shell.WithCatchOutput(false))
	if err != nil {
		return fmt.Errorf("failed to run docker-compose to establish connection: %w", err)
	}

	return nil
}

// RunCmdDBImport imports a database from stdin to the environment's database container.
func (c *Client) RunCmdDBImport(cmd *cobra.Command, args []string) error {
	runAsRootUser, err := cmd.Flags().GetBool("root")
	if err != nil {
		return fmt.Errorf("failed to get flag: %w", err)
	}

	mysqlDBParam := "--database=$(printenv MYSQL_DATABASE)"

	var mysqlUserParam, mysqlPasswordParam string
	if runAsRootUser {
		mysqlUserParam = "-uroot"
		mysqlPasswordParam = "-p$(printenv MYSQL_ROOT_PASSWORD)" //nolint:gosec
	} else {
		mysqlUserParam = "-u$(printenv MYSQL_USER)"
		mysqlPasswordParam = "-p$(printenv MYSQL_PASSWORD)" //nolint:gosec
	}

	// FIXME: ExtractUnknownArgs not working here
	passedArgs := []string{
		"exec",
		"-T",
		c.DBContainer(),
		"sh",
		"-c",
		fmt.Sprintf("%s %s %s %s %s",
			c.DBCommand(),
			mysqlUserParam,
			mysqlPasswordParam,
			mysqlDBParam,
			strings.Join(util.ExtractUnknownArgs(cmd.Flags(), args), " "),
		),
	}

	err = c.RunCmdDBDockerCompose(passedArgs, false)
	if err != nil {
		return fmt.Errorf("failed to run docker-compose to import database: %w", err)
	}

	return nil
}

// RunCmdDBDump dumps the database from the environment's database container.
func (c *Client) RunCmdDBDump(cmd *cobra.Command, args []string) error {
	runAsRootUser, err := cmd.Flags().GetBool("root")
	if err != nil {
		return fmt.Errorf("failed to get flag: %w", err)
	}

	mysqlDBParam := "$(printenv MYSQL_DATABASE)"

	var mysqlUserParam, mysqlPasswordParam string
	if runAsRootUser {
		mysqlUserParam = "-uroot"
		mysqlPasswordParam = "-p$(printenv MYSQL_ROOT_PASSWORD)" //nolint:gosec
	} else {
		mysqlUserParam = "-u$(printenv MYSQL_USER)"
		mysqlPasswordParam = "-p$(printenv MYSQL_PASSWORD)" //nolint:gosec
	}

	passedArgs := []string{
		"exec",
		"-T",
		c.DBContainer(),
		"sh",
		"-c",
		fmt.Sprintf(
			"%s %s %s %s %s",
			c.DBDumpCommand(),
			mysqlUserParam,
			mysqlPasswordParam,
			mysqlDBParam,
			strings.Join(util.ExtractUnknownArgs(cmd.Flags(), args), " "),
		),
	}

	err = c.RunCmdDBDockerCompose(passedArgs, false)
	if err != nil {
		return fmt.Errorf("failed to run docker-compose to dump database: %w", err)
	}

	return nil
}

// RunCmdDBDockerCompose function is a wrapper around the docker-compose command.
// It appends the current directory and current project name to the args.
// It also changes the output if the OS StdOut is suppressed.
func (c *Client) RunCmdDBDockerCompose(args []string, suppressOsStdOut ...bool) error {
	passedArgs := []string{
		"--project-directory",
		c.Cwd(),
		"--project-name",
		c.EnvName(),
	}
	passedArgs = append(passedArgs, args...)

	// run docker-compose command
	out, err := c.RunCmdDBBuildDockerComposeCommand(passedArgs, suppressOsStdOut...)
	out = regexp.MustCompile("(?m)[\r\n]+^.*--file.*$").ReplaceAllString(out, "")
	out = regexp.MustCompile("(?m)[\r\n]+^.*--project-name.*$").ReplaceAllString(out, "")
	out = regexp.MustCompile("(?m)[\r\n]+^.*--project-directory.*$").ReplaceAllString(out, "")
	out = regexp.MustCompile("(?m)[\r\n]+^.*default: docker-compose.yml.*$").ReplaceAllString(out, "")
	out = regexp.MustCompile("(?m)[\r\n]+^.*default: directory name.*$").ReplaceAllString(out, "")
	out = strings.ReplaceAll(out, "docker-compose", "env")

	_, _ = fmt.Fprint(os.Stdout, out)

	if err != nil {
		return fmt.Errorf("failed to run docker-compose: %w", err)
	}

	return nil
}

// DBBuildDockerComposeCommand builds up the docker-compose command's templates.
func (c *Client) RunCmdDBBuildDockerComposeCommand(args []string, suppressOsStdOut ...bool) (string, error) {
	dbTemplate := new(template.Template)
	dbTemplateList := list.New()

	err := c.RunCmdEnvBuildDockerComposeTemplate(dbTemplate, dbTemplateList)
	if err != nil {
		return "", err
	}

	dockerComposeConfigs, err := templates.New().ConvertTemplateToComposeConfig(dbTemplate, dbTemplateList)
	if err != nil {
		return "", err
	}

	out, err := c.RunCmdDBDockerComposeWithConfig(args, dockerComposeConfigs, suppressOsStdOut...)
	if err != nil {
		return out, err
	}

	return out, nil
}

// RunCmdDBDockerComposeWithConfig calls docker-compose with the previously built docker-compose configuration.
func (c *Client) RunCmdDBDockerComposeWithConfig(
	args []string,
	details compose.ConfigDetails,
	suppressOsStdOut ...bool,
) (string, error) {
	tmpFiles := make([]string, len(details.ConfigFiles))

	for i, conf := range details.ConfigFiles {
		bs, err := yaml.Marshal(conf.Config)
		if err != nil {
			return "", fmt.Errorf("failed to marshal config: %w", err)
		}

		tmpFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s-", c.AppName()))
		if err != nil {
			return "", fmt.Errorf("failed to create temporary file: %w", err)
		}

		c.TmpFiles.PushBack(tmpFile.Name())
		tmpFiles[i] = tmpFile.Name()

		_, err = tmpFile.Write(bs)
		if err != nil {
			return "", fmt.Errorf("failed to write to temporary file: %w", err)
		}

		err = tmpFile.Close()
		if err != nil {
			return "", fmt.Errorf("failed to close temporary file: %w", err)
		}
	}

	composeArgs := make([]string, 0, len(tmpFiles))
	for _, file := range tmpFiles {
		composeArgs = append(composeArgs, "-f", file)
	}

	composeArgs = append(composeArgs, args...)

	out, err := c.RunCmdDBDockerComposeCommandModifyStdin(composeArgs, suppressOsStdOut...)
	if err != nil {
		return out, fmt.Errorf("failed to run docker-compose: %w", err)
	}

	return out, nil
}

// RunCmdDBDockerComposeCommandModifyStdin runs the passed parameters with docker-compose and returns the output.
func (c *Client) RunCmdDBDockerComposeCommandModifyStdin(args []string, suppressOsStdOut ...bool) (string, error) {
	cmd := exec.Command("docker-compose", args...)

	var combinedOutBuf bytes.Buffer

	r, w := io.Pipe()
	definerRegex := regexp.MustCompile("DEFINER[ ]*=[ ]*`[^`]+`@`[^`]+`")
	globalRegex := regexp.MustCompile(`@@(GLOBAL\.GTID_PURGED|SESSION\.SQL_LOG_BIN)`)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)

		maxCapacity := c.GetInt("db_import_line_buffer_size") * 1024 * 1024 // max capacity for buffer is 10MB/line
		bs := make([]byte, 0, 1024*1024)
		scanner.Buffer(bs, maxCapacity)

		for scanner.Scan() {
			_, _ = fmt.Fprintln(
				w, globalRegex.ReplaceAllString(
					definerRegex.ReplaceAllString(
						scanner.Text(),
						"DEFINER=CURRENT_USER",
					),
					"",
				),
			)
		}

		err := scanner.Err()
		if err != nil {
			log.Errorf("An error occurred: %s", err)

			os.Exit(1)
		}

		defer func(w *io.PipeWriter) {
			_ = w.Close()
		}(w)
	}()

	cmd.Stdin = r
	if len(suppressOsStdOut) > 0 && suppressOsStdOut[0] {
		cmd.Stdout = io.Writer(&combinedOutBuf)
		cmd.Stderr = io.Writer(&combinedOutBuf)
	} else {
		cmd.Stdout = io.Writer(os.Stdout)
		cmd.Stderr = io.Writer(os.Stderr)
	}

	err := cmd.Run()
	outStr := combinedOutBuf.String()

	return outStr, err //nolint:wrapcheck
}
