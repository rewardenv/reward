package config

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"reward/internal/docker"
	"reward/internal/dockercompose"
	"reward/internal/shell"
	"reward/internal/util"
)

var (
	// ErrEnvNameIsInvalid occurs when the environment name is invalid. It should be a valid hostname.
	ErrEnvNameIsInvalid = fmt.Errorf("environment name is invalid, it should match RFC1178")
	// ErrEnvIsEmpty occurs when environment name is empty.
	ErrEnvIsEmpty = fmt.Errorf("env name is empty. please run `reward env-init`")
	// ErrUnknownAction occurs when an unknown actions is called.
	ErrUnknownAction = fmt.Errorf("unknown action error")

	// ErrInvokedAsRootUser occurs when the Application was called by Root user.
	ErrInvokedAsRootUser = fmt.Errorf(
		"in most cases, you should not run as root user except for `self-update`. if you are sure you want to do this, use REWARD_ALLOW_SUPERUSER=1",
	)

	ErrHostnameRequired   = fmt.Errorf("hostname is required")
	ErrCaCertDoesNotExist = fmt.Errorf("the root CA certificate is missing, please run 'reward install' and try again")

	// ErrUnknownEnvType occurs when an unknown environment type is specified.
	ErrUnknownEnvType = fmt.Errorf("unknown env type")
)

var (
	// FS is the implementation of Afero Filesystem. It's a filesystem wrapper and used for testing.
	FS = &afero.Afero{Fs: afero.NewOsFs()}
)

type Config struct {
	*viper.Viper
	Shell               shell.Shell
	Docker              *docker.Client
	DockerCompose       *dockercompose.Client
	ShellUser           string
	ShellContainer      string
	DefaultShellCommand string
	TmpFiles            *list.List
}

func New(name, ver string) *Config {
	c := &Config{
		Viper:    viper.GetViper(),
		Shell:    &shell.LocalShell{},
		TmpFiles: list.New(),
	}

	c.SetDefault("app_name", name)
	c.SetDefault(fmt.Sprintf("%s_version", name), version.Must(version.NewVersion(ver)).String())

	return c
}

