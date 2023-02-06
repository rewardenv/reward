package logic

import (
	"container/list"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"text/template"

	"github.com/rewardenv/reward/internal/docker"
	"github.com/rewardenv/reward/internal/shell"
	"github.com/rewardenv/reward/internal/templates"
	"github.com/rewardenv/reward/pkg/util"
)

// RunCmdEnv build up the contents for the env command.
func (c *Client) RunCmdEnv(args []string) error {
	// Run docker-compose help command if no args are passed.
	if len(args) == 0 {
		//nolint:gocritic
		passedArgs := append(args, "--help")

		err := c.RunCmdEnvDockerCompose(passedArgs, shell.WithCatchOutput(false))
		if err != nil {
			return err
		}

		return nil
	}

	// down: disconnect peered service containers from environment network
	err := c.configureCmdDown(args)
	if err != nil {
		return fmt.Errorf("an error occurred while configuring the `down` command: %w", err)
	}

	// up: connect peered service containers to environment network
	args, err = c.configureCmdUp(args)
	if err != nil {
		return fmt.Errorf("an error occurred while configuring the `up` command: %w", err)
	}

	err = c.configureCmdCommon(args)
	if err != nil {
		return fmt.Errorf("an error occurred while configuring the command: %w", err)
	}

	err = c.CheckAndCreateLocalAppDirs()
	if err != nil {
		return fmt.Errorf("cannot create local app directories: %w", err)
	}

	// pass orchestration through to docker-compose
	err = c.RunCmdEnvDockerCompose(args, shell.WithCatchOutput(false))
	if err != nil {
		return err
	}

	err = c.updateMutagen(args)
	if err != nil {
		return fmt.Errorf("an error occurred while updating mutagen: %w", err)
	}

	return nil
}

// RunCmdEnvDockerCompose function is a wrapper around the docker-compose command.
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

	// run docker-compose command
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

