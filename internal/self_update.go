package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/inconshreveable/go-update"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const RepoURL = "https://github.com/rewardenv/reward/releases/latest/download"

func SelfUpdateCmd(cmd *cobra.Command) error {
	forceUpdate := isForce(cmd)

	needsUpdate, err := isNotLatest()
	if err != nil {
		return err
	}

	if needsUpdate || forceUpdate {
		log.Println("Your version is not the latest.")

		if !isDryRun(cmd) {
			if isAssumeYes(cmd) || AskForConfirmation("Would you like to update?") {
				err = selfUpdate()
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	log.Println("Your version is the latest.")

	return nil
}

func isNotLatest() (bool, error) {
	remoteData, err := getContentFromURL(RepoURL + "/VERSION.txt")
	if err != nil {
		return true, err
	}

	remoteVersion, err := version.NewVersion(strings.TrimSpace(remoteData))
	if err != nil {
		return true, err
	}

	currentVersion := GetAppVersion()

	log.Printf("Current Version: %v, Remote Version: %v", currentVersion.String(), remoteVersion.String())

	return remoteVersion.GreaterThan(currentVersion), nil
}

func getContentFromURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	log.Debugln(url, "status: ", resp.Status)

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("%w: %v", http.ErrMissingFile, url)
	}

	buf := new(strings.Builder)

	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func isForce(cmd *cobra.Command) bool {
	force, err := cmd.Flags().GetBool("force")
	if force || err != nil {
		return true
	}

	return false
}
func isDryRun(cmd *cobra.Command) bool {
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if dryRun || err != nil {
		return true
	}

	return false
}

func isAssumeYes(cmd *cobra.Command) bool {
	yes, err := cmd.Flags().GetBool("assume-yes")
	if yes || err != nil {
		return true
	}

	return false
}

func selfUpdate() error {
	binaryName := AppName
	binaryPath, err := os.Executable()

	if err != nil {
		return err
	}

	if GetOSDistro() == "windows" {
		binaryName = binaryName + ".exe"

		if !strings.HasSuffix(binaryPath, ".exe") {
			binaryPath = binaryPath + ".exe"
		}
	}

	log.Debugln("binary path:", binaryPath)

	binaryPath, err = EvalSymlinkPath(binaryPath)
	if err != nil {
		return err
	}

	log.Debugln("resolved binary path:", binaryPath)

	updateURL := getUpdateURL(RepoURL)
	fileURL, err := url.Parse(updateURL)

	if err != nil {
		return err
	}

	path := fileURL.Path
	segments := strings.Split(path, "/")
	fileName := segments[len(segments)-1]

	req, err := http.NewRequestWithContext(context.Background(), "GET", updateURL, nil)
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

	newBinary, err := decompressFileFromArchive(src, fileName, binaryName)
	if err != nil {
		return err
	}

	err = update.Apply(newBinary, update.Options{TargetPath: binaryPath})
	if err != nil {
		return err
	}

	return nil
}

func getUpdateURL(url string) string {
	replacements := map[string]string{
		"darwin":  "Darwin",
		"linux":   "Linux",
		"windows": "Windows",
		"386":     "i386",
		"amd64":   "x86_64",
	}

	goOS := runtime.GOOS
	goArch := runtime.GOARCH

	switch goOS {
	case "windows":
		return url + "/" + AppName + "_" + replacements[goOS] + "_" + replacements[goArch] + ".zip"
	default:
		return url + "/" + AppName + "_" + replacements[goOS] + "_" + replacements[goArch] + ".tar.gz"
	}
}