func (c *Config) Init() *Config {
	c.AddConfigPath(".")

	cfg := c.GetString(fmt.Sprintf("%s_config_file", c.AppName()))
	if cfg != "" {
		c.AddConfigPath(filepath.Dir(cfg))
		c.SetConfigName(filepath.Base(cfg))
		c.SetConfigType("yaml")
	}

	c.AutomaticEnv()

	if err := c.ReadInConfig(); err != nil {
		log.Debugf("%s", err)
	}

	c.AddConfigPath(".")
	c.SetConfigName(".env")
	c.SetConfigType("dotenv")
	c.SetTypeByDefaultValue(true)

	if err := c.MergeInConfig(); err != nil {
		log.Debugf("%s", err)
	}

	c.SetDefault("silence_errors", true)
	c.SetDefault(fmt.Sprintf("%s_ssl_dir", c.AppName()), filepath.Join(c.AppHomeDir(), "ssl"))
	c.SetDefault(fmt.Sprintf("%s_composer_dir", c.AppName()), filepath.Join(util.HomeDir(), ".composer"))
	c.SetDefault(fmt.Sprintf("%s_ssh_dir", c.AppName()), filepath.Join(util.HomeDir(), ".ssh"))
	c.SetDefault(fmt.Sprintf("%s_runtime_os", c.AppName()), runtime.GOOS)
	c.SetDefault(fmt.Sprintf("%s_runtime_arch", c.AppName()), runtime.GOARCH)
	// c.SetDefault(fmt.Sprintf("%s_repo_url", c.AppName()),
	// 	"https://github.com/rewardenv/reward/releases/latest/download")
	c.SetDefault(fmt.Sprintf("%s_repo_url", c.AppName()),
		"https://api.github.com/repos/rewardenv/reward/releases")
	c.SetDefault(fmt.Sprintf("%s_ssl_base_dir", c.AppName()), "ssl")
	c.SetDefault(fmt.Sprintf("%s_ssl_dir", c.AppName()), filepath.Join(c.AppHomeDir(), c.SSLBaseDir()))
	c.SetDefault(fmt.Sprintf("%s_ssl_ca_base_dir", c.AppName()), "rootca")
	c.SetDefault(fmt.Sprintf("%s_ssl_ca_dir", c.AppName()), filepath.Join(c.SSLDir(), c.SSLCABaseDir()))
	c.SetDefault(fmt.Sprintf("%s_ssl_cert_base_dir", c.AppName()), "certs")
	c.SetDefault(fmt.Sprintf("%s_ssl_cert_dir", c.AppName()), filepath.Join(c.SSLDir(), c.SSLCertBaseDir()))
	c.SetDefault(fmt.Sprintf("%s_resolve_domain_to_traefik", c.AppName()), true)
	c.SetDefault(fmt.Sprintf("%s_plugins_dir", c.AppName()), filepath.Join(c.AppHomeDir(), "plugins"))
	c.SetDefault(fmt.Sprintf("%s_plugins_available", c.AppName()), []string{"cloud"})

	// Default Shortcuts
	c.SetDefault(fmt.Sprintf("%s_shortcuts", c.AppName()), map[string]string{
		"up":           "svc up && env up",
		"down":         "env down ; svc down",
		"restart":      "env down ; env up",
		"sync-restart": "sync stop && sync start",
	})

	// Sync
	c.SetDefault(fmt.Sprintf("%s_mutagen_url", c.AppName()),
		"https://github.com/mutagen-io/mutagen/releases/download/v0.14.0/mutagen_windows_amd64_v0.14.0.zip")
	c.SetDefault(fmt.Sprintf("%s_mutagen_required_version", c.AppName()), "0.11.8")

	if util.OSDistro() == "windows" || util.OSDistro() == "darwin" {
		c.SetDefault(fmt.Sprintf("%s_sync_enabled", c.AppName()), true)
	}

	// SVC Defaults
	c.SetDefault(fmt.Sprintf("%s_portainer", c.AppName()), true)
	c.SetDefault(fmt.Sprintf("%s_dnsmasq", c.AppName()), true)
	c.SetDefault(fmt.Sprintf("%s_mailhog", c.AppName()), true)
	c.SetDefault(fmt.Sprintf("%s_phpmyadmin", c.AppName()), true)
	c.SetDefault(fmt.Sprintf("%s_tunnel", c.AppName()), true)
	c.SetDefault(fmt.Sprintf("%s_elastichq", c.AppName()), true)
	c.SetDefault(fmt.Sprintf("%s_adminer", c.AppName()), false)

	// Env Defaults
	c.SetDefault(fmt.Sprintf("%s_env_synced_container", c.AppName()), "php-fpm")
	c.SetDefault(fmt.Sprintf("%s_env_synced_dir", c.AppName()), "/var/www/html")
	c.SetDefault(fmt.Sprintf("%s_shared_composer", c.AppName()), true)
	c.SetDefault(fmt.Sprintf("%s_opensearch_dashboards", c.AppName()), true)

	// Bootstrap
	c.SetDefault(fmt.Sprintf("%s_full_bootstrap", c.AppName()), false)
	c.SetDefault(fmt.Sprintf("%s_composer_no_parallel", c.AppName()), true)
	c.SetDefault(fmt.Sprintf("%s_skip_composer_install", c.AppName()), false)
	c.SetDefault(fmt.Sprintf("%s_no_pull", c.AppName()), false)
	c.SetDefault(fmt.Sprintf("%s_with_sampledata", c.AppName()), false)
	c.SetDefault(fmt.Sprintf("%s_magento_disable_tfa", c.AppName()), false)
	c.SetDefault(fmt.Sprintf("%s_reset_admin_url", c.AppName()), false)

	if c.EnvType() == "magento1" {
		c.SetDefault(fmt.Sprintf("%s_magento_version", c.AppName()), "1.9.4")
	} else {
		c.SetDefault(fmt.Sprintf("%s_magento_version", c.AppName()), "2.4.5-p1")
	}

	c.SetDefault(fmt.Sprintf("%s_magento_type", c.AppName()), "community")
	c.SetDefault(fmt.Sprintf("%s_magento_mode", c.AppName()), "developer")
	c.SetDefault(fmt.Sprintf("%s_db_prefix", c.AppName()), "")
	c.SetDefault(fmt.Sprintf("%s_crypt_key", c.AppName()), "")
	c.SetDefault(fmt.Sprintf("%s_shopware_version", c.AppName()), "6.4.18.0")
	c.SetDefault(fmt.Sprintf("%s_shopware_mode", c.AppName()), "production")

	c.SetDefault(fmt.Sprintf("%s_env_db_command", c.AppName()), "mysql")
	c.SetDefault(fmt.Sprintf("%s_env_db_dump_command", c.AppName()), "mysqldump")
	c.SetDefault(fmt.Sprintf("%s_env_db_container", c.AppName()), "db")
	c.SetDefault(fmt.Sprintf("%s_single_web_container", c.AppName()), false)

	c.SetLogging()

	c.Docker = docker.Must(docker.NewClient(c.DockerHost()))
	c.DockerCompose = dockercompose.NewClient(c.Shell, c.TmpFiles)

	return c
}

// SetLogging sets the logging level based on the command line flags and environment variables.
func (c *Config) SetLogging() {
	switch {
	case c.GetString("log_level") == "trace":
		c.Set("silence_errors", false)
		log.SetLevel(log.TraceLevel)
		log.SetReportCaller(true)
	case c.IsDebug(), c.GetString("log_level") == "debug":
		c.Set("silence_errors", false)
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
	case c.GetString("log_level") == "info":
		log.SetLevel(log.InfoLevel)
	case c.GetString("log_level") == "warning":
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(log.ErrorLevel)
	}

	log.SetFormatter(
		&log.TextFormatter{
			DisableColors:          c.GetBool("disable_colors"),
			PadLevelText:           true,
			ForceColors:            true,
			DisableLevelTruncation: true,
			FullTimestamp:          true,
			DisableTimestamp:       !c.GetBool("debug"),
			QuoteEmptyFields:       true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := strings.ReplaceAll(path.Base(f.File), "reward/", "")

				return fmt.Sprintf("%s()", f.Function), fmt.Sprintf(" %s:%d", filename, f.Line)
			},
		},
	)
}

func (c *Config) SilenceErrors() bool {
	return c.GetBool("silence_errors")
}

func (c *Config) Check(cmd *cobra.Command, args []string) error {
	err := c.CheckInvokerUser(cmd)
	if err != nil {
		return fmt.Errorf("error checking invoker user: %w", err)
	}

	if !c.Installed() && cmd.Name() != "install" {
		return fmt.Errorf("reward is not installed")
	}

	err = c.Docker.Check()
	if err != nil {
		return fmt.Errorf("error checking docker: %w", err)
	}

	err = c.DockerCompose.Check()
	if err != nil {
		return fmt.Errorf("error checking docker-compose: %w", err)
	}

	err = c.EnvCheck()
	if err != nil {
		return fmt.Errorf("error checking env: %w", err)
	}

	return nil
}

