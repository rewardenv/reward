//go:build linux || darwin
// +build linux darwin

package core

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// IsAdmin returns true if the user who runs the command is root.
func IsAdmin() bool {
	uid := os.Geteuid()

	log.Debugf("uid: %v", uid)

	return uid == 0
}

// RunMeElevated do nothing on unix systems.
func RunMeElevated() {
}
