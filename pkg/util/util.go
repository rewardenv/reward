package util

import (
	tarpkg "archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	dockerClient "github.com/docker/docker/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	xzpkg "github.com/ulikunitz/xz"
	"gopkg.in/ini.v1"
)

// FS is the implementation of Afero Filesystem. It's a filesystem wrapper and used for testing.
var FS = &afero.Afero{Fs: afero.NewOsFs()}

var (
	// ErrFileWithThisDirNameExist occurs when file already exist.
	ErrFileWithThisDirNameExist = errors.New("file with the same name exists")
	// ErrFileNotFound occurs when file is not found.
	ErrFileNotFound = func(s string) error {
		return errors.Errorf("file not found: %s", s)
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
		return errors.Wrap(err, "cannot determine absolute path for directory")
	}

	dirMode := os.FileMode(0o755)
	if perm != nil {
		dirMode = *perm
	}

	stat, err := FS.Stat(dirPath)

	switch {
	case os.IsNotExist(err):
		log.Tracef("path: %s, mode: %s\n", dirPath, dirMode)

		err = FS.MkdirAll(dirPath, dirMode)
		if err != nil {
			return errors.Wrap(err, "creating directory")
		}
	case stat.Mode().IsDir():
		if stat.Mode().Perm() != dirMode {
			err = FS.Chmod(dirPath, dirMode)
			if err != nil {
				return errors.Wrap(err, "changing directory permission")
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
		return errors.Wrap(err, "determining absolute path for directory")
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
		return err
	}

	err = FS.WriteFile(filePath, bytes, fileMode)
	if err != nil {
		return errors.Wrap(err, "writing to file")
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
		return errors.Wrap(err, "determining absolute path for directory")
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
		return err
	}

	f, err := FS.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fileMode)
	if err != nil {
		return errors.Wrap(err, "opening file")
	}
	defer f.Close()

	_, err = f.Write(bytes)
	if err != nil {
		return errors.Wrap(err, "writing to file")
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

	//nolint:gocritic
	filePath := filepath.Join(file)
	confirmation := false

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		log.Printf("File already exists: %s", filePath)

		conf := AskForConfirmation("Would you like to recreate it?")

		// Have to swap bool, if the user doesn't want to recreate it, than return as it exist (true).
		confirmation = !conf
	}

	return confirmation
}

// AskForConfirmation ask msg from the user and returns the answer.
func AskForConfirmation(msg string) bool {
	if viper.GetBool("assume_yes") {
		return true
	}

	//nolint:forbidigo
	fmt.Printf("%s (y)es, (n)o\n", msg)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := scanner.Text()

	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		//nolint:forbidigo
		fmt.Println("I'm sorry but I didn't get what you meant, please type (y)es or (n)o and then press enter:")

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

	//nolint:gocritic
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
		return "", errors.Wrap(err, "cannot stat file")
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
		return "", nil, errors.Wrap(err, "cannot evaluate symlink")
	}

	fi, err := fs.Stat(link)
	if err != nil {
		return "", nil, errors.Wrap(err, "cannot stat symlink")
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

	//nolint:tagliatelle
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
		arg := args[i]

		var flag *pflag.Flag

		if arg[0] == '-' {
			if arg[1] == '-' {
				flag = flags.Lookup(strings.SplitN(arg[2:], "=", 2)[0])
			} else {
				for _, s := range arg[1:] {
					flag = flags.ShorthandLookup(string(s))
					if flag == nil {
						break
					}
				}
			}
		}

		if flag != nil {
			if flag.NoOptDefVal == "" && i+1 < len(args) && flag.Value.String() == args[i+1] {
				i++
			}

			continue
		}

		unknownArgs = append(unknownArgs, arg)
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

// CheckRegexInFile checks if the file contains content.
func CheckRegexInFile(regex, filePath string) (bool, error) {
	//nolint:gocritic
	file, err := FS.Open(filepath.Join(filePath))
	if err != nil {
		return false, errors.Wrap(err, "cannot open file")
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
		return false, errors.Wrap(err, "cannot scan file")
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

// ContainsString checks if a slice of string contains a string.
func ContainsString(slice []string, vals ...string) bool {
	for _, item := range slice {
		for _, val := range vals {
			if item == val {
				return true
			}
		}
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

func DecompressFileFromArchive(src io.Reader, archive, filename string) (io.Reader, error) {
	switch {
	case strings.HasSuffix(archive, ".zip"):
		log.Debugf("Decompressing zip file %s...", archive)

		buf, err := io.ReadAll(src)
		if err != nil {
			return nil, errors.Wrap(err, "cannot read file")
		}

		r := bytes.NewReader(buf)

		z, err := zip.NewReader(r, r.Size())
		if err != nil {
			return nil, errors.Wrap(err, "cannot read zip file")
		}

		for _, file := range z.File {
			_, name := filepath.Split(file.Name)
			if !file.FileInfo().IsDir() && matchExecutableName(filename, name) {
				log.Debugf("...%s found in zip file %s.", name, archive)

				f, err := file.Open()
				if err != nil {
					return nil, errors.Wrap(err, "cannot open file in zip")
				}

				return f, nil
			}
		}

		return nil, ErrFileNotFound(filename)
	case strings.HasSuffix(archive, ".tar.gz"), strings.HasSuffix(archive, ".tgz"):
		log.Debugf("Decompressing tar.gz file %s...", archive)

		gz, err := gzip.NewReader(src)
		if err != nil {
			return nil, errors.Wrap(err, "cannot read gzip file")
		}

		return unarchiveTar(gz, archive, filename)
	case strings.HasSuffix(archive, ".gzip"), strings.HasSuffix(archive, ".gz"):
		log.Debugf("Decompressing gzip file %s...", archive)

		r, err := gzip.NewReader(src)
		if err != nil {
			return nil, errors.Wrap(err, "cannot read gzip file")
		}

		name := r.Header.Name
		if !matchExecutableName(filename, name) {
			return nil, ErrFileNotFound(filename)
		}

		log.Debugf("...%s found in gzip file.", name)

		return r, nil
	case strings.HasSuffix(archive, ".tar.xz"):
		log.Debugf("Decompressing tar.xz file %s...", archive)

		xz, err := xzpkg.NewReader(src)
		if err != nil {
			return nil, errors.Wrap(err, "cannot read txz file")
		}

		return unarchiveTar(xz, archive, filename)
	case strings.HasSuffix(archive, ".xz"):
		log.Debugf("Decompressing xz file %s...", archive)

		xz, err := xzpkg.NewReader(src)
		if err != nil {
			return nil, errors.Wrap(err, "cannot read xz file")
		}

		log.Debugf("...%s found in xz file.", filename)

		return xz, nil
	}

	log.Debugln("...decompression is not needed.", filename)

	return src, nil
}

func unarchiveTar(src io.Reader, archive, filename string) (io.Reader, error) {
	tar := tarpkg.NewReader(src)

	for {
		h, err := tar.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, errors.Wrap(err, "cannot read tar file")
		}

		_, name := filepath.Split(h.Name)
		if matchExecutableName(filename, name) {
			log.Debugln("Executable file", h.Name, "was found in tar archive")

			return tar, nil
		}
	}

	return nil, errors.Errorf("file named '%s' is not found in %s", filename, archive)
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

// Unzip will decompress a zip archive, moving all files and folders within the zip file (parameter 1) to an
// output directory (parameter 2).
func Unzip(src io.Reader, dest string) ([]string, error) {
	body, err := io.ReadAll(src)
	if err != nil {
		log.Panicln(err)
	}

	z, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		log.Panicln(err)
	}

	filenames := make([]string, 0, len(z.File))

	//nolint:varnamelen
	for _, f := range z.File {
		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name) //nolint:gosec

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, errors.Wrap(err, "illegal file path")
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			err = FS.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return []string{}, errors.Wrap(err, "creating directory")
			}

			continue
		}

		// Make File
		if err = FS.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, errors.Wrap(err, "creating directory")
		}

		outFile, err := FS.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, errors.Wrap(err, "opening file")
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, errors.Wrap(err, "opening file")
		}

		for {
			_, err := io.CopyN(outFile, rc, 1024)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				return []string{}, errors.Wrap(err, "copying file")
			}
		}

		// Close the file without defer to close before next iteration of loop
		_ = outFile.Close()
		_ = rc.Close()
	}

	return filenames, nil
}

// InsertStringBeforeOccurrence inserts insertStr before occurrence of searchStr (if exist) to args and returns args.
// If searchStr doesn't exist it will append to the end of args.
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
// If searchStr does not exist it will append to the end of args.
func InsertStringAfterOccurrence(args []string, insertStr, searchStr string) []string {
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

func BoolPtr(b bool) *bool {
	return &b
}

func RemoveStringFromSlice(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}

	return s
}