func (c *Config) SkipCleanup() bool {
	return c.GetBool(fmt.Sprintf("%s_skip_cleanup", c.AppName()))
}

// Cleanup removes all the temporary template files.
func (c *Config) Cleanup() error {
	log.Debugln("Cleaning up temporary files...")

	if c.SkipCleanup() {
		log.Debugln("...skipping cleanup.")

		return nil
	}

	if c.TmpFiles.Len() == 0 {
		log.Debugln("...no temporary files to clean up.")

		return nil
	}

	for e := c.TmpFiles.Front(); e != nil; e = e.Next() {
		log.Tracef("Cleaning up: %s", e.Value)

		err := os.Remove(fmt.Sprint(e.Value))
		if err != nil {
			return fmt.Errorf("failed to remove temporary file: %w", err)
		}
	}

	log.Debugln("...cleanup done.")

	return nil
}

// AppName returns the application's name.
func (c *Config) AppName() string {
	return c.GetString("app_name")
}

// AppHomeDir returns the application's home directory.
func (c *Config) AppHomeDir() string {
	return c.GetString(fmt.Sprintf("%s_home_dir", c.AppName()))
}

// AppVersion returns the application's version.
func (c *Config) AppVersion() string {
	return c.GetString(fmt.Sprintf("%s_version", c.AppName()))
}

// EnvName returns the environment name in lowercase format.
func (c *Config) EnvName() string {
	return strings.ToLower(c.GetString(fmt.Sprintf("%s_env_name", c.AppName())))
}

// EnvType returns the environment type in lowercase format.
func (c *Config) EnvType() string {
	return strings.ToLower(c.GetString(fmt.Sprintf("%s_env_type", c.AppName())))
}

func (c *Config) EnvInitialized() bool {
	_, err := FS.Open(".env")

	return err == nil
}

// IsDebug returns true if debug mode is set.
func (c *Config) IsDebug() bool {
	return c.GetBool("debug")
}

// Cwd returns the current working directory.
func (c *Config) Cwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Panicln(err)
	}

	return cwd
}

// EnvCheck returns an error if the env name is empty (.env file does not contain an env name).
func (c *Config) EnvCheck() error {
	if !c.EnvInitialized() {
		return nil
	}

	if len(strings.TrimSpace(c.EnvName())) == 0 {
		return ErrEnvIsEmpty
	}

	return nil
}

// RepoURL returns the repository URL for self-update.
func (c *Config) RepoURL() string {
	return c.GetString(fmt.Sprintf("%s_repo_url", c.AppName()))
}

// SuperuserAllowed returns true if the application is allowed to be invoked by root.
func (c *Config) SuperuserAllowed() bool {
	return c.GetBool(fmt.Sprintf("%s_allow_superuser", c.AppName()))
}

// BlackfireEnabled returns true if the blackfire container is enabled.
func (c *Config) BlackfireEnabled() bool {
	return c.GetBool(fmt.Sprintf("%s_blackfire", c.AppName()))
}

// BlackfireCommand returns the command which is called when the application manipulates blackfire.
func (c *Config) BlackfireCommand() string {
	c.SetDefault(fmt.Sprintf("%s_env_blackfire_command", c.AppName()), "blackfire")

	return c.GetString(fmt.Sprintf("%s_env_blackfire_command", c.AppName()))
}

// BlackfireContainer returns the container name of the Blackfire debug container.
func (c *Config) BlackfireContainer() string {
	c.SetDefault(fmt.Sprintf("%s_env_blackfire_container", c.AppName()), "php-blackfire")

	return c.GetString(fmt.Sprintf("%s_env_blackfire_container", c.AppName()))
}

// IsDBEnabled returns true if the database service is enabled for the current environment.
func (c *Config) IsDBEnabled() bool {
	return c.GetBool(fmt.Sprintf("%s_db", c.AppName()))
}

// CheckInvokerUser returns an error if the invoker user is root.
func (c *Config) CheckInvokerUser(cmd *cobra.Command) error {
	// If the REWARD_ALLOW_SUPERUSER=true is set or the Distro is Windows then we can skip this.
	if c.SuperuserAllowed() || util.OSDistro() == "windows" {
		return nil
	}

	// Most of the commands should run by normal users except `self-update`.
	if cmd.Name() != "self-update" && util.IsAdmin() {
		return ErrInvokedAsRootUser
	}

	return nil
}

// SyncedContainer returns the container name of the synced container from REWARD_ENV_SYNCED_CONTAINER variable.
func (c *Config) SyncedContainer() string {
	return c.GetString(fmt.Sprintf("%s_env_synced_container", c.AppName()))
}

// SetSyncedContainer sets the synced container name in REWARD_ENV_SYNCED_CONTAINER variable.
func (c *Config) SetSyncedContainer(s string) {
	c.Set(fmt.Sprintf("%s_env_synced_container", c.AppName()), s)
}

func (c *Config) DefaultSyncedDir(envType string) string {
	conf := c.GetString(fmt.Sprintf("%s_sync_path", c.AppName()))
	if conf != "" {
		return conf
	}

	switch envType {
	case "pwa-studio":
		return "/usr/src/app"
	default:
		return "/var/www/html"
	}
}

