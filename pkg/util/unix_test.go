package util

import (
	"os/user"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (suite *UtilTestSuite) TestIsAdmin() {
	usr, _ := user.Current()

	tests := []struct {
		name string
		want bool
	}{
		{
			name: "running as non root should return false",
			want: usr.Uid == "0",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsAdmin(), "IsAdmin()")
		})
	}
}

func (suite *UtilTestSuite) TestRunMeElevated() {
	tests := []struct {
		name string
	}{
		{},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			RunMeElevated()
		})
	}
}
