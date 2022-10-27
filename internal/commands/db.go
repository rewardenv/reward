package commands

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

	"github.com/rewardenv/reward/internal/core"

	compose "github.com/docker/cli/cli/compose/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// DBConnectCmd connects to the environment's database container.
func DBConnectCmd(cmd *cobra.Command, args []string) error {
	log.Debugln()

	runAsRootUser, err := cmd.Flags().GetBool("root")
	if err != nil {
		return err
	}

	var mysqlCommand, mysqlUserParam, mysqlPasswordParam, mysqlDBParam string

	command := []string{"sh", "-c"}
	mysqlCommand = core.DBCommand()

	if runAsRootUser {
		mysqlUserParam = "-uroot"
		mysqlPasswordParam = "-p$(printenv MYSQL_ROOT_PASSWORD)" //nolint:gosec
	} else {
		mysqlUserParam = "-u$(printenv MYSQL_USER)"
		mysqlPasswordParam = "-p$(printenv MYSQL_PASSWORD)" //nolint:gosec
	}

	mysqlDBParam = "--database=$(printenv MYSQL_DATABASE)"

	params := fmt.Sprintf("%v %v %v %v", mysqlCommand, mysqlUserParam, mysqlPasswordParam, mysqlDBParam)

	log.Debugln("command:", command)
	log.Debugln("container:", core.DBContainer())

	var passedArgs []string

	passedArgs = append(passedArgs, "exec")
	passedArgs = append(passedArgs, core.DBContainer())
	passedArgs = append(passedArgs, command...)
	params = params + " " + strings.Join(core.ExtractUnknownArgs(cmd.Flags(), args), " ")
	passedArgs = append(passedArgs, params)

	err = EnvRunDockerCompose(passedArgs, false)
	if err != nil {
		return err
	}

	return nil
}

// DBImportCmd imports a database from stdin to the environment's database container.
func DBImportCmd(cmd *cobra.Command, args []string) error {
	log.Debugln()

	runAsRootUser, err := cmd.Flags().GetBool("root")
	if err != nil {
		return err
	}

	var mysqlCommand, mysqlUserParam, mysqlPasswordParam, mysqlDBParam string

	command := []string{"sh", "-c"}
	mysqlCommand = core.DBCommand()

	if runAsRootUser {
		mysqlUserParam = "-uroot"
		mysqlPasswordParam = "-p$(printenv MYSQL_ROOT_PASSWORD)" //nolint:gosec
	} else {
		mysqlUserParam = "-u$(printenv MYSQL_USER)"
		mysqlPasswordParam = "-p$(printenv MYSQL_PASSWORD)" //nolint:gosec
	}

	mysqlDBParam = "--database=$(printenv MYSQL_DATABASE)"
	params := fmt.Sprintf("%v %v %v %v", mysqlCommand, mysqlUserParam, mysqlPasswordParam, mysqlDBParam)

	log.Debugln("command:", command)
	log.Debugln("container:", core.DBContainer())

	var passedArgs []string

	passedArgs = append(passedArgs, "exec")
	passedArgs = append(passedArgs, "-T")
	passedArgs = append(passedArgs, core.DBContainer())
	passedArgs = append(passedArgs, command...)
	// FIXME: ExtractUnknownArgs not working here
	params = params + " " + strings.Join(core.ExtractUnknownArgs(cmd.Flags(), args), " ")
	passedArgs = append(passedArgs, params)

	err = DBRunDockerCompose(passedArgs, false)
	if err != nil {
		return err
	}

	return nil
}

// DBDumpCmd dumps the database from the environment's database container.
func DBDumpCmd(cmd *cobra.Command, args []string) error {
	log.Debugln()

	runAsRootUser, err := cmd.Flags().GetBool("root")
	if err != nil {
		return err
	}

	var mysqlDumpCommand, mysqlUserParam, mysqlPasswordParam, mysqlDBParam string

	command := []string{"sh", "-c"}
	mysqlDumpCommand = core.DBDumpCommand()

	if runAsRootUser {
		mysqlUserParam = "-uroot"
		mysqlPasswordParam = "-p$(printenv MYSQL_ROOT_PASSWORD)" //nolint:gosec
	} else {
		mysqlUserParam = "-u$(printenv MYSQL_USER)"
		mysqlPasswordParam = "-p$(printenv MYSQL_PASSWORD)" //nolint:gosec
	}

	mysqlDBParam = "$(printenv MYSQL_DATABASE)"
	params := fmt.Sprintf("%v %v %v %v", mysqlDumpCommand, mysqlUserParam, mysqlPasswordParam, mysqlDBParam)

	log.Debugln("command:", command)
	log.Debugln("container:", core.DBContainer())

	var passedArgs []string

	passedArgs = append(passedArgs, "exec")
	passedArgs = append(passedArgs, "-T")
	passedArgs = append(passedArgs, core.DBContainer())
	passedArgs = append(passedArgs, command...)
	// FIXME: ExtractUnknownArgs not working here
	params = params + " " + strings.Join(core.ExtractUnknownArgs(cmd.Flags(), args), " ")
	passedArgs = append(passedArgs, params)

	err = DBRunDockerCompose(passedArgs, false)
	if err != nil {
		return err
	}

	return nil
}

