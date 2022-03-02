//go:build windows
// +build windows

package core_test

import (
	"testing"

	reward "github.com/rewardenv/reward/internal/core"
)

func TestIsAdmin(t *testing.T) {
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
		t.Run(
			tt.name, func(t *testing.T) {
				if got := reward.IsAdmin(); got != tt.want {
					t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestRunMeElevated(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			"Default test case",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
			},
		)
	}
}
