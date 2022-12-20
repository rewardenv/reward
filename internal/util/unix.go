//go:build linux || darwin
// +build linux darwin

package util

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// IsAdmin returns true if the user who runs the command is root.
func IsAdmin() bool {
	uid := os.Geteuid()

	log.Tracef("Caller UID: %d.", uid)

	return uid == 0
}

// RunMeElevated does nothing on unix systems.
func RunMeElevated() {
	// But it needs to be implemented for the testing.
}
