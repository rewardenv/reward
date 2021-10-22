package core_test

import (
	"errors"
	reward "github.com/rewardenv/reward/internal/core"
	"os"
	"runtime"
	"testing"

	"github.com/spf13/afero"
)

var (
	existingDir               = "/testdata/a-directory-that-exists"
	existingDirWithWrongPerms = "/testdata/a-directory-that-exists"
	nonExistingDir            = "/testdata/a-directory-that-does-not-exists"
	existingFile              = "/testdata/a-file-that-exists"
	nonExistingFile           = "/testdata/a-file-that-does-not-exist"

	existingRegexFile = "/testdata/a-file-with-regex-data-exists"
	existingEmptyFile = "/testdata/an-empty-file"
)

func init() {
	reward.FS = afero.NewMemMapFs()
	reward.AFS = &afero.Afero{Fs: reward.FS}

	initCommonTests()
	initCertTests()
	initCryptoTests()
	initTemplateTests()
}

func initCommonTests() {
	_ = reward.AFS.MkdirAll(existingDir, 0o755)
	_ = reward.AFS.MkdirAll(existingDirWithWrongPerms, 0o700)
	_ = reward.AFS.Mkdir("/tmpdir3", 0o755)
	_ = afero.WriteFile(reward.AFS, existingFile, []byte("test file"), 0o644)
	_ = afero.WriteFile(reward.AFS, existingRegexFile, []byte("test line 1\nexisting data\ntest line2"), 0o644)
	_ = afero.WriteFile(reward.AFS, existingEmptyFile, []byte(""), 0o644)
}

