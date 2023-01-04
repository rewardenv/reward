package core

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	dockerClient "github.com/docker/docker/client"
	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/ulikunitz/xz"
	"gopkg.in/ini.v1"

	"github.com/rewardenv/reward/internal"
)

const (
	appname = "reward"
)

var (
	// AppName is the lowercase Application Name, used globally.
	AppName               = strings.ToLower(appname)
	versionFileContent, _ = internal.Asset("VERSION.txt")
)

var (
	// FS is a Memory Map Filesystem.
	FS = afero.NewOsFs()
	// AFS is the implementation of Afero Filesystem.
	AFS = &afero.Afero{Fs: FS}
)

var (
	// ErrEmptyDirName occurs when directory name is empty.
	ErrEmptyDirName = errors.New("empty directory name")
	// ErrFileWithThisDirNameExist occurs when file already exist.
	ErrFileWithThisDirNameExist = errors.New("file with the same name exists")
	// ErrFileNotFound occurs when file is not found.
	ErrFileNotFound = errors.New("file not found")
	// ErrUnknownEnvType occurs when an unknown environment type is specified.
	ErrUnknownEnvType = errors.New("unknown env type")
	// ErrUnknownAction occurs when an unknown actions is called.
	ErrUnknownAction = errors.New("unknown action error")
	// ErrDockerAPIIsUnreachable occurs when Docker is not running, Docker version is too old,
	// or the user who runs the application cannot call Docker API.
	ErrDockerAPIIsUnreachable = errors.New("docker api is unreachable")
	// ErrDockerVersionMismatch occurs when Docker version is too old.
	ErrDockerVersionMismatch = errors.New("docker version is too old")
	// ErrDockerComposeVersionMismatch occurs when docker-compose version is too old.
	ErrDockerComposeVersionMismatch = errors.New("docker-compose version is too old")
	// ErrEnvNameIsInvalid occurs when the environment name is invalid. It should be a valid hostname.
	ErrEnvNameIsInvalid = errors.New("environment name is invalid, it should match RFC1178")
	// ErrEnvIsEmpty occurs when environment name is empty.
	ErrEnvIsEmpty = fmt.Errorf("env name is empty. please run `%v env-init`", AppName)
	// ErrCaCertDoesNotExist occurs when the Signing CA Certificate is not yet created.
	ErrCaCertDoesNotExist = fmt.Errorf(
		"the root CA certificate is missing, please run '%v install' and try again",
		AppName,
	)
	// ErrCannotFindContainer occurs when the application cannot find the requested container.
	ErrCannotFindContainer = errors.New("container cannot be found")
	// ErrTooManyContainersFound occurs when the application found more than 1 container.
	ErrTooManyContainersFound = errors.New("too many containers found")
	// ErrArgumentRequired occurs when the function is called without a required argument.
	ErrArgumentRequired = errors.New("argument required")
	// ErrInvokedAsRootUser occurs when the Application was called by Root user.
	ErrInvokedAsRootUser = errors.New(
		"In most cases, you should not run " +
			AppName + " as root user except for `self-update`. " + "If you are sure you want to do this, use " +
			strings.ToUpper(AppName) + "_ALLOW_SUPERUSER=1.",
	)
)

// FileNotFoundError is a wrapper function for ErrFileNotFound error.
func FileNotFoundError(op string) error {
	return fmt.Errorf("ErrFileNotFound: %w, %v", ErrFileNotFound, op)
}

// CannotFindContainerError is a wrapper function for ErrCannotFindContainer error.
func CannotFindContainerError(op string) error {
	log.Debugln()

	return fmt.Errorf("ErrCannotFindContainer: %w: %s", ErrCannotFindContainer, op)
}

// TooManyContainersFoundError is a wrapper function for ErrCannotFindContainer error.
func TooManyContainersFoundError(op string) error {
	log.Debugln()

	return fmt.Errorf("ErrTooManyContainersFound: %w: %s", ErrTooManyContainersFound, op)
}

// ArgumentRequiredError is a wrapper function for ErrArgumentRequired error.
func ArgumentRequiredError(op string) error {
	return fmt.Errorf("ErrArgumentRequired: %w: %s", ErrArgumentRequired, op)
}

// DockerVersionMismatchError is a wrapper function for ErrDockerVersionMismatch error.
func DockerVersionMismatchError(op string) error {
	return fmt.Errorf("ErrDockerVersionMismatch: %w: %s", ErrDockerVersionMismatch, op)
}

