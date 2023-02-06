package logic

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/inconshreveable/go-update"
	log "github.com/sirupsen/logrus"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/pkg/util"
)

func (c *Client) RunCmdPluginList() error {
	plugins := c.Plugins()

	if len(plugins) > 0 {
		log.Println("The following plugins are installed:")
	} else {
		log.Println("No plugins are installed.")
	}

	for _, plugin := range plugins {
		printPlugins(plugin)
	}

	return nil
}

func (c *Client) RunCmdPluginListAvailable() error {
	plugins := c.PluginsAvailable()

	if len(plugins) > 0 {
		log.Println("The following plugins are available online:")
	} else {
		log.Println("No plugins are available online.")
	}

	for _, plugin := range plugins {
		printPlugins(plugin)
	}

	return nil
}

func (c *Client) RunCmdPluginRemove(cmd *cmdpkg.Command, args []string) error {
	for _, plugin := range args {
		log.Printf("Removing plugin %s...", plugin)

		if util.AskForConfirmation(fmt.Sprintf("Would you like to remove plugin %s?", plugin)) {
			err := c.pluginRemove(plugin)
			if err != nil {
				return err
			}
		}

		log.Print("...plugin removed.")
	}

	return nil
}

func (c *Client) RunCmdPluginInstall(cmd *cmdpkg.Command, args []string) error {
	err := c.checkPlugins(args)
	if err != nil {
		return err
	}

	for _, plugin := range args {
		log.Printf("Installing plugin %s...", plugin)

		needsUpdate, err := c.pluginIsNotLatest(cmd, plugin)
		if err != nil {
			return err
		}

		if flag(cmd, "force") || needsUpdate {
			if util.AskForConfirmation(fmt.Sprintf("Would you like to install plugin %s?", plugin)) {
				err = c.pluginInstall(cmd, plugin)
				if err != nil {
					return err
				}
			}

			log.Print("...plugin installed.")
		}
	}

	return nil
}

func (c *Client) checkPlugins(args []string) error {
	for _, plugin := range args {
		if _, ok := c.PluginsAvailable()[plugin]; !ok {
			return fmt.Errorf("plugin %s is not available", plugin)
		}
	}

	return nil
}

func (c *Client) pluginIsNotLatest(cmd *cmdpkg.Command, name string) (bool, error) {
	pluginURL, err := c.pluginURL(name)
	if err != nil {
		return false, err
	}

	currentRelease, err := c.fetchRelease(cmd, pluginURL)
	if err != nil {
		return false, fmt.Errorf("cannot fetch latest release: %w", err)
	}

	if currentRelease == nil {
		return false, fmt.Errorf("cannot find latest release")
	}

	remoteVersion := version.Must(version.NewVersion(strings.TrimSpace(currentRelease.TagName)))

	pluginVersion, err := c.pluginVersion(name)
	if pluginVersion == "" || err != nil {
		log.Debugf("Cannot get plugin version. Error: %s", err)
		log.Printf("Cannot determine plugin version. Remote version: %s", remoteVersion.String())

		return true, nil
	}

	currentVersion := version.Must(version.NewVersion(pluginVersion))

	log.Printf("Current version: %s, Remote version: %s",
		currentVersion.String(),
		remoteVersion.String())

	return remoteVersion.GreaterThan(currentVersion), nil
}

