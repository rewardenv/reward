package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UtilTestSuite struct {
	suite.Suite
}

func (suite *UtilTestSuite) SetupTest() {
	FS = &afero.Afero{Fs: afero.NewMemMapFs()}

	_ = FS.Mkdir("/path/to/existing-directory", os.FileMode(0o750))
	_ = FS.WriteFile("/path/to/existing-file", []byte("non-empty-file"), os.FileMode(0o644))
	f, _ := FS.Create(".env")

	defer f.Close()
}

func TestUtilTestSuite(t *testing.T) {
	suite.Run(t, new(UtilTestSuite))
}

func (suite *UtilTestSuite) TestCreateDir() {
	dirMode := os.FileMode(0o755)

	type args struct {
		dir  string
		perm *os.FileMode
	}

	tests := []struct {
		name      string
		args      args
		want      os.FileMode
		wantErr   error
		wantPanic bool
	}{
		{
			name: "valid test",
			args: args{
				dir:  "/path/to/directory-with-permission",
				perm: &dirMode,
			},
			want:    dirMode,
			wantErr: nil,
		},
		{
			name: "valid test with empty permission",
			args: args{
				dir: "/path/to/directory2",
			},
			want:    os.FileMode(0o755),
			wantErr: nil,
		},
		{
			name: "test overwriting existing directory perms",
			args: args{
				dir: "/path/to/existing-directory",
			},
			want:    os.FileMode(0o755),
			wantErr: nil,
		},
		{
			name: "test overwriting existing file",
			args: args{
				dir: "/path/to/existing-file",
			},
			want:    os.FileMode(0o755),
			wantErr: ErrFileWithThisDirNameExist,
		},
		{
			name: "empty dir name",
			args: args{
				dir: "",
			},
			want:      os.FileMode(0o755),
			wantErr:   nil,
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() {
					_ = CreateDir(tt.args.dir, tt.args.perm)
				})

				return
			}

			err := CreateDir(tt.args.dir, tt.args.perm)

			if err != nil || tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())

				return
			}

			got, _ := FS.Stat(tt.args.dir)
			assert.Equal(t, tt.want, got.Mode().Perm())
		})
	}
}

func (suite *UtilTestSuite) TestCreateDirAndWriteToFile() {
	type args struct {
		bytes []byte
		file  string
		perms []os.FileMode
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "test with not-existing dir with perms",
			args: args{
				bytes: []byte("test"),
				file:  "/path/to/not-existing-dir/file",
				perms: []os.FileMode{os.FileMode(0o644), os.FileMode(0o750)},
			},
			wantErr: nil,
		},
		{
			name: "test with specific permissions",
			args: args{
				bytes: []byte("test"),
				file:  "/path/to/file",
				perms: []os.FileMode{os.FileMode(0o644), os.FileMode(0o750)},
			},
			wantErr: nil,
		},
		{
			name: "test without specific permissions",
			args: args{
				bytes: []byte("test"),
				file:  "/path/to/file2",
			},
			wantErr: nil,
		},
		{
			name: "test with existing file",
			args: args{
				bytes: []byte("test"),
				file:  "/path/to/existing-file",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			err := CreateDirAndWriteToFile(tt.args.bytes, tt.args.file, tt.args.perms...)
			if err != nil || tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)

				return
			}

			dirStat, _ := FS.Stat(filepath.Dir(tt.args.file))
			fileStat, _ := FS.Stat(tt.args.file)
			file, _ := FS.ReadFile(tt.args.file)

			if len(tt.args.perms) > 0 {
				perm := fileStat.Mode().Perm()
				assert.Equal(t, tt.args.perms[0], perm)

				if len(tt.args.perms) > 1 {
					perm = dirStat.Mode().Perm()
					assert.Equal(t, tt.args.perms[1], perm)
				}
			}

			assert.Equal(t, tt.args.bytes, file)
		})
	}
}

func (suite *UtilTestSuite) TestAppendToFileOrCreateDirAndWriteToFile() {
	type args struct {
		bytes []byte
		file  string
		perms []os.FileMode
	}

	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr error
	}{
		{
			name: "test with not existing file",
			args: args{
				bytes: []byte("test"),
				file:  "/path/to/not-existing-file",
			},
			want:    []byte("test"),
			wantErr: nil,
		},
		{
			name: "test with existing file",
			args: args{
				bytes: []byte("test"),
				file:  "/path/to/existing-file",
			},
			want:    []byte("non-empty-filetest"),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			err := AppendToFileOrCreateDirAndWriteToFile(tt.args.bytes, tt.args.file, tt.args.perms...)
			if err != nil || tt.wantErr != nil {
				assert.ErrorIs(t, tt.wantErr, err)

				return
			}

			dirStat, _ := FS.Stat(filepath.Dir(tt.args.file))
			fileStat, _ := FS.Stat(tt.args.file)
			file, _ := FS.ReadFile(tt.args.file)

			if len(tt.args.perms) > 0 {
				perm := fileStat.Mode().Perm()
				assert.Equal(t, tt.args.perms[0], perm)

				if len(tt.args.perms) > 1 {
					perm = dirStat.Mode().Perm()
					assert.Equal(t, tt.args.perms[1], perm)
				}
			}

			assert.Equal(t, tt.want, file)
		})
	}
}