// DockerComposeVersionMismatchError is a wrapper function for ErrDockerComposeVersionMismatch error.
func DockerComposeVersionMismatchError(op string) error {
	return fmt.Errorf("ErrDockerComposeVersionMismatch: %w: %s", ErrDockerComposeVersionMismatch, op)
}

// AppVersion returns a version.Version object contains the application version.
func AppVersion() *version.Version {
	v, _ := version.NewVersion(strings.TrimSpace(string(versionFileContent)))

	return v
}

// EnvName returns the environment name in lowercase format.
func EnvName() string {
	return strings.ToLower(viper.GetString(AppName + "_env_name"))
}

// EnvType returns the environment type in lowercase format.
func EnvType() string {
	return strings.ToLower(viper.GetString(AppName + "_env_type"))
}

// EnvNetworkName returns the environments docker network name in lowercase format.
func EnvNetworkName() string {
	return strings.ToLower(EnvName() + "_default")
}

// Cwd returns the current working directory.
func Cwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	return cwd
}

// HomeDir returns the invoking user's home directory.
func HomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}

	return home
}

// AppHomeDir returns the application's home directory.
func AppHomeDir() string {
	return viper.GetString(AppName + "_home_dir")
}

// ServiceDomain returns the application's service domain.
func ServiceDomain() string {
	return viper.GetString(AppName + "_service_domain")
}

// MutagenSyncFile returns the file path of the mutagen sync file.
func MutagenSyncFile() string {
	return filepath.Join(Cwd(), "."+AppName, "mutagen.yml")
}

// MutagenSyncIgnore returns the additional mutagen ignored files from Viper settings.
func MutagenSyncIgnore() string {
	return viper.GetString(AppName + "_sync_ignore")
}

// WebRoot returns the content of the WEB_ROOT variable from Viper settings.
func WebRoot() string {
	return viper.GetString(AppName + "_web_root")
}

// ComposerVersion returns the Composer Version defined in Viper settings.
func ComposerVersion() (*version.Version, error) {
	log.Debugln()

	var (
		ver *version.Version
		err error
	)

	if viper.GetFloat64("composer_version") >= 2 {
		v, _ := version.NewVersion("2.0")
		ver = v
	} else {
		v, _ := version.NewVersion("1.0")
		ver = v
	}

	if err != nil {
		return nil, err
	}

	return ver, nil
}

// IsDBEnabled returns true if the database service is enabled for the current environment.
func IsDBEnabled() bool {
	return viper.GetString(AppName+"_db") == "1"
}

// DBContainer returns the name of the database container.
func DBContainer() string {
	log.Debugln()

	if viper.IsSet(AppName + "_env_db_container") {
		return viper.GetString(AppName + "_env_db_container")
	}

	return "db"
}

// DBCommand returns the command which is called when the application manipulates the database.
func DBCommand() string {
	log.Debugln()

	if viper.IsSet(AppName + "_env_db_command") {
		return viper.GetString(AppName + "_env_db_command")
	}

	return "mysql"
}

// DBDumpCommand returns the command which is called when the application dumps a database.
func DBDumpCommand() string {
	if viper.IsSet(AppName + "_env_db_command") {
		return viper.GetString(AppName + "_env_db_command")
	}

	return "mysqldump"
}

// BlackfireContainer returns the container name of the Blackfire debug container.
func BlackfireContainer() string {
	if viper.IsSet(AppName + "_env_blackfire_container") {
		return viper.GetString(AppName + "_env_blackfire_container")
	}

	return "php-blackfire"
}

// BlackfireCommand returns the command which is called when the application manipulates blackfire.
func BlackfireCommand() string {
	if viper.IsSet(AppName + "_env_blackfire_command") {
		return viper.GetString(AppName + "_env_blackfire_command")
	}

	return "blackfire"
}

// BlackfireEnabled returns true if the blackfire container is enabled.
func BlackfireEnabled() bool {
	return viper.GetString(AppName+"_blackfire") == "1"
}

// ResolveDomainToTraefik returns true if an environment's service containers should resolve their "main"
// domains to the Traefik container's address. If this is enabled, the nginx/php containers will be able to
// reach themselves through the traefik proxy.
func ResolveDomainToTraefik() bool {
	if viper.IsSet(AppName + "_resolve_domain_to_traefik") {
		return viper.GetBool(AppName + "_resolve_domain_to_traefik")
	}

	return true
}

