//go:build linux || darwin
// +build linux darwin

package core_test

import (
	reward "github.com/rewardenv/reward/internal/core"
	"testing"
)

func TestIsAdmin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		{
			"Test should be run by normal user",
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := reward.IsAdmin(); got != tt.want {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunMeElevated(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			"Default test case",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
		})
	}
}
