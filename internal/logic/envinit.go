package logic

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/pkg/util"
)

// RunCmdEnvInit creates a .env file for envType based on envName.
func (c *Client) RunCmdEnvInit(cmd *cobra.Command, args []string) error {
	if len(args) == 0 && len(strings.TrimSpace(c.EnvName())) == 0 {
		log.Println("Please provide an environment name.")

		_ = cmd.Help()

		os.Exit(1)
	}

	if len(args) > 0 {
		c.Set(fmt.Sprintf("%s_env_name", c.AppName()), args[0])

		if len(args) > 1 {
			if util.ContainsString(c.ValidEnvTypes(), args[1]) {
				c.Set(fmt.Sprintf("%s_env_type", c.AppName()), args[1])
			} else {
				return config.ErrUnknownEnvType
			}
		}
	}

	path := c.Cwd()
	envType := c.EnvType()
	envName := c.EnvName()

	if !validateEnvName(envName) {
		return config.ErrEnvNameIsInvalid
	}

	if !util.ContainsString(c.ValidEnvTypes(), envType) {
		return config.ErrUnknownEnvType
	}

	log.Debugln("name:", envName)
	log.Debugln("type:", envType)

	envFilePath := filepath.Join(path, ".env")

	envFileExist := util.CheckFileExistsAndRecreate(envFilePath)

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

`, strings.ToUpper(c.AppName()), envName, envType, webRoot,
	)
	envFileContent := strings.Join([]string{envBase, c.EnvTypes()[envType]}, "")

	if !envFileExist {
		err := util.CreateDirAndWriteToFile([]byte(envFileContent), envFilePath)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	err := c.CheckAndCreateLocalAppDirs()
	if err != nil {
		return fmt.Errorf("cannot create local app dirs: %w", err)
	}

	return nil
}

func (c *Client) CheckAndCreateLocalAppDirs() error {
	localAppDir := filepath.Join(c.Cwd(), fmt.Sprintf(".%s", c.AppName()))

	_, err := util.FS.Stat(localAppDir)
	if !os.IsNotExist(err) {
		return nil
	}

	err = util.CreateDir(localAppDir, nil)
	if err != nil {
		return fmt.Errorf("cannot create directory: %w", err)
	}

	if c.SvcEnabledPermissive("nginx") {
		err = util.CreateDir(filepath.Join(localAppDir, "nginx"), nil)
		if err != nil {
			return fmt.Errorf("cannot create directory: %w", err)
		}
	}

	if c.SvcEnabledStrict("varnish") {
		err = util.CreateDir(filepath.Join(localAppDir, "varnish"), nil)
		if err != nil {
			return fmt.Errorf("cannot create directory: %w", err)
		}
	}

	return nil
}

func validateEnvName(name string) bool {
	validatorRegex := `^[A-Za-z0-9](?:[A-Za-z0-9\-]{0,61}[A-Za-z0-9])?$`
	if !util.CheckRegexInString(validatorRegex, name) {
		log.Debugln("Environment name validator regex is not matching.")

		return false
	}

	log.Debugln("Environment name validator regex matches.")

	return true
}