// MagentoVersion returns a *version.Version object which contains the Magento version.
func MagentoVersion() (*version.Version, error) {
	v := new(version.Version)

	type ComposerJSON struct {
		Name    string            `json:"name"`
		Version string            `json:"version"`
		Require map[string]string `json:"require"`
	}

	var composerJSON ComposerJSON

	if CheckFileExists("composer.json") {
		data, err := AFS.ReadFile("composer.json")
		if err != nil {
			v, err = MagentoVersionFromViper()
			if err != nil {
				return nil, err
			}
		}

		if err = json.Unmarshal(data, &composerJSON); err != nil {
			v, err = MagentoVersionFromViper()
			if err != nil {
				return nil, err
			}
		}

		if CheckRegexInString(`^magento/magento2(ce|ee)$`, composerJSON.Name) && composerJSON.Version != "" {
			re := regexp.MustCompile(semver.SemVerRegex)
			ver := re.Find([]byte(composerJSON.Version))

			v, _ = version.NewVersion(string(ver))
		}

		if v.String() == "" {
			for key, val := range composerJSON.Require {
				if CheckRegexInString(`^magento/product-(enterprise|community)-edition$`, key) {
					re := regexp.MustCompile(semver.SemVerRegex)
					ver := re.Find([]byte(val))

					v, err = version.NewVersion(string(ver))
					if err != nil {
						return nil, err
					}

					log.Debugln(string(ver))
				} else if CheckRegexInString(`^magento/magento-cloud-metapackage$`, key) {
					re := regexp.MustCompile(semver.SemVerRegex)
					ver := re.Find([]byte(val))

					v, err = version.NewVersion(string(ver))
					if err != nil {
						return nil, err
					}

					log.Debugln(string(ver))
				}
			}
		}

		return v, nil
	}

	v, err := MagentoVersionFromViper()
	if err != nil {
		return nil, err
	}

	return v, nil
}

// MagentoVersionFromViper returns a *version.Version object from Viper settings.
// Note: If it's unset, it will return a dedicated latest version.
func MagentoVersionFromViper() (*version.Version, error) {
	const (
		magentoOneDefaultVersion = "1.9.4"
		magentoTwoDefaultVersion = "2.4.4"
	)

	var (
		v   *version.Version
		err error
	)

	if EnvType() == "magento1" {
		if viper.IsSet(AppName + "_magento_version") {
			v, err = version.NewVersion(viper.GetString(AppName + "_magento_version"))
			if err != nil {
				return nil, err
			}
		} else {
			v, err = version.NewVersion(magentoOneDefaultVersion)
			if err != nil {
				return nil, err
			}
		}
	} else {
		if viper.IsSet(AppName + "_magento_version") {
			v, err = version.NewVersion(viper.GetString(AppName + "_magento_version"))
			if err != nil {
				return nil, err
			}
		} else {
			v, err = version.NewVersion(magentoTwoDefaultVersion)
			if err != nil {
				return nil, err
			}
		}
	}

	return v, nil
}

// TraefikDomain returns traefik domain from Viper settings.
func TraefikDomain() string {
	domain := viper.GetString("traefik_domain")

	log.Debugln("Traefik Domain:", domain)

	return domain
}

// TraefikSubdomain returns traefik subdomain from Viper settings.
func TraefikSubdomain() string {
	subDomain := viper.GetString("traefik_subdomain")

	log.Debugln("Traefik Subdomain:", subDomain)

	return subDomain
}

// TraefikFullDomain returns traefik full domain (subdomain + domain merged).
func TraefikFullDomain() string {
	var fullDomain string

	if TraefikSubdomain() == "" {
		fullDomain = TraefikDomain()
	} else {
		fullDomain = TraefikSubdomain() + "." + TraefikDomain()
	}

	log.Debugln("Traefik Full Domain:", fullDomain)

	return fullDomain
}

// MagentoBackendFrontname returns Magento admin path from Viper settings.
func MagentoBackendFrontname() string {
	if viper.IsSet("magento_backend_frontname") {
		return viper.GetString("magento_backend_frontname")
	}

	return "admin"
}

// ServiceEnabled returns true if service is enabled in Viper settings.
func ServiceEnabled(service string) bool {
	if viper.IsSet(AppName + "_" + service) {
		return viper.GetBool(AppName + "_" + service)
	}

	return false
}

// ContainerRunning returns true if container is running.
func ContainerRunning(container string) (bool, error) {
	_, err := ContainerIDByName(container)

	return err == nil, err
}

// SuperuserAllowed returns true if the application is allowed to be invoked by root.
func SuperuserAllowed() bool {
	if viper.IsSet(AppName + "_allow_superuser") {
		return viper.GetBool(AppName + "_allow_superuser")
	}

	return false
}

// SingleWebContainer returns true if Single Web Container setting is enabled in Viper settings.
func SingleWebContainer() bool {
	if viper.IsSet(AppName + "_single_web_container") {
		return viper.GetBool(AppName + "_single_web_container")
	}

	return false
}

