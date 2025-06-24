package logic

import (
	"container/list"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"github.com/rewardenv/reward/internal/docker"
	"github.com/rewardenv/reward/internal/shell"
	"github.com/rewardenv/reward/internal/templates"
	"github.com/rewardenv/reward/pkg/util"
)

// RunCmdEnv build up the contents for the env command.
func (c *Client) RunCmdEnv(args []string) error {
	// Run docker compose help command if no args are passed.
	if len(args) == 0 {
		//nolint:gocritic
		passedArgs := append(args, "--help")

		// Don't catch stdout
		if err := c.RunCmdEnvDockerCompose(passedArgs); err != nil {
			return err
		}

		return nil
	}

	// down: disconnect peered service containers from environment network
	if err := c.configureCmdDown(args); err != nil {
		return errors.Wrap(err, "configuring the `down` command")
	}

	// up: connect peered service containers to environment network
	args, err := c.configureCmdUp(args)
	if err != nil {
		return errors.Wrap(err, "configuring the `up` command")
	}

	if err := c.configureCmdCommon(args); err != nil {
		return errors.Wrap(err, "configuring the command")
	}

	if err := c.CheckAndCreateLocalAppDirs(); err != nil {
		return errors.Wrap(err, "creating local app directories")
	}

	// Pass orchestration through to docker compose
	// Don't catch stdout
	if err := c.RunCmdEnvDockerCompose(args); err != nil {
		return err
	}

	if err := c.updateMutagen(args); err != nil {
		return errors.Wrap(err, "updating mutagen")
	}

	return nil
}

// RunCmdEnvDockerCompose function is a wrapper around the docker compose command.
// It appends the current directory and current project name to the args.
// It also changes the output if the OS StdOut is suppressed.
func (c *Client) RunCmdEnvDockerCompose(args []string, opts ...shell.Opt) error {
	passedArgs := []string{
		"--project-directory",
		c.Cwd(),
		"--project-name",
		c.EnvName(),
	}
	passedArgs = append(passedArgs, args...)

	// run docker compose command
	out, err := c.RunCmdEnvBuildDockerCompose(passedArgs, opts...)
	out = regexp.MustCompile("(?m)[\r\n]+^.*--file.*$").ReplaceAllString(out, "")
	out = regexp.MustCompile("(?m)[\r\n]+^.*--project-name.*$").ReplaceAllString(out, "")
	out = regexp.MustCompile("(?m)[\r\n]+^.*--project-directory.*$").ReplaceAllString(out, "")
	out = regexp.MustCompile("(?m)[\r\n]+^.*default: docker-compose.yml.*$").ReplaceAllString(out, "")
	out = regexp.MustCompile("(?m)[\r\n]+^.*default: directory name.*$").ReplaceAllString(out, "")
	out = strings.ReplaceAll(out, "docker-compose", "env")
	_, _ = fmt.Fprint(os.Stdout, out)

	if err != nil {
		return err
	}

	return nil
}

// RunCmdEnvBuildDockerComposeTemplate builds the templates which are used to invoke docker compose.
func (c *Client) RunCmdEnvBuildDockerComposeTemplate(tpl *template.Template, templateList *list.List) error {
	envType := c.EnvType()

	switch {
	// pwa-studio: everything is disabled, except node container
	case util.CheckRegexInString("^pwa-studio", envType):
		c.SetPWADefaults()
	// magento 1,2, shopware, wordpress have their own php-fpm containers
	case util.CheckRegexInString("^magento|wordpress|shopware", envType):
		c.SetPHPDefaults(envType)

		fallthrough
	default:
		if util.CheckRegexInString("^local", envType) {
			// local: empty env, varnish, elasticsearch/opensearch, rabbitmq can be enabled
			c.SetLocalDefaults()
		} else {
			// not local: only nginx, db and redis are enabled, php-fpm is running locally
			c.SetNonLocalDefaults()
		}
	}

	c.SetSyncSettings()

	// windows
	//nolint:goconst
	if runtime.GOOS == "windows" {
		c.SetDefault("xdebug_connect_back_host", "host.docker.internal")
	}

	// For linux, if UID is 1000, there is no need to use the socat proxy.
	if runtime.GOOS == "linux" && os.Geteuid() == 1000 {
		c.SetDefault("ssh_auth_sock_path_env", "/run/host-services/ssh-auth.sock")
	}

	if err := templates.New().AppendEnvironmentTemplates(tpl, templateList, "networks", envType); err != nil {
		return errors.Wrap(err, "appending network templates")
	}

	svcs := []string{
		"php-fpm",
		"nginx",
		"db",
		"elasticsearch",
		"opensearch",
		"varnish",
		"rabbitmq",
		"redis",
		"node",
		"mercure",
		"valkey",
	}
	for _, svc := range svcs {
		if c.GetBool(fmt.Sprintf("%s_%s", c.AppName(), strings.ReplaceAll(svc, "-", "_"))) {
			if err := templates.New().AppendEnvironmentTemplates(tpl, templateList, svc, envType); err != nil {
				return errors.Wrapf(err, "appending %s service templates", svc)
			}
		}
	}

	if err := templates.New().AppendEnvironmentTemplates(tpl, templateList, envType, envType); err != nil {
		return errors.Wrapf(err, "appending %s environment templates", envType)
	}

	additionalMagentoSvcs := map[string]string{
		fmt.Sprintf("%s_test_db", c.AppName()):        fmt.Sprintf("%s.tests", envType),
		fmt.Sprintf("%s_split_sales", c.AppName()):    fmt.Sprintf("%s.splitdb.sales", envType),
		fmt.Sprintf("%s_split_checkout", c.AppName()): fmt.Sprintf("%s.splitdb.checkout", envType),
	}

	for k, v := range additionalMagentoSvcs {
		if c.GetBool(k) {
			if err := templates.New().AppendEnvironmentTemplates(tpl, templateList, v, envType); err != nil {
				return errors.Wrapf(err, "appending %s additional magento templates", v)
			}
		}
	}

	externalSVCs := map[string][]string{
		"blackfire": {"blackfire", fmt.Sprintf("%s.blackfire", envType)},
		"allure":    {"allure"},
		"selenium":  {"selenium"},
		"magepack":  {fmt.Sprintf("%s.magepack", envType)},
	}
	for name, svcs := range externalSVCs {
		if c.GetBool(fmt.Sprintf("%s_%s", c.AppName(), name)) {
			for _, svc := range svcs {
				if err := templates.New().AppendEnvironmentTemplates(tpl, templateList, svc, envType); err != nil {
					return errors.Wrapf(err, "appending %s external service templates", svc)
				}
			}
		}
	}

	// ./.reward/reward-env.yml
	// ./.reward/reward-env.os.yml
	additionalTemplates := []string{
		fmt.Sprintf("%s-env.yml", c.AppName()),
		fmt.Sprintf("%[1]v-env.%[2]v.yml", c.AppName(), runtime.GOOS),
	}

	if err := templates.New().AppendTemplatesFromPaths(tpl, templateList, additionalTemplates); err != nil {
		return errors.Wrap(err, "appending templates from current directory")
	}

	c.SetSeleniumDefaults()

	return nil
}

