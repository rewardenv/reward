package logic

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	currentRelease, err := c.fetchRelease(cmd, c.RepoURL())
	if err != nil {
		return false, fmt.Errorf("cannot fetch latest release: %w", err)
	}

	if currentRelease == nil {
		return false, fmt.Errorf("cannot find latest release")
	}

	remoteVersion := version.Must(version.NewVersion(strings.TrimSpace(currentRelease.TagName)))
	currentVersion := version.Must(version.NewVersion(c.AppVersion()))

	log.Printf("Current version: %s, Remote version: %s",
		currentVersion.String(),
		remoteVersion.String())

	return remoteVersion.GreaterThan(currentVersion), nil
}

func (c *Client) fetchRelease(cmd *cmdpkg.Command, url string) (*release, error) {
	pageQuery := ""
	if strings.Contains(url, "api.github.com") {
		pageQuery = "?page=1&per_page=10"
	}

	url += pageQuery

	remoteData, err := c.getContentFromURL(url)
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

func (c *Client) getContentFromURL(url string) ([]byte, error) {
	req, err := c.prepareRequest(url, false)
	if err != nil {
		return nil, err
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
	if name == "force" {
		cmd.Config.Set("assume_yes", true)
	}

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

	updateAsset, err := c.updateURL(cmd)
	if err != nil {
		return fmt.Errorf("cannot get update url: %w", err)
	}

	req, err := c.prepareRequest(updateAsset.URL, true)
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot run http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("%s URL not found", updateAsset.Name)
	}

	newBinary, err := util.DecompressFileFromArchive(resp.Body, updateAsset.Name, binaryName)
	if err != nil {
		return err
	}

	err = update.Apply(newBinary, update.Options{TargetPath: binaryPath})
	if err != nil {
		return fmt.Errorf("cannot apply update: %w", err)
	}

	return nil
}

func (c *Client) updateURL(cmd *cmdpkg.Command) (*asset, error) {
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

	release, err := c.fetchRelease(cmd, c.RepoURL())
	if err != nil {
		return nil, fmt.Errorf("cannot fetch latest release: %w", err)
	}

	// url := strings.Replace(release.HTMLURL, "/releases/tag/", "/releases/download/", 1)

	var packagename string

	switch goOS {
	case "windows":
		packagename = fmt.Sprintf("%s_%s_%s.zip",
			c.AppName(),
			replacements[goOS][goOS],
			replacements[goOS][goArch],
		)
	default:
		packagename = fmt.Sprintf("%s_%s_%s.tar.gz",
			c.AppName(),
			replacements[goOS][goOS],
			replacements[goOS][goArch],
		)
	}

	for _, asset := range release.Assets {
		if asset.Name == packagename {
			return &asset, nil
		}
	}

	return nil, fmt.Errorf("cannot find asset %s", packagename)
}

//nolint:tagliatelle
type release struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	TagName    string  `json:"tag_name"`
	Prerelease bool    `json:"prerelease"`
	URL        string  `json:"url"`
	AssetsURL  string  `json:"assets_url"`
	HTMLURL    string  `json:"html_url"`
	Assets     []asset `json:"assets"`
}

//nolint:tagliatelle
type asset struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
}
