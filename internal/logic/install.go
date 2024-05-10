package logic

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	cryptopkg "github.com/rewardenv/reward/internal/crypto"
	"github.com/rewardenv/reward/pkg/util"
)

type installer struct {
	*Client
}

func newInstaller(c *Client) installer {
	return installer{
		Client: c,
	}
}

// RunCmdInstall opens a shell in the environment's default application container.
func (c *Client) RunCmdInstall() error {
	installer := newInstaller(c)

	if installer.reinstallFlag() || installer.uninstallFlag() {
		return installer.uninstall()
	}

	if !installer.uninstallFlag() {
		return installer.install()
	}

	return nil
}

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
#reward_portainer: false
#reward_dnsmasq: false
#reward_tunnel: false
#reward_mailbox: false
#reward_phpmyadmin: false
#reward_elastichq: false

# If you want to enable the following services, uncomment any of them.
# These services are disabled by default.
#reward_adminer: true

############
# SERVICE CONTAINERS
# It's possible to change service container images using these vars:
#reward_traefik_image: "traefik"

#reward_portainer_image: "portainer/portainer-ce"

# Reward < v0.2.33 uses "jpillora/dnsmasq" as the default dnsmasq image.
# Reward >= v0.2.34 uses the internally built "docker.io/rewardenv/dnsmasq"
#reward_dnsmasq_image: "docker.io/rewardenv/dnsmasq"

#reward_mailbox_image: "docker.io/rewardenv/mailbox:1.0"

#reward_phpmyadmin_image: "phpmyadmin"

#reward_adminer_image: "dehy/adminer"

#reward_elastichq_image: "elastichq/elasticsearch-hq"

# Reward < v0.2.33 uses "panubo/sshd:1.1.0" as the default dnsmasq image.
# Reward >= v0.2.34 uses the internally built "docker.io/rewardenv/sshd"
#reward_tunnel_image: "docker.io/rewardenv/sshd"

# Override default listen address and ports for traefik
#reward_traefik_listen: "0.0.0.0"
#reward_traefik_http_port: 80
#reward_traefik_https_port: 443

# You can configure Traefik to bind additional http ports on top of the default port (80).
# reward_traefik_bind_additional_http_ports: [8080]
reward_traefik_bind_additional_http_ports: []

# You can configure Traefik to bind additional https ports on top of the default port (443).
# reward_traefik_bind_additional_https_ports: [8443,9443]
reward_traefik_bind_additional_https_ports: []

# By default, Reward redirects all HTTP traffic to HTTPS. If you want to disable this behaviour, you can
# uncomment the following line.
#reward_traefik_allow_http: false

# Default logging level for traefik is info. Enable more detailed logs by uncommenting the next line.
#reward_traefik_log_level: debug

# By default Reward makes it possible to resolve the environment's domain to the nginx container's IP address
# inside the docker network. To disable this behaviour you can uncomment the following line.
#reward_resolve_domain_to_traefik: false

# Override default dnsmasq listen address and ports
#reward_dnsmasq_listen: "0.0.0.0"
#reward_dnsmasq_tcp_port: "53"
#reward_dnsmasq_udp_port: "53"

# By default, only the UDP port 53 is exposed from the dnsmasq container. Sometimes it doesn't seem to be enough, and
# the TCP port 53 has to be exposed as well. To do so enable the "reward_dnsmasq_bind_tcp" variable.
#reward_dnsmasq_bind_tcp: true
#reward_dnsmasq_bind_udp: true

# Override default tunnel listen address and ports
#reward_tunnel_listen: "0.0.0.0"
#reward_tunnel_port: "2222"

# By default Reward is not allowed to run commands as root.
# To disable this check you can uncomment the following line.
#reward_allow_superuser: true

# By default Reward uses Debian Bullseye based images. You can experiment with Ubuntu based images with uncommenting
# this.
#reward_docker_image_base: ubuntu-jammy

# By default Reward uses separated nginx + php-fpm containers.Enabling this setting will merge
# them to one "web" container
#reward_single_web_container: true