func (c *Config) DefaultSyncedContainer(envType string) string {
	conf := c.GetString(fmt.Sprintf("%s_sync_container", c.AppName()))
	if conf != "" {
		return conf
	}

	switch envType {
	case "pwa-studio":
		return "node"
	default:
		return "php-fpm"
	}
}

func (c *Config) SetPHPDefaults(envType string) {
	if !c.SingleWebContainer() {
		c.Set(
			fmt.Sprintf("%s_svc_php_variant", c.AppName()),
			fmt.Sprintf("-%s", envType),
		)
		c.Set(
			fmt.Sprintf("%s_svc_php_debug_variant", c.AppName()),
			fmt.Sprintf("-%s", envType),
		)
	} else {
		c.Set(
			fmt.Sprintf("%s_svc_php_variant", c.AppName()),
			fmt.Sprintf("-%s-web", envType),
		)
		c.Set(
			fmt.Sprintf("%s_svc_php_debug_variant", c.AppName()),
			fmt.Sprintf("-%s", envType),
		)
	}
}

func (c *Config) SetPWADefaults() {
	viper.SetDefault(fmt.Sprintf("%s_node", c.AppName()), true)
	viper.SetDefault(fmt.Sprintf("%s_db", c.AppName()), false)
	viper.SetDefault(fmt.Sprintf("%s_nginx", c.AppName()), false)
	viper.SetDefault(fmt.Sprintf("%s_php_fpm", c.AppName()), false)
	viper.SetDefault(fmt.Sprintf("%s_redis", c.AppName()), false)
	viper.SetDefault(fmt.Sprintf("%s_varnish", c.AppName()), false)
	viper.SetDefault(fmt.Sprintf("%s_elasticsearch", c.AppName()), false)
	viper.SetDefault(fmt.Sprintf("%s_opensearch", c.AppName()), false)
	viper.SetDefault(fmt.Sprintf("%s_opensearch_dashboards", c.AppName()), false)
	viper.SetDefault(fmt.Sprintf("%s_rabbitmq", c.AppName()), false)
}

func (c *Config) SetNonLocalDefaults() {
	viper.SetDefault(fmt.Sprintf("%s_php_fpm", c.AppName()), true)
	viper.SetDefault(fmt.Sprintf("%s_nginx", c.AppName()), true)
	viper.SetDefault(fmt.Sprintf("%s_db", c.AppName()), true)
	viper.SetDefault(fmt.Sprintf("%s_redis", c.AppName()), true)
}

func (c *Config) SetLocalDefaults() {
	viper.SetDefault(fmt.Sprintf("%s_varnish", c.AppName()), false)
	viper.SetDefault(fmt.Sprintf("%s_elasticsearch", c.AppName()), false)
	viper.SetDefault(fmt.Sprintf("%s_opensearch", c.AppName()), false)
	viper.SetDefault(fmt.Sprintf("%s_rabbitmq", c.AppName()), false)
}

// TODO: test if this works as expected.
func (c *Config) SetSeleniumDefaults() {
	if c.GetBool(fmt.Sprintf("%s_selenium_debug", c.AppName())) {
		c.Set(fmt.Sprintf("%s_selenium_debug", c.AppName()), "-debug")

		return
	}

	c.Set(fmt.Sprintf("%s_selenium_debug", c.AppName()), "")
}

// SetSyncSettings sets the settings for synchronization.
func (c *Config) SetSyncSettings() {
	c.SetSyncedContainer(c.DefaultSyncedContainer(c.EnvType()))
	c.SetSyncedDir(c.DefaultSyncedDir(c.EnvType()))
}

// SyncedDir returns the directory which is synced with the host stored in the REWARD_ENV_SYNCED_DIR environment variable.
func (c *Config) SyncedDir() string {
	return c.GetString(fmt.Sprintf("%s_env_synced_dir", c.AppName()))
}

// SetSyncedDir sets the REWARD_ENV_SYNCED_DIRECTORY variable.
func (c *Config) SetSyncedDir(s string) {
	c.Set(fmt.Sprintf("%s_env_synced_dir", c.AppName()), s)
}

// MutagenSyncFile returns the file path of the mutagen sync file.
func (c *Config) MutagenSyncFile() string {
	return filepath.Join(c.Cwd(), fmt.Sprintf(".%s", c.AppName()), "mutagen.yml")
}

// MutagenSyncIgnore returns the additional mutagen ignored files.
func (c *Config) MutagenSyncIgnore() string {
	return c.GetString(fmt.Sprintf("%s_sync_ignore", c.AppName()))
}

// WebRoot returns the content of the WEB_ROOT variable.
func (c *Config) WebRoot() string {
	return c.GetString(fmt.Sprintf("%s_web_root", c.AppName()))
}

// MutagenURL returns the content of the REWARD_MUTAGEN_URL variable.
func (c *Config) MutagenURL() string {
	return c.GetString(fmt.Sprintf("%s_mutagen_url", c.AppName()))
}

// MutagenRequiredVersion returns the content of the REWARD_MUTAGEN_VERSION variable.
func (c *Config) MutagenRequiredVersion() string {
	return c.GetString(fmt.Sprintf("%s_mutagen_required_version", c.AppName()))
}

