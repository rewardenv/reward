// +build windows

package internal_test

import (
	"testing"

	. "reward/internal"
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
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAdmin(); got != tt.want {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
			}
		})
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
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}
