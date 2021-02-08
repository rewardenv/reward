package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const defaultConfig = `---
log_level: info
debug: false

# Uncomment the following settings to disable default services.
# All of them are enabled by default.
#reward_portainer: 0
#reward_dnsmasq: 0
#reward_tunnel: 0
#reward_mailhog: 0
#reward_phpmyadmin: 0
#reward_elastichq: 0

#reward_allow_superuser: 0
`

func InstallCmd() error {
	if getReinstallFlag() || getUninstallFlag() {
		return uninstall()
	}

	if !getUninstallFlag() {
		return install()
	}

	return nil
}

func uninstall() error {
	appHomeDir := GetAppHomeDir()

	// Make sure we are not deleting "/"
	match, _ := regexp.MatchString("^/$", appHomeDir)
	if !match {
		if _, err := os.Stat(appHomeDir); !os.IsNotExist(err) {
			if confirmation := AskForConfirmation(
				fmt.Sprintf("Are you sure you want to delete %v?", appHomeDir)); confirmation {
				log.Debugf("Deleting: %v\n", appHomeDir)

				err = os.RemoveAll(appHomeDir)

				if err != nil {
					return err
				}
			}

			if confirmation := AskForConfirmation(
				fmt.Sprintf("Are you sure you want to delete %v?", viper.GetString(AppName+"_config_file"))); confirmation {
				log.Debugf("Deleting: %v\n", viper.GetString(AppName+"_config_file"))

				err = os.Remove(viper.GetString(AppName + "_config_file"))

				if err != nil {
					return err
				}
			}
		} else {
			return fmt.Errorf(AppName + " is not installed")
		}
	}

	return nil
}

func install() error {
	// On windows this command should run in elevated command prompt
	osDistro := GetOSDistro()
	if osDistro == "windows" {
		if !IsAdmin() {
			log.Printf("Running %v in an Elevated command prompt...", AppName)
			RunMeElevated()
		}
	}

	appHomeDir := GetAppHomeDir()

	err := CreateDir(appHomeDir, getInstallModeFlag())
	if err != nil {
		return err
	}

	log.Debugf("Chmod %v dir: %v to %v\n", AppName, appHomeDir, os.FileMode(getInstallModeFlag()))

	err = os.Chmod(appHomeDir, os.FileMode(getInstallModeFlag()))
	if err != nil {
		return err
	}

	if !getInstallCaCertFlag() && !getInstallDNSFlag() && !getInstallSSHKeyFlag() && !getInstallSSHConfigFlag() {
		configFile := viper.GetString(AppName + "_config_file")
		log.Debugln("Creating default config:", configFile)

		if !CheckFileExistsAndRecreate(configFile) {
			err = CreateDirAndWriteBytesToFile([]byte(defaultConfig), configFile)
			if err != nil {
				return err
			}
		}
	}

	// CA Cert
	if !getInstallDNSFlag() && !getInstallSSHKeyFlag() && !getInstallSSHConfigFlag() {
		sslDir := filepath.Join(appHomeDir, "ssl")
		caDir := filepath.Join(sslDir, caBaseDir)

		log.Debugln("Creating certificate to directory:", caDir)

		caCertExist := CheckCaCertificateExistInDir(caDir)

		if !caCertExist {
			err = CreateCaCertificate(caDir)
			if err != nil {
				return err
			}

			err = InstallCaCertificate(caDir)
			if err != nil {
				return err
			}
		}
	}

	// Dns Resolver
	if !getInstallCaCertFlag() && !getInstallSSHKeyFlag() && !getInstallSSHConfigFlag() {
		installDNSResolver()
	}

	// SSH Key
	if !getInstallCaCertFlag() && !getInstallDNSFlag() && !getInstallSSHConfigFlag() {
		keyPath := filepath.Join(appHomeDir, "tunnel", "ssh_key")

		keyFileExist := CheckFileExistsAndRecreate(keyPath)
		if !keyFileExist {
			err = GenerateAndSaveSSHKeys(2048, keyPath)
			if err != nil {
				return err
			}
		}

		// Since bind mounts are native on linux to use .pub file as authorized_keys file in tunnel it
		//   must have proper perms.
		if runtime.GOOS == "linux" {
			cmdChown := fmt.Sprintf("sudo chown -v %v:%v %v", 0, 0, filepath.Join(appHomeDir, "tunnel", "ssh_key.pub"))
			cmd := exec.Command("/bin/sh", "-c", cmdChown)

			log.Debugf("Running command: %v", cmd)

			out, err := cmd.CombinedOutput()

			log.Debugf("output: %v", string(out))

			if err != nil {
				log.Fatalln(err)
			}
		}
	}

	if osDistro != "windows" {
		if !getInstallCaCertFlag() && !getInstallDNSFlag() && !getInstallSSHKeyFlag() {
			// Install SSH settings for the SSH tunnel
			err = installSSHConfig()
			if err != nil {
				return err
			}
		}

		// Create composer directory
		home, err := homedir.Dir()
		if err != nil {
			return err
		}

		err = CreateDir(home + "/.composer")
		if err != nil {
			return err
		}
	} else {
		_ = AskForConfirmation("Press ENTER to continue...", false)
	}

	return nil
}

func getReinstallFlag() bool {
	return viper.GetBool(AppName + "_install_reinstall")
}
func getUninstallFlag() bool {
	return viper.GetBool(AppName + "_install_uninstall")
}
func getInstallCaCertFlag() bool {
	return viper.GetBool(AppName + "_install_ca_cert")
}
func getInstallDNSFlag() bool {
	return viper.GetBool(AppName + "_install_dns")
}
func getInstallSSHKeyFlag() bool {
	return viper.GetBool(AppName + "_install_ssh_key")
}
func getInstallSSHConfigFlag() bool {
	return viper.GetBool(AppName + "_install_ssh_config")
}
func getInstallModeFlag() int {
	return viper.GetInt(AppName + "_install_app_home_mode")
}