// SyncEnabled returns true for macOS and Windows if it's not disabled explicitly (or if the WSL2 Direct Mount
// option is not enabled on Windows).
func (c *Config) SyncEnabled() bool {
	if util.OSDistro() == "windows" || util.OSDistro() == "darwin" {
		return c.GetBool(fmt.Sprintf("%s_sync_enabled", c.AppName()))
	}

	return false
}

// ValidEnvTypes return a list of valid environment types based on the predefined EnvTypes.
func (c *Config) ValidEnvTypes() []string {
	envTypes := c.EnvTypes()

	validEnvTypes := make([]string, 0, len(envTypes))
	for key := range envTypes {
		validEnvTypes = append(validEnvTypes, key)
	}

	return validEnvTypes
}

func (c *Config) EnvTypes() map[string]string {
	return map[string]string{
		"generic-php": fmt.Sprintf(
			`%[1]v_DB=true
%[1]v_REDIS=true

MARIADB_VERSION=10.4
NODE_VERSION=16
PHP_VERSION=7.4
REDIS_VERSION=6.0
COMPOSER_VERSION=2

MYSQL_ROOT_PASSWORD=app
MYSQL_DATABASE=app
MYSQL_USER=app
MYSQL_PASSWORD=app

NGINX_ROOT=/var/www/html
NGINX_PUBLIC=
`, strings.ToUpper(c.AppName()),
		),

		"magento1": fmt.Sprintf(
			`%[1]v_DB=true
%[1]v_REDIS=true

MARIADB_VERSION=10.3
NODE_VERSION=16
PHP_VERSION=7.2
REDIS_VERSION=5.0
COMPOSER_VERSION=1

%[1]v_SELENIUM=false
%[1]v_SELENIUM_DEBUG=false
%[1]v_BLACKFIRE=false

BLACKFIRE_CLIENT_ID=
BLACKFIRE_CLIENT_TOKEN=
BLACKFIRE_SERVER_ID=
BLACKFIRE_SERVER_TOKEN=
`, strings.ToUpper(c.AppName()),
		),

		"magento2": fmt.Sprintf(
			`%[1]v_DB=true
%[1]v_ELASTICSEARCH=false
%[1]v_OPENSEARCH=true
%[1]v_OPENSEARCH_DASHBOARDS=false
%[1]v_VARNISH=true
%[1]v_RABBITMQ=true
%[1]v_REDIS=true
%[1]v_MERCURE=false

ELASTICSEARCH_VERSION=7.16
OPENSEARCH_VERSION=1.2
MARIADB_VERSION=10.4
NODE_VERSION=16
PHP_VERSION=8.1
RABBITMQ_VERSION=3.9
REDIS_VERSION=6.0
VARNISH_VERSION=7.0
COMPOSER_VERSION=2.1

%[1]v_SYNC_IGNORE=

%[1]v_ALLURE=false
%[1]v_SELENIUM=false
%[1]v_SELENIUM_DEBUG=false
%[1]v_BLACKFIRE=false
%[1]v_SPLIT_SALES=false
%[1]v_SPLIT_CHECKOUT=false
%[1]v_TEST_DB=false
%[1]v_MAGEPACK=false

BLACKFIRE_CLIENT_ID=
BLACKFIRE_CLIENT_TOKEN=
BLACKFIRE_SERVER_ID=
BLACKFIRE_SERVER_TOKEN=

XDEBUG_VERSION=
`, strings.ToUpper(c.AppName()),
		),

		"laravel": fmt.Sprintf(
			`MARIADB_VERSION=10.4
NODE_VERSION=16
PHP_VERSION=7.4
REDIS_VERSION=6.0
COMPOSER_VERSION=2

%[1]v_DB=true
%[1]v_REDIS=true
%[1]v_MERCURE=false

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
`, strings.ToUpper(c.AppName()),
		),

		"pwa-studio": fmt.Sprintf(
			`NODE_VERSION=16
%[1]v_VARNISH=false
VARNISH_VERSION=6.5

`, strings.ToUpper(c.AppName()),
		),

		"symfony": fmt.Sprintf(
			`%[1]v_DB=true
%[1]v_REDIS=true
%[1]v_RABBITMQ=false
%[1]v_ELASTICSEARCH=false
%[1]v_OPENSEARCH=false
%[1]v_OPENSEARCH_DASHBOARDS=false
%[1]v_VARNISH=false
%[1]v_MERCURE=false

MARIADB_VERSION=10.4
NODE_VERSION=16
PHP_VERSION=7.4
RABBITMQ_VERSION=3.8
REDIS_VERSION=6.0
VARNISH_VERSION=6.5
COMPOSER_VERSION=2
`, strings.ToUpper(c.AppName()),
		),

		"shopware": fmt.Sprintf(
			`%[1]v_DB=true
%[1]v_REDIS=true
%[1]v_RABBITMQ=false
%[1]v_ELASTICSEARCH=false
%[1]v_OPENSEARCH=true
%[1]v_VARNISH=false

MARIADB_VERSION=10.4
NODE_VERSION=16
PHP_VERSION=7.4
RABBITMQ_VERSION=3.8
REDIS_VERSION=6.0
VARNISH_VERSION=6.5
COMPOSER_VERSION=2.4.4
`, strings.ToUpper(c.AppName()),
		),

		"wordpress": fmt.Sprintf(
			`MARIADB_VERSION=10.4
NODE_VERSION=16
PHP_VERSION=7.4
COMPOSER_VERSION=2

%[1]v_DB=true
%[1]v_REDIS=false

APP_ENV=local
APP_DEBUG=true

DB_CONNECTION=mysql
DB_HOST=db
DB_PORT=3306
DB_DATABASE=wordpress
DB_USERNAME=wordpress
DB_PASSWORD=wordpress
`, strings.ToUpper(c.AppName()),
		),

		"local": fmt.Sprintf(
			`
%[1]v_SHELL_CONTAINER=php-fpm
%[1]v_SHELL_COMMAND=bash
%[1]v_SHELL_USER=www-data
%[1]v_SYNC_CONTAINER=php-fpm
%[1]v_SYNC_PATH=/var/www/html
%[1]v_SYNC_ENABLED=true

%[1]v_RABBITMQ=false
%[1]v_ELASTICSEARCH=false
%[1]v_OPENSEARCH=false
%[1]v_VARNISH=false

RABBITMQ_VERSION=3.8
ELASTICSEARCH_VERSION=7.16
OPENSEARCH_VERSION=1.2
REDIS_VERSION=6.0
VARNISH_VERSION=6.5`, strings.ToUpper(c.AppName()),
		),
	}
}

