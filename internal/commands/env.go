package commands

import (
	"container/list"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"

	"github.com/rewardenv/reward/internal/core"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	envTypes = map[string]string{
		"generic-php": fmt.Sprintf(
			`%[1]v_DB=1
%[1]v_REDIS=1

MARIADB_VERSION=10.4
NODE_VERSION=10
PHP_VERSION=7.4
REDIS_VERSION=6.0
COMPOSER_VERSION=2

MYSQL_ROOT_PASSWORD=app
MYSQL_DATABASE=app
MYSQL_USER=app
MYSQL_PASSWORD=app

NGINX_ROOT=/var/www/html
NGINX_PUBLIC=
`, strings.ToUpper(core.AppName),
		),

		"magento1": fmt.Sprintf(
			`%[1]v_DB=1
%[1]v_REDIS=1

MARIADB_VERSION=10.3
NODE_VERSION=10
PHP_VERSION=7.2
REDIS_VERSION=5.0
COMPOSER_VERSION=1

%[1]v_SELENIUM=0
%[1]v_SELENIUM_DEBUG=0
%[1]v_BLACKFIRE=0

BLACKFIRE_CLIENT_ID=
BLACKFIRE_CLIENT_TOKEN=
BLACKFIRE_SERVER_ID=
BLACKFIRE_SERVER_TOKEN=
`, strings.ToUpper(core.AppName),
		),

		"magento2": fmt.Sprintf(
			`%[1]v_DB=1
%[1]v_ELASTICSEARCH=0
%[1]v_OPENSEARCH=1
%[1]v_OPENSEARCH_DASHBOARDS=0
%[1]v_VARNISH=1
%[1]v_RABBITMQ=1
%[1]v_REDIS=1
%[1]v_MERCURE=0

ELASTICSEARCH_VERSION=7.16
OPENSEARCH_VERSION=1.2
MARIADB_VERSION=10.4
NODE_VERSION=10
PHP_VERSION=7.4
RABBITMQ_VERSION=3.8
REDIS_VERSION=6.0
VARNISH_VERSION=6.5
COMPOSER_VERSION=2

%[1]v_SYNC_IGNORE=

%[1]v_ALLURE=0
%[1]v_SELENIUM=0
%[1]v_SELENIUM_DEBUG=0
%[1]v_BLACKFIRE=0
%[1]v_SPLIT_SALES=0
%[1]v_SPLIT_CHECKOUT=0
%[1]v_TEST_DB=0
%[1]v_MAGEPACK=0

BLACKFIRE_CLIENT_ID=
BLACKFIRE_CLIENT_TOKEN=
BLACKFIRE_SERVER_ID=
BLACKFIRE_SERVER_TOKEN=

XDEBUG_VERSION=
`, strings.ToUpper(core.AppName),
		),

		"laravel": fmt.Sprintf(
			`MARIADB_VERSION=10.4
NODE_VERSION=10
PHP_VERSION=7.4
REDIS_VERSION=6.0
COMPOSER_VERSION=2

%[1]v_DB=1
%[1]v_REDIS=1
%[1]v_MERCURE=0

## Laravel Config
APP_URL=https://${%[1]v_ENV_NAME}.test
APP_KEY=

APP_ENV=local
APP_DEBUG=true

DB_CONNECTION=mysql
DB_HOST=db
DB_PORT=3306
DB_DATABASE=laravel
DB_USERNAME=laravel
DB_PASSWORD=laravel

CACHE_DRIVER=redis
SESSION_DRIVER=redis

REDIS_HOST=redis
REDIS_PORT=6379

MAIL_DRIVER=sendmail
`, strings.ToUpper(core.AppName),
		),

		"pwa-studio": fmt.Sprintf(
			`NODE_VERSION=10
%[1]v_VARNISH=0
VARNISH_VERSION=6.5

`, strings.ToUpper(core.AppName),
		),

		"symfony": fmt.Sprintf(
			`%[1]v_DB=1
%[1]v_REDIS=1
%[1]v_RABBITMQ=0
%[1]v_ELASTICSEARCH=0
%[1]v_OPENSEARCH=0
%[1]v_OPENSEARCH_DASHBOARDS=0
%[1]v_VARNISH=0
%[1]v_MERCURE=0

MARIADB_VERSION=10.4
NODE_VERSION=10
PHP_VERSION=7.4
RABBITMQ_VERSION=3.8
REDIS_VERSION=6.0
VARNISH_VERSION=6.5
COMPOSER_VERSION=2
`, strings.ToUpper(core.AppName),
		),

		"shopware": fmt.Sprintf(
			`%[1]v_DB=1
%[1]v_REDIS=1
%[1]v_RABBITMQ=0
%[1]v_ELASTICSEARCH=0
%[1]v_OPENSEARCH=0
%[1]v_VARNISH=0

MARIADB_VERSION=10.4
NODE_VERSION=10
PHP_VERSION=7.4
RABBITMQ_VERSION=3.8
REDIS_VERSION=6.0
VARNISH_VERSION=6.5
COMPOSER_VERSION=2
`, strings.ToUpper(core.AppName),
		),

		"wordpress": fmt.Sprintf(
			`MARIADB_VERSION=10.4
NODE_VERSION=10
PHP_VERSION=7.4
COMPOSER_VERSION=2

%[1]v_DB=1
%[1]v_REDIS=0

APP_ENV=local
APP_DEBUG=true

DB_CONNECTION=mysql
DB_HOST=db
DB_PORT=3306
DB_DATABASE=wordpress
DB_USERNAME=wordpress
DB_PASSWORD=wordpress
`, strings.ToUpper(core.AppName),
		),
	}

	validEnvTypes []string

	syncedContainer = "php-fpm"
)

