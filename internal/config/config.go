package config

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"reward/internal/util"
)

var (
	// ErrEnvNameIsInvalid occurs when the environment name is invalid. It should be a valid hostname.
	ErrEnvNameIsInvalid = fmt.Errorf("environment name is invalid, it should match RFC1178")
	// ErrEnvIsEmpty occurs when environment name is empty.
	ErrEnvIsEmpty = fmt.Errorf("env name is empty. please run `reward env-init`")

	// ErrInvokedAsRootUser occurs when the Application was called by Root user.
	ErrInvokedAsRootUser = fmt.Errorf(
		"in most cases, you should not run as root user except for `self-update`. if you are sure you want to do this, use REWARD_ALLOW_SUPERUSER=1",
	)
)

type Config struct {
	appName string
	*viper.Viper
}

func New(name, ver string) *Config {
	c := &Config{
		appName: name,
		Viper:   viper.GetViper(),
	}

	c.SetDefault(fmt.Sprintf("%s_version", name), version.Must(version.NewVersion(ver)).String())

	return c
}

func (c *Config) Init() *Config {
	c.AddConfigPath(".")

	cfg := c.GetString(fmt.Sprintf("%s_config_file", c.appName))
	if cfg != "" {
		c.AddConfigPath(filepath.Dir(cfg))
		c.SetConfigName(filepath.Base(cfg))
		c.SetConfigType("yaml")
	}
	c.AutomaticEnv()

	if err := c.ReadInConfig(); err != nil {
		log.Debugf("%v", err)
	}

	c.AddConfigPath(".")
	c.SetConfigName(".env")
	c.SetConfigType("dotenv")

	if err := c.MergeInConfig(); err != nil {
		log.Debugf("%v", err)
	}

	c.SetDefault(fmt.Sprintf("%s_ssl_dir", c.appName), filepath.Join(c.AppHomeDir(), "ssl"))
	c.SetDefault(fmt.Sprintf("%s_composer_dir", c.appName), filepath.Join(util.HomeDir(), ".composer"))
	c.SetDefault(fmt.Sprintf("%s_ssh_dir", c.appName), filepath.Join(util.HomeDir(), ".ssh"))
	c.SetDefault(fmt.Sprintf("%s_runtime_os", c.appName), runtime.GOOS)
	c.SetDefault(fmt.Sprintf("%s_runtime_arch", c.appName), runtime.GOARCH)

	c.SetLogging()

	return c
}

func (c *Config) SetLogging() {
	switch {
	case c.GetString("log_level") == "trace":
		log.SetLevel(log.TraceLevel)
		log.SetReportCaller(true)
	case c.IsDebug(), c.GetString("log_level") == "debug":
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

func (c *Config) AppName() string {
	return c.appName
}

// AppHomeDir returns the application's home directory.
func (c *Config) AppHomeDir() string {
	return c.GetString(fmt.Sprintf("%s_home_dir", c.appName))
}

func (c *Config) AppVersion() string {
	return c.GetString(fmt.Sprintf("%s_version", c.appName))
}

// EnvName returns the environment name in lowercase format.
func (c *Config) EnvName() string {
	return strings.ToLower(c.GetString(fmt.Sprintf("%s_env_name", c.appName)))
}

// EnvType returns the environment type in lowercase format.
func (c *Config) EnvType() string {
	return strings.ToLower(c.GetString(fmt.Sprintf("%s_env_type", c.appName)))
}

// IsDebug returns true if debug mode is set.
func (c *Config) IsDebug() bool {
	return c.GetBool("debug")
}

// EnvCheck returns an error if the env name is empty (.env file does not contain an env name).
func (c *Config) EnvCheck() error {
	if len(strings.TrimSpace(c.EnvName())) == 0 {
		return ErrEnvIsEmpty
	}

	return nil
}

// SuperuserAllowed returns true if the application is allowed to be invoked by root.
func (c *Config) SuperuserAllowed() bool {
	return c.GetBool(fmt.Sprintf("%s_allow_superuser", c.appName))
}

// BlackfireEnabled returns true if the blackfire container is enabled.
func (c *Config) BlackfireEnabled() bool {
	return c.GetBool(fmt.Sprintf("%s_blackfire", c.appName))
}

// BlackfireCommand returns the command which is called when the application manipulates blackfire.
func (c *Config) BlackfireCommand() string {
	c.SetDefault(fmt.Sprintf("%s_env_blackfire_command", c.AppName), "blackfire")

	return c.GetString(fmt.Sprintf("%s_env_blackfire_command", c.AppName))
}

// BlackfireContainer returns the container name of the Blackfire debug container.
func (c *Config) BlackfireContainer() string {
	c.SetDefault(fmt.Sprintf("%_env_blackfire_container", c.AppName), "php-blackfire")

	return c.GetString(fmt.Sprintf("%s_env_blackfire_container", c.AppName))
}

// IsDBEnabled returns true if the database service is enabled for the current environment.
func (c *Config) IsDBEnabled() bool {
	return viper.GetBool(fmt.Sprintf("%s_db", c.appName))
}

// CheckInvokerUser returns an error if the invoker user is root.
func (c *Config) CheckInvokerUser(cmd *cobra.Command) error {
	// If the REWARD_ALLOW_SUPERUSER=1 is set or the Distro is Windows then we can skip this.
	if c.SuperuserAllowed() || util.OSDistro() == "windows" {
		return nil
	}

	// Most of the commands should run by normal users except `self-update`.
	if cmd.Name() != "self-update" && util.IsAdmin() {
		return ErrInvokedAsRootUser
	}

	return nil
}