// DBRunDockerCompose function is a wrapper around the docker-compose command.
//
//	It appends the current directory and current project name to the args.
//	It also changes the output if the OS StdOut is suppressed.
func DBRunDockerCompose(args []string, suppressOsStdOut ...bool) error {
	log.Debugln()

	passedArgs := []string{
		"--project-directory",
		core.Cwd(),
		"--project-name",
		core.EnvName(),
	}
	passedArgs = append(passedArgs, args...)

	// run docker-compose command
	out, err := DBBuildDockerComposeCommand(passedArgs, suppressOsStdOut...)
	re := regexp.MustCompile("(?m)[\r\n]+^.*--file.*$")
	out = re.ReplaceAllString(out, "")
	re = regexp.MustCompile("(?m)[\r\n]+^.*--project-name.*$")
	out = re.ReplaceAllString(out, "")
	re = regexp.MustCompile("(?m)[\r\n]+^.*--project-directory.*$")
	out = re.ReplaceAllString(out, "")
	re = regexp.MustCompile("(?m)[\r\n]+^.*default: docker-compose.yml.*$")
	out = re.ReplaceAllString(out, "")
	re = regexp.MustCompile("(?m)[\r\n]+^.*default: directory name.*$")
	out = re.ReplaceAllString(out, "")
	out = strings.ReplaceAll(out, "docker-compose", "env")

	_, _ = fmt.Fprint(os.Stdout, out)

	if err != nil {
		return err
	}

	return nil
}

// DBBuildDockerComposeCommand builds up the docker-compose command's templates.
func DBBuildDockerComposeCommand(args []string, suppressOsStdOut ...bool) (string, error) {
	log.Debugln()

	dbTemplate := new(template.Template)
	dbTemplateList := list.New()

	err := EnvBuildDockerComposeTemplate(dbTemplate, dbTemplateList)
	if err != nil {
		return "", err
	}

	dockerComposeConfigs, err := core.ConvertTemplateToComposeConfig(dbTemplate, dbTemplateList)
	if err != nil {
		return "", err
	}

	out, err := DBRunDockerComposeWithConfig(args, dockerComposeConfigs, suppressOsStdOut...)
	if err != nil {
		return out, err
	}

	return out, nil
}

// DBRunDockerComposeWithConfig calls docker-compose with the previously built docker-compose configuration.
func DBRunDockerComposeWithConfig(
	args []string, details compose.ConfigDetails, suppressOsStdOut ...bool,
) (string, error) {
	log.Debugln("Reading configs...")

	tmpFiles := make([]string, 0, len(details.ConfigFiles))

	for _, conf := range details.ConfigFiles {
		bs, err := yaml.Marshal(conf.Config)
		if err != nil {
			return "", err
		}

		log.Traceln("Reading config:")
		log.Traceln(conf.Filename)
		log.Traceln(string(bs))

		tmpFile, err := os.CreateTemp(os.TempDir(), core.AppName+"-")
		if err != nil {
			return "", err
		}

		core.TmpFilesList.PushBack(tmpFile.Name())
		// defer os.Remove(tmpFile.Name())

		tmpFiles = append(tmpFiles, tmpFile.Name())

		if _, err = tmpFile.Write(bs); err != nil {
			return "", err
		}

		if err := tmpFile.Close(); err != nil {
			return "", err
		}
	}

	composeArgs := make([]string, 0, len(tmpFiles))
	for _, file := range tmpFiles {
		composeArgs = append(composeArgs, "-f")
		composeArgs = append(composeArgs, file)
	}

	composeArgs = append(composeArgs, args...)

	out, err := DBRunDockerComposeCommandModifyStdin(composeArgs, suppressOsStdOut...)
	if err != nil {
		return out, err
	}

	return out, nil
}

// DBRunDockerComposeCommandModifyStdin runs the passed parameters with docker-compose and returns the output.
func DBRunDockerComposeCommandModifyStdin(args []string, suppressOsStdOut ...bool) (string, error) {
	log.Tracef("args: %#v", args)
	log.Debugf("Running command: docker-compose %v", strings.Join(args, " "))

	cmd := exec.Command("docker-compose", args...)

	var combinedOutBuf bytes.Buffer

	r, w := io.Pipe()
	definerRegex := regexp.MustCompile("DEFINER[ ]*=[ ]*`[^`]+`@`[^`]+`")
	globalRegex := regexp.MustCompile(`@@(GLOBAL\.GTID_PURGED|SESSION\.SQL_LOG_BIN)`)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)

		maxCapacity := viper.GetInt("db_import_line_buffer_size") * 1024 * 1024 // max capacity for buffer is 10MB/line
		buf := make([]byte, 0, 1024*1024)
		scanner.Buffer(buf, maxCapacity)

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

		if err := scanner.Err(); err != nil {
			log.Debugln("error:", err)
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
