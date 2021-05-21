package internal

import (
	"bytes"
	"container/list"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	envTypes = map[string]string{
		"generic-php": fmt.Sprintf(`%[1]v_DB=1
%[1]v_REDIS=1

MARIADB_VERSION=10.4
NODE_VERSION=10
PHP_VERSION=7.4
REDIS_VERSION=5.0
COMPOSER_VERSION=2

MYSQL_ROOT_PASSWORD=app
MYSQL_DATABASE=app
MYSQL_USER=app
MYSQL_PASSWORD=app

NGINX_ROOT=/var/www/html
NGINX_PUBLIC=
`, strings.ToUpper(AppName)),

		"magento1": fmt.Sprintf(`%[1]v_DB=1
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
`, strings.ToUpper(AppName)),

		"magento2": fmt.Sprintf(`%[1]v_DB=1
%[1]v_ELASTICSEARCH=1
%[1]v_VARNISH=1
%[1]v_RABBITMQ=1
%[1]v_REDIS=1

ELASTICSEARCH_VERSION=7.12
MARIADB_VERSION=10.3
NODE_VERSION=10
PHP_VERSION=7.4
RABBITMQ_VERSION=3.8
REDIS_VERSION=5.0
VARNISH_VERSION=6.0
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
`, strings.ToUpper(AppName)),

		"laravel": fmt.Sprintf(`MARIADB_VERSION=10.4
NODE_VERSION=10
PHP_VERSION=7.4
REDIS_VERSION=5.0
COMPOSER_VERSION=2

%[1]v_DB=1
%[1]v_REDIS=1

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
`, strings.ToUpper(AppName)),

		"pwa-studio": `NODE_VERSION=10

`,

		"symfony": fmt.Sprintf(`%[1]v_DB=1
%[1]v_REDIS=1
%[1]v_RABBITMQ=0
%[1]v_ELASTICSEARCH=0
%[1]v_VARNISH=0

MARIADB_VERSION=10.4
NODE_VERSION=10
PHP_VERSION=7.4
RABBITMQ_VERSION=3.8
REDIS_VERSION=5.0
VARNISH_VERSION=6.0
COMPOSER_VERSION=2
`, strings.ToUpper(AppName)),

		"shopware": fmt.Sprintf(`%[1]v_DB=1
%[1]v_REDIS=1
%[1]v_RABBITMQ=0
%[1]v_ELASTICSEARCH=0
%[1]v_VARNISH=0

MARIADB_VERSION=10.4
NODE_VERSION=10
PHP_VERSION=7.4
RABBITMQ_VERSION=3.8
REDIS_VERSION=5.0
VARNISH_VERSION=6.0
COMPOSER_VERSION=2
`, strings.ToUpper(AppName)),

		"wordpress": fmt.Sprintf(`MARIADB_VERSION=10.4
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
`, strings.ToUpper(AppName)),
	}

	validEnvTypes []string
	composeBuffer bytes.Buffer

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
	if ContainsString(args, "down") {
		err := DockerPeeredServices("disconnect", GetEnvNetworkName())
		if err != nil {
			return err
		}
	}

	// up: connect peered service containers to environment network
	if ContainsString(args, "up") {
		// check if network already exist
		networkExist, err := CheckDockerNetworkExist(GetEnvNetworkName())
		if err != nil {
			return err
		}

		if !networkExist {
			log.Println("Creating network...")

			var passedArgs []string

			if ContainsString(args, "--") {
				passedArgs = InsertStringBeforeOccurrence(args, "--no-start", "--")
			} else {
				passedArgs = append(args, "--no-start")
			}

			log.Tracef("args: %#v, updated args: %#v", args, passedArgs)

			err = EnvRunDockerCompose(passedArgs)
			if err != nil {
				return err
			}
		}

		err = DockerPeeredServices("connect", GetEnvNetworkName())
		if err != nil {
			return err
		}

		if !ContainsString(args, "-d") && !ContainsString(args, "--detach") {
			args = InsertStringAfterOccurrence(args, "--detach", "up")
		}
	}

	// traefik: lookup address of traefik container in the environment network
	traefikAddress, err := LookupContainerAddressInNetwork("traefik", GetEnvNetworkName())
	if err != nil {
		return CannotFindContainerError("traefik")
	}

	viper.Set("traefik_address", traefikAddress)

	log.Tracef("Traefik container address in network %v: %v", GetEnvNetworkName(), traefikAddress)

	// mutagen: sync file
	if IsMutagenSyncEnabled() {
		err = GenerateMutagenTemplateFileIfNotExist()
		if err != nil {
			return err
		}
	}

	// mutagen: pause sync if needed
	if ContainsString(args, "stop") {
		if IsMutagenSyncEnabled() {
			err := SyncPauseCmd()
			if err != nil {
				return err
			}
		}
	}

	// pass orchestration through to docker-compose
	err = EnvRunDockerCompose(args, false)
	if err != nil {
		return err
	}

	// mutagen: resume mutagen sync if available and php-fpm container id hasn't changed
	if ContainsString(args, "up") || ContainsString(args, "start") {
		if IsMutagenSyncEnabled() && !IsContainerChanged(GetSyncedContainer()) && !ContainsString(args, "--") {
			err := SyncResumeCmd()
			if err != nil {
				return err
			}
		}
	}

	// mutagen: start mutagen sync if needed (container id changed or previously didn't exist
	if ContainsString(args, "up") || ContainsString(args, "start") {
		if IsMutagenSyncEnabled() && IsContainerChanged(GetSyncedContainer()) && !ContainsString(args, "--") {
			err := SyncStartCmd()
			if err != nil {
				return err
			}
		}
	}

	// mutagen: stop mutagen sync if needed
	if ContainsString(args, "down") {
		if IsMutagenSyncEnabled() {
			err := SyncStopCmd()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// EnvCheck returns an error if the env name is empty (.env file does not contain an env name).
func EnvCheck() error {
	if len(strings.TrimSpace(GetEnvName())) == 0 {
		return ErrEnvIsEmpty
	}

	return nil
}

func validateEnvName(name string) bool {
	validatorRegex := `^[A-Za-z0-9](?:[A-Za-z0-9\-]{0,61}[A-Za-z0-9])?$`
	if !CheckRegexInString(validatorRegex, name) {
		log.Debugln("Environment name validator regex is not matching.")

		return false
	}

	log.Debugln("Environment name validator regex matches.")

	return true
}

// EnvInitCmd creates a .env file for envType based on envName.
func EnvInitCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 && len(strings.TrimSpace(GetEnvName())) == 0 {
		log.Println("Please provide an environment name.")

		_ = cmd.Help()

		os.Exit(1)
	}

	if len(args) > 0 {
		viper.Set(AppName+"_env_name", args[0])

		log.Debugf("args(%v): %v", len(args), args)

		if len(args) > 1 {
			if ContainsString(GetValidEnvTypes(), args[1]) {
				viper.Set(AppName+"_env_type", args[1])
			} else {
				return ErrUnknownEnvType
			}
		}
	}

	path := GetCwd()
	envType := GetEnvType()
	envName := GetEnvName()

	if !validateEnvName(envName) {
		return ErrEnvNameIsInvalid
	}

	if !ContainsString(GetValidEnvTypes(), envType) {
		return ErrUnknownEnvType
	}

	log.Debugln("name:", envName)
	log.Debugln("type:", envType)

	envFilePath := filepath.Join(path, ".env")

	envFileExist := CheckFileExistsAndRecreate(envFilePath)

	envBase := fmt.Sprintf(`%[1]v_ENV_NAME=%[2]v
%[1]v_ENV_TYPE=%[3]v
%[1]v_WEB_ROOT=/

TRAEFIK_DOMAIN=%[2]v.test
TRAEFIK_SUBDOMAIN=
TRAEFIK_EXTRA_HOSTS=

`, strings.ToUpper(AppName), envName, envType)
	envFileContent := strings.Join([]string{envBase, envTypes[envType]}, "")

	if !envFileExist {
		err := CreateDirAndWriteBytesToFile([]byte(envFileContent), envFilePath)
		if err != nil {
			return fmt.Errorf("%w", err)
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
		GetCwd(),
		"--project-name",
		GetEnvName(),
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
	envType := GetEnvType()

	log.Debugln("ENV_TYPE:", envType)

	// magento1,2, shopware wordpress have their own php-fpm containers
	if CheckRegexInString(`^magento|wordpress|shopware`, envType) {
		log.Debugln("Setting SVC_PHP_VARIANT.")

		if !IsSingleWebContainer() {
			viper.Set(AppName+"_svc_php_variant", "-"+envType)
			viper.Set(AppName+"_svc_php_debug_variant", "-"+envType)
		} else {
			viper.Set(AppName+"_svc_php_variant", "-"+envType+"-web")
			viper.Set(AppName+"_svc_php_debug_variant", "-"+envType+"")
		}
	}

	log.Debugln("SVC_PHP_VARIANT:", viper.GetString(AppName+"_svc_php_variant"))
	log.Debugln("SVC_PHP_DEBUG_VARIANT:", viper.GetString(AppName+"_svc_php_debug_variant"))

	SetSyncSettingsByEnvType()

	// pwa-studio: everything is disabled, except node container
	if CheckRegexInString("^pwa-studio", envType) {
		if !viper.IsSet(AppName + "_node") {
			viper.Set(AppName+"_node", "1")
		}

		if !viper.IsSet(AppName + "_db") {
			viper.Set(AppName+"_db", "0")
		}

		if !viper.IsSet(AppName + "_nginx") {
			viper.Set(AppName+"_nginx", "0")
		}

		if !viper.IsSet(AppName + "_php_fpm") {
			viper.Set(AppName+"_php_fpm", "0")
		}

		if !viper.IsSet(AppName + "_redis") {
			viper.Set(AppName+"_redis", "0")
		}

		if !viper.IsSet(AppName + "_varnish") {
			viper.Set(AppName+"_varnish", "0")
		}

		if !viper.IsSet(AppName + "_elasticsearch") {
			viper.Set(AppName+"_elasticsearch", "0")
		}

		if !viper.IsSet(AppName + "_rabbitmq") {
			viper.Set(AppName+"_rabbitmq", "0")
		}
	}

	// not local: only nginx, db and redis are enabled, php-fpm is running locally
	if !CheckRegexInString(`^local`, envType) {
		if !viper.IsSet(AppName + "_php_fpm") {
			viper.Set(AppName+"_php_fpm", "1")
		}

		if !viper.IsSet(AppName + "_nginx") {
			viper.Set(AppName+"_nginx", "1")
		}

		if !viper.IsSet(AppName + "_db") {
			viper.Set(AppName+"_db", "1")
		}

		if !viper.IsSet(AppName + "_redis") {
			viper.Set(AppName+"_redis", "1")
		}
	}

	// local: varnish, elasticsearch and rabbitmq only
	if CheckRegexInString("^local", envType) {
		if !viper.IsSet(AppName + "_varnish") {
			viper.Set(AppName+"_varnish", "1")
		}

		if !viper.IsSet(AppName + "_elasticsearch") {
			viper.Set(AppName+"_elasticsearch", "1")
		}

		if !viper.IsSet(AppName + "_rabbitmq") {
			viper.Set(AppName+"_rabbitmq", "1")
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

	err := AppendEnvironmentTemplates(t, templateList, "networks")
	if err != nil {
		return err
	}

	svcs := []string{
		"php-fpm",
		"nginx",
		"db",
		"elasticsearch",
		"varnish",
		"rabbitmq",
		"redis",
		"node",
	}
	for _, svc := range svcs {
		if viper.GetString(AppName+"_"+strings.Replace(svc, "-", "_", -1)) == "1" {
			err = AppendEnvironmentTemplates(t, templateList, svc)
			if err != nil {
				return err
			}
		}
	}

	err = AppendEnvironmentTemplates(t, templateList, envType)
	if err != nil {
		return err
	}

	additionalMagentoSvcs := map[string]string{
		AppName + "_test_db":        envType + ".tests",
		AppName + "_split_sales":    envType + ".splitdb.sales",
		AppName + "_split_checkout": envType + ".splitdb.checkout",
	}
	for k, v := range additionalMagentoSvcs {
		if viper.GetString(k) == "1" {
			err = AppendEnvironmentTemplates(t, templateList, v)
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
		if viper.GetString(AppName+"_"+k) == "1" {
			for _, tpl := range v {
				err = AppendEnvironmentTemplates(t, templateList, tpl)
				if err != nil {
					return err
				}
			}
		}
	}

	// ./.reward/reward-env.yml
	// ./.reward/reward-env.os.yml
	additionalTemplates := []string{
		AppName + "-env.yml",
		fmt.Sprintf("%[1]v-env.%[2]v.yml", AppName, runtime.GOOS),
	}

	log.Traceln("AdditionalTemplatesPath: ", additionalTemplates)

	err = AppendTemplatesFromPaths(t, templateList, additionalTemplates)
	if err != nil {
		return err
	}

	// selenium
	if viper.GetString(AppName+"_selenium_debug") == "1" {
		viper.Set(AppName+"_selenium_debug", "-debug")
	} else {
		viper.Set(AppName+"_selenium_debug", "")
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

	dockerComposeConfigs, err := ConvertTemplateToComposeConfig(envTemplate, envTemplateList)
	if err != nil {
		return "", err
	}

	out, err := RunDockerComposeWithConfig(args, dockerComposeConfigs, suppressOsStdOut...)
	if err != nil {
		return out, err
	}

	return out, nil
}
