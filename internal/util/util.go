package util

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	dockerClient "github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"gopkg.in/ini.v1"
)

var (
	// FS is the implementation of Afero Filesystem. It's a filesystem wrapper and used for testing.
	FS = &afero.Afero{Fs: afero.NewOsFs()}
)

var (
	// ErrFileWithThisDirNameExist occurs when file already exist.
	ErrFileWithThisDirNameExist = fmt.Errorf("file with the same name exists")
	// ErrFileNotFound occurs when file is not found.
	ErrFileNotFound = func(s string) error {
		return fmt.Errorf("file not found: %s", s)
	}
)

// CreateDir creates the directory if not exist.
func CreateDir(dir string, perm *os.FileMode) error {
	log.Debugf("Creating directory %s...", dir)

	if dir == "" {
		log.Panicln("Directory path is empty.")
	}

	dirPath, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("cannot determine absolute path for directory: %w", err)
	}

	dirMode := os.FileMode(0o755)
	if perm != nil {
		dirMode = *perm
	}

	stat, err := FS.Stat(dirPath)
	switch {
	case os.IsNotExist(err):
		log.Tracef("path: %v, mode: %v\n", dirPath, dirMode)

		err = FS.MkdirAll(dirPath, dirMode)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	case stat.Mode().IsDir():
		if stat.Mode().Perm() != dirMode {
			err = FS.Chmod(dirPath, dirMode)
			if err != nil {
				return fmt.Errorf("%w", err)
			}

			return nil
		}
	default:
		return ErrFileWithThisDirNameExist
	}

	log.Debugf("...directory %s created successfully.", dir)

	return nil
}

// CreateDirAndWriteToFile creates the base directory and writes bytes to a file in it.
// Perms are optional. The first permission will be the file permission, the second will be the directory permission.
func CreateDirAndWriteToFile(bytes []byte, file string, perms ...os.FileMode) error {
	log.Debugf("Writing file %s...", file)

	filePath, err := filepath.Abs(file)
	if err != nil {
		return fmt.Errorf("cannot determine absolute path for directory: %w", err)
	}

	fileMode := os.FileMode(0o640)
	if len(perms) > 0 {
		fileMode = perms[0]
	}

	dirMode := os.FileMode(0o755)
	if len(perms) > 1 {
		dirMode = perms[1]
	}

	err = CreateDir(filepath.Dir(filePath), &dirMode)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	err = FS.WriteFile(filePath, bytes, fileMode)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	log.Debugf("...file %s created successfully.", file)

	return nil
}

// AppendToFileOrCreateDirAndWriteToFile creates the base directory and writes to a file in it.
// Perms are optional. The first permission will be the file permission, the second will be the directory permission.
func AppendToFileOrCreateDirAndWriteToFile(bytes []byte, file string, perms ...os.FileMode) error {
	log.Debugf("Appending to file %s...", file)

	filePath, err := filepath.Abs(file)
	if err != nil {
		return fmt.Errorf("cannot determine absolute path for directory: %w", err)
	}

	fileMode := os.FileMode(0o640)
	if len(perms) > 0 {
		fileMode = perms[0]
	}

	dirMode := os.FileMode(0o755)
	if len(perms) > 1 {
		dirMode = perms[1]
	}

	err = CreateDir(filepath.Dir(filePath), &dirMode)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	f, err := FS.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fileMode)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()

	_, err = f.Write(bytes)
	if err != nil {
		return fmt.Errorf("cannot write to file: %w", err)
	}

	log.Debugf("...file %s created successfully.", file)

	return nil
}

// CheckFileExistsAndRecreate checks if the file already exists and ask the user if he'd like to recreate it.
// If user input is yes - return false (as if the file does not existing).
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

// AskForConfirmation ask msg from the user and returns the answer.
func AskForConfirmation(msg string) bool {
	log.Printf("%v (y)es, (n)o", msg)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := scanner.Text()

	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		log.Println("I'm sorry but I didn't get what you meant, please type (y)es or (n)o and then press enter:")

		return AskForConfirmation(msg)
	}
}

// FileExists checks if the file already exists and ask the user if he'd like to recreate it.
func FileExists(file string) bool {
	log.Tracef("Checking if file exist: %s...", file)

	if file == "" {
		log.Panicln("Path is empty.")
	}

	exist := false
	if _, err := FS.Stat(filepath.Join(file)); !os.IsNotExist(err) {
		log.Traceln("...file exists.")

		exist = true
	} else {
		log.Traceln("...file does not exist.")
	}

	return exist
}

// EvalSymlinkPath checks if file exists and returns the resolved path if a symlink if file is a symlink,
// else it returns the filepath.
func EvalSymlinkPath(file string) (string, error) {
	log.Tracef("Evaluating symlink path for file %s...", file)

	if !FileExists(file) {
		return "", ErrFileNotFound(file)
	}

	stat, err := os.Lstat(file)
	if err != nil {
		return "", err
	}

	var resolvedPath string
	if isSymlink(stat) {
		resolvedPath, _, err = evalSymlink(FS, file)
		if err != nil {
			return "", err
		}

		log.Tracef("...symlink path for file %s is %s.", file, resolvedPath)
	} else {
		resolvedPath = file

		log.Tracef("...file %s is not a symlink.", file)
	}

	return resolvedPath, nil
}

func isSymlink(fi os.FileInfo) bool {
	return fi != nil && fi.Mode()&os.ModeSymlink == os.ModeSymlink
}

func evalSymlink(fs afero.Fs, filename string) (string, os.FileInfo, error) {
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
	_, err := exec.LookPath(name)

	return err == nil
}

// OSDistro returns the linux distro name if GOOS is linux, else "darwin" or "windows".
func OSDistro() string {
	if runtime.GOOS == "linux" {
		cfg, err := ini.Load("/etc/os-release")
		if err != nil {
			log.Panicln("Fail to read file: ", err)
		}

		ConfigParams := make(map[string]string)
		ConfigParams["ID"] = cfg.Section("").Key("ID").String()

		return strings.ToLower(ConfigParams["ID"])
	}

	return runtime.GOOS
}

// HomeDir returns the invoking user's home directory.
func HomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Panicln(err)
	}

	return home
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

	var contexts []struct {
		Current        bool   `json:"Current"`
		DockerEndpoint string `json:"DockerEndpoint"`
		Name           string `json:"Name"`
	}

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

func Insert(array []string, index int, value string) []string {
	if len(array) == index {
		return append(array, value)
	}
	array = append(array[:index+1], array[index:]...)
	array[index] = value
	return array
}