# By default Reward mounts the ~/.composer directory from the host computer. Using this method you only have to set
# the Composer credentials once. In some situations you may want to use different Composer credentials per project.
# You can disable this sharing mechanism by setting the variable REWARD_SHARED_COMPOSER=false in the project's
# .env file. Or you can disable it globally by setting the following variable to false.
reward_shared_composer: true

# By default mutagen sync is enabled in macOS and Windows, but you can disable it globally (here) or adding
# REWARD_SYNC_ENABLED=false to the environment's .env file.
#reward_sync_enabled: false

# It's possible to add additional shortcuts to the Reward CLI. These are the default shortcuts.
# Feel free to add your own. If you chain the commands, you can use the "&&" or the ";" operators.
# Using the "&&" operator Reward will execute the next command only if the previous one was successful.
# Using the ";" operator will execute the next command even if the previous one failed.
#reward_shortcuts:
#  up: "svc up && env up"
#  down: "env down ; svc down"
#  restart: "env down && env up"
#  sync-restart: "sync stop && sync start"
`

// uninstall removes the application's settings.
func (c *installer) uninstall() error {
	appHomeDir := c.AppHomeDir()

	log.Printf("Uninstalling %s...", cases.Title(language.English).String(c.AppName()))

	// Make sure we are not deleting "/"
	if !regexp.MustCompile("^/$").MatchString(appHomeDir) {
		if _, err := os.Stat(appHomeDir); !os.IsNotExist(err) {
			if confirmation := util.AskForConfirmation(
				fmt.Sprintf("Are you sure you want to delete %s?", appHomeDir),
			); confirmation {
				log.Debugf("Deleting: %s\n", appHomeDir)

				err = os.RemoveAll(appHomeDir)
				if err != nil {
					return fmt.Errorf("failed to delete %s: %w", appHomeDir, err)
				}
			}

			if confirmation := util.AskForConfirmation(
				fmt.Sprintf(
					"Are you sure you want to delete %s?",
					c.GetString(c.AppName()+"_config_file"),
				),
			); confirmation {
				f := c.GetString(c.AppName() + "_config_file")
				log.Debugf("Deleting: %s\n", f)

				err = os.Remove(f)
				if err != nil {
					return fmt.Errorf("failed to delete %s: %w", f, err)
				}
			}
		} else {
			return fmt.Errorf(c.AppName() + " is not installed")
		}
	}

	log.Println("...uninstall complete.")

	return nil
}

// install configures the application default settings.
func (c *installer) install() error {
	log.Printf("Installing %s...", cases.Title(language.English).String(c.AppName()))

	// On windows this command should run in elevated command prompt
	c.runElevated()
	c.checkInstalled()

	err := c.installAppDirectories()
	if err != nil {
		return err
	}

	err = c.installConfig()
	if err != nil {
		return err
	}

	err = c.installCACertificate()
	if err != nil {
		return err
	}

	err = c.installDNSResolver()
	if err != nil {
		return err
	}

	err = c.installSSHKey()
	if err != nil {
		return fmt.Errorf("failed to install SSH key: %w", err)
	}

	err = c.installSSHConfig()
	if err != nil {
		return err
	}

	err = c.installComposerDirectory()
	if err != nil {
		return err
	}

	err = c.touchInstallMarkerFile()
	if err != nil {
		return err
	}

	err = c.postInstall()
	if err != nil {
		return err
	}

	return nil
}

func (c *installer) postInstall() error {
	// If the install command is not called with --ignore-svcs or the specific install options directly, then
	// run `reward svc up`.
	if c.installInitServicesFlag() &&
		!c.installCaCertFlag() &&
		!c.installDNSFlag() &&
		!c.installSSHKeyFlag() &&
		!c.installSSHConfigFlag() {
		log.Println("...installation finished. Starting common services. Press ENTER to continue...")
		_, _ = bufio.NewReader(os.Stdin).ReadString('\n')

		if err := c.RunCmdSvc([]string{"up"}); err != nil {
			return err
		}
	} else {
		log.Println("...installation finished. Press ENTER to continue...")
		_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
	}

	return nil
}

func (c *installer) installComposerDirectory() error {
	if util.OSDistro() != "windows" {
		// Create composer directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}

		err = util.CreateDir(home+"/.composer", nil)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	return nil
}

func (c *installer) installCACertificate() error {
	// Install CA Certificate
	if !c.installDNSFlag() && !c.installSSHKeyFlag() && !c.installSSHConfigFlag() {
		sslDir := filepath.Join(c.AppHomeDir(), "ssl")
		caDir := filepath.Join(sslDir, c.SSLCABaseDir())

		log.Debugf("Installing CA certificate to directory %s...", caDir)

		crypto := cryptopkg.New(c.Config)
		if !crypto.CheckCACertificateExistInDir(caDir) {
			if err := crypto.CreateCACertificate(caDir); err != nil {
				return err
			}

			if err := crypto.InstallCACertificate(caDir); err != nil {
				return err
			}
		}

		log.Print("...CA certificate installed.")
	}

	return nil
}

func (c *installer) installConfig() error {
	// If we are not directly call installation for cacert, dns, ssh then create the app's default config file.
	if !c.installCaCertFlag() && !c.installDNSFlag() && !c.installSSHKeyFlag() && !c.installSSHConfigFlag() {
		configFile := viper.GetString(c.AppName() + "_config_file")
		log.Debugf("Creating default config: %s...", configFile)

		if !util.CheckFileExistsAndRecreate(configFile) {
			if err := util.CreateDirAndWriteToFile([]byte(defaultConfig), configFile); err != nil {
				return err
			}
		}

		log.Debugf("...%s created.", configFile)
	}

	return nil
}

func (c *installer) checkInstalled() {
	// If we are not directly call installation for cacert, dns, ssh then check if the install marker already exists.
	if !c.installCaCertFlag() && !c.installDNSFlag() && !c.installSSHKeyFlag() && !c.installSSHConfigFlag() {
		if util.FileExists(c.InstallMarkerFilePath()) {
			if !util.AskForConfirmation(
				fmt.Sprintf(
					"%s is already installed. Would you like to reinstall?",
					cases.Title(language.English).String(c.AppName()),
				),
			) {
				log.Println("...installation aborted.")
				os.Exit(0)
			}
		}
	}
}

func (c *installer) runElevated() {
	if util.OSDistro() == "windows" {
		if !util.IsAdmin() {
			log.Printf("Running %s in an Elevated command prompt...", c.AppName())

			util.RunMeElevated()
		}
	}
}

func (c *installer) installAppDirectories() error {
	log.Printf("Creating %s app directories...", cases.Title(language.English).String(c.AppName()))

	dirs := []string{
		c.AppHomeDir(),
		c.PluginsDir(),
		c.PluginsConfigDir(),
	}

	for _, dir := range dirs {
		// Create application's config directory
		if err := util.CreateDir(dir, c.installModeFlag()); err != nil {
			return fmt.Errorf("failed to create %s directory: %w", c.AppName(), err)
		}

		log.Debugf("Chmod %s dir: %s to %s\n", c.AppName(), dir, c.installModeFlag())

		// Change mode for it
		if err := os.Chmod(dir, *c.installModeFlag()); err != nil {
			return fmt.Errorf("failed to chmod %s directory: %w", c.AppName(), err)
		}
	}

	log.Print("...directories created.")

	return nil
}

// reinstallFlag returns true if the --reinstall flag is set during the execution.
func (c *installer) reinstallFlag() bool {
	return c.GetBool(c.AppName() + "_install_reinstall")
}

// uninstallFlag returns true if the --uninstall flag is set during the execution.
func (c *installer) uninstallFlag() bool {
	return c.GetBool(c.AppName() + "_install_uninstall")
}

// installCaCertFlag returns true if --install-ca-cert flag is set during the execution.
func (c *installer) installCaCertFlag() bool {
	return c.GetBool(c.AppName() + "_install_ca_cert")
}

// installDNSFlag returns true if --install-dns flag is set during the execution.
func (c *installer) installDNSFlag() bool {
	return c.GetBool(c.AppName() + "_install_dns")
}

// installSSHKeyFlag returns true if --install-ssh-key flag is set during the execution.
func (c *installer) installSSHKeyFlag() bool {
	return c.GetBool(c.AppName() + "_install_ssh_key")
}

// installSSHConfigFlag returns true if --install-ssh-config flag is set during the execution.
func (c *installer) installSSHConfigFlag() bool {
	return c.GetBool(c.AppName() + "_install_ssh_config")
}

// installModeFlag returns an int which represents the app home directory permissions.
func (c *installer) installModeFlag() *os.FileMode {
	m := os.FileMode(c.GetInt(fmt.Sprintf("%s_install_app_home_mode", c.AppName())))

	return &m
}

// installInitServicesFlag returns true if the common services should started during the installation.
func (c *installer) installInitServicesFlag() bool {
	return !c.GetBool(c.AppName() + "_install_ignore_init_svcs")
}

// touchInstallMarkerFile creates a flag file after the application is installed.
func (c *installer) touchInstallMarkerFile() error {
	// Create file which marks the application as already installed (~/.reward/.installed)
	markerFile := c.InstallMarkerFilePath()
	timeNow := time.Now().String()

	if err := util.CreateDirAndWriteToFile([]byte(timeNow+"\n"), markerFile); err != nil {
		return err
	}

	return nil
}

// installDNSResolver configures local DNS resolution based on the operating system.
func (c *installer) installDNSResolver() error {
	if !c.installCaCertFlag() && !c.installSSHKeyFlag() && !c.installSSHConfigFlag() {
		log.Print("Configuring DNS resolver...")

		var err error

		switch util.OSDistro() {
		case "windows":
			log.Warnln("On Windows you should configure YogaDNS or add DNS records to hosts file manually.")
		case "darwin":
			err = c.darwinInstallDNSResolver()
		case "ubuntu", "debian", "pop", "linuxmint", "fedora", "centos", "elementary", "manjaro", "arch":
			err = c.linuxInstallDNSResolver()
		default:
			log.Panicln("Your Operating System is not supported. Yet. :(")
		}

		if err != nil {
			log.Warnf("...failed to configure DNS resolver: %s", err)

			return err
		}

		log.Println("...DNS resolver configured.")
	}

	return nil
}

// linuxInstallDNSResolver configures dns resolver for linux.
func (c *installer) linuxInstallDNSResolver() error {
	networkManagerStatus := c.Shell.ExitCodeOfCommand("systemctl status NetworkManager | grep 'active (running)'")

	resolvConfUsesLocalNs, err := util.CheckRegexInFile("nameserver 127.0.0.1", "/etc/resolv.conf")
	if err != nil {
		log.Panicln(err)
	}

	if networkManagerStatus == 0 && !resolvConfUsesLocalNs { //nolint:nestif
		dhclientConfigFilePath := filepath.Join("/", "etc", "dhcp", "dhclient.conf")

		//nolint:gosec
		sudoMkdirCmd := exec.Command(
			"sudo", "install", "-vdm", "0755", filepath.Dir(dhclientConfigFilePath),
		)
		log.Printf("Running command: %s", sudoMkdirCmd)

		out, err := sudoMkdirCmd.CombinedOutput()

		log.Debugf("output: %s, error: %s", string(out), err)

		dhclientConfig := "prepend domain-name-servers 127.0.0.1;"

		dhclientConfigContextExist := false
		if util.FileExists(dhclientConfigFilePath) {
			dhclientConfigContextExist, err = util.CheckRegexInFile(
				fmt.Sprintf("^%s$", dhclientConfig), dhclientConfigFilePath,
			)
			if err != nil {
				log.Panicln(err)
			}
		}

		if !dhclientConfigContextExist {
			log.Println("Updating dhclient config file...")

			dhclientConfigFileContent, err := os.ReadFile(dhclientConfigFilePath)
			if err != nil {
				log.Debug(err)
			}

			newDhclientContent := append([]byte(dhclientConfig+"\n"), dhclientConfigFileContent...)

			log.Debugf("path: %s", dhclientConfigFilePath)
			log.Debugf("new content: %s\n", string(newDhclientContent))

			echoCmd := exec.Command("echo", "-e", string(newDhclientContent)) //nolint:gosec
			sudoTeeCmd := exec.Command("sudo", "tee", dhclientConfigFilePath)
			stdout, stderr, err := c.Shell.Pipeline(echoCmd, sudoTeeCmd)
			log.Debugln(string(stdout), string(stderr))

			if err != nil {
				return fmt.Errorf("cannot update dhclient config file: %w", err)
			}

			log.Println("Dhclient configuration updated.")

			restartNetworkManagerCmd := "sudo systemctl restart NetworkManager.service"
			networkManagerStatus = c.Shell.ExitCodeOfCommand(restartNetworkManagerCmd)
			log.Debugln(restartNetworkManagerCmd, networkManagerStatus)
		} else {
			log.Println("Dhclient config file was already set.")
		}
	}

	systemdResolvedStatus := c.Shell.ExitCodeOfCommand("systemctl status systemd-resolved | grep 'active (running)'")

	resolveConfFileInfo, err := os.Lstat("/etc/resolv.conf")
	if err != nil {
		log.Panicln(err)
	}

	log.Debugf("resolve conf file mode: %s", resolveConfFileInfo.Mode())

	if systemdResolvedStatus == 0 && resolveConfFileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		systemdResolvedContent := `[Resolve]
