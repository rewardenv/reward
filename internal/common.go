package internal

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
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/hashicorp/go-version"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/ulikunitz/xz"
	"gopkg.in/ini.v1"
)

const (
	APPNAME = "reward"
)

var (
	AppName               = strings.ToLower(APPNAME)
	VersionFileContent, _ = Asset("VERSION.txt")
)

var (
	// FS is a Memory Map Filesystem.
	FS = afero.NewOsFs()
	// AFS is the implementation of Afero Filesystem.
	AFS = &afero.Afero{Fs: FS}
)

var (
	ErrEmptyDirName                 = errors.New("empty directory name")
	ErrFileWithThisDirNameExist     = errors.New("file with the same name exists")
	ErrFileNotFound                 = errors.New("file not found")
	ErrUnknownEnvType               = errors.New("unknown env type")
	ErrUnknownAction                = errors.New("unknown action error")
	ErrDockerIsNotRunning           = errors.New("docker is not running or docker version is too old")
	ErrDockerVersionMismatch        = errors.New("docker version is too old")
	ErrDockerComposeVersionMismatch = errors.New("docker-compose version is too old")
	ErrEnvNameIsInvalid             = errors.New("environment name is invalid, it should match RFC1178")
	ErrEnvIsEmpty                   = fmt.Errorf("env name is empty. please run `%v env-init`", AppName)
	ErrCaCertDoesNotExist           = fmt.Errorf(
		"the root CA certificate is missing, please run '%v install' and try again",
		AppName)
	ErrCannotFindContainer = errors.New("container cannot be found")
	ErrArgumentRequired    = errors.New("argument required")
	ErrInvokedAsRootUser   = errors.New("In most cases, you should not run " +
		AppName + " as root user except for `self-update`. " + "If you are sure you want to do this, use " +
		strings.ToUpper(AppName) + "_ALLOW_SUPERUSER=1.")
)

func FileNotFoundError(op string) error {
	return fmt.Errorf("ErrFileNotFound: %w, %v", ErrFileNotFound, op)
}

func CannotFindContainerError(op string) error {
	return fmt.Errorf("ErrCannotFindContainer: %w: %s", ErrCannotFindContainer, op)
}

func ArgumentRequiredError(op string) error {
	return fmt.Errorf("ErrArgumentRequired: %w: %s", ErrArgumentRequired, op)
}
func DockerVersionMismatchError(op string) error {
	return fmt.Errorf("ErrDockerVersionMismatch: %w: %s", ErrDockerVersionMismatch, op)
}
func DockerComposeVersionMismatchError(op string) error {
	return fmt.Errorf("ErrDockerComposeVersionMismatch: %w: %s", ErrDockerComposeVersionMismatch, op)
}

func GetAppVersion() *version.Version {
	v, _ := version.NewVersion(strings.TrimSpace(string(VersionFileContent)))

	return v
}

// GetEnvName returns the environment name in lowercase format.
func GetEnvName() string {
	return strings.ToLower(viper.GetString(AppName + "_env_name"))
}

// GetEnvType returns the environment type in lowercase format.
func GetEnvType() string {
	return strings.ToLower(viper.GetString(AppName + "_env_type"))
}

// GetEnvNetworkName returns the environments docker network name in lowercase format.
func GetEnvNetworkName() string {
	return strings.ToLower(GetEnvName() + "_default")
}

// GetCwd return the current working directory.
func GetCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	return cwd
}

func GetHomeDir() string {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatalln(err)
	}

	return home
}

// GetAppHomeDir returns the application's home directory.
func GetAppHomeDir() string {
	return viper.GetString(AppName + "_home_dir")
}

func GetServiceDomain() string {
	return viper.GetString(AppName + "_service_domain")
}

func GetMutagenSyncFile() string {
	return filepath.Join(GetCwd(), "."+AppName, "mutagen.yml")
}

func GetMutagenSyncIgnore() string {
	return viper.GetString(AppName + "_sync_ignore")
}

func GetWebRoot() string {
	return viper.GetString(AppName + "_web_root")
}

func IsDBEnabled() bool {
	return viper.GetString(AppName+"_db") == "1"
}

func GetDBContainer() string {
	if viper.IsSet(AppName + "_env_db_container") {
		return viper.GetString(AppName + "_env_db_container")
	}

	return "db"
}

func GetDBCommand() string {
	if viper.IsSet(AppName + "_env_db_command") {
		return viper.GetString(AppName + "_env_db_command")
	}

	return "mysql"
}

func GetBlackfireContainer() string {
	if viper.IsSet(AppName + "_env_blackfire_container") {
		return viper.GetString(AppName + "_env_blackfire_container")
	}

	return "php-blackfire"
}

func GetBlackfireCommand() string {
	if viper.IsSet(AppName + "_env_blackfire_command") {
		return viper.GetString(AppName + "_env_blackfire_command")
	}

	return "blackfire"
}

func IsBlackfireEnabled() bool {
	return viper.GetString(AppName+"_blackfire") == "1"
}

