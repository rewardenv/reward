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
	"github.com/pkg/errors"
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
				if err := c.selfUpdate(cmd); err != nil {
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
		return false, errors.Wrap(err, "fetching latest release")
	}

	if currentRelease == nil {
		return false, errors.New("cannot find latest release")
	}

	remoteVersion := version.Must(version.NewVersion(strings.TrimSpace(currentRelease.TagName))).Core()
	currentVersion := version.Must(version.NewVersion(c.AppVersion())).Core()

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
		return nil, errors.Wrap(err, "getting remote version")
	}

	var releases []*release

	if err := json.Unmarshal(remoteData, &releases); err != nil {
		return nil, errors.Wrap(err, "unmarshalling remote data")
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
		return nil, errors.Wrap(err, "running request")
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.Wrapf(http.ErrMissingFile, "downloading file from url %s", url)
	}

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading response body")
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
		return errors.Wrap(err, "finding executable path")
	}

	if util.OSDistro() == "windows" {
		binaryName = fmt.Sprintf("%s.exe", binaryName)

		if !strings.HasSuffix(binaryPath, ".exe") {
			binaryPath += ".exe"
		}
	}

	binaryPath, err = util.EvalSymlinkPath(binaryPath)
	if err != nil {
		return errors.Wrap(err, "evaluating symlink path")
	}

	updateAsset, err := c.updateURL(cmd)
	if err != nil {
		return errors.Wrap(err, "getting update url")
	}

	req, err := c.prepareRequest(updateAsset.URL, true)
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "running http request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.Errorf("%s url not found", updateAsset.Name)
	}

	newBinary, err := util.DecompressFileFromArchive(resp.Body, updateAsset.Name, binaryName)
	if err != nil {
		return err
	}

	if err := update.Apply(newBinary, update.Options{TargetPath: binaryPath}); err != nil {
		return errors.Wrap(err, "applying update")
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
		return nil, errors.Wrap(err, "fetching latest release")
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

	return nil, errors.Errorf("cannot find asset %s", packagename)
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