// ContainsString checks if a slice of string contains an exact match of string.
func TestContainsString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		str  string
		want bool
	}{
		{
			"Returns true if []string contains str",
			"contains-me",
			true,
		},
		{
			"Returns false if []string doesn't contain str",
			"contains-me-not",
			false,
		},
		{
			"Returns false if []string doesn't contain substring of str",
			"con",
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := reward.ContainsString([]string{"test", "contains-me", "test2"}, tt.str); got != tt.want {
				t.Errorf("ContainsString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOSDistro(t *testing.T) {
	t.Parallel()

	_ = reward.AFS.MkdirAll("/etc", 0o755)
	_ = afero.WriteFile(reward.AFS, "/etc/os-release", []byte("ID: ubuntu"), 0o644)

	tests := []struct {
		name string
		os   string
		want string
	}{
		{
			"Returns windows if os is windows",
			"windows",
			"windows",
		},
		{
			"Returns darwin if os is darwin",
			"darwin",
			"darwin",
		},
		{
			"Returns ubuntu if os is linux",
			"linux",
			"ubuntu",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if runtime.GOOS == tt.os {
				if got := reward.GetOSDistro(); got != tt.want {
					t.Errorf("GetOSDistro() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestCheckFileExists(t *testing.T) {
	t.Parallel()

	type args struct {
		path string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Returns false when provided path is empty",
			args{path: ""},
			false,
		},
		{
			"Returns true when given file exists",
			args{path: existingFile},
			true,
		},
		{
			"Returns false when given file does not exist",
			args{path: nonExistingFile},
			false,
		},
		{
			"Returns true when given directory exists",
			args{path: existingDir},
			true,
		},
		{
			"Returns false when given directory does not exist",
			args{path: nonExistingDir},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := reward.CheckFileExists(tt.args.path); got != tt.want {
				t.Errorf("CheckFileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCommandAvailable(t *testing.T) {
	t.Parallel()

	type args struct {
		command string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Returns true when command exist",
			args{command: "ls"},
			true,
		},
		{
			"Returns false when command does not exist",
			args{command: "test-command-which-does-not-exist"},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := reward.IsCommandAvailable(tt.args.command); got != tt.want {
				t.Errorf("IsCommandAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateDir(t *testing.T) {
	t.Parallel()

	type args struct {
		path  string
		perms int
	}

	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr error
	}{
		{
			"Check for filemode 0755",
			args{
				path: "/tmpdir1",
			},
			os.FileMode(0o755),
			nil,
		},
		{
			"Check for filemode 0750",
			args{
				path:  "/tmpdir2",
				perms: 0o750,
			},
			os.FileMode(0o750),
			nil,
		},
		{
			"Override filemode to 0700 for existing dir",
			args{
				path:  existingDirWithWrongPerms,
				perms: 0o755,
			},
			os.FileMode(0o755),
			nil,
		},
		{
			"Returns error, because the file exist and it's not a directory",
			args{
				path:  existingFile,
				perms: 0o755,
			},
			nil,
			reward.ErrFileWithThisDirNameExist,
		},
		{
			"Empty directory name",
			args{
				path:  "",
				perms: 0o755,
			},
			nil,
			reward.ErrEmptyDirName,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var err error
			if tt.args.perms == 0 {
				err = reward.CreateDir(tt.args.path)
			} else {
				err = reward.CreateDir(tt.args.path, tt.args.perms)
			}
			if err != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("CreateDir() = %v, want %v", err, tt.want)
				} else {
					return
				}
			}
			got, _ := reward.AFS.Stat(tt.args.path)
			if got.Mode().Perm() != tt.want {
				t.Errorf("CreateDir() = %v, want %v", got.Mode().Perm(), tt.want)
			}
		})
	}
}

func TestCheckExitCodeOfCommand(t *testing.T) {
	t.Parallel()

	type args struct {
		command string
	}

	tests := []struct {
		name string
		args args
		want int
	}{
		{
			"Return 0 for true",
			args{"true"},
			0,
		},
		{
			"Return 1 for false",
			args{"false"},
			1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := reward.CheckExitCodeOfCommand(tt.args.command); got != tt.want {
				t.Errorf("CheckExitCodeOfCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckRegexInFile(t *testing.T) {
	t.Parallel()

	type args struct {
		regex    string
		filePath string
	}

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"Return true for regex pattern exist in file",
			args{
				"^exi.",
				existingRegexFile,
			},
			true,
			false,
		},
		{
			"Return false for regex pattern does not exist in file",
			args{
				"^exist$",
				existingRegexFile,
			},
			false,
			false,
		},
		{
			"Return error if file does not exist",
			args{
				".*",
				nonExistingFile,
			},
			false,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := reward.CheckRegexInFile(tt.args.regex, tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckRegexInFile() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if got != tt.want {
				t.Errorf("CheckRegexInFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAskForConfirmation(t *testing.T) {
	type args struct {
		msg             string
		suppressMessage []bool
	}

	tests := []struct {
		name      string
		args      args
		userInput string
		want      bool
	}{
		{
			"Returns true if userInput is yes",
			args{
				"Random message",
				[]bool{},
			},
			"yes",
			true,
		},
		{
			"Returns false if userInput is no",
			args{
				"Random message 2",
				[]bool{},
			},
			"no",
			false,
		},
		{
			"Returns true if suppressMessage is []bool{true}, userInput can be anything",
			args{
				"Random message 3",
				[]bool{true},
			},
			"anything",
			true,
		},
		{
			"Returns false if suppressMessage is []bool{false}, userInput can be anything",
			args{
				"Random message 3",
				[]bool{false},
			},
			"anything",
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			answer := []byte(tt.userInput)
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}

			_, err = w.Write(answer)
			if err != nil {
				t.Error(err)
			}
			_ = w.Close()

			stdin := os.Stdin
			// Restore stdin right after the test.
			defer func() { os.Stdin = stdin }()
			os.Stdin = r

			got := reward.AskForConfirmation(tt.args.msg, tt.args.suppressMessage...)
			if got != tt.want {
				t.Errorf("AskForConfirmation() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckFileExistsAndRecreate(t *testing.T) {
	type args struct {
		file string
	}

	tests := []struct {
		name      string
		args      args
		userInput string
		want      bool
	}{
		{
			"Return false if file does not exist",
			args{nonExistingFile},
			"anything",
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			answer := []byte(tt.userInput)
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}

			_, err = w.Write(answer)
			if err != nil {
				t.Error(err)
			}
			_ = w.Close()

			stdin := os.Stdin
			// Restore stdin right after the test.
			defer func() { os.Stdin = stdin }()
			os.Stdin = r

			if got := reward.CheckFileExistsAndRecreate(tt.args.file); got != tt.want {
				t.Errorf("CheckFileExistsAndRecreate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCommandAvailable1(t *testing.T) {
	type args struct {
		name string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Return true for command ls",
			args{"ls"},
			true,
		},
		{
			"Return true for command ls",
			args{"ls123"},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := reward.IsCommandAvailable(tt.args.name); got != tt.want {
				t.Errorf("IsCommandAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}
