package commands

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/rewardenv/reward/internal/core"
	log "github.com/sirupsen/logrus"
)

const (
	mutagenRequiredVersion = "0.11.8"
	mutagenURL             = "https://github.com/mutagen-io/mutagen/releases/download/v0.14.0/mutagen_windows_amd64_v0.14.0.zip"
)

var syncedDir = "/var/www/html"

// GetSyncedDir returns the directory which is synchronized with mutagen.
func GetSyncedDir() string {
	return syncedDir
}

// SetSyncedDir sets the directory to be synchronized with mutagen.
func SetSyncedDir(s string) {
	syncedDir = s
}

// SetSyncSettingsByEnvType sets the settings for synchronization.
func SetSyncSettingsByEnvType() {
	if core.CheckRegexInString("^pwa-studio", core.GetEnvType()) {
		SetSyncedContainer("node")
		SetSyncedDir("/usr/src/app")
	}
}

// SyncCheck checks if mutagen configuration is ok. If it doesn't exists, this function is going to generate one.
func SyncCheck() error {
	if core.IsMutagenSyncEnabled() {
		err := CheckAndInstallMutagen()
		if err != nil {
			return err
		}

		log.Debugln("Checking mutagen version.")

		mutagenVersion, err := core.RunOsCommand([]string{"mutagen", "version"}, true)
		if err != nil {
			return err
		}

		v1, err := version.NewVersion(strings.TrimSpace(mutagenVersion))
		if err != nil {
			return err
		}

		v2, err := version.NewVersion(mutagenRequiredVersion)
		if err != nil {
			return err
		}

		log.Debugf("Mutagen version: %v.", mutagenVersion)

		if v1.LessThan(v2) {
			log.Printf(
				"Mutagen version %v or greater is required (version %v is installed).",
				mutagenRequiredVersion, mutagenVersion,
			)
			log.Printf("Please update Mutagen:\n  brew upgrade mutagen-io/mutagen/mutagen")
		}
	}

	err := core.GenerateMutagenTemplateFileIfNotExist()
	if err != nil {
		return err
	}

	return nil
}

// SyncStartCmd represents the sync start command.
func SyncStartCmd() error {
	// Terminate previous sync if it ran.
	cmd := []string{
		"mutagen", "sync", "terminate",
		"--label-selector",
		fmt.Sprintf("%v-sync=%v", core.AppName, core.GetEnvName()),
	}

	_, err := core.RunOsCommand(cmd)
	if err != nil {
		return err
	}

	containerID, err := core.GetContainerIDByName(GetSyncedContainer())
	if err != nil {
		return err
	}

	// Create sync session
	// mutagen sync create -c /path/to/config/file.yml --label reward-sync=env --ignore xyz path docker://container/path
	cmd = []string{
		"mutagen", "sync", "create", "-c",
		core.Quote(fmt.Sprintf(`%v`, core.GetMutagenSyncFile())),
		"--label",
		fmt.Sprintf(`%v-sync=%v`, core.AppName, core.GetEnvName()),
	}

	// Append --ignore flag only if it's not empty
	if strings.TrimSpace(core.GetMutagenSyncIgnore()) != "" {
		ignoreFlag := "--ignore " + core.GetMutagenSyncIgnore()
		cmd = append(cmd, fmt.Sprintf(`%v`, ignoreFlag))
	}

	// Append rest of the command line flags
	cmd = append(
		cmd,
		core.Quote(fmt.Sprintf(`%v%v`, core.GetCwd(), core.GetWebRoot())),
		core.Quote(fmt.Sprintf(`docker://%v%v`, containerID, GetSyncedDir())),
	)

	log.Println("Syncing environment with mutagen...")

	out, err := core.RunOsCommand(cmd)
	log.Debugln(out)

	if err != nil {
		return err
	}

	cmd = []string{
		"mutagen", "sync", "list", "--label-selector",
		fmt.Sprintf("%v-sync=%v", core.AppName, core.GetEnvName()),
	}

	for {
		out, err := core.RunOsCommand(cmd, true)

		log.Debugln(out)

		if err != nil {
			return err
		}

		if strings.Contains(strings.ToLower(out), strings.ToLower("last error")) {
			return fmt.Errorf("mutagen encountered an error: %v, %w", out, err)
		}

		if strings.Contains(strings.ToLower(out), strings.ToLower("watching for changes")) {
			break
		}

		time.Sleep(time.Second * 5)
	}

	return nil
}