// RunCmdEnvBuildDockerComposeTemplate builds the templates which are used to invoke docker-compose.
func (c *Client) RunCmdEnvBuildDockerComposeTemplate(tpl *template.Template, templateList *list.List) error {
	envType := c.EnvType()

	// magento 1,2, shopware, wordpress have their own php-fpm containers
	if util.CheckRegexInString(`^magento|wordpress|shopware`, envType) {
		c.SetPHPDefaults(envType)
	}

	c.SetSyncSettings()

	// pwa-studio: everything is disabled, except node container
	if util.CheckRegexInString("^pwa-studio", envType) {
		c.SetPWADefaults()
	}

	// not local: only nginx, db and redis are enabled, php-fpm is running locally
	if !util.CheckRegexInString(`^local`, envType) {
		c.SetNonLocalDefaults()
	}

	// local: empty env, varnish, elasticsearch/opensearch, rabbitmq can be enabled
	if util.CheckRegexInString("^local", envType) {
		c.SetLocalDefaults()
	}

	// windows
	//nolint:goconst
	if runtime.GOOS == "windows" {
		c.SetDefault("xdebug_connect_back_host", "host.docker.internal")
	}

	// For linux, if UID is 1000, there is no need to use the socat proxy.
	if runtime.GOOS == "linux" && os.Geteuid() == 1000 {
		c.SetDefault("ssh_auth_sock_path_env", "/run/host-services/ssh-auth.sock")
	}

	err := templates.New().AppendEnvironmentTemplates(tpl, templateList, "networks", envType)
	if err != nil {
		return fmt.Errorf("an error occurred while appending network templates: %w", err)
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
	}
	for _, svc := range svcs {
		if c.GetBool(fmt.Sprintf("%s_%s", c.AppName(), strings.ReplaceAll(svc, "-", "_"))) {
			err = templates.New().AppendEnvironmentTemplates(tpl, templateList, svc, envType)
			if err != nil {
				return fmt.Errorf("an error occurred while appending %s service templates: %w",
					svc,
					err)
			}
		}
	}

	err = templates.New().AppendEnvironmentTemplates(tpl, templateList, envType, envType)
	if err != nil {
		return fmt.Errorf("an error occurred while appending %s environment templates: %w", envType, err)
	}

	additionalMagentoSvcs := map[string]string{
		fmt.Sprintf("%s_test_db", c.AppName()):        fmt.Sprintf("%s.tests", envType),
		fmt.Sprintf("%s_split_sales", c.AppName()):    fmt.Sprintf("%s.splitdb.sales", envType),
		fmt.Sprintf("%s_split_checkout", c.AppName()): fmt.Sprintf("%s.splitdb.checkout", envType),
	}

	for k, v := range additionalMagentoSvcs {
		if c.GetBool(k) {
			err = templates.New().AppendEnvironmentTemplates(tpl, templateList, v, envType)
			if err != nil {
				return fmt.Errorf("an error occurred while appending %s additional magento templates: %w",
					v,
					err)
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
				err = templates.New().AppendEnvironmentTemplates(tpl, templateList, svc, envType)
				if err != nil {
					return fmt.Errorf(
						"an error occurred while appending %s external service templates: %w",
						svc,
						err,
					)
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

	err = templates.New().AppendTemplatesFromPaths(tpl, templateList, additionalTemplates)
	if err != nil {
		return fmt.Errorf("an error occurred while appending templates from current directory: %w", err)
	}

	c.SetSeleniumDefaults()

	return nil
}

// RunCmdEnvBuildDockerCompose builds up the docker-compose command by passing it the previously built templates.
func (c *Client) RunCmdEnvBuildDockerCompose(args []string, opts ...shell.Opt) (string, error) {
	var (
		envTemplate     = new(template.Template)
		envTemplateList = list.New()
	)

	err := c.RunCmdEnvBuildDockerComposeTemplate(envTemplate, envTemplateList)
	if err != nil {
		return "", err
	}

	dockerComposeConfigs, err := templates.New().ConvertTemplateToComposeConfig(envTemplate, envTemplateList)
	if err != nil {
		return "", err
	}

	out, err := c.DockerCompose.RunWithConfig(args, dockerComposeConfigs, opts...)
	if err != nil {
		return out, err
	}

	return out, nil
}

func (c *Client) configureCmdDown(args []string) error {
	if util.ContainsString(args, "down") {
		err := c.DockerPeeredServices("disconnect", c.EnvNetworkName())
		if err != nil {
			return fmt.Errorf("an error occurred while disconnecting peered services: %w", err)
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

			err = c.RunCmdEnvDockerCompose(passedArgs, shell.WithCatchOutput(false))
			if err != nil {
				return nil, fmt.Errorf("an error occurred while running `docker compose --no-start` to create network: %w",
					err)
			}
		}

		err = c.DockerPeeredServices("connect", c.EnvNetworkName())
		if err != nil {
			return nil, fmt.Errorf("an error occurred while connecting peered services to docker network: %w",
				err)
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
	err := c.RunCmdSyncCheck()
	if err != nil {
		return fmt.Errorf("an error occurred while checking mutagen sync: %w", err)
	}

	// mutagen: pause sync if needed
	if util.ContainsString(args, "stop") {
		err := c.RunCmdSyncPause()
		if err != nil {
			return fmt.Errorf("an error occurred while pausing mutagen sync: %w", err)
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
			err := c.RunCmdSyncStart()
			if err != nil {
				return fmt.Errorf("an error occurred while starting mutagen sync: %w", err)
			}

			return nil
		}

		// mutagen: resume mutagen sync if php-fpm container id hasn't changed
		err := c.RunCmdSyncResume()
		if err != nil {
			return fmt.Errorf("an error occurred while resuming mutagen sync: %w", err)
		}

		return nil
	}

	// mutagen: stop mutagen sync if needed
	if util.ContainsString(args, "down") {
		err := c.RunCmdSyncStop()
		if err != nil {
			return fmt.Errorf("an error occurred while stopping mutagen sync: %w", err)
		}
	}

	return nil
}