// EnvNetworkName returns the environments docker network name in lowercase format.
func (c *Config) EnvNetworkName() string {
	return strings.ToLower(fmt.Sprintf("%s_default", c.EnvName()))
}

func (c *Config) DockerHost() string {
	return c.GetString("docker_host")
}

func (c *Config) ShopwareVersion() (*version.Version, error) {
	return version.NewVersion(c.GetString(fmt.Sprintf("%s_shopware_version", c.AppName())))
}

func (c *Config) ShopwareMode() string {
	return c.GetString(fmt.Sprintf("%s_shopware_mode", c.AppName()))
}

// MagentoVersion returns a *version.Version object which contains the Magento version.
func (c *Config) MagentoVersion() (*version.Version, error) {
	log.Debugln("Looking up Magento version...")

	v := new(version.Version)

	type ComposerJSON struct {
		Name    string            `json:"name"`
		Version string            `json:"version"`
		Require map[string]string `json:"require"`
	}

	var composerJSON ComposerJSON

	if util.FileExists("composer.json") {
		data, err := FS.ReadFile("composer.json")
		if err != nil {
			log.Debugln("...cannot read composer.json. Using .env settings.")

			v = c.MagentoVersionFromConfig()
		}

		if err = json.Unmarshal(data, &composerJSON); err != nil {
			log.Debugln("...cannot unmarshal composer.json. Using .env settings.")

			v = c.MagentoVersionFromConfig()
		}

		if util.CheckRegexInString(`^magento/magento2(ce|ee)$`,
			composerJSON.Name) && composerJSON.Version != "" {
			re := regexp.MustCompile(semver.SemVerRegex)
			ver := re.Find([]byte(composerJSON.Version))

			log.Debugf("...using magento/magento2(ce|ee) package version from composer.json. Found version: %s.",
				ver)

			v, err = version.NewVersion(string(ver))
			if err != nil {
				return nil, fmt.Errorf("cannot parse Magento version from composer.json: %w", err)
			}
		}

		if v.String() == "" {
			for key, val := range composerJSON.Require {
				if util.CheckRegexInString(`^magento/product-(enterprise|community)-edition$`, key) {
					re := regexp.MustCompile(semver.SemVerRegex)
					ver := re.Find([]byte(val))

					log.Debugf("...using magento/product-(enterprise-community)-edition package version from composer.json. Found version: %s.",
						ver)

					v, err = version.NewVersion(string(ver))
					if err != nil {
						return nil, fmt.Errorf("cannot parse Magento version from composer.json: %w",
							err)
					}
				} else if util.CheckRegexInString(`^magento/magento-cloud-metapackage$`, key) {
					re := regexp.MustCompile(semver.SemVerRegex)
					ver := re.Find([]byte(val))

					log.Debugf("...using magento/magento-cloud-metapackage package version from composer.json. Found version: %s.",
						ver)

					v, err = version.NewVersion(string(ver))
					if err != nil {
						return nil, fmt.Errorf("cannot parse Magento version from composer.json: %w",
							err)
					}
				}
			}
		}

		return v, nil
	}

	v = c.MagentoVersionFromConfig()

	log.Debugf("...cannot find Magento version in composer.json, using .env settings. Version: %s.", v.String())

	return v, nil
}

// MagentoVersionFromConfig returns a *version.Version object from Config settings.
// Note: If it's unset, it will return a dedicated latest version.
func (c *Config) MagentoVersionFromConfig() *version.Version {
	return version.Must(version.NewVersion(c.GetString(fmt.Sprintf("%s_magento_version",
		c.AppName()))))
}

// ServiceDomain returns the application's service domain.
func (c *Config) ServiceDomain() string {
	return c.GetString(fmt.Sprintf("%s_service_domain", c.AppName()))
}

func (c *Config) SSLBaseDir() string {
	return c.GetString(fmt.Sprintf("%s_ssl_base_dir", c.AppName()))
}

func (c *Config) SSLDir() string {
	return c.GetString(fmt.Sprintf("%s_ssl_dir", c.AppName()))
}

func (c *Config) SSLCABaseDir() string {
	return c.GetString(fmt.Sprintf("%s_ssl_ca_base_dir", c.AppName()))
}

func (c *Config) SSLCADir() string {
	return c.GetString(fmt.Sprintf("%s_ssl_ca_dir", c.AppName()))
}

func (c *Config) SSLCertBaseDir() string {
	return c.GetString(fmt.Sprintf("%s_ssl_cert_base_dir", c.AppName()))
}

