package core_test

import (
	"github.com/rewardenv/reward/internal/core"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

var (
	existingCaDir     = filepath.Join(existingDir, "ssl", "rootca")
	existingCaCertDir = filepath.Join(existingCaDir, "certs")
	existingCaFile    = filepath.Join(existingCaCertDir, "ca.cert.pem")
)

func initCertTests() {
	_ = core.AFS.MkdirAll(existingCaCertDir, 0o755)
	_ = afero.WriteFile(core.AFS, existingCaFile, []byte("test ca file"), 0o644)
}

func TestCaCertificateFilePath(t *testing.T) {
	t.Parallel()

	type args struct {
		caDir string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"Returns a cacert filepath for give path",
			args{caDir: "dir"},
			"dir/certs/ca.cert.pem",
			false,
		},
		{
			"Returns an error if path is empty",
			args{caDir: ""},
			"",
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := core.GetCaCertificateFilePath(tt.args.caDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCaCertificateFilePath() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if got != tt.want {
				t.Errorf("GetCaCertificateFilePath() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCaPrivKeyFilePath(t *testing.T) {
	t.Parallel()

	type args struct {
		caDir string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"Returns a privkey filepath for give path",
			args{caDir: "dir"},
			"dir/private/ca.key.pem",
			false,
		},
		{
			"Returns an error if path is empty",
			args{caDir: ""},
			"",
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := core.GetCaPrivKeyFilePath(tt.args.caDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCaPrivKeyFilePath() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if got != tt.want {
				t.Errorf("GetCaPrivKeyFilePath() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckCaCertificateExistInDir(t *testing.T) {
	t.Parallel()

	type args struct {
		caDir           string
		dontAskRecreate []bool
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Check existing CA file in given dir",
			args{
				existingCaDir,
				[]bool{true},
			},
			true,
		},
		{
			"Check non existing CA file in given dir",
			args{
				nonExistingDir,
				[]bool{true},
			},
			false,
		},
		{
			"Check for empty given dir",
			args{
				"",
				[]bool{true},
			},
			false,
		},
		{
			"Check for recreating ca cert in given dir",
			args{
				nonExistingDir,
				[]bool{false},
			},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := core.CheckCaCertificateExistInDir(tt.args.caDir, tt.args.dontAskRecreate...)
			if got != tt.want {
				t.Errorf("CheckCaCertificateExistInDir() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateCaCertificate(t *testing.T) {
	t.Parallel()

	type args struct {
		caDir string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"Create CA Certificate inside test-ca-dir directory",
			args{"/test-ca-dir"},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := core.CreateCaCertificate(tt.args.caDir); (err != nil) != tt.wantErr {
				t.Errorf("CreateCaCertificate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