func ResolveDomainToTraefik() bool {
	if viper.IsSet(AppName + "_resolve_domain_to_traefik") {
		return viper.GetBool(AppName + "_resolve_domain_to_traefik")
	}

	return true
}

// func GetSeleniumEnabled() bool {
// 	return viper.GetString(AppName+"_selenium") == "1" && viper.GetString(AppName+"_selenium_debug") == "1"
// }

func GetMagentoVersion() *version.Version {
	v := new(version.Version)

	type ComposerJSON struct {
		Require map[string]string `json:"require"`
	}

	var composerJSON ComposerJSON

	if CheckFileExists("composer.json") {
		data, err := AFS.ReadFile("composer.json")
		if err != nil {
			v = GetMagentoVersionFromViper()
		}

		if err = json.Unmarshal(data, &composerJSON); err != nil {
			v = GetMagentoVersionFromViper()
		}

		for key, val := range composerJSON.Require {
			if CheckRegexInString(`^magento/product-(enterprise|community)-edition$`, key) {
				v, _ = version.NewVersion(val)
			}
		}

		return v
	}

	v = GetMagentoVersionFromViper()

	return v
}

func GetMagentoVersionFromViper() *version.Version {
	var v *version.Version

	if GetEnvType() == "magento1" {
		if viper.IsSet(AppName + "_magento_version") {
			v, _ = version.NewVersion(viper.GetString(AppName + "_magento_version"))
		} else {
			v, _ = version.NewVersion("1.9.4")
		}
	} else {
		if viper.IsSet(AppName + "_magento_version") {
			v, _ = version.NewVersion(viper.GetString(AppName + "_magento_version"))
		} else {
			v, _ = version.NewVersion("2.4.2")
		}
	}

	return v
}

func GetTraefikDomain() string {
	domain := viper.GetString("traefik_domain")

	log.Debugln("Traefik Domain:", domain)

	return domain
}
func GetTraefikSubdomain() string {
	subDomain := viper.GetString("traefik_subdomain")

	log.Debugln("Traefik Subdomain:", subDomain)

	return subDomain
}

func GetTraefikFullDomain() string {
	var fullDomain string

	if GetTraefikSubdomain() == "" {
		fullDomain = GetTraefikDomain()
	} else {
		fullDomain = GetTraefikSubdomain() + "." + GetTraefikDomain()
	}

	log.Debugln("Traefik Full Domain:", fullDomain)

	return fullDomain
}

func GetMagentoBackendFrontname() string {
	if viper.IsSet("magento_backend_frontname") {
		return viper.GetString("magento_backend_frontname")
	}

	return "admin"
}

func IsServiceEnabled(service string) bool {
	if viper.IsSet(AppName + "_" + service) {
		return viper.GetBool(AppName + "_" + service)
	}

	return false
}

func IsContainerRunning(name string) bool {
	_, err := GetContainerIDByName(name)

	return err == nil
}

func IsAllowedSuperuser() bool {
	if viper.IsSet(AppName + "_allow_superuser") {
		return viper.GetBool(AppName + "_allow_superuser")
	}

	return false
}

func IsWSL2DirectMount() bool {
	if viper.IsSet(AppName + "_wsl2_direct_mount") {
		return viper.GetBool(AppName + "_wsl2_direct_mount")
	}

	return false
}