// ContainsString checks if a slice of string contains a string.
func ContainsString(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}

	return false
}

// AskForConfirmation ask msg from the user and returns the answer.
func AskForConfirmation(msg string, suppressMessage ...bool) bool {
	if len(suppressMessage) == 0 {
		log.Printf("%v (y)es, (n)o", msg)
	} else {
		log.Println(msg)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := scanner.Text()

	if len(suppressMessage) == 0 {
		switch strings.ToLower(response) {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			log.Println("I'm sorry but I didn't get what you meant, please type (y)es or (n)o and then press enter:")

			return AskForConfirmation(msg)
		}
	} else {
		return suppressMessage[0]
	}
}

// OSDistro returns the linux distro name if GOOS is linux, else "darwin" or "windows".
func OSDistro() string {
	if runtime.GOOS == "linux" {
		cfg, err := ini.Load("/etc/os-release")
		if err != nil {
			log.Fatal("Fail to read file: ", err)
		}

		ConfigParams := make(map[string]string)
		ConfigParams["ID"] = cfg.Section("").Key("ID").String()

		return strings.ToLower(ConfigParams["ID"])
	}

	return runtime.GOOS
}

// MutagenSyncEnabled returns true for macOS and Windows if it's not disabled explicitly (or if the WSL2 Direct Mount
// option is not enabled on Windows).
func MutagenSyncEnabled() bool {
	switch OSDistro() {
	case "darwin":
		if viper.IsSet(AppName + "_mutagen_enabled") {
			return viper.GetBool(AppName + "_mutagen_enabled")
		}

		return true
	case "windows":
		if viper.IsSet(AppName + "_mutagen_enabled") {
			return viper.GetBool(AppName + "_mutagen_enabled")
		}

		if viper.IsSet(AppName + "_wsl2_direct_mount") {
			return !viper.GetBool(AppName + "_wsl2_direct_mount")
		}

		return false
	}

	return false
}

// CheckFileExistsAndRecreate checks if the file already exists and ask the user if he'd like to recreate it.
//
//	If user input is yes - return false (as if the file does not existing).
func CheckFileExistsAndRecreate(file string) bool {
	log.Traceln("Checking if file exist:", file)

	if file == "" {
		log.Debug("Path is empty")

		return false
	}

	filePath := filepath.Join(file)
	confirmation := false

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		log.Printf("File already exists: %v", filePath)

		conf := AskForConfirmation("Would you like to recreate it?")

		// Have to swap bool, if the user doesn't want to recreate it, than return as it exist (true).
		confirmation = !conf
	}

	return confirmation
}

// CheckFileExists checks if the file already exists and ask the user if he'd like to recreate it.
func CheckFileExists(file string) bool {
	log.Traceln("Checking if file exist:", file)

	if file == "" {
		log.Debug("Path is empty.")

		return false
	}

	filePath := filepath.Join(file)
	exist := false

	if _, err := AFS.Stat(filePath); !os.IsNotExist(err) {
		log.Traceln("File exist:", file)

		exist = true
	} else {
		log.Traceln("File does not exist:", file)
	}

	return exist
}

// EvalSymlinkPath checks if file exists and returns the resolved path if a symlink if file is a symlink,
//
//	else it returns the filepath.
func EvalSymlinkPath(file string) (string, error) {
	if !CheckFileExists(file) {
		return "", FileNotFoundError(file)
	}

	stat, err := os.Lstat(file)
	if err != nil {
		return "", err
	}

	var resolvedPath string

	symlink := symlink(stat)
	log.Debugln(symlink)

	if symlink {
		resolvedPath, _, err = evalSymlinks(FS, file)
		if err != nil {
			return "", err
		}
	} else {
		resolvedPath = file
	}

	log.Debugln(resolvedPath)

	return resolvedPath, nil
}

func symlink(fi os.FileInfo) bool {
	return fi != nil && fi.Mode()&os.ModeSymlink == os.ModeSymlink
}

func evalSymlinks(fs afero.Fs, filename string) (string, os.FileInfo, error) {
	link, err := filepath.EvalSymlinks(filename)
	if err != nil {
		return "", nil, err
	}

	fi, err := fs.Stat(link)
	if err != nil {
		return "", nil, err
	}

	return link, fi, nil
}

// CommandAvailable returns if the parameter can be find in $PATH.
func CommandAvailable(name string) bool {
	log.Debugln()

	_, err := exec.LookPath(name)

	return err == nil
}

