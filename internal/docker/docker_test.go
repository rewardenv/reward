package docker

import (
	"testing"

	dockerpkg "github.com/docker/docker/client"
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DockerTestSuite struct {
	suite.Suite
}

func (suite *DockerTestSuite) SetupTest() {
}

func TestDockerTestSuite(t *testing.T) {
	suite.Run(t, new(DockerTestSuite))
}

func (suite *DockerTestSuite) TestClient_dockerVersion() {
	tests := []struct {
		name     string
		optional bool
		want     *version.Version
		wantErr  bool
	}{
		{
			name:     "test",
			optional: true,
			want:     version.Must(version.NewVersion("20.10.00")),
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			c := Must(NewClient(""))

			got, err := c.dockerVersion()
			if (err != nil) != tt.wantErr {
				if tt.optional && dockerpkg.IsErrConnectionFailed(err) {
					t.Skipf("skipping test: %s", err)
				}
				t.Errorf("dockerVersion() error = %s, wantErr %t", err, tt.wantErr)

				return
			}

			assert.GreaterOrEqual(t, got.String(), tt.want.String())
		})
	}
}

func (suite *DockerTestSuite) TestClient_isMinimumVersionInstalled() {
	// Change the requiredVersion of Docker to a fake the behaviour of the test.
	tests := []struct {
		name            string
		optional        bool
		requiredVersion string
		want            bool
	}{
		{
			name:            "should be valid (using the real required docker version)",
			optional:        true,
			requiredVersion: "20.4.0",
			want:            true,
		},
		{
			name:            "should fail (using a fake required docker version)",
			optional:        true,
			requiredVersion: "30.1.1",
			want:            false,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			requiredVersion = tt.requiredVersion
			c := Must(NewClient(""))

			// assert.Equal(t, c.isMinimumVersionInstalled(), tt.want)
			assert.Condition(t, func() bool {
				if c.isMinimumVersionInstalled() == tt.want {
					return true
				}
				if tt.optional {
					t.Skipf("skipping test: %s", tt.name)
				}
				return false
			}, "isMinimumVersionInstalled() = %t, want %t", c.isMinimumVersionInstalled(), tt.want)
		})
	}
}

func (suite *DockerTestSuite) TestClient_Check() {
	// Change the requiredVersion of Docker to a fake the behaviour of the test.
	tests := []struct {
		name            string
		optional        bool
		requiredVersion string
		wantErr         bool
	}{
		{
			name:            "should be valid (using the real required docker version)",
			optional:        true,
			requiredVersion: "20.4.0",
			wantErr:         false,
		},
		{
			name:            "should fail (using a fake required docker version)",
			optional:        true,
			requiredVersion: "30.1.1",
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			requiredVersion = tt.requiredVersion
			c := Must(NewClient(""))

			if err := c.Check(); (err != nil) != tt.wantErr {
				if tt.optional {
					t.Skipf("skipping test: %s", err)
				}
				assert.Failf(t, "Check() error = %s, wantErr %s", err.Error(), tt.wantErr)
			}
		})
	}
}