func (c *Config) SSLCertDir() string {
	return c.GetString(fmt.Sprintf("%s_ssl_cert_dir", c.AppName()))
}

// DockerPeeredServices attaches/detaches the common services to the current environment's docker network.
func (c *Config) DockerPeeredServices(action, networkName string) error {
	if action != "connect" && action != "disconnect" {
		return ErrUnknownAction
	}

	var (
		ctx                  = context.Background()
		dockerPeeredServices = []string{"traefik"}

		// Enabled by default
		dockerAdditionalServices = []string{
			"tunnel",
			"mailhog",
			"phpmyadmin",
			"elastichq",
		}

		// Disabled by default
		dockerOptionalServices = []string{
			"adminer",
		}
	)

	for _, svc := range dockerAdditionalServices {
		if c.SvcEnabledPermissive(svc) {
			dockerPeeredServices = append(dockerPeeredServices, svc)
		}
	}

	for _, svc := range dockerOptionalServices {
		if c.SvcEnabledStrict(svc) {
			dockerPeeredServices = append(dockerPeeredServices, svc)
		}
	}

	for _, v := range dockerPeeredServices {
		networkSettings := new(network.EndpointSettings)

		if v == "traefik" && c.ResolveDomainToTraefik() {
			networkSettings.Aliases = []string{
				c.TraefikDomain(),
				c.TraefikFullDomain(),
			}

			log.Debugln("Network aliases for Traefik container:", networkSettings.Aliases)
		}

		containers, err := c.Docker.ContainerList(ctx, types.ContainerListOptions{
			Filters: filters.NewArgs(
				filters.KeyValuePair{
					Key:   "name",
					Value: v,
				},
			),
		})
		if err != nil {
			return fmt.Errorf("cannot list containers: %w", err)
		}

		for _, container := range containers {
			if action == "connect" {
				log.Debugf("Connecting container: %s to network %s...", container.Names, networkName)

				err = c.Docker.NetworkConnect(ctx, networkName, container.ID, networkSettings)
				if err != nil {
					log.Debugf("%s", err)
				}

				log.Debugln("...connected.")
			}

			if action == "disconnect" {
				log.Debugf("Disconnecting container: %s from network %s.", container.Names, networkName)

				err = c.Docker.NetworkDisconnect(ctx, networkName, container.ID, false)
				if err != nil {
					log.Debugf("%s", err)
				}

				log.Debugln("...disconnected.")
			}
		}
	}

	return nil
}

func (c *Config) ResolveDomainToTraefik() bool {
	return c.GetBool(fmt.Sprintf("%s_resolve_domain_to_traefik", c.AppName()))
}

// TraefikDomain returns traefik domain from Viper settings.
func (c *Config) TraefikDomain() string {
	return c.GetString("traefik_domain")
}

// TraefikSubdomain returns traefik subdomain from Viper settings.
func (c *Config) TraefikSubdomain() string {
	return c.GetString("traefik_subdomain")
}

// TraefikFullDomain returns traefik full domain (subdomain + domain merged).
func (c *Config) TraefikFullDomain() string {
	if c.TraefikSubdomain() == "" {
		return c.TraefikDomain()
	}

	return fmt.Sprintf("%s.%s", c.TraefikSubdomain(), c.TraefikDomain())
}

// SvcEnabledPermissive returns true if the s service is enabled in Viper settings. This function is also going to
// return true if the service is not mentioned in Viper settings (defaults to true).
func (c *Config) SvcEnabledPermissive(s string) bool {
	if c.IsSet(fmt.Sprintf("%s_%s", c.AppName(), s)) {
		return c.GetBool(fmt.Sprintf("%s_%s", c.AppName(), s))
	}

	return true
}

// SvcEnabledStrict returns true if the s service is enabled in Viper settings. This function is going to
// return false if the service is not mentioned in Viper settings (defaults to false).
func (c *Config) SvcEnabledStrict(s string) bool {
	if c.IsSet(fmt.Sprintf("%s_%s", c.AppName(), s)) {
		return c.GetBool(fmt.Sprintf("%s_%s", c.AppName(), s))
	}

	return false
}

func (c *Config) PluginsAvailable() []string {
	return c.GetStringSlice(fmt.Sprintf("%s_plugins_available", c.AppName()))
}

func (c *Config) PluginsDir() string {
	return c.GetString(fmt.Sprintf("%s_plugins_dir", c.AppName()))
}

func (c *Config) Plugins() []string {
	content, err := FS.ReadDir(c.PluginsDir())
	if err != nil {
		return nil
	}

	var files []string

	for _, v := range content {
		if v.IsDir() {
			continue
		}

		if strings.HasPrefix(v.Name(), ".") {
			continue
		}

		files = append(files, filepath.Join(c.PluginsDir(), v.Name()))
	}

	return files
}

func (c *Config) Shortcuts() map[string]string {
	return c.GetStringMapString(fmt.Sprintf("%s_shortcuts", c.AppName()))
}

// ComposerVersion returns the Composer Version defined in Config settings.
func (c *Config) ComposerVersion() *version.Version {
	if c.GetString("composer_version") != "1" {
		return version.Must(version.NewVersion(c.GetString("composer_version")))
	}

	return version.Must(version.NewVersion("1.0"))
}

// ServiceEnabled returns true if service is enabled in Config settings.
func (c *Config) ServiceEnabled(servicename string) bool {
	if c.IsSet(fmt.Sprintf("%s_%s", c.AppName(), servicename)) {
		return c.GetBool(fmt.Sprintf("%s_%s", c.AppName(), servicename))
	}

	return false
}