// CreateDir creates the directory if not exist.
func CreateDir(dir string, perms ...int) error {
	log.Debugln()

	if dir == "" {
		return ErrEmptyDirName
	}

	dirPath, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	dirMode := os.FileMode(0o755)

	if len(perms) > 0 {
		dirMode = os.FileMode(perms[0])
	}

	if stat, err := AFS.Stat(dirPath); os.IsNotExist(err) { //nolint:nestif
		log.Printf("Creating %v directory...", dirPath)
		log.Debugf("path: %v, mode: %v\n", dirPath, dirMode)
		err = AFS.MkdirAll(dirPath, dirMode)

		if err != nil {
			return fmt.Errorf("%w", err)
		}
	} else if stat.Mode().IsDir() {
		if stat.Mode().Perm() != dirMode {
			err = AFS.Chmod(dirPath, dirMode)
			if err != nil {
				return fmt.Errorf("%w", err)
			}

			return nil
		}
	} else {
		return ErrFileWithThisDirNameExist
	}

	return nil
}

// CreateDirAndWriteBytesToFile creates the base directory and writes bytes to a file in it.
func CreateDirAndWriteBytesToFile(bytes []byte, file string, perms ...int) error {
	dirPath, err := filepath.Abs(filepath.Dir(file))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	filePath, err := filepath.Abs(file)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	fileMode := os.FileMode(0o640)
	if len(perms) > 0 {
		fileMode = os.FileMode(perms[0])
	}

	dirMode := os.FileMode(0o755)
	if len(perms) > 1 {
		dirMode = os.FileMode(perms[1])
	}

	if _, err := AFS.Stat(dirPath); os.IsNotExist(err) {
		err = CreateDir(dirPath, int(dirMode))
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	err = AFS.WriteFile(filePath, bytes, fileMode)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	log.Debugf("File saved: %v", filePath)

	return nil
}

// InstallSSHConfig updates the ssh config file to use our key if the target host is the tunnel.
func InstallSSHConfig() error {
	var sshConfig = fmt.Sprintf(
		`## %v START ##
Host tunnel.%v.test
  HostName 127.0.0.1
  User user
  Port 2222
  IdentityFile %v/tunnel/ssh_key
## %v END ##`, strings.ToUpper(AppName), AppName, AppHomeDir(), strings.ToUpper(AppName),
	)

	log.Println("Updating SSH config file...")

	sshConfigFile := filepath.Join("/etc/ssh/ssh_config")

	content, err := os.ReadFile(sshConfigFile)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	re := regexp.MustCompile(fmt.Sprintf("## %v START ##", strings.ToUpper(AppName)))
	matches := re.FindStringSubmatch(string(content))
	log.Debugf("regex match: %+v", matches)

	if len(matches) == 0 {
		log.Println("Updating SSH config file...")
		log.Debugf("path: %v", sshConfigFile)
		log.Debugf("content: %v", sshConfig)

		cmdAppend := fmt.Sprintf("echo '%v' | sudo tee -a %v", sshConfig, sshConfigFile)
		cmd := exec.Command("/bin/sh", "-c", cmdAppend)

		log.Debugf("Running command: %v", cmd)

		out, err := cmd.CombinedOutput()

		log.Debugf("output: %v", string(out))

		if err != nil {
			return fmt.Errorf("%w", err)
		}

		log.Println("SSH config file updated.")
	} else {
		log.Println("SSH config file was already set.")
	}

	return nil
}

// CheckRegexInFile checks if the file contains content.
func CheckRegexInFile(regex, filePath string) (bool, error) {
	file, err := AFS.Open(filepath.Join(filePath))
	if err != nil {
		return false, fmt.Errorf("%w", err)
	}

	defer func(file afero.File) {
		_ = file.Close()
	}(file)

	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(regex)

	var matches []string

	for scanner.Scan() {
		if re.MatchString(scanner.Text()) {
			matches = append(matches, scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("%w", err)
	}

	if len(matches) > 0 {
		return true, nil
	}

	return false, nil
}

// CheckRegexInString checks if the string contains content.
func CheckRegexInString(regex, str string) bool {
	re := regexp.MustCompile(regex)

	return re.MatchString(str)
}

// RunOsCommand is going to run a command depending on the caller's operating system.
func RunOsCommand(args []string, suppressOsStdOut ...bool) (string, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		args = append([]string{"/c"}, args...)

		cmd = exec.Command("cmd", args...)
		log.Debugf("Running command: %v", cmd.String())
	} else {
		args = append([]string{"-c"}, strings.Join(args, " "))

		cmd = exec.Command("sh", args...)
		log.Debugf("Running command: %v", cmd.String())
	}

	var combinedOutBuf bytes.Buffer

	cmd.Stdin = os.Stdin
	if len(suppressOsStdOut) > 0 && suppressOsStdOut[0] {
		cmd.Stdout = io.Writer(&combinedOutBuf)
		cmd.Stderr = io.Writer(&combinedOutBuf)
	} else {
		cmd.Stdout = io.Writer(os.Stdout)
		cmd.Stderr = io.Writer(os.Stderr)
	}

	err := cmd.Run()
	outStr := combinedOutBuf.String()

	return outStr, err //nolint:wrapcheck
}

// CheckExitCodeOfCommand runs a command and returns its exit code.
func CheckExitCodeOfCommand(command string) int {
	cmd := exec.Command(command)
	log.Printf("Running command: %v", cmd)
	out, err := cmd.CombinedOutput()
	log.Debugf("output: %v", string(out))

	var status int

	if err != nil {
		var exitError *exec.ExitError
		if ok := errors.As(err, &exitError); ok {
			status = exitError.ExitCode()
		}
	}

	return status
}

// Pipeline runs cmds piped after each other.
func Pipeline(cmds ...*exec.Cmd) (pipeLineOutput, collectedStandardError []byte, pipeLineError error) {
	// Require at least one command
	if len(cmds) < 1 {
		return nil, nil, nil
	}

	// Collect the output from the command(s)
	var output, stderr bytes.Buffer

	last := len(cmds) - 1
	for i, cmd := range cmds[:last] {
		var err error
		// Connect each command's stdin to the previous command's stdout
		if cmds[i+1].Stdin, err = cmd.StdoutPipe(); err != nil {
			return nil, nil, fmt.Errorf("%w", err)
		}
		// Connect each command's stderr to a buffer
		cmd.Stderr = &stderr
	}

	// Connect the output and error for the last command
	cmds[last].Stdout, cmds[last].Stderr = &output, &stderr

	// Start each command
	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			return output.Bytes(), stderr.Bytes(), fmt.Errorf("%w", err)
		}
	}

	// Wait for each command to complete
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			return output.Bytes(), stderr.Bytes(), fmt.Errorf("%w", err)
		}
	}

	// Return the pipeline output and the collected standard error
	return output.Bytes(), stderr.Bytes(), nil
}

