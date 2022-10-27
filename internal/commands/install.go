package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/rewardenv/reward/internal/core"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const defaultConfig = `---
# Logging level, can be: trace, debug, info, warn, error
log_level: info

# Enable debugging (set log level to debug). Can be used as environment variable too.
# eg: DEBUG=true reward env up
debug: false

# Override default docker socket. Can be used as environment variable too.
# Default: Reward tries to extract the default socket using "docker context list" command
# Docker Desktop for Linux
#docker_host: unix://home/_USERNAME_/.docker/desktop/docker.sock
# Docker Desktop for macOS
#docker_host: unix://Users/_USERNAME_/.docker/run/docker.sock

# Uncomment the following settings to disable default services.
# These services are enabled by default.
#reward_portainer: 0
#reward_dnsmasq: 0
#reward_tunnel: 0
#reward_mailhog: 0
#reward_phpmyadmin: 0
#reward_elastichq: 0

# If you want to enable the following services, uncomment any of them.
# These services are disabled by default.
#reward_adminer: 1

############
# SERVICE CONTAINERS
# It's possible to change service container images using these vars:
#reward_traefik_image: "traefik"

#reward_portainer_image: "portainer/portainer-ce"

# Reward < v0.2.33 uses "jpillora/dnsmasq" as the default dnsmasq image.
# Reward >= v0.2.34 uses the internally built "docker.io/rewardenv/dnsmasq"
#reward_dnsmasq_image: "docker.io/rewardenv/dnsmasq"

#reward_mailhog_image: "docker.io/rewardenv/mailhog:1.0"

#reward_phpmyadmin_image: "phpmyadmin"

#reward_adminer_image: "dehy/adminer"

#reward_elastichq_image: "elastichq/elasticsearch-hq"

# Reward < v0.2.33 uses "panubo/sshd:1.1.0" as the default dnsmasq image.
# Reward >= v0.2.34 uses the internally built "docker.io/rewardenv/sshd"
#reward_tunnel_image: "docker.io/rewardenv/sshd"

# You can configure Traefik to bind additional http ports on top of the default port (80).
# reward_traefik_bind_additional_http_ports: [8080]
reward_traefik_bind_additional_http_ports: []

# You can configure Traefik to bind additional https ports on top of the default port (443).
# reward_traefik_bind_additional_https_ports: [8443,9443]
reward_traefik_bind_additional_https_ports: []

# By default Reward makes it possible to resolve the environment's domain to the nginx container's IP address
# inside the docker network. To disable this behaviour you can uncomment the following line.
#reward_resolve_domain_to_traefik: 0

# By default Reward is not allowed to run commands as root.
# To disable this check you can uncomment the following line.
#reward_allow_superuser: 1

# By default Reward is going to use sync session for Windows. With WSL2 it's possible to use well performing
# direct mount from WSL2's drive. It is disabled by default.
# To enable this functionality, uncomment the following line.
#reward_wsl2_direct_mount: 1

# By default Reward uses CentOS 7 based images. You can experiment with Debian based images with uncommenting this.
#reward_docker_image_base: debian

# By default Reward uses separated nginx + php-fpm containers.Enabling this setting will merge
# them to one "web" container
#reward_single_web_container: 1

# By default Reward mounts the ~/.composer directory from the host computer. Using this method you only have to set
# the Composer credentials once. In some situations you may want to use different Composer credentials per project.
# You can disable this sharing mechanism by setting the variable REWARD_SHARED_COMPOSER=0 in the project's .env file.
# Or you can disable it globally by setting the following variable to 0.
reward_shared_composer: 1

