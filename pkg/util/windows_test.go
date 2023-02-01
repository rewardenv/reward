package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAdmin(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsAdmin(), "IsAdmin()")
		})
	}
}

func TestRunMeElevated(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RunMeElevated()
		})
	}
}