func IsSingleWebContainer() bool {
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

// GetOSDistro returns the linux distro name if GOOS is linux, else "darwin" or "windows".
func GetOSDistro() string {
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

func IsMutagenSyncEnabled() bool {
	return GetOSDistro() == "darwin" || (GetOSDistro() == "windows" && !IsWSL2DirectMount())
}

// // CopyFile copies src file to dst path
// func CopyFile(src, dst string) (int64, error) {
//	sourceFileStat, err := os.Stat(src)
//	if err != nil {
//		return 0, err
//	}
//
//	if !sourceFileStat.Mode().IsRegular() {
//		return 0, fmt.Errorf("%s is not a regular file", src)
//	}
//
//	source, err := os.Open(src)
//	if err != nil {
//		return 0, err
//	}
//	defer source.Close()
//
//	destination, err := os.Create(dst)
//	if err != nil {
//		return 0, err
//	}
//	defer destination.Close()
//	nBytes, err := io.Copy(destination, source)
//	return nBytes, err
// }

// CheckFileExistsAndRecreate checks if the file already exists and ask the user if he'd like to recreate it.
//   If user input is yes - return false (as if the file does not existing).
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
//   else it returns the filepath.
func EvalSymlinkPath(file string) (string, error) {
	if !CheckFileExists(file) {
		return "", FileNotFoundError(file)
	}

	stat, err := os.Lstat(file)
	if err != nil {
		return "", err
	}

	var resolvedPath string

	isSymlink := isSymlink(stat)
	log.Debugln(isSymlink)

	if isSymlink {
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

func isSymlink(fi os.FileInfo) bool {
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

// IsCommandAvailable returns if the parameter can be find in $PATH.
func IsCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// CreateDir creates the directory if not exist.
func CreateDir(dir string, perms ...int) error {
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

	log.Printf("File saved: %v", filePath)

	return nil
}

// installSSHConfig updates the ssh config file to use our key if the target host is the tunnel.
func installSSHConfig() error {
	var sshConfig = fmt.Sprintf(`## %v START ##
Host tunnel.%v.test
  HostName 127.0.0.1
  User user
  Port 2222
  IdentityFile %v/tunnel/ssh_key
## %v END ##`, strings.ToUpper(AppName), AppName, GetAppHomeDir(), strings.ToUpper(AppName))

	log.Println("Updating SSH config file...")

	sshConfigFile := filepath.Join("/etc/ssh/ssh_config")

	content, err := ioutil.ReadFile(sshConfigFile)
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

		log.Printf("Running command: %v", cmd)

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

// // AppendStringToFile appends content to a file
// func AppendStringToFile(content, filePath string) {
//	f, err := os.OpenFile(filePath,
//		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
//	if err != nil {
//		log.Println(err)
//	}
//	defer f.Close()
//	if _, err := f.WriteString(content); err != nil {
//		log.Println(err)
//	}
// }

// CheckRegexInFile checks if the file contains content.
func CheckRegexInFile(regex, filePath string) (bool, error) {
	file, err := AFS.Open(filepath.Join(filePath))
	if err != nil {
		return false, fmt.Errorf("%w", err)
	}
	defer file.Close()

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

func RunOsCommand(args string, suppressOsStdOut ...bool) (string, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		log.Debugf("Running command: cmd /c %v", args)

		cmd = exec.Command("cmd", "/c", args)
	} else {
		log.Debugf("Running command: sh -c %v", args)

		cmd = exec.Command("sh", "-c", args)
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
				GetTraefikDomain(),
				GetTraefikFullDomain(),
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

func IsContainerChanged(container string) bool {
	containerState, err := GetContainerStateByName(container)
	if err != nil {
		return true
	}

	log.Debugf("Current container (%v) state is: %v", container, containerState)

	containerID, err := GetContainerIDByName(container)
	if err != nil {
		return true
	}

	log.Debugf("Current container (%v) ID is: %v", container, containerID)

	if containerState != "running" {
		return true
	}

	syncListOut, err := SyncListCmd(true)
	if err != nil {
		return true
	}
	// log.Debugln(syncListOut)

	re := regexp.MustCompile(`URL: docker://[^/]*`)
	stripRe := regexp.MustCompile(`(URL: docker://)([^/\\]*)`)
	previousContainerID := stripRe.ReplaceAllString(re.FindString(syncListOut), "$2")

	log.Debugln("Previously synced container ID:", previousContainerID)

	return previousContainerID != containerID
}

func CheckDockerNetworkExist(networkName string) (bool, error) {
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

// func ListAllFiles() {
//	AFS.Walk("/",
//		func(path string, info os.FileInfo, err error) error {
//			if err != nil {
//				return err
//			}
//			fmt.Println(path, info.Size())
//			return nil
//		})
//	return
// }

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

func DockerComposeCompleter() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) { //nolint:lll
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
//   If searchStr is not exists it will append to the end of args.
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
//   If searchStr is not exists it will append to the end of args.
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

func decompressFileFromArchive(src io.Reader, archive, filename string) (io.Reader, error) {
	if strings.HasSuffix(archive, ".zip") {
		log.Println("Decompressing zip file", archive)

		buf, err := ioutil.ReadAll(src)
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
		log.Println("Decompressing tar.gz file", archive)

		gz, err := gzip.NewReader(src)
		if err != nil {
			return nil, err
		}

		return unarchiveTar(gz, archive, filename)
	} else if strings.HasSuffix(archive, ".gzip") || strings.HasSuffix(archive, ".gz") {
		log.Println("Decompressing gzip file", archive)

		r, err := gzip.NewReader(src)
		if err != nil {
			return nil, err
		}

		name := r.Header.Name
		if !matchExecutableName(filename, name) {
			return nil, fmt.Errorf("file name '%v' does not match to command '%v' found in %v", name, filename, archive)
		}

		log.Debugln("Executable file", name, "was found in gzip file")
		return r, nil
	} else if strings.HasSuffix(archive, ".tar.xz") {
		log.Println("Decompressing tar.xz file...", archive)

		xzip, err := xz.NewReader(src)
		if err != nil {
			return nil, err
		}

		return unarchiveTar(xzip, archive, filename)
	} else if strings.HasSuffix(archive, ".xz") {
		log.Println("Decompressing xzip file", archive)

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
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		_, name := filepath.Split(h.Name)
		if matchExecutableName(filename, name) {
			log.Println("Executable file", h.Name, "was found in tar archive")
			return t, nil
		}
	}

	return nil, fmt.Errorf("file named '%v' is not found in %v", filename, archive)
}

// Unzip will decompress a zip archive, moving all files and folders
//   within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src io.Reader, dest string) ([]string, error) {
	var filenames []string

	body, err := ioutil.ReadAll(src)
	if err != nil {
		log.Fatal(err)
	}

	z, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		log.Fatal(err)
	}

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

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

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
