package logic

import (
	"container/list"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/rewardenv/reward/internal/shell"
	"github.com/rewardenv/reward/internal/templates"
	"github.com/rewardenv/reward/pkg/util"
)

// RunCmdSvc builds up the contents for the svc command.
func (c *Client) RunCmdSvc(args []string) error {
	log.Debugln("Running svc command...")

	if len(args) == 0 {
		args = append(args, "--help")

		// Don't catch stdout
		if err := c.RunCmdSvcDockerCompose(args); err != nil {
			return errors.Wrap(err, "running docker compose help command")
		}

		return nil
	}

	tplgen := templates.New()

	if util.ContainsString(args, "up") {
		serviceDomain := c.ServiceDomain()

		if !util.FileExists(filepath.Join(c.SSLDir(), "certs", serviceDomain+".crt.pem")) {
			if err := c.RunCmdSignCertificate([]string{serviceDomain}); err != nil {
				return errors.Wrap(err, "signing certificate")
			}
		}

		if err := tplgen.SvcGenerateTraefikConfig(); err != nil {
			return errors.Wrap(err, "generating traefik config")
		}

		if err := tplgen.SvcGenerateTraefikDynamicConfig(c.ServiceDomain()); err != nil {
			return errors.Wrap(err, "generating traefik dynamic config")
		}

		// Add --detach to the args (to run in background) if the user didn't specify it.
		newArgs := args

		if !util.ContainsString(args, "-d", "--detach") {
			for i, arg := range args {
				if arg == "up" {
					newArgs = []string{}
					newArgs = append(newArgs, args[:i+1]...)
					newArgs = append(newArgs, "--detach")
					newArgs = append(newArgs, args[i+1:]...)
				}
			}
		}

		args = newArgs
	}

	if util.ContainsString(args, "restart") {
		serviceDomain := c.ServiceDomain()

		if !util.FileExists(filepath.Join(c.SSLDir(), "certs", serviceDomain+".crt.pem")) {
			if err := c.RunCmdSignCertificate([]string{serviceDomain}); err != nil {
				return errors.Wrapf(err,
					"signing certificate for service domain %s",
					serviceDomain,
				)
			}
		}

		if err := tplgen.SvcGenerateTraefikConfig(); err != nil {
			return errors.Wrap(err, "generating traefik config")
		}

		if err := tplgen.SvcGenerateTraefikDynamicConfig(c.ServiceDomain()); err != nil {
			return errors.Wrap(err, "generating traefik dynamic config")
		}
	}

	// Pass orchestration through to docker compose
	// Don't catch stdout
	if err := c.RunCmdSvcDockerCompose(args); err != nil {
		return err
	}

	// connect peered service containers to environment networks when 'svc up' is run
	networks, err := c.Docker.NetworkNamesByLabel(fmt.Sprintf("dev.%s.environment.name", c.AppName()))
	if err != nil {
		return errors.Wrap(err, "getting environment networks")
	}

	for _, network := range networks {
		if err := c.DockerPeeredServices("connect", network); err != nil {
			return errors.Wrap(err, "connecting peered services")
		}
	}

	log.Debugln("...finished running svc command.")

	return nil
}

// RunCmdSvcDockerCompose function is a wrapper around the docker compose command.
// It appends the current directory and current project name to the args.
// It also changes the output if the OS StdOut is suppressed.
func (c *Client) RunCmdSvcDockerCompose(args []string, opts ...shell.Opt) error {
	passedArgs := []string{
		"--project-directory",
		c.AppHomeDir(),
		"--project-name",
		c.AppName(),
	}
	passedArgs = append(passedArgs, args...)

	// run docker compose command
	out, err := c.RunCmdSvcBuildDockerComposeCommand(passedArgs, opts...)
	out = regexp.MustCompile("(?m)[\r\n]+^.*--file.*$").ReplaceAllString(out, "")
	out = regexp.MustCompile("(?m)[\r\n]+^.*--project-name.*$").ReplaceAllString(out, "")
	out = regexp.MustCompile("(?m)[\r\n]+^.*--project-directory.*$").ReplaceAllString(out, "")
	out = regexp.MustCompile("(?m)[\r\n]+^.*default: docker-compose.yml.*$").ReplaceAllString(out, "")
	out = regexp.MustCompile("(?m)[\r\n]+^.*default: directory name.*$").ReplaceAllString(out, "")
	out = strings.ReplaceAll(out, "docker-compose", "env")

	log.Debugf("Command output: %s", out)

	if err != nil {
		return errors.Wrap(err, "running docker compose `svc` command")
	}

	return nil
}

// RunCmdSvcBuildDockerComposeCommand builds up the docker-compose command by passing it the previously built
// templates for the common services..
func (c *Client) RunCmdSvcBuildDockerComposeCommand(args []string, opts ...shell.Opt) (string, error) {
	var (
		tpl     = &template.Template{}
		tplList = list.New()
		tplgen  = templates.New()
	)

	if err := tplgen.RunCmdSvcBuildDockerComposeTemplate(tpl, tplList); err != nil {
		return "", err
	}

	svcDockerComposeConfigs, err := tplgen.ConvertTemplateToComposeConfig(tpl, tplList)
	if err != nil {
		return "", err
	}

	out, err := c.Compose.RunWithConfig(args, svcDockerComposeConfigs, opts...)
	if err != nil {
		return out, err
	}

	return out, nil
}
