package logic

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

	"reward/internal/util"
)

// RunCmdSelfUpdate represents the self-update command.
func (c *Client) RunCmdSelfUpdate(cmd *cobra.Command) error {
	needsUpdate, err := c.isNotLatest()
	if err != nil {
		return err
	}

	if needsUpdate || force(cmd) {
		log.Println("Your version is not the latest.")

		if !dryRun(cmd) {
			if util.AskForConfirmation("Would you like to update?") {
				err = c.selfUpdate()
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

func (c *Client) isNotLatest() (bool, error) {
	remoteData, err := getContentFromURL(c.RepoURL() + "/VERSION.txt")
	if err != nil {
		return true, fmt.Errorf("cannot get remote version: %w", err)
	}

	remoteVersion, err := version.NewVersion(strings.TrimSpace(remoteData))
	if err != nil {
		return true, fmt.Errorf("cannot parse remote version: %w", err)
	}

	currentVersion := version.Must(version.NewVersion(c.AppVersion()))

	log.Printf("Current Version: %s, Remote Version: %s",
		currentVersion.String(),
		remoteVersion.String())

	return remoteVersion.GreaterThan(currentVersion), nil
}

func getContentFromURL(url string) (string, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot run request: %w", err)
	}

	// defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("cannot download file from url %s: %w", url, http.ErrMissingFile)
	}

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read response body: %w", err)
	}

	return string(out), nil
}

func force(cmd *cobra.Command) bool {
	force, err := cmd.Flags().GetBool("force")
	if force || err != nil {
		return true
	}

	return false
}

func dryRun(cmd *cobra.Command) bool {
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if dryRun || err != nil {
		return true
	}

	return false
}

func (c *Client) selfUpdate() error {
	binaryName := c.AppName()

	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find executable path: %w", err)
	}

	if util.OSDistro() == "windows" {
		binaryName = fmt.Sprintf("%s.exe", binaryName)

		if !strings.HasSuffix(binaryPath, ".exe") {
			binaryPath = binaryPath + ".exe"
		}
	}

	binaryPath, err = util.EvalSymlinkPath(binaryPath)
	if err != nil {
		return fmt.Errorf("cannot evaluate symlink path: %w", err)
	}

	updateURL := c.updateURL(c.RepoURL())

	fileURL, err := url.Parse(updateURL)
	if err != nil {
		return fmt.Errorf("cannot parse url: %w", err)
	}

	path := fileURL.Path
	segments := strings.Split(path, "/")
	fileName := segments[len(segments)-1]

	req, err := http.NewRequestWithContext(context.Background(), "GET", updateURL, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/octet-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot run request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("%s URL not found", fileURL.String())
	}

	newBinary, err := util.DecompressFileFromArchive(resp.Body, fileName, binaryName)
	if err != nil {
		return err
	}

	err = update.Apply(newBinary, update.Options{TargetPath: binaryPath})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) updateURL(url string) string {
	replacements := map[string]map[string]string{
		"darwin": {
			"darwin": "Darwin",
			"arm64":  "arm64",
			"amd64":  "x86_64",
		},
		"linux": {
			"linux": "Linux",
			"arm64": "aarch64",
			"amd64": "x86_64",
			"386":   "i386",
		},
		"windows": {
			"windows": "Windows",
			"amd64":   "x86_64",
			"386":     "i386",
		},
	}

	goOS := runtime.GOOS
	goArch := runtime.GOARCH

	switch goOS {
	case "windows":
		return url + "/" + c.AppName() + "_" + replacements[goOS][goOS] + "_" + replacements[goOS][goArch] + ".zip"
	default:
		return url + "/" + c.AppName() + "_" + replacements[goOS][goOS] + "_" + replacements[goOS][goArch] + ".tar.gz"
	}
}