// DockerPeeredServices attaches/detaches the common services to the current environment's docker network.
func DockerPeeredServices(action, networkName string) error {
	if action != "connect" && action != "disconnect" {
		return ErrUnknownAction
	}

	dockerPeeredServices := []string{"traefik"}

	// Enabled by default
	dockerAdditionalServices := []string{
		"tunnel",
		"mailhog",
		"phpmyadmin",
		"elastichq",
	}

	for _, svc := range dockerAdditionalServices {
		if SvcEnabledPermissive(svc) {
			dockerPeeredServices = append(dockerPeeredServices, svc)
		}
	}

	// Disabled by default
	dockerOptionalServices := []string{
		"adminer",
	}

	for _, svc := range dockerOptionalServices {
		if SvcEnabledStrict(svc) {
			dockerPeeredServices = append(dockerPeeredServices, svc)
		}
	}

	ctx := context.Background()

	client, err := NewDockerClient()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	for _, v := range dockerPeeredServices {
		networkSettings := new(network.EndpointSettings)

		if v == "traefik" && ResolveDomainToTraefik() {
			networkSettings.Aliases = []string{
				TraefikDomain(),
				TraefikFullDomain(),
			}

			log.Debugln("Network aliases for Traefik container:", networkSettings.Aliases)
		}

		f := filters.NewArgs()

		f.Add("name", v)

		filterName := types.ContainerListOptions{
			Filters: f,
		}

		containers, err := client.ContainerList(ctx, filterName)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		for _, container := range containers {
			if action == "connect" {
				log.Debugf("Connecting container: %v to network %v.", container.Names, networkName)

				err = client.NetworkConnect(ctx, networkName, container.ID, networkSettings)

				if err != nil {
					log.Debugf("%v", err)
				}
			}

			if action == "disconnect" {
				log.Debugf("Disconnecting container: %v from network %v.", container.Names, networkName)
				err = client.NetworkDisconnect(ctx, networkName, container.ID, false)

				if err != nil {
					log.Debugf("%v", err)
				}
			}
		}
	}

	return nil
}