// MagentoBackendFrontname returns Magento admin path from Config settings.
func (c *Config) MagentoBackendFrontname() string {
	if c.IsSet("magento_backend_frontname") {
		return c.GetString("magento_backend_frontname")
	}

	return "admin"
}

// FullBootstrap checks if full bootstrap is enabled in configs.
func (c *Config) FullBootstrap() bool {
	return c.GetBool(fmt.Sprintf("%s_full_bootstrap", c.AppName()))
}

// Parallel checks if composer parallel mode is enabled in configs.
func (c *Config) Parallel() bool {
	return !c.GetBool(fmt.Sprintf("%s_composer_no_parallel", c.AppName()))
}

// SkipComposerInstall checks if composer install is disabled in configs.
func (c *Config) SkipComposerInstall() bool {
	return c.GetBool(fmt.Sprintf("%s_skip_composer_install", c.AppName()))
}

// NoPull checks if docker-compose pull is disabled in configs.
func (c *Config) NoPull() bool {
	return c.GetBool(fmt.Sprintf("%s_no_pull", c.AppName()))
}

// WithSampleData checks if Magento 2 sample data is enabled in configs.
func (c *Config) WithSampleData() bool {
	return c.GetBool(fmt.Sprintf("%s_with_sampledata", c.AppName()))
}

// MagentoDisableTFA checks if the installer should Disable TwoFactorAuth module in configs.
func (c *Config) MagentoDisableTFA() bool {
	return c.GetBool(fmt.Sprintf("%s_magento_disable_tfa", c.AppName()))
}

// ResetAdminURL checks if the installer should Reset the Admin URLs in Viper settings.
func (c *Config) ResetAdminURL() bool {
	return c.GetBool(fmt.Sprintf("%s_reset_admin_url", c.AppName()))
}

// MagentoType returns Magento type: enterprise or community (default: community).
func (c *Config) MagentoType() string {
	if c.GetString(fmt.Sprintf("%s_magento_type", c.AppName())) == "enterprise" ||
		c.GetString(fmt.Sprintf("%s_magento_type", c.AppName())) == "commerce" {
		c.Set(fmt.Sprintf("%s_magento_type", c.AppName()), "enterprise")
	}

	return c.GetString(fmt.Sprintf("%s_magento_type", c.AppName()))
}

// MagentoMode returns Magento mode: developer or production (default: developer).
func (c *Config) MagentoMode() string {
	return c.GetString(fmt.Sprintf("%s_magento_mode", c.AppName()))
}

func (c *Config) DBPrefix() string {
	return c.GetString(fmt.Sprintf("%s_db_prefix", c.AppName()))
}

func (c *Config) CryptKey() string {
	return c.GetString(fmt.Sprintf("%s_crypt_key", c.AppName()))
}

// DBCommand returns the command which is called when the application manipulates the database.
func (c *Config) DBCommand() string {
	return c.GetString(fmt.Sprintf("%s_env_db_command", c.AppName()))
}

// DBDumpCommand returns the command which is called when the application dumps a database.
func (c *Config) DBDumpCommand() string {
	return c.GetString(fmt.Sprintf("%s_env_db_dump_command", c.AppName()))
}

// DBContainer returns the name of the database container.
func (c *Config) DBContainer() string {
	return c.GetString(fmt.Sprintf("%s_env_db_container", c.AppName()))
}

// SingleWebContainer returns true if Single Web Container setting is enabled in Viper settings.
func (c *Config) SingleWebContainer() bool {
	return c.GetBool(fmt.Sprintf("%s_single_web_container", c.AppName()))
}

// SetShellContainer changes the container used for the reward shell command.
func (c *Config) SetShellContainer(envType string) {
	c.ShellContainer = c.defaultShellContainer(envType)
}

// SetDefaultShellCommand changes the command invoked by reward shell command.
func (c *Config) SetDefaultShellCommand(containerName string) {
	c.DefaultShellCommand = c.defaultShellCommand(containerName)
}

// SetShellUser changes the user of the reward shell command.
func (c *Config) SetShellUser(containerName string) {
	c.ShellUser = c.defaultShellUser(containerName)
}

func (c *Config) defaultShellContainer(envType string) string {
	conf := c.GetString(c.AppName() + "_shell_container")
	if conf != "" {
		return conf
	}

	switch envType {
	case "pwa-studio":
		return "node"
	default:
		return "php-fpm"
	}
}

func (c *Config) defaultShellCommand(containerName string) string {
	conf := viper.GetString(c.AppName() + "_shell_command")
	if conf != "" {
		return conf
	}

	switch containerName {
	case "php-fpm":
		return "bash"
	default:
		return "sh"
	}
}

func (c *Config) defaultShellUser(containerName string) string {
	conf := c.GetString(c.AppName() + "_shell_user")
	if conf != "" {
		return conf
	}

	switch containerName {
	case "php-fpm":
		return "www-data"
	case "node":
		return "node"
	default:
		return "root"
	}
}

// Installed returns true if the application is installed, false anyway.
func (c *Config) Installed() bool {
	return util.FileExists(c.InstallMarkerFilePath())
}

// InstallMarkerFilePath returns the filepath of the Install Marker file.
func (c *Config) InstallMarkerFilePath() string {
	return filepath.Join(c.AppHomeDir(), ".installed")
}