DNS=127.0.0.1 1.1.1.1
Domains=~.`

		systemdResolvedConfigFilePath := filepath.Join("/", "etc", "systemd", "resolved.conf")

		echoCmd := exec.Command("echo", "-e", systemdResolvedContent)
		sudoTeeCmd := exec.Command("sudo", "tee", systemdResolvedConfigFilePath)
		stdout, stderr, err := c.Shell.Pipeline(echoCmd, sudoTeeCmd)
		log.Debugln(string(stdout), string(stderr))

		if err != nil {
			return fmt.Errorf("cannot update systemd resolved config file: %w", err)
		}

		link, err := os.Readlink("/etc/resolv.conf")

		log.Debugln(link)

		if err != nil {
			log.Panicln(err)
		}

		if link != "../run/systemd/resolve/resolv.conf" {
			sudoLnCmd := exec.Command(
				"sudo", "ln", "-fsn", "../run/systemd/resolve/resolv.conf", "/etc/resolv.conf",
			)
			log.Printf("Running command: %s", sudoLnCmd)
			out, err := sudoLnCmd.CombinedOutput()
			log.Debugf("output: %s", string(out))

			if err != nil {
				return fmt.Errorf("%w", err)
			}
		}

		restartSystemdResolvedCmd := "sudo systemctl restart systemd-resolved.service"
		systemdResolvedStatus = c.Shell.ExitCodeOfCommand(restartSystemdResolvedCmd)
		log.Debugln(restartSystemdResolvedCmd, systemdResolvedStatus)
	}

	return nil
}

// darwinInstallDNSResolver configures dns zone for mac.
func (c *installer) darwinInstallDNSResolver() error {
	resolverFilePath := filepath.Join("/", "etc", "resolver", "test")
	dirPath := filepath.Dir(resolverFilePath)

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		log.Printf("Creating %s directory...", dirPath)
		log.Debugf("path: %s", dirPath)

		cmdCreateDir := "sudo mkdir -v " + dirPath
		cmd := exec.Command("/bin/sh", "-c", cmdCreateDir)
		log.Printf("Running command: %si", cmd)
		out, err := cmd.CombinedOutput()
		log.Debugf("output: %s", string(out))

		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	content, err := os.ReadFile(resolverFilePath)
	if err != nil {
		log.Debug(err)
	}

	resolverConfig := "nameserver 127.0.0.1"
	re := regexp.MustCompile(resolverConfig)
	matches := re.FindStringSubmatch(string(content))
	log.Debugf("regex match: %+v", matches)

	if len(matches) == 0 {
		log.Printf("Creating DNS resolver config...")

		cmdWriteFile := fmt.Sprintf("echo '%s' | sudo tee %s", resolverConfig, resolverFilePath)
		cmd := exec.Command("/bin/sh", "-c", cmdWriteFile)

		log.Printf("Running command: %s", cmd)

		out, err := cmd.CombinedOutput()

		log.Debugf("output: %s", string(out))

		if err != nil {
			return fmt.Errorf("%w", err)
		}

		log.Printf("Done.")
	}

	return nil
}

func (c *installer) installSSHKey() error {
	// Install common SSH Key for Tunnel
	if !c.installCaCertFlag() && !c.installDNSFlag() && !c.installSSHConfigFlag() {
		log.Print("Installing SSH key...")

		crypto := cryptopkg.New(c.Config)
		appHomeDir := c.AppHomeDir()
		keyPath := filepath.Join(appHomeDir, "tunnel", "ssh_key")

		// On linux, if we want to reinstall the pubfile we have to revert its permissions first
		if runtime.GOOS == "linux" && util.FileExists(keyPath) {
			cmdChown := fmt.Sprintf(
				"sudo chown -v %d:%d %s", os.Getuid(), 0,
				filepath.Join(appHomeDir, "tunnel", "ssh_key.pub"),
			)
			cmd := exec.Command("/bin/sh", "-c", cmdChown)

			log.Debugf("Running command: %s", cmd)

			out, err := cmd.CombinedOutput()

			log.Debugf("Command output: %s", string(out))

			if err != nil {
				log.Fatalln(err)
			}
		}

		keyFileExist := util.CheckFileExistsAndRecreate(keyPath)
		if !keyFileExist {
			if err := crypto.GenerateSSHKeys(2048, keyPath); err != nil {
				return err
			}
		}

		// Since bind mounts are native on linux to use .pub file as authorized_keys file in tunnel it
		// must have proper perms.
		if runtime.GOOS == "linux" {
			cmdChown := fmt.Sprintf(
				"sudo chown -v %d:%d %s", 0, 0, filepath.Join(appHomeDir, "tunnel", "ssh_key.pub"),
			)
			cmd := exec.Command("/bin/sh", "-c", cmdChown)

			log.Debugf("Running command: %s", cmd)

			out, err := cmd.CombinedOutput()

			log.Debugf("Command output: %s", string(out))

			if err != nil {
				log.Fatalln(err)
			}
		}

		log.Print("...SSH key installed.")
	}

	return nil
}

// InstallSSHConfig updates the ssh config file to use our key if the target host is the tunnel.
func (c *installer) installSSHConfig() error {
	if util.OSDistro() != "windows" {
		if !c.installCaCertFlag() && !c.installDNSFlag() && !c.installSSHKeyFlag() {
			log.Println("Updating SSH config file...")

			var (
				sshConfig = fmt.Sprintf(
					`## %[1]s START ##