// DockerNetworkExist returns true if the docker network exists.
func DockerNetworkExist(networkName string) (bool, error) {
	ctx := context.Background()

	client, err := NewDockerClient()
	if err != nil {
		return false, fmt.Errorf("%w", err)
	}

	f := filters.NewArgs()

	f.Add("name", networkName)

	filterName := types.NetworkListOptions{
		Filters: f,
	}

	networks, err := client.NetworkList(ctx, filterName)
	if err != nil {
		return false, fmt.Errorf("%w", err)
	}

	log.Traceln("networks: ", networks)

	if len(networks) == 0 {
		return false, nil
	}

	return true, nil
}

// ExtractUnknownArgs returns []string arguments which are not used by pflags.
func ExtractUnknownArgs(flags *pflag.FlagSet, args []string) []string {
	var unknownArgs []string

	for i := 0; i < len(args); i++ {
		a := args[i]

		var f *pflag.Flag

		if a[0] == '-' {
			if a[1] == '-' {
				f = flags.Lookup(strings.SplitN(a[2:], "=", 2)[0])
			} else {
				for _, s := range a[1:] {
					f = flags.ShorthandLookup(string(s))
					if f == nil {
						break
					}
				}
			}
		}

		if f != nil {
			if f.NoOptDefVal == "" && i+1 < len(args) && f.Value.String() == args[i+1] {
				i++
			}

			continue
		}

		unknownArgs = append(unknownArgs, a)
	}

	return unknownArgs
}

// DockerComposeCompleter returns a completer function for docker-compose.
func DockerComposeCompleter() func(cmd *cobra.Command, args []string, toComplete string) (
	[]string, cobra.ShellCompDirective,
) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		args = append(args, "--help")
		out, _ := RunDockerComposeCommand(args, true)

		commandsMatched := false
		scanner := bufio.NewScanner(strings.NewReader(out))

		var words []string

		for scanner.Scan() {
			if commandsMatched {
				fields := strings.Fields(scanner.Text())
				if len(fields) > 0 {
					words = append(words, fields[0])
				}
			}

			if strings.Contains(strings.ToLower(scanner.Text()), "commands:") {
				commandsMatched = true
			}
		}

		return words, cobra.ShellCompDirectiveNoFileComp
	}
}

// IsDebug returns true if debug mode is set.
func IsDebug() bool {
	return viper.GetBool("debug")
}

// InsertStringBeforeOccurrence inserts insertStr before occurrence of searchStr (if exist) to args and returns args.
//
//	If searchStr is not exists it will append to the end of args.
func InsertStringBeforeOccurrence(args []string, insertStr, searchStr string) []string {
	if ContainsString(args, searchStr) {
		var newArgs []string

		for i, arg := range args {
			if arg == searchStr {
				newArgs = append(newArgs, args[:i]...)
				newArgs = append(newArgs, insertStr)
				newArgs = append(newArgs, args[i:]...)
			}
		}

		return newArgs
	}

	return append(args, insertStr)
}

// InsertStringAfterOccurrence inserts insertStr after the occurrence of searchStr to args and returns args.
//
//	If searchStr is not exists it will append to the end of args.
func InsertStringAfterOccurrence(args []string, insertStr, searchStr string) []string {
	log.Debugln(args)
	log.Debugln(ContainsString(args, "up"))

	if ContainsString(args, searchStr) {
		var newArgs []string

		for i, arg := range args {
			if arg == searchStr {
				newArgs = append(newArgs, args[:i+1]...)
				newArgs = append(newArgs, insertStr)
				newArgs = append(newArgs, args[i+1:]...)
			}
		}

		return newArgs
	}

	return append(args, insertStr)
}

func DecompressFileFromArchive(src io.Reader, archive, filename string) (io.Reader, error) {
	if strings.HasSuffix(archive, ".zip") {
		log.Debugln("Decompressing zip file", archive)

		buf, err := io.ReadAll(src)
		if err != nil {
			return nil, err
		}

		r := bytes.NewReader(buf)

		z, err := zip.NewReader(r, r.Size())
		if err != nil {
			return nil, err
		}

		log.Debugln("files in archive:")

		for _, file := range z.File {
			log.Debugln(file.Name)

			_, name := filepath.Split(file.Name)
			if !file.FileInfo().IsDir() && matchExecutableName(filename, name) {
				log.Debugln("Executable file", file.Name, "was found in zip archive")

				return file.Open()
			}
		}

		return nil, FileNotFoundError(filename)
	} else if strings.HasSuffix(archive, ".tar.gz") || strings.HasSuffix(archive, ".tgz") {
		log.Debugln("Decompressing tar.gz file", archive)

		gz, err := gzip.NewReader(src)
		if err != nil {
			return nil, err
		}

		return unarchiveTar(gz, archive, filename)
	} else if strings.HasSuffix(archive, ".gzip") || strings.HasSuffix(archive, ".gz") {
		log.Debugln("Decompressing gzip file", archive)

		r, err := gzip.NewReader(src)
		if err != nil {
			return nil, err
		}

		name := r.Header.Name
		if !matchExecutableName(filename, name) {
			return nil, fmt.Errorf(
				"file name '%v' does not match to command '%v' found in %v", name, filename, archive,
			)
		}

		log.Debugln("Executable file", name, "was found in gzip file")

		return r, nil
	} else if strings.HasSuffix(archive, ".tar.xz") {
		log.Debugln("Decompressing tar.xz file...", archive)

		xzip, err := xz.NewReader(src)
		if err != nil {
			return nil, err
		}

		return unarchiveTar(xzip, archive, filename)
	} else if strings.HasSuffix(archive, ".xz") {
		log.Debugln("Decompressing xzip file", archive)

		xzip, err := xz.NewReader(src)
		if err != nil {
			return nil, err
		}

		log.Println("Decompressed file from xzip is assumed to be an executable", filename)

		return xzip, nil
	}

	log.Debugln("Decompression is not needed", filename)

	return src, nil
}