func (c *Client) pluginInstall(cmd *cmdpkg.Command, name string) error {
	binaryName := fmt.Sprintf("%s-%s", c.AppName(), name)

	binaryPath := filepath.Join(c.PluginsDir(), binaryName)

	if util.OSDistro() == "windows" {
		binaryName = fmt.Sprintf("%s.exe", binaryName)

		if !strings.HasSuffix(binaryPath, ".exe") {
			binaryPath += ".exe"
		}
	}

	symlinkPath, _ := util.EvalSymlinkPath(binaryPath)
	if symlinkPath != "" {
		binaryPath = symlinkPath
	}

	asset, err := c.pluginNormalizedURL(cmd, name)
	if err != nil {
		return fmt.Errorf("cannot get update url: %w", err)
	}

	fileURL, err := url.Parse(asset.URL)
	if err != nil {
		return fmt.Errorf("cannot parse url: %w", err)
	}

	req, err := c.prepareRequest(asset.URL, true)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot run request: response %s, error: %w", resp.Body, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("%s URL not found", fileURL.String())
	}

	newBinary, err := util.DecompressFileFromArchive(resp.Body, asset.Name, binaryName)
	if err != nil {
		return err
	}

	_, err = os.Open(binaryPath)
	if errors.Is(err, os.ErrNotExist) {
		_ = util.CreateDirAndWriteToFile([]byte{}, binaryPath)
	}

	err = update.Apply(newBinary, update.Options{TargetPath: binaryPath})
	if err != nil {
		return fmt.Errorf("cannot apply update: %w", err)
	}

	return nil
}

func (c *Client) prepareRequest(downloadURL string, binary bool) (*http.Request, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	if c.GitHubToken() != "" {
		req.Header.Add("Authorization", fmt.Sprintf("token %s", c.GitHubToken()))
	}

	if binary {
		req.Header.Set("Accept", "application/octet-stream")

		return req, nil
	}

	req.Header.Add("Accept", "application/json")

	return req, nil
}

func (c *Client) pluginRemove(name string) error {
	binaryName := fmt.Sprintf("%s-%s", c.AppName(), name)
	binaryPath := filepath.Join(c.PluginsDir(), binaryName)

	if util.OSDistro() == "windows" {
		if !strings.HasSuffix(binaryPath, ".exe") {
			binaryPath += ".exe"
		}
	}

	err := os.Remove(binaryPath)
	if err != nil {
		return fmt.Errorf("cannot remove file: %w", err)
	}

	return nil
}

func (c *Client) pluginURL(name string) (string, error) {
	plugin, ok := c.PluginsAvailable()[name]
	if !ok {
		return "", fmt.Errorf("plugin %s is not available", name)
	}

	return plugin.URL, nil
}

func (c *Client) pluginNormalizedURL(cmd *cmdpkg.Command, name string) (*asset, error) {
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

	pluginURL, err := c.pluginURL(name)
	if err != nil {
		return nil, err
	}

	release, err := c.fetchRelease(cmd, pluginURL)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch latest release: %w", err)
	}

	var packagename string

	switch goOS {
	case "windows":
		packagename = fmt.Sprintf("%s-%s_%s_%s.zip",
			c.AppName(),
			name,
			replacements[goOS][goOS],
			replacements[goOS][goArch],
		)
	default:
		packagename = fmt.Sprintf("%s-%s_%s_%s.tar.gz",
			c.AppName(),
			name,
			replacements[goOS][goOS],
			replacements[goOS][goArch],
		)
	}

	for _, asset := range release.Assets {
		if asset.Name == packagename {
			return &asset, nil
		}
	}

	return nil, fmt.Errorf("cannot find asset %s", name)
}

func (c *Client) pluginVersion(name string) (string, error) {
	var combinedOutBuf bytes.Buffer

	//nolint:gosec
	cmd := exec.Command(filepath.Join(c.PluginsDir(), fmt.Sprintf("%s-%s", c.AppName(), name)))
	cmd.Args = append(cmd.Args, "--version")
	cmd.Stdout = io.Writer(&combinedOutBuf)
	cmd.Stderr = io.Writer(&combinedOutBuf)
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run command: %w", err)
	}

	out := strings.Split(strings.TrimSpace(combinedOutBuf.String()), " ")

	return out[len(out)-1], nil
}

func printPlugins(plugin *config.Plugin) {
	tabs := "\t\t\t"

	switch x := len(filepath.Base(plugin.Name)); {
	case x >= 14:
		tabs = strings.TrimPrefix(tabs, "\t\t")
	case x >= 6:
		tabs = strings.TrimPrefix(tabs, "\t")
	}

	//nolint:forbidigo
	fmt.Printf("- %s%s%s\n", filepath.Base(plugin.Name), tabs, plugin.Description)
}
