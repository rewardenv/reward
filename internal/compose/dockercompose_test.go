package compose

import (
	"container/list"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/suite"

	"github.com/rewardenv/reward/internal/shell"
)

type DockerComposeTestSuite struct {
	suite.Suite
}

func (suite *DockerComposeTestSuite) SetupTest() {
}

func TestDockerComposeTestSuite(t *testing.T) {
	suite.Run(t, new(DockerComposeTestSuite))
}

func (suite *DockerComposeTestSuite) TestClient_Version() {
	tests := []struct {
		name    string
		want    *version.Version
		wantErr bool
	}{
		{
			name:    "test version with mock shell client",
			want:    version.Must(version.NewVersion("1.25.0")),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			c := NewClient(shell.NewLocalShellWithOpts(), list.New())
			got, err := c.Version()
			if (err != nil) != tt.wantErr {
				t.Errorf("Version() error = %s, wantErr %t", err, tt.wantErr)

				return
			}

			if tt.want.GreaterThan(got) {
				t.Errorf("Version() got = %s, want %s", got, tt.want)
			}
		})
	}
}

func (suite *DockerComposeTestSuite) TestMockClient_Version() {
	tests := []struct {
		name    string
		want    *version.Version
		wantErr bool
	}{
		{
			name:    "test version with mock shell client",
			want:    version.Must(version.NewVersion("2.13.0")),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			c := NewClient(shell.NewMockShell("", []byte("2.13.0"), nil), list.New())
			got, err := c.Version()
			if (err != nil) != tt.wantErr {
				t.Errorf("Version() error = %s, wantErr %t", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Version() got = %s, want %s", got, tt.want)
			}
		})
	}
}

func (suite *DockerComposeTestSuite) TestClient_minimumVersionInstalled() {
	type fields struct {
		Shell shell.Shell
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "version below required version",
			fields: fields{
				shell.NewMockShell("", []byte("1.0"), nil),
			},
			want: false,
		},
		{
			name: "version above required version",
			fields: fields{
				shell.NewMockShell("", []byte("2.0.0"), nil),
			},
			want: true,
		},
		{
			name: "version is not valid",
			fields: fields{
				shell.NewMockShell("", []byte("not-a-valid-version"), nil),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			c := &Client{
				Shell: tt.fields.Shell,
			}
			if got := c.isMinimumVersionInstalled(); got != tt.want {
				t.Errorf("minimumVersionInstalled() = %t, want %t", got, tt.want)
			}
		})
	}
}

func (suite *DockerComposeTestSuite) TestClient_Check() {
	type fields struct {
		Shell shell.Shell
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				shell.NewMockShell("", []byte("2.13.0"), nil),
			},
			wantErr: false,
		},
		{
			name: "in valid version",
			fields: fields{
				shell.NewMockShell("", []byte("1.24.0"), nil),
			},
			wantErr: true,
		},
		{
			name: "docker compose is not installed",
			fields: fields{
				shell.NewMockShell("", []byte("1.0.0"), fmt.Errorf("command not found")),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			c := &Client{
				Shell: tt.fields.Shell,
			}
			if err := c.Check(); (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %s, wantErr %t", err, tt.wantErr)
			}
		})
	}
}

func (suite *DockerComposeTestSuite) TestClient_RunCommand() {
	// Cannot run in parallel execution as it uses the os.stdout
	type fields struct {
		Shell shell.Shell
	}

	type args struct {
		args []string
		opts []shell.Opt
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "using mock shell client",
			fields: fields{
				shell.NewMockShell("", []byte("2.13.0"), nil),
			},
			args: args{
				args: []string{"--version", "--short"},
			},
			want:    []byte("2.13.0"),
			wantErr: false,
		},
		{
			name: "using real shell client",
			fields: fields{
				Shell: &shell.LocalShell{},
			},
			args: args{
				args: []string{"version", "--short"},
				opts: []shell.Opt{shell.WithCatchOutput(true)},
			},
			want:    []byte("1.25.0"),
			wantErr: false,
		},
		{
			name: "using real shell client with suppress stdout opt",
			fields: fields{
				Shell: &shell.LocalShell{},
			},
			args: args{
				args: []string{"version", "--short"},
				opts: []shell.Opt{shell.WithCatchOutput(true)},
			},
			want:    []byte("1.25.0"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			rescueStdout := os.Stdout
			//nolint:varnamelen
			r, w, _ := os.Pipe()
			os.Stdout = w

			c := &Client{
				Shell: tt.fields.Shell,
			}
			got, err := c.RunCommand(tt.args.args, tt.args.opts...)

			w.Close()

			if (err != nil) != tt.wantErr {
				t.Errorf("RunCommand() error = %s, wantErr %t", err, tt.wantErr)

				return
			}

			if version.Must(version.NewVersion(strings.TrimSpace(string(tt.want)))).
				GreaterThan(version.Must(version.NewVersion(strings.TrimSpace(string(got))))) {
				t.Errorf("Version() got = %s, want %s", got, tt.want)
			}

			if len(tt.args.opts) > 0 && reflect.DeepEqual(tt.args.opts[0], shell.WithCatchOutput(true)) {
				out, _ := io.ReadAll(r)
				// The io.ReadAll() function will return an empty byte slice but the want variable is nil
				// if !reflect.DeepEqual(out, tt.want) {
				if len(out) != len(tt.want) {
					t.Errorf("ExecuteWithOptions() got = %s, want %s", out, tt.want)
				}
			}

			os.Stdout = rescueStdout
		})
	}
}
