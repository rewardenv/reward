package logic

import (
	"context"
	"encoding/json"
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

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/pkg/util"
)

// RunCmdSelfUpdate represents the self-update command.
func (c *Client) RunCmdSelfUpdate(cmd *cmdpkg.Command) error {
	needsUpdate, err := c.isNotLatest(cmd)
	if err != nil {
		return err
	}

	if needsUpdate || flag(cmd, "force") {
		log.Println("Your version is not the latest.")

		if !flag(cmd, "dry-run") {
			if util.AskForConfirmation("Would you like to update?") {
				err = c.selfUpdate(cmd)
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

func (c *Client) isNotLatest(cmd *cmdpkg.Command) (bool, error) {
	currentRelease, err := c.fetchRelease(cmd)
	if err != nil {
		return false, fmt.Errorf("cannot fetch latest release: %w", err)
	}

	remoteVersion := version.Must(version.NewVersion(strings.TrimSpace(currentRelease.TagName)))
	currentVersion := version.Must(version.NewVersion(c.AppVersion()))

	log.Printf("Current Version: %s, Remote Version: %s",
		currentVersion.String(),
		remoteVersion.String())

	return remoteVersion.GreaterThan(currentVersion), nil
}

func (c *Client) fetchRelease(cmd *cmdpkg.Command) (*release, error) {
	remoteData, err := getContentFromURL(c.RepoURL())
	if err != nil {
		return nil, fmt.Errorf("cannot get remote version: %w", err)
	}

	var releases []*release

	err = json.Unmarshal(remoteData, &releases)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal remote data: %w", err)
	}

	var currentRelease *release

out:
	for _, r := range releases {
		switch {
		case flag(cmd, "prerelease"):
			currentRelease = r

			break out
		case !flag(cmd, "prerelease") && !r.Prerelease:
			currentRelease = r

			break out
		default:
			continue
		}
	}

	return currentRelease, nil
}

func getContentFromURL(url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot run request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("cannot download file from url %s: %w", url, http.ErrMissingFile)
	}

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response body: %w", err)
	}

	return out, nil
}

func flag(cmd *cmdpkg.Command, name string) bool {
	flag, err := cmd.Flags().GetBool(name)
	if err != nil {
		return false
	}

	return flag
}

func (c *Client) selfUpdate(cmd *cmdpkg.Command) error {
	binaryName := c.AppName()

	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find executable path: %w", err)
	}

	if util.OSDistro() == "windows" {
		binaryName = fmt.Sprintf("%s.exe", binaryName)

		if !strings.HasSuffix(binaryPath, ".exe") {
			binaryPath += ".exe"
		}
	}

	binaryPath, err = util.EvalSymlinkPath(binaryPath)
	if err != nil {
		return fmt.Errorf("cannot evaluate symlink path: %w", err)
	}

	updateURL, err := c.updateURL(cmd)
	if err != nil {
		return fmt.Errorf("cannot get update url: %w", err)
	}

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

func (c *Client) updateURL(cmd *cmdpkg.Command) (string, error) {
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

	release, err := c.fetchRelease(cmd)
	if err != nil {
		return "", fmt.Errorf("cannot fetch latest release: %w", err)
	}

	url := strings.Replace(release.HTMLURL, "/releases/tag/", "/releases/download/", 1)

	switch goOS {
	case "windows":
		return url + "/" + c.AppName() + "_" + replacements[goOS][goOS] + "_" + replacements[goOS][goArch] + ".zip", nil
	default:
		return url + "/" + c.AppName() + "_" + replacements[goOS][goOS] + "_" + replacements[goOS][goArch] + ".tar.gz", nil
	}
}

//nolint:tagliatelle
type release struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	TagName    string `json:"tag_name"`
	Prerelease bool   `json:"prerelease"`
	URL        string `json:"url"`
	AssetsURL  string `json:"assets_url"`
	HTMLURL    string `json:"html_url"`
}