Host tunnel.%[2]s.test
  HostName 127.0.0.1
  User user
  Port 2222
  IdentityFile %[3]s/tunnel/ssh_key
## %[1]s END ##`,
					strings.ToUpper(c.AppName()),
					c.AppName(),
					c.AppHomeDir(),
				)
				//nolint:gocritic
				sshConfigFile = filepath.Join("/etc/ssh/ssh_config")
			)

			content, err := os.ReadFile(sshConfigFile)
			if err != nil {
				return fmt.Errorf("%w", err)
			}

			log.Debugf("Searching for configuration regex in ssh config file: %s...", sshConfigFile)

			matches := regexp.MustCompile(
				fmt.Sprintf(
					"## %s START ##",
					strings.ToUpper(c.AppName()),
				),
			).FindStringSubmatch(string(content))

			log.Debugf("...regex match: %+v", matches)

			if len(matches) == 0 {
				log.Println("Updating SSH config file...")
				log.Debugf("SSH config file path: %s", sshConfigFile)
				log.Debugf("SSH config file content: %s", sshConfig)

				cmdAppend := fmt.Sprintf("echo '%s' | sudo tee -a %s", sshConfig, sshConfigFile)
				cmd := exec.Command("/bin/sh", "-c", cmdAppend)

				log.Debugf("Running command: %s", cmd)

				out, err := cmd.CombinedOutput()

				log.Debugf("Command output: %s", string(out))

				if err != nil {
					return fmt.Errorf("%w", err)
				}

				log.Println("...SSH config file updated.")
			} else {
				log.Println("...SSH config file was already set.")
			}
		}
	}

	return nil
}