func unarchiveTar(src io.Reader, archive, filename string) (io.Reader, error) {
	t := tar.NewReader(src)

	for {
		h, err := t.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}

		_, name := filepath.Split(h.Name)
		if matchExecutableName(filename, name) {
			log.Debugln("Executable file", h.Name, "was found in tar archive")

			return t, nil
		}
	}

	return nil, fmt.Errorf("file named '%v' is not found in %v", filename, archive)
}

// Unzip will decompress a zip archive, moving all files and folders
//
//	within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src io.Reader, dest string) ([]string, error) {
	body, err := io.ReadAll(src)
	if err != nil {
		log.Fatal(err)
	}

	z, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		log.Fatal(err)
	}

	filenames := make([]string, 0, len(z.File))

	for _, f := range z.File {
		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name) //nolint:gosec

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			err = AFS.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return []string{}, err
			}

			continue
		}

		// Make File
		if err = AFS.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := AFS.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		for {
			_, err := io.CopyN(outFile, rc, 1024)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				return []string{}, err
			}
		}

		// Close the file without defer to close before next iteration of loop
		_ = outFile.Close()
		_ = rc.Close()

		if err != nil {
			return filenames, err
		}
	}

	return filenames, nil
}

func matchExecutableName(cmd, target string) bool {
	if cmd == target {
		return true
	}

	o, a := runtime.GOOS, runtime.GOARCH

	// When the contained executable name is full name (e.g. foo_darwin_amd64),
	// it is also regarded as a target executable file. (#19)
	for _, d := range []rune{'_', '-'} {
		c := fmt.Sprintf("%s%c%s%c%s", cmd, d, o, d, a)

		if o == "windows" {
			c += ".exe"
		}

		if c == target {
			return true
		}
	}

	return false
}

// SvcEnabledPermissive returns true if the s service is enabled in Viper settings. This function is also going to
// return true if the service is not mentioned in Viper settings (defaults to true).
func SvcEnabledPermissive(s string) bool {
	key := AppName + "_" + s
	if viper.IsSet(key) {
		return viper.GetBool(key)
	}

	return true
}

// SvcEnabledStrict returns true if the s service is enabled in Viper settings. This function is going to
// return false if the service is not mentioned in Viper settings (defaults to false).
func SvcEnabledStrict(name string) bool {
	key := AppName + "_" + name
	if viper.IsSet(key) {
		return viper.GetBool(key)
	}

	return false
}

// Quote puts a quote around s string in Unix-like systems and returns it, while it just returns s as-is on Windows.
func Quote(s string) string {
	switch OSDistro() {
	case "windows":
		return s
	default:
		return fmt.Sprintf("%q", s)
	}
}

func DockerHost() string {
	if runtime.GOOS == "windows" {
		return dockerClient.DefaultDockerHost
	}

	cmd := exec.Command("/bin/sh", "-c", "docker context list --format json")

	out, err := cmd.Output()
	if err != nil {
		return dockerClient.DefaultDockerHost
	}

	var contexts []dockerContext

	err = json.Unmarshal(out, &contexts)
	if err != nil {
		return dockerClient.DefaultDockerHost
	}

	for _, v := range contexts {
		if v.Current {
			return v.DockerEndpoint
		}
	}

	return dockerClient.DefaultDockerHost
}

// nolint: tagliatelle
type dockerContext struct {
	Current        bool   `json:"Current"`
	DockerEndpoint string `json:"DockerEndpoint"`
	Name           string `json:"Name"`
}