// SyncStopCmd represents the sync stop command.
func SyncStopCmd() error {
	cmd := []string{
		"mutagen", "sync", "terminate", "--label-selector",
		fmt.Sprintf("%v-sync=%v", core.AppName, core.GetEnvName()),
	}

	_, err := core.RunOsCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

// SyncResumeCmd represents the sync resume command.
func SyncResumeCmd() error {
	cmd := []string{
		"mutagen", "sync", "resume", "--label-selector",
		fmt.Sprintf("%v-sync=%v", core.AppName, core.GetEnvName()),
	}

	_, err := core.RunOsCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

// SyncPauseCmd represents the sync pause command.
func SyncPauseCmd() error {
	cmd := []string{
		"mutagen", "sync", "pause", "--label-selector",
		fmt.Sprintf("%v-sync=%v", core.AppName, core.GetEnvName()),
	}

	_, err := core.RunOsCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

// SyncListCmd represents the sync list command.
func SyncListCmd(suppressOsStdOut ...bool) (string, error) {
	cmd := []string{
		"mutagen", "sync", "list", "--label-selector",
		fmt.Sprintf("%v-sync=%v", core.AppName, core.GetEnvName()),
	}

	out, err := core.RunOsCommand(cmd, suppressOsStdOut...)
	if err != nil {
		return "", err
	}

	return out, nil
}

// SyncFlushCmd represents the sync flush command.
func SyncFlushCmd() error {
	cmd := []string{
		"mutagen", "sync", "flush", "--label-selector",
		fmt.Sprintf("%v-sync=%v", core.AppName, core.GetEnvName()),
	}

	_, err := core.RunOsCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

// SyncMonitorCmd represents the sync monitor command.
func SyncMonitorCmd() error {
	cmd := []string{
		"mutagen", "sync", "monitor", "--label-selector",
		fmt.Sprintf("%v-sync=%v", core.AppName, core.GetEnvName()),
	}

	_, err := core.RunOsCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

// SyncResetCmd represents the sync reset command.
func SyncResetCmd() error {
	cmd := []string{
		"mutagen", "sync", "reset", "--label-selector",
		fmt.Sprintf("%v-sync=%v", core.AppName, core.GetEnvName()),
	}

	_, err := core.RunOsCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

// CheckAndInstallMutagen checks if mutagen is available. If not, it's going to install mutagen.
func CheckAndInstallMutagen() error {
	log.Debugln("Checking for mutagen.")

	if !core.IsCommandAvailable("mutagen") {
		err := InstallMutagen()
		if err != nil {
			return err
		}
	}

	return nil
}

// InstallMutagen installs mutagen.
func InstallMutagen() error {
	switch core.GetOSDistro() {
	case "darwin":
		if core.AskForConfirmation("Mutagen could not be found; would you like to install it via Homebrew?") {
			_, err := core.RunOsCommand([]string{"brew", "install", "mutagen-io/mutagen/mutagen"}, false)
			if err != nil {
				return err
			}
		}
	case "windows":
		if core.AskForConfirmation("Mutagen could not be found; would you like to download it?") {
			err := InstallMutagenForWindows()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// InstallMutagenForWindows installs mutagen for Windows.
func InstallMutagenForWindows() error {
	binaryPath, err := os.Executable()

	if err != nil {
		return err
	}

	installDir := filepath.Dir(binaryPath)

	log.Debugln("path to install:", installDir)

	req, err := http.NewRequestWithContext(context.Background(), "GET", mutagenURL, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/octet-stream")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return err
	}

	src := res.Body
	defer func(src io.ReadCloser) {
		_ = src.Close()
	}(src)

	files, err := core.Unzip(src, installDir)
	if err != nil {
		return err
	}

	log.Debugln(files)

	return nil
}

// IsContainerChanged returns true if the container's state is not "running" or the container's ID is changed since
//   the last sync session.
func IsContainerChanged(container string) bool {
	log.Debugln()
	containerState, err := core.GetContainerStateByName(container)
	if err != nil {
		return true
	}

	log.Debugf("Current container (%v) state is: %v", container, containerState)

	containerID, err := core.GetContainerIDByName(container)
	if err != nil {
		return true
	}

	log.Debugf("Current container (%v) ID is: %v", container, containerID)

	if containerState != "running" {
		return true
	}

	syncListOut, err := SyncListCmd(true)
	if err != nil {
		return true
	}

	re := regexp.MustCompile(`URL: docker://[^/]*`)
	stripRe := regexp.MustCompile(`(URL: docker://)([^/\\]*)`)
	previousContainerID := stripRe.ReplaceAllString(re.FindString(syncListOut), "$2")

	log.Debugln("Previously synced container ID:", previousContainerID)

	return previousContainerID != containerID
}
