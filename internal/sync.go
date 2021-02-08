package internal

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
)

const (
	mutagenRequiredVersion = "0.11.8"
)

var syncedDir = "/var/www/html"

func GetSyncedDir() string {
	return syncedDir
}
func SetSyncedDir(s string) {
	syncedDir = s
}

func SetSyncVarsByEnvType() {
	if CheckRegexInString("^pwa-studio", GetEnvType()) {
		SetSyncedContainer("node")
		SetSyncedDir("/usr/src/app")
	}
}

func SyncCheck() error {
	if IsMutagenSyncEnabled() {
		err := CheckAndInstallMutagen()
		if err != nil {
			return err
		}

		log.Debugln("Checking mutagen version.")

		mutagenVersion, err := RunOsCommand("mutagen version", true)
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
			log.Printf("Mutagen version %v or greater is required (version %v is installed).",
				mutagenRequiredVersion, mutagenVersion)
			log.Printf("Please update Mutagen:\n  brew upgrade havoc-io/mutagen/mutagen")
		}
	}

	err := GenerateMutagenTemplateFileIfNotExist()
	if err != nil {
		return err
	}

	return nil
}

func SyncStartCmd() error {
	// Terminate previous sync if it ran.
	cmd := fmt.Sprintf("mutagen sync terminate --label-selector %v-sync=%v", AppName, GetEnvName())

	_, err := RunOsCommand(cmd)
	if err != nil {
		return err
	}

	containerID, err := GetContainerIDByName(GetSyncedContainer())
	if err != nil {
		return err
	}

	ignoreFlag := ""
	if strings.TrimSpace(GetMutagenSyncIgnore()) != "" {
		ignoreFlag = "--ignore " + GetMutagenSyncIgnore()
	}
	// Create sync session
	cmd = fmt.Sprintf(
		"mutagen sync create -c %v --label %v-sync=%v %v %v%v docker://%v%v",
		GetMutagenSyncFile(),
		AppName,
		GetEnvName(),
		ignoreFlag,
		GetCwd(),
		GetWebRoot(),
		containerID,
		GetSyncedDir(),
	)

	log.Println("Syncing environment with mutagen...")

	out, err := RunOsCommand(cmd)
	log.Debugln(out)

	if err != nil {
		return err
	}

	cmd = fmt.Sprintf(
		"mutagen sync list --label-selector %v-sync=%v",
		AppName,
		GetEnvName())

	for {
		out, err := RunOsCommand(cmd, true)

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

func SyncStopCmd() error {
	cmd := fmt.Sprintf("mutagen sync terminate --label-selector %v-sync=%v", AppName, GetEnvName())

	_, err := RunOsCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

func SyncResumeCmd() error {
	cmd := fmt.Sprintf("mutagen sync resume --label-selector %v-sync=%v", AppName, GetEnvName())

	_, err := RunOsCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

func SyncPauseCmd() error {
	cmd := fmt.Sprintf("mutagen sync pause --label-selector %v-sync=%v", AppName, GetEnvName())

	_, err := RunOsCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

func SyncListCmd(suppressOsStdOut ...bool) (string, error) {
	cmd := fmt.Sprintf("mutagen sync list --label-selector %v-sync=%v", AppName, GetEnvName())

	out, err := RunOsCommand(cmd, suppressOsStdOut...)
	if err != nil {
		return "", err
	}

	return out, nil
}

func SyncFlushCmd() error {
	cmd := fmt.Sprintf("mutagen sync flush --label-selector %v-sync=%v", AppName, GetEnvName())

	_, err := RunOsCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

func SyncMonitorCmd() error {
	cmd := fmt.Sprintf("mutagen sync monitor --label-selector %v-sync=%v", AppName, GetEnvName())

	_, err := RunOsCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

func SyncResetCmd() error {
	cmd := fmt.Sprintf("mutagen sync reset --label-selector %v-sync=%v", AppName, GetEnvName())

	_, err := RunOsCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

func CheckAndInstallMutagen() error {
	log.Debugln("Checking for mutagen.")

	if !IsCommandAvailable("mutagen") {
		err := InstallMutagen()
		if err != nil {
			return err
		}
	}

	return nil
}

func InstallMutagen() error {
	switch GetOSDistro() {
	case "darwin":
		if AskForConfirmation("Mutagen could not be found; would you like to install it via Homebrew?") {
			_, err := RunOsCommand("brew install havoc-io/mutagen/mutagen", false)
			if err != nil {
				return err
			}
		}
	case "windows":
		if AskForConfirmation("Mutagen could not be found; would you like to download it?") {
			err := InstallMutagenForWindows()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func InstallMutagenForWindows() error {
	const mutagenURL = "https://github.com/mutagen-io/mutagen/releases/download/v0.11.8/mutagen_windows_amd64_v0.11.8.zip"

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
	defer src.Close()

	files, err := Unzip(src, installDir)
	if err != nil {
		return err
	}

	log.Debugln(files)

	return nil
}
