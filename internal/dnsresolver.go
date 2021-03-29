package internal

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	log "github.com/sirupsen/logrus"
)

// installDNSResolver configures local DNS resolution based on the operating system.
func installDNSResolver() {
	switch GetOSDistro() {
	case "windows":
		log.Warnln("On Windows you should configure YogaDNS or add DNS records to hosts file manually.")
	case "darwin":
		err := MacOsInstallDNSResolver()
		if err != nil {
			log.Debugln(err)
		}
	case "ubuntu", "debian", "pop":
		err := LinuxInstallDNSResolver()
		if err != nil {
			log.Debugln(err)
		}
	case "fedora", "centos":
		err := LinuxInstallDNSResolver()
		if err != nil {
			log.Debugln(err)
		}
	default:
		log.Fatalln("Your Operating System is not supported. Yet. :(")
	}
}

// LinuxInstallDNSResolver configures dns resolver for linux.
func LinuxInstallDNSResolver() error {
	networkManagerStatus := CheckExitCodeOfCommand("systemctl status NetworkManager | grep 'active (running)'")

	resolvConfUsesLocalNs, err := CheckRegexInFile("nameserver 127.0.0.1", "/etc/resolv.conf")
	if err != nil {
		log.Fatalln(err)
	}

	if networkManagerStatus == 0 && !resolvConfUsesLocalNs { //nolint:nestif
		dhclientConfigFilePath := filepath.Join("/", "etc", "dhcp", "dhclient.conf")

		sudoMkdirCmd := exec.Command("sudo", "install", "-vdm", "0755", filepath.Dir(dhclientConfigFilePath)) //nolint:gosec
		log.Printf("Running command: %v", sudoMkdirCmd)

		out, err := sudoMkdirCmd.CombinedOutput()

		log.Debugf("output: %v, error: %v", string(out), err)

		dhclientConfig := "prepend domain-name-servers 127.0.0.1;"

		dhclientConfigContextExist := false
		if CheckFileExists(dhclientConfigFilePath) {
			dhclientConfigContextExist, err = CheckRegexInFile(fmt.Sprintf("^%v$", dhclientConfig), dhclientConfigFilePath)
			if err != nil {
				log.Fatalln(err)
			}
		}

		if !dhclientConfigContextExist {
			log.Println("Updating dhclient config file...")

			dhclientConfigFileContent, err := ioutil.ReadFile(dhclientConfigFilePath)
			if err != nil {
				log.Debug(err)
			}

			newDhclientContent := append([]byte(dhclientConfig+"\n"), dhclientConfigFileContent...)

			log.Debugf("path: %v", dhclientConfigFilePath)
			log.Debugf("new content: %v\n", string(newDhclientContent))

			echoCmd := exec.Command("echo", "-e", string(newDhclientContent)) //nolint:gosec
			sudoTeeCmd := exec.Command("sudo", "tee", dhclientConfigFilePath)
			stdout, stderr, err := Pipeline(echoCmd, sudoTeeCmd)
			log.Debugln(string(stdout), string(stderr))

			if err != nil {
				return err
			}

			log.Println("Dhclient configuration updated.")

			restartNetworkManagerCmd := "sudo systemctl restart NetworkManager.service"
			networkManagerStatus = CheckExitCodeOfCommand(restartNetworkManagerCmd)
			log.Debugln(restartNetworkManagerCmd, networkManagerStatus)
		} else {
			log.Println("Dhclient config file was already set.")
		}
	}

	systemdResolvedStatus := CheckExitCodeOfCommand("systemctl status systemd-resolved | grep 'active (running)'")

	resolveConfFileInfo, err := os.Lstat("/etc/resolv.conf")
	if err != nil {
		log.Fatalln(err)
	}

	log.Debugf("resolve conf file mode: %v", resolveConfFileInfo.Mode())

	if systemdResolvedStatus == 0 && resolveConfFileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		systemdResolvedContent := `[Resolve]
DNS=127.0.0.1
DNS=1.1.1.1
DNS=1.0.0.1`

		systemdResolvedConfigFilePath := filepath.Join("/", "etc", "systemd", "resolved.conf")

		echoCmd := exec.Command("echo", "-e", systemdResolvedContent)
		sudoTeeCmd := exec.Command("sudo", "tee", systemdResolvedConfigFilePath)
		stdout, stderr, err := Pipeline(echoCmd, sudoTeeCmd)
		log.Debugln(string(stdout), string(stderr))

		if err != nil {
			return err
		}

		link, err := os.Readlink("/etc/resolv.conf")

		log.Debugln(link)

		if err != nil {
			log.Fatalln(err)
		}

		if link != "../run/systemd/resolve/resolv.conf" {
			sudoLnCmd := exec.Command("sudo", "ln", "-fsn", "../run/systemd/resolve/resolv.conf", "/etc/resolv.conf")
			log.Printf("Running command: %v", sudoLnCmd)
			out, err := sudoLnCmd.CombinedOutput()
			log.Debugf("output: %v", string(out))

			if err != nil {
				return fmt.Errorf("%w", err)
			}
		}

		restartSystemdResolvedCmd := "sudo systemctl restart systemd-resolved.service"
		systemdResolvedStatus = CheckExitCodeOfCommand(restartSystemdResolvedCmd)
		log.Debugln(restartSystemdResolvedCmd, systemdResolvedStatus)
	}

	return nil
}

// MacOsInstallDNSResolver configures dns zone for mac.
func MacOsInstallDNSResolver() error {
	tld := "test"
	resolverFilePath := filepath.Join("/", "etc", "resolver", tld)
	dirPath := filepath.Dir(resolverFilePath)

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		log.Printf("Creating %v directory...", dirPath)
		log.Debugf("path: %v", dirPath)

		cmdCreateDir := "sudo mkdir -v " + dirPath
		cmd := exec.Command("/bin/sh", "-c", cmdCreateDir)
		log.Printf("Running command: %vi", cmd)
		out, err := cmd.CombinedOutput()
		log.Debugf("output: %v", string(out))

		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	content, err := ioutil.ReadFile(resolverFilePath)
	if err != nil {
		log.Debug(err)
	}

	resolverConfig := "nameserver 127.0.0.1"
	re := regexp.MustCompile(resolverConfig)
	matches := re.FindStringSubmatch(string(content))
	log.Debugf("regex match: %+v", matches)

	if len(matches) == 0 {
		log.Printf("Creating DNS resolver config...")

		cmdWriteFile := fmt.Sprintf("echo '%v' | sudo tee %v", resolverConfig, resolverFilePath)
		cmd := exec.Command("/bin/sh", "-c", cmdWriteFile)

		log.Printf("Running command: %v", cmd)

		out, err := cmd.CombinedOutput()

		log.Debugf("output: %v", string(out))

		if err != nil {
			return fmt.Errorf("%w", err)
		}

		log.Printf("Done.")
	}

	return nil
}