# By default mutagen sync is enabled in macOS and Windows, but you can disable it globally (here) or adding 
# REWARD_MUTAGEN_ENABLED=0 to the environment's .env file.
#reward_mutagen_enabled: 0
`

// InstallCmd represents the install command.
func InstallCmd() error {
	if reinstallFlag() || uninstallFlag() {
		return uninstall()
	}

	if !uninstallFlag() {
		return install()
	}

	return nil
}

// uninstall removes the application's settings.
func uninstall() error {
	appHomeDir := core.AppHomeDir()

	// Make sure we are not deleting "/"
	match, _ := regexp.MatchString("^/$", appHomeDir)
	if !match {
		if _, err := os.Stat(appHomeDir); !os.IsNotExist(err) {
			if confirmation := core.AskForConfirmation(
				fmt.Sprintf("Are you sure you want to delete %v?", appHomeDir),
			); confirmation {
				log.Debugf("Deleting: %v\n", appHomeDir)

				err = os.RemoveAll(appHomeDir)

				if err != nil {
					return err
				}
			}

			if confirmation := core.AskForConfirmation(
				fmt.Sprintf(
					"Are you sure you want to delete %v?",
					viper.GetString(core.AppName+"_config_file"),
				),
			); confirmation {
				log.Debugf("Deleting: %v\n", viper.GetString(core.AppName+"_config_file"))

				err = os.Remove(viper.GetString(core.AppName + "_config_file"))

				if err != nil {
					return err
				}
			}
		} else {
			return fmt.Errorf(core.AppName + " is not installed")
		}
	}

	return nil
}

// install configures the application default settings.
func install() error {
	// On windows this command should run in elevated command prompt
	osDistro := core.OSDistro()
	if osDistro == "windows" {
		if !core.IsAdmin() {
			log.Printf("Running %v in an Elevated command prompt...", core.AppName)
			core.RunMeElevated()
		}
	}

	appHomeDir := core.AppHomeDir()

	caser := cases.Title(language.English)
	// If we are not directly call installation for cacert, dns, ssh then check if the install marker already exists.
	if !installCaCertFlag() && !installDNSFlag() && !installSSHKeyFlag() && !installSSHConfigFlag() {
		if core.CheckFileExists(installMarkerFilePath()) {
			if !core.AskForConfirmation(caser.String(core.AppName) + " is already installed. Would you like to reinstall?") {
				return nil
			}
		}
	}

	// Create application's config directory
	if err := core.CreateDir(appHomeDir, installModeFlag()); err != nil {
		return err
	}

	log.Debugf("Chmod %v dir: %v to %v\n", core.AppName, appHomeDir, os.FileMode(installModeFlag()))

	// Change mode for it
	if err := os.Chmod(appHomeDir, os.FileMode(installModeFlag())); err != nil {
		return err
	}

	// If we are not directly call installation for cacert, dns, ssh then create the app's default config file.
	if !installCaCertFlag() && !installDNSFlag() && !installSSHKeyFlag() && !installSSHConfigFlag() {
		configFile := viper.GetString(core.AppName + "_config_file")
		log.Debugln("Creating default config:", configFile)

		if !core.CheckFileExistsAndRecreate(configFile) {
			if err := core.CreateDirAndWriteBytesToFile([]byte(defaultConfig), configFile); err != nil {
				return err
			}
		}
	}

	// Install CA Certificate
	if !installDNSFlag() && !installSSHKeyFlag() && !installSSHConfigFlag() {
		sslDir := filepath.Join(appHomeDir, "ssl")
		caDir := filepath.Join(sslDir, core.CaBaseDir)

		log.Debugln("Creating certificate to directory:", caDir)

		caCertExist := core.CheckCaCertificateExistInDir(caDir)

		if !caCertExist {
			if err := core.CreateCaCertificate(caDir); err != nil {
				return err
			}

			if err := core.InstallCaCertificate(caDir); err != nil {
				return err
			}
		}
	}

	// Install DNS resolver
	if !installCaCertFlag() && !installSSHKeyFlag() && !installSSHConfigFlag() {
		core.InstallDNSResolver()
	}

	// Install common SSH Key for Tunnel
	if !installCaCertFlag() && !installDNSFlag() && !installSSHConfigFlag() {
		keyPath := filepath.Join(appHomeDir, "tunnel", "ssh_key")

		// On linux, if we want to reinstall the pubfile we have to revert its permissions first
		if runtime.GOOS == "linux" && core.CheckFileExists(keyPath) {
			cmdChown := fmt.Sprintf(
				"sudo chown -v %v:%v %v", os.Getuid(), 0,
				filepath.Join(appHomeDir, "tunnel", "ssh_key.pub"),
			)
			cmd := exec.Command("/bin/sh", "-c", cmdChown)

			log.Debugf("Running command: %v", cmd)

			out, err := cmd.CombinedOutput()

			log.Debugf("output: %v", string(out))

			if err != nil {
				log.Fatalln(err)
			}
		}

		keyFileExist := core.CheckFileExistsAndRecreate(keyPath)
		if !keyFileExist {
			if err := core.GenerateAndSaveSSHKeys(2048, keyPath); err != nil {
				return err
			}
		}

		// Since bind mounts are native on linux to use .pub file as authorized_keys file in tunnel it
		//   must have proper perms.
		if runtime.GOOS == "linux" {
			cmdChown := fmt.Sprintf(
				"sudo chown -v %v:%v %v", 0, 0, filepath.Join(appHomeDir, "tunnel", "ssh_key.pub"),
			)
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
		if !installCaCertFlag() && !installDNSFlag() && !installSSHKeyFlag() {
			// Install SSH settings for the SSH tunnel
			if err := core.InstallSSHConfig(); err != nil {
				return err
			}
		}

		// Create composer directory
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		err = core.CreateDir(home + "/.composer")
		if err != nil {
			return err
		}
	} else {
		_ = core.AskForConfirmation("Press ENTER to continue...", false)
	}

	// Create file which marks the application as already installed (~/.reward/.installed)
	if err := putInstallMarkerFile(); err != nil {
		return err
	}

	// If the install command is not called with --ignore-svcs or the specific install options directly, then
	//    run `reward svc up`.
	if installInitServicesFlag() && !installCaCertFlag() &&
		!installDNSFlag() && !installSSHKeyFlag() && !installSSHConfigFlag() {
		if err := core.CheckDocker(); err != nil {
			return err
		}

		if err := SvcCmd([]string{"up"}); err != nil {
			return err
		}
	}

	return nil
}

// reinstallFlag returns true if the --reinstall flag is set during the execution.
func reinstallFlag() bool {
	return viper.GetBool(core.AppName + "_install_reinstall")
}

// uninstallFlag returns true if the --uninstall flag is set during the execution.
func uninstallFlag() bool {
	return viper.GetBool(core.AppName + "_install_uninstall")
}

// installCaCertFlag returns true if --install-ca-cert flag is set during the execution.
func installCaCertFlag() bool {
	return viper.GetBool(core.AppName + "_install_ca_cert")
}

// installDNSFlag returns true if --install-dns flag is set during the execution.
func installDNSFlag() bool {
	return viper.GetBool(core.AppName + "_install_dns")
}

// installSSHKeyFlag returns true if --install-ssh-key flag is set during the execution.
func installSSHKeyFlag() bool {
	return viper.GetBool(core.AppName + "_install_ssh_key")
}

// installSSHConfigFlag returns true if --install-ssh-config flag is set during the execution.
func installSSHConfigFlag() bool {
	return viper.GetBool(core.AppName + "_install_ssh_config")
}

// installModeFlag returns an int which represents the app home directory permissions.
func installModeFlag() int {
	return viper.GetInt(core.AppName + "_install_app_home_mode")
}

// installInitServicesFlag returns true if the common services should started during the installation.
func installInitServicesFlag() bool {
	return !viper.GetBool(core.AppName + "_install_ignore_init_svcs")
}

// installMarkerFilePath returns the filepath of the Install Marker file.
func installMarkerFilePath() string {
	return filepath.Join(core.AppHomeDir(), ".installed")
}

// putInstallMarkerFile writes an Install Marker file after the application is installed.
func putInstallMarkerFile() error {
	markerFile := installMarkerFilePath()
	timeNow := time.Now().String()

	if err := core.CreateDirAndWriteBytesToFile([]byte(timeNow+"\n"), markerFile); err != nil {
		return err
	}

	return nil
}

// CheckIfInstalled returns an error if the application is not yet installed.
func CheckIfInstalled() error {
	if !core.CheckFileExists(installMarkerFilePath()) {
		if err := InstallCmd(); err != nil {
			return err
		}
	}

	return nil
}