// RunCmdEnvBuildDockerCompose builds up the docker compose command by passing it the previously built templates.
func (c *Client) RunCmdEnvBuildDockerCompose(args []string, opts ...shell.Opt) (string, error) {
	var (
		envTemplate     = new(template.Template)
		envTemplateList = list.New()
	)

	if err := c.RunCmdEnvBuildDockerComposeTemplate(envTemplate, envTemplateList); err != nil {
		return "", err
	}

	dockerComposeConfigs, err := templates.New().ConvertTemplateToComposeConfig(envTemplate, envTemplateList)
	if err != nil {
		return "", err
	}

	out, err := c.Compose.RunWithConfig(args, dockerComposeConfigs, opts...)
	if err != nil {
		return out, err
	}

	return out, nil
}

func (c *Client) configureCmdDown(args []string) error {
	if util.ContainsString(args, "down") {
		if err := c.DockerPeeredServices("disconnect", c.EnvNetworkName()); err != nil {
			return errors.Wrap(err, "disconnecting peered services")
		}
	}

	return nil
}

func (c *Client) configureCmdUp(args []string) ([]string, error) {
	if util.ContainsString(args, "up") {
		// check if network already exist
		networkExist, err := c.Docker.NetworkExist(c.EnvNetworkName())
		if err != nil {
			return nil, err
		}

		if !networkExist {
			var passedArgs []string

			//nolint:gocritic
			if util.ContainsString(args, "--") {
				passedArgs = util.InsertStringBeforeOccurrence(args, "--no-start", "--")
			} else {
				passedArgs = append(args, "--no-start")
			}

			// Don't catch stdout
			if err := c.RunCmdEnvDockerCompose(passedArgs); err != nil {
				return nil, errors.Wrap(err, "running `docker compose --no-start` to create network")
			}
		}

		if err := c.DockerPeeredServices("connect", c.EnvNetworkName()); err != nil {
			return nil, errors.Wrap(err, "connecting peered services to docker network")
		}

		if !util.ContainsString(args, "-d", "--detach") {
			args = util.InsertStringAfterOccurrence(args, "--detach", "up")
		}
	}

	return args, nil
}

func (c *Client) configureCmdCommon(args []string) error {
	// If the command is 'env config' then skip traefik address lookup, mutagen settings, etc.
	if util.ContainsString([]string{args[0]}, "config", "down") {
		return nil
	}

	// mutagen: sync file
	if err := c.RunCmdSyncCheck(); err != nil {
		return errors.Wrap(err, "checking mutagen sync")
	}

	// mutagen: pause sync if needed
	if util.ContainsString(args, "stop") {
		if err := c.RunCmdSyncPause(); err != nil {
			return errors.Wrap(err, "pausing mutagen sync")
		}
	}

	if !util.ContainsString([]string{args[0]}, "up", "start") {
		return nil
	}

	// traefik: lookup address of traefik container in the environment network
	traefikAddress, err := c.Docker.ContainerAddressInNetwork(
		"traefik", c.AppName(), c.EnvNetworkName(),
	)
	if err != nil {
		return docker.ErrCannotFindContainer("traefik", err)
	}

	c.Set("traefik_address", traefikAddress)

	return nil
}

func (c *Client) updateMutagen(args []string) error {
	if !c.SyncEnabled() {
		return nil
	}

	if util.ContainsString(args, "up", "start") {
		if util.ContainsString(args, "--") {
			return nil
		}

		// mutagen: start mutagen sync if container id changed (or previously didn't exist)
		if c.ContainerChanged(c.SyncedContainer()) {
			if err := c.RunCmdSyncStart(); err != nil {
				return errors.Wrap(err, "starting mutagen sync")
			}

			return nil
		}

		// mutagen: resume mutagen sync if php-fpm container id hasn't changed
		if err := c.RunCmdSyncResume(); err != nil {
			return errors.Wrap(err, "resuming mutagen sync")
		}

		return nil
	}

	// mutagen: stop mutagen sync if needed
	if util.ContainsString(args, "down") {
		if err := c.RunCmdSyncStop(); err != nil {
			return errors.Wrap(err, "stopping mutagen sync")
		}
	}

	return nil
}