// GetSyncedContainer returns the name of the container which is used for syncing.
func GetSyncedContainer() string {
	return syncedContainer
}

// SetSyncedContainer sets the synced container.
func SetSyncedContainer(s string) {
	syncedContainer = s
}

// GetValidEnvTypes return a list of valid environment types based on the predefined EnvTypes.
func GetValidEnvTypes() []string {
	validEnvTypes = make([]string, 0, len(envTypes))
	for key := range envTypes {
		validEnvTypes = append(validEnvTypes, key)
	}

	return validEnvTypes
}

// EnvCmd build up the contents for the env command.
func EnvCmd(args []string) error {
	if len(args) == 0 {
		args = append(args, "--help")

		err := EnvRunDockerCompose(args, true)
		if err != nil {
			return err
		}

		return nil
	}

	// down: disconnect peered service containers from environment network
	if core.ContainsString(args, "down") {
		err := core.DockerPeeredServices("disconnect", core.GetEnvNetworkName())
		if err != nil {
			return err
		}
	}

	// up: connect peered service containers to environment network
	if core.ContainsString(args, "up") {
		// check if network already exist
		networkExist, err := core.CheckDockerNetworkExist(core.GetEnvNetworkName())
		if err != nil {
			return err
		}

		if !networkExist {
			log.Println("Creating network...")

			var passedArgs []string

			if core.ContainsString(args, "--") {
				passedArgs = core.InsertStringBeforeOccurrence(args, "--no-start", "--")
			} else {
				passedArgs = append(args, "--no-start")
			}

			log.Tracef("args: %#v, updated args: %#v", args, passedArgs)

			err = EnvRunDockerCompose(passedArgs)
			if err != nil {
				return err
			}
		}

		err = core.DockerPeeredServices("connect", core.GetEnvNetworkName())
		if err != nil {
			return err
		}

		if !core.ContainsString(args, "-d") && !core.ContainsString(args, "--detach") {
			args = core.InsertStringAfterOccurrence(args, "--detach", "up")
		}
	}

	// If the command is 'env config' then skip traefik address lookup, mutagen settings, etc.
	if !core.ContainsString([]string{args[0]}, "config") {
		// traefik: lookup address of traefik container in the environment network
		traefikAddress, err := core.LookupContainerAddressInNetwork("traefik", core.AppName, core.GetEnvNetworkName())
		if err != nil {
			return core.CannotFindContainerError("traefik")
		}

		viper.Set("traefik_address", traefikAddress)

		log.Tracef("Traefik container address in network %v: %v", core.GetEnvNetworkName(), traefikAddress)

		// mutagen: sync file
		if core.IsMutagenSyncEnabled() {
			err = core.GenerateMutagenTemplateFileIfNotExist()
			if err != nil {
				return err
			}
		}

		// mutagen: pause sync if needed
		if core.ContainsString(args, "stop") {
			if core.IsMutagenSyncEnabled() {
				err := SyncPauseCmd()
				if err != nil {
					return err
				}
			}
		}
	}

	if err := CheckAndCreateLocalAppDirs(); err != nil {
		return err
	}

	// pass orchestration through to docker-compose
	if err := EnvRunDockerCompose(args, false); err != nil {
		return err
	}

	// mutagen: resume mutagen sync if available and php-fpm container id hasn't changed
	if core.ContainsString(args, "up") || core.ContainsString(args, "start") {
		if core.IsMutagenSyncEnabled() && !IsContainerChanged(GetSyncedContainer()) && !core.ContainsString(
			args, "--",
		) {
			err := CheckAndInstallMutagen()
			if err != nil {
				return err
			}

			err = SyncResumeCmd()
			if err != nil {
				return err
			}
		}
	}

	// mutagen: start mutagen sync if needed (container id changed or previously didn't exist
	if core.ContainsString(args, "up") || core.ContainsString(args, "start") {
		if core.IsMutagenSyncEnabled() && IsContainerChanged(GetSyncedContainer()) && !core.ContainsString(args, "--") {
			err := CheckAndInstallMutagen()
			if err != nil {
				return err
			}

			err = SyncStartCmd()
			if err != nil {
				return err
			}
		}
	}

	// mutagen: stop mutagen sync if needed
	if core.ContainsString(args, "down") {
		if core.IsMutagenSyncEnabled() {
			err := CheckAndInstallMutagen()
			if err != nil {
				return err
			}

			err = SyncStopCmd()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// EnvCheck returns an error if the env name is empty (.env file does not contain an env name).
func EnvCheck() error {
	log.Debugln()
	if len(strings.TrimSpace(core.GetEnvName())) == 0 {
		return core.ErrEnvIsEmpty
	}

	return nil
}

func validateEnvName(name string) bool {
	validatorRegex := `^[A-Za-z0-9](?:[A-Za-z0-9\-]{0,61}[A-Za-z0-9])?$`
	if !core.CheckRegexInString(validatorRegex, name) {
		log.Debugln("Environment name validator regex is not matching.")

		return false
	}

	log.Debugln("Environment name validator regex matches.")

	return true
}

// EnvInitCmd creates a .env file for envType based on envName.
func EnvInitCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 && len(strings.TrimSpace(core.GetEnvName())) == 0 {
		log.Println("Please provide an environment name.")

		_ = cmd.Help()

		os.Exit(1)
	}

	if len(args) > 0 {
		viper.Set(core.AppName+"_env_name", args[0])

		log.Debugf("args(%v): %v", len(args), args)

		if len(args) > 1 {
			if core.ContainsString(GetValidEnvTypes(), args[1]) {
				viper.Set(core.AppName+"_env_type", args[1])
			} else {
				return core.ErrUnknownEnvType
			}
		}
	}

	path := core.GetCwd()
	envType := core.GetEnvType()
	envName := core.GetEnvName()

	if !validateEnvName(envName) {
		return core.ErrEnvNameIsInvalid
	}

	if !core.ContainsString(GetValidEnvTypes(), envType) {
		return core.ErrUnknownEnvType
	}

	log.Debugln("name:", envName)
	log.Debugln("type:", envType)

	envFilePath := filepath.Join(path, ".env")

	envFileExist := core.CheckFileExistsAndRecreate(envFilePath)

	webRoot := "/"
	if envType == "shopware" {
		webRoot = "/webroot"
	}

	envBase := fmt.Sprintf(
		`%[1]v_ENV_NAME=%[2]v
%[1]v_ENV_TYPE=%[3]v
%[1]v_WEB_ROOT=%[4]v

TRAEFIK_DOMAIN=%[2]v.test
TRAEFIK_SUBDOMAIN=
TRAEFIK_EXTRA_HOSTS=

`, strings.ToUpper(core.AppName), envName, envType, webRoot,
	)
	envFileContent := strings.Join([]string{envBase, envTypes[envType]}, "")

	if !envFileExist {
		err := core.CreateDirAndWriteBytesToFile([]byte(envFileContent), envFilePath)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	if err := CheckAndCreateLocalAppDirs(); err != nil {
		return err
	}

	return nil
}

func CheckAndCreateLocalAppDirs() error {
	log.Debugln()
	path := core.GetCwd()
	localAppDir := filepath.Join(path, "."+core.AppName)

	if _, err := core.AFS.Stat(localAppDir); !os.IsNotExist(err) {
		return nil
	}

	if err := core.CreateDir(localAppDir); err != nil {
		return err
	}

	if core.SvcEnabledPermissive("nginx") {
		if err := core.CreateDir(filepath.Join(localAppDir, "nginx")); err != nil {
			return err
		}
	}

	if core.SvcEnabledStrict("varnish") {
		if err := core.CreateDir(filepath.Join(localAppDir, "varnish")); err != nil {
			return err
		}
	}

	return nil
}

// EnvRunDockerCompose function is a wrapper around the docker-compose command.
//   It appends the current directory and current project name to the args.
//   It also changes the output if the OS StdOut is suppressed.
func EnvRunDockerCompose(args []string, suppressOsStdOut ...bool) error {
	passedArgs := []string{
		"--project-directory",
		core.GetCwd(),
		"--project-name",
		core.GetEnvName(),
	}
	passedArgs = append(passedArgs, args...)

	// run docker-compose command
	out, err := EnvBuildDockerComposeCommand(passedArgs, suppressOsStdOut...)
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

// EnvBuildDockerComposeTemplate builds the templates which are used to invoke docker-compose.
func EnvBuildDockerComposeTemplate(t *template.Template, templateList *list.List) error {
	envType := core.GetEnvType()

	log.Debugln("ENV_TYPE:", envType)

	// magento1,2, shopware wordpress have their own php-fpm containers
	if core.CheckRegexInString(`^magento|wordpress|shopware`, envType) {
		log.Debugln("Setting SVC_PHP_VARIANT.")

		if !core.IsSingleWebContainer() {
			viper.Set(core.AppName+"_svc_php_variant", "-"+envType)
			viper.Set(core.AppName+"_svc_php_debug_variant", "-"+envType)
		} else {
			viper.Set(core.AppName+"_svc_php_variant", "-"+envType+"-web")
			viper.Set(core.AppName+"_svc_php_debug_variant", "-"+envType+"")
		}
	}

	log.Debugln("SVC_PHP_VARIANT:", viper.GetString(core.AppName+"_svc_php_variant"))
	log.Debugln("SVC_PHP_DEBUG_VARIANT:", viper.GetString(core.AppName+"_svc_php_debug_variant"))

	SetSyncSettingsByEnvType()

	// pwa-studio: everything is disabled, except node container
	if core.CheckRegexInString("^pwa-studio", envType) {
		if !viper.IsSet(core.AppName + "_node") {
			viper.Set(core.AppName+"_node", "1")
		}

		if !viper.IsSet(core.AppName + "_db") {
			viper.Set(core.AppName+"_db", "0")
		}

		if !viper.IsSet(core.AppName + "_nginx") {
			viper.Set(core.AppName+"_nginx", "0")
		}

		if !viper.IsSet(core.AppName + "_php_fpm") {
			viper.Set(core.AppName+"_php_fpm", "0")
		}

		if !viper.IsSet(core.AppName + "_redis") {
			viper.Set(core.AppName+"_redis", "0")
		}

		if !viper.IsSet(core.AppName + "_varnish") {
			viper.Set(core.AppName+"_varnish", "0")
		}

		if !viper.IsSet(core.AppName + "_elasticsearch") {
			viper.Set(core.AppName+"_elasticsearch", "0")
		}

		if !viper.IsSet(core.AppName + "_opensearch") {
			viper.Set(core.AppName+"_opensearch", "0")
		}
		if !viper.IsSet(core.AppName + "_opensearch_dashboards") {
			viper.Set(core.AppName+"_opensearch_dashboards", "0")
		}

		if !viper.IsSet(core.AppName + "_rabbitmq") {
			viper.Set(core.AppName+"_rabbitmq", "0")
		}
	}

	// not local: only nginx, db and redis are enabled, php-fpm is running locally
	if !core.CheckRegexInString(`^local`, envType) {
		if !viper.IsSet(core.AppName + "_php_fpm") {
			viper.Set(core.AppName+"_php_fpm", "1")
		}

		if !viper.IsSet(core.AppName + "_nginx") {
			viper.Set(core.AppName+"_nginx", "1")
		}

		if !viper.IsSet(core.AppName + "_db") {
			viper.Set(core.AppName+"_db", "1")
		}

		if !viper.IsSet(core.AppName + "_redis") {
			viper.Set(core.AppName+"_redis", "1")
		}
	}

	// local: varnish, elasticsearch and rabbitmq only
	if core.CheckRegexInString("^local", envType) {
		if !viper.IsSet(core.AppName + "_varnish") {
			viper.Set(core.AppName+"_varnish", "1")
		}

		if !viper.IsSet(core.AppName + "_elasticsearch") {
			viper.Set(core.AppName+"_elasticsearch", "0")
		}

		if !viper.IsSet(core.AppName + "_opensearch") {
			viper.Set(core.AppName+"_opensearch", "1")
		}

		if !viper.IsSet(core.AppName + "_rabbitmq") {
			viper.Set(core.AppName+"_rabbitmq", "1")
		}
	}

	// windows
	if runtime.GOOS == "windows" && !viper.IsSet("xdebug_connect_back_host") {
		viper.Set("xdebug_connect_back_host", "host.docker.internal")
	}

	// For linux, if UID is 1000, there is no need to use the socat proxy.
	if runtime.GOOS == "linux" && os.Geteuid() == 1000 && !viper.IsSet("ssh_auth_sock_path_env") {
		viper.Set("ssh_auth_sock_path_env", "/run/host-services/ssh-auth.sock")
	}

	err := core.AppendEnvironmentTemplates(t, templateList, "networks")
	if err != nil {
		return err
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
		if viper.GetString(core.AppName+"_"+strings.Replace(svc, "-", "_", -1)) == "1" {
			err = core.AppendEnvironmentTemplates(t, templateList, svc)
			if err != nil {
				return err
			}
		}
	}

	err = core.AppendEnvironmentTemplates(t, templateList, envType)
	if err != nil {
		return err
	}

	additionalMagentoSvcs := map[string]string{
		core.AppName + "_test_db":        envType + ".tests",
		core.AppName + "_split_sales":    envType + ".splitdb.sales",
		core.AppName + "_split_checkout": envType + ".splitdb.checkout",
	}
	for k, v := range additionalMagentoSvcs {
		if viper.GetString(k) == "1" {
			err = core.AppendEnvironmentTemplates(t, templateList, v)
			if err != nil {
				return err
			}
		}
	}

	externalSVCs := map[string][]string{
		"blackfire": {"blackfire", envType + ".blackfire"},
		"allure":    {"allure"},
		"selenium":  {"selenium"},
		"magepack":  {envType + ".magepack"},
	}
	for k, v := range externalSVCs {
		if viper.GetString(core.AppName+"_"+k) == "1" {
			for _, tpl := range v {
				err = core.AppendEnvironmentTemplates(t, templateList, tpl)
				if err != nil {
					return err
				}
			}
		}
	}

	// ./.reward/reward-env.yml
	// ./.reward/reward-env.os.yml
	additionalTemplates := []string{
		core.AppName + "-env.yml",
		fmt.Sprintf("%[1]v-env.%[2]v.yml", core.AppName, runtime.GOOS),
	}

	log.Traceln("AdditionalTemplatesPath: ", additionalTemplates)

	err = core.AppendTemplatesFromPaths(t, templateList, additionalTemplates)
	if err != nil {
		return err
	}

	// selenium
	if viper.GetString(core.AppName+"_selenium_debug") == "1" {
		viper.Set(core.AppName+"_selenium_debug", "-debug")
	} else {
		viper.Set(core.AppName+"_selenium_debug", "")
	}

	return nil
}

// EnvBuildDockerComposeCommand builds up the docker-compose command by passing it the previously built templates.
func EnvBuildDockerComposeCommand(args []string, suppressOsStdOut ...bool) (string, error) {
	envTemplate := new(template.Template)

	envTemplateList := list.New()

	err := EnvBuildDockerComposeTemplate(envTemplate, envTemplateList)
	if err != nil {
		return "", err
	}

	dockerComposeConfigs, err := core.ConvertTemplateToComposeConfig(envTemplate, envTemplateList)
	if err != nil {
		return "", err
	}

	out, err := core.RunDockerComposeWithConfig(args, dockerComposeConfigs, suppressOsStdOut...)
	if err != nil {
		return out, err
	}

	return out, nil
}
