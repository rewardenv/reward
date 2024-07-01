package logic

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/rewardenv/reward/internal/shell"
	"github.com/rewardenv/reward/internal/templates"
	"github.com/rewardenv/reward/pkg/util"
)

// RunCmdSyncStart represents the sync start command.
func (c *Client) RunCmdSyncStart() error {
	if !c.SyncEnabled() {
		return nil
	}

	if err := c.CheckAndInstallMutagen(); err != nil {
		return errors.Wrap(err, "checking mutagen installation")
	}

	if err := c.RunCmdSyncTerminate(); err != nil {
		return errors.Wrap(err, "terminating mutagen sync session")
	}

	log.Debugln("Looking up synced container...")

	containerID, err := c.Docker.ContainerIDByName(c.Config.SyncedContainer())
	if err != nil {
		return errors.Wrap(err, "looking up synced container")
	}

	log.Debugf("...synced container found: %s.", containerID)
	log.Debugln("Creating mutagen sync session...")

	// Create sync session
	// mutagen sync create -c /path/to/config/file.yml --label reward-sync=env --ignore xyz path docker://container/path
	cmd := []string{
		"mutagen", "sync", "create", "-c",
		util.Quote(c.MutagenSyncFile()),
		"--label",
		fmt.Sprintf(`%s-sync=%s`, c.AppName(), c.EnvName()),
	}

	// Append --ignore flag only if it's not empty
	if strings.TrimSpace(c.MutagenSyncIgnore()) != "" {
		cmd = append(cmd, fmt.Sprintf(`--ignore %s`, c.Config.MutagenSyncIgnore()))
	}

	// Append rest of the command line flags
	cmd = append(
		cmd,
		util.Quote(fmt.Sprintf(`%s%s`, c.Config.Cwd(), c.Config.WebRoot())),
		util.Quote(fmt.Sprintf(`docker://%s%s`, containerID, c.Config.SyncedDir())),
	)

	out, err := c.Shell.RunCommand(cmd, shell.WithSuppressOutput())
	log.Debugf("Mutagen sync start command output: %s", out)

	if err != nil {
		return errors.Wrap(err, "creating mutagen sync session")
	}

	log.Debugln("...mutagen sync session created.")
	log.Println("Waiting for sync to be ready...")

	cmd = []string{
		"mutagen", "sync", "list", "--label-selector",
		fmt.Sprintf("%s-sync=%s", c.AppName(), c.EnvName()),
	}

	for {
		out, err := c.Shell.RunCommand(
			cmd,
			shell.WithCatchOutput(),
			shell.WithSuppressOutput(),
		)
		log.Debugf("Mutagen sync list command output: %s", out)

		if err != nil {
			return errors.Wrap(err, "listing mutagen sync session")
		}

		if strings.Contains(strings.ToLower(string(out)), strings.ToLower("last error")) {
			return errors.Wrapf(err, "mutagen encountered an error: %s", out)
		}

		if strings.Contains(strings.ToLower(string(out)), strings.ToLower("watching for changes")) {
			break
		}

		time.Sleep(time.Second * 5)
	}

	log.Println("...mutagen sync is ready.")

	return nil
}

// RunCmdSyncStop represents the sync stop command.
func (c *Client) RunCmdSyncStop() error {
	if !c.SyncEnabled() {
		return nil
	}

	if err := c.CheckAndInstallMutagen(); err != nil {
		return errors.Wrap(err, "checking mutagen installation")
	}

	log.Println("Terminating mutagen sync session...")

	cmd := []string{
		"mutagen", "sync", "terminate", "--label-selector",
		fmt.Sprintf("%s-sync=%s", c.AppName(), c.EnvName()),
	}

	out, err := c.Shell.RunCommand(cmd)

	log.Debugf("Mutagen sync stop command output: %s", out)

	if err != nil {
		return errors.Wrap(err, "terminating mutagen sync session")
	}

	log.Println("...mutagen sync session terminated.")

	return nil
}

// RunCmdSyncResume represents the sync status command.
func (c *Client) RunCmdSyncResume() error {
	if !c.SyncEnabled() {
		return nil
	}

	if err := c.CheckAndInstallMutagen(); err != nil {
		return errors.Wrap(err, "checking mutagen installation")
	}

	log.Debugln("Resuming mutagen sync session...")

	cmd := []string{
		"mutagen", "sync", "resume", "--label-selector",
		fmt.Sprintf("%s-sync=%s", c.AppName(), c.EnvName()),
	}

	out, err := c.Shell.RunCommand(cmd, shell.WithSuppressOutput())

	log.Debugf("Mutagen sync resume command output: %s", out)

	if err != nil {
		return errors.Wrap(err, "resuming mutagen sync session")
	}

	log.Debugln("...mutagen sync session resumed.")

	return nil
}

// RunCmdSyncPause represents the sync pause command.
func (c *Client) RunCmdSyncPause() error {
	if !c.SyncEnabled() {
		return nil
	}

	if err := c.CheckAndInstallMutagen(); err != nil {
		return errors.Wrap(err, "checking mutagen installation")
	}

	log.Println("Pausing mutagen sync session...")

	cmd := []string{
		"mutagen", "sync", "pause", "--label-selector",
		fmt.Sprintf("%s-sync=%s", c.Config.AppName(), c.Config.EnvName()),
	}

	out, err := c.Shell.RunCommand(cmd)

	log.Debugf("Mutagen sync pause command output: %s", out)

	if err != nil {
		return errors.Wrap(err, "pausing mutagen sync session")
	}

	log.Println("...mutagen sync session paused.")

	return nil
}

// RunCmdSyncList represents the sync list command.
func (c *Client) RunCmdSyncList(opts ...shell.Opt) (string, error) {
	if !c.SyncEnabled() {
		return "", nil
	}

	if err := c.CheckAndInstallMutagen(); err != nil {
		return "", errors.Wrap(err, "checking mutagen installation")
	}

	log.Debugln("Listing mutagen sync sessions...")

	cmd := []string{
		"mutagen", "sync", "list", "--label-selector",
		fmt.Sprintf("%s-sync=%s", c.Config.AppName(), c.Config.EnvName()),
	}

	out, err := c.Shell.RunCommand(cmd, opts...)

	log.Debugf("Mutagen sync pause command output: %s", out)

	if err != nil {
		return "", errors.Wrap(err, "listing mutagen sync sessions")
	}

	log.Debugln("...mutagen sync sessions listed.")

	return string(out), nil
}

// RunCmdSyncFlush represents the sync flush command.
func (c *Client) RunCmdSyncFlush() error {
	if !c.SyncEnabled() {
		return nil
	}

	if err := c.CheckAndInstallMutagen(); err != nil {
		return errors.Wrap(err, "checking mutagen installation")
	}

	log.Println("Flushing mutagen sync session...")

	cmd := []string{
		"mutagen", "sync", "flush", "--label-selector",
		fmt.Sprintf("%s-sync=%s", c.Config.AppName(), c.Config.EnvName()),
	}

	out, err := c.Shell.RunCommand(cmd)

	log.Debugf("Mutagen sync flush command output: %s", out)

	if err != nil {
		return errors.Wrap(err, "flushing mutagen sync session")
	}

	log.Println("...mutagen sync session flushed.")

	return nil
}

// RunCmdSyncMonitor represents the sync monitor command.
func (c *Client) RunCmdSyncMonitor() error {
	if !c.SyncEnabled() {
		return nil
	}

	if err := c.CheckAndInstallMutagen(); err != nil {
		return errors.Wrap(err, "checking mutagen installation")
	}

	log.Println("Monitoring mutagen sync session...")

	cmd := []string{
		"mutagen", "sync", "monitor", "--label-selector",
		fmt.Sprintf("%s-sync=%s", c.Config.AppName(), c.Config.EnvName()),
	}

	out, err := c.Shell.RunCommand(cmd)

	log.Debugf("Mutagen sync monitor command output: %s", out)

	if err != nil {
		return errors.Wrap(err, "monitoring mutagen sync session")
	}

	log.Println("...mutagen sync session monitored.")

	return nil
}

// RunCmdSyncReset represents the sync reset command.
func (c *Client) RunCmdSyncReset() error {
	if !c.SyncEnabled() {
		return nil
	}

	if err := c.CheckAndInstallMutagen(); err != nil {
		return errors.Wrap(err, "checking mutagen installation")
	}

	log.Println("Resetting mutagen sync session...")

	cmd := []string{
		"mutagen", "sync", "reset", "--label-selector",
		fmt.Sprintf("%s-sync=%s", c.Config.AppName(), c.Config.EnvName()),
	}

	out, err := c.Shell.RunCommand(cmd)

	log.Debugf("Mutagen sync reset command output: %s", out)

	if err != nil {
		return errors.Wrap(err, "resetting mutagen sync session")
	}

	log.Println("...mutagen sync session reset.")

	return nil
}

// RunCmdSyncCheck checks if mutagen configuration is ok. If it doesn't exist, this function is going to generate one.
func (c *Client) RunCmdSyncCheck() error {
	if !c.SyncEnabled() {
		return nil
	}

	if err := c.CheckAndInstallMutagen(); err != nil {
		return errors.Wrap(err, "checking mutagen installation")
	}

	log.Debugln("Checking mutagen sync configuration...")

	if err := templates.New().GenerateMutagenTemplateFile(c.MutagenSyncFile(), c.EnvType()); err != nil {
		return errors.Wrap(err, "generating mutagen template file")
	}

	log.Debugln("...mutagen sync configuration checked.")

	return nil
}

// RunCmdSyncTerminate represents the sync start command.
func (c *Client) RunCmdSyncTerminate() error {
	if !c.SyncEnabled() {
		return nil
	}

	if err := c.CheckAndInstallMutagen(); err != nil {
		return errors.Wrap(err, "checking mutagen installation")
	}

	log.Debugln("Terminating sync session...")

	// Terminate previous sync if it ran.
	cmd := []string{
		"mutagen", "sync", "terminate",
		"--label-selector",
		fmt.Sprintf("%s-sync=%s", c.Config.AppName(), c.Config.EnvName()),
	}

	out, err := c.Shell.RunCommand(cmd)

	log.Debugf("Mutagen sync terminate command output: %s", out)

	if err != nil {
		return errors.Wrap(err, "terminating previous sync session")
	}

	log.Debugln("...successfully terminated sync sessions.")

	return nil
}

// CheckAndInstallMutagen checks if mutagen is available. If not, it's going to install mutagen.
func (c *Client) CheckAndInstallMutagen() error {
	if !c.SyncEnabled() {
		return nil
	}

	log.Debugln("Checking for mutagen...")

	if !util.CommandAvailable("mutagen") {
		if err := c.InstallMutagen(); err != nil {
			return errors.Wrap(err, "installing mutagen")
		}
	}

	log.Debugln("...mutagen is available.")
	log.Debugln("Checking mutagen version...")

	mutagenVersion, err := c.Shell.RunCommand([]string{"mutagen", "version"},
		shell.WithCatchOutput(),
		shell.WithSuppressOutput(),
	)
	if err != nil {
		return errors.Wrap(err, "getting mutagen version")
	}

	if version.Must(version.NewVersion(strings.TrimSpace(string(mutagenVersion)))).LessThan(
		version.Must(version.NewVersion(c.Config.MutagenRequiredVersion()))) {
		log.Printf(
			"Mutagen version %s or greater is required (version %s is installed).",
			c.Config.MutagenRequiredVersion(), mutagenVersion,
		)
		log.Printf("Please update Mutagen:\n  brew upgrade mutagen-io/mutagen/mutagen")
	}

	log.Debugln("...mutagen version is ok.")

	return nil
}

// InstallMutagen installs mutagen.
func (c *Client) InstallMutagen() error {
	switch util.OSDistro() {
	case "darwin":
		log.Println("Installing mutagen...")

		if util.AskForConfirmation("Mutagen could not be found; would you like to install it via Homebrew?") {
			if _, err := c.Shell.RunCommand([]string{"brew", "install", "mutagen-io/mutagen/mutagen"}); err != nil {
				return errors.Wrap(err, "installing mutagen")
			}
		}

		log.Println("...mutagen installed.")
	case "windows":
		log.Println("Installing mutagen...")

		if util.AskForConfirmation("Mutagen could not be found; would you like to download it?") {
			if err := c.InstallMutagenForWindows(); err != nil {
				return errors.Wrap(err, "installing mutagen")
			}
		}

		log.Println("...mutagen installed.")
	}

	return nil
}

// InstallMutagenForWindows installs mutagen for Windows.
func (c *Client) InstallMutagenForWindows() error {
	log.Debugln("Installing mutagen for Windows...")

	binaryPath, err := os.Executable()
	if err != nil {
		return errors.Wrap(err, "getting executable path")
	}

	installDir := filepath.Dir(binaryPath)

	log.Tracef("determined path to install mutagen: %s", installDir)

	log.Debugln("Downloading mutagen...")

	req, err := c.prepareRequest(c.Config.MutagenURL(), true)
	if err != nil {
		return errors.Wrap(err, "creating HTTP request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "downloading mutagen")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.Errorf("cannot download mutagen, http response status: %s", resp.Status)
	}

	log.Debugln("...mutagen downloaded.")
	log.Debugln("Extracting mutagen...")

	files, err := util.Unzip(resp.Body, installDir)
	if err != nil {
		return errors.Wrap(err, "extracting mutagen")
	}

	log.Tracef("Extracted files: %s", files)
	log.Debugln("...mutagen extracted.")
	log.Debugln("...mutagen installed for Windows.")

	return nil
}

// ContainerChanged returns true if the container's state is not "running" or the container's ID is changed since
// the last sync session.
func (c *Client) ContainerChanged(container string) bool {
	log.Println("Checking if synced container is changed...")
	log.Debugln("Getting container state...")

	containerState, err := c.Docker.ContainerStateByName(container)
	if err != nil {
		log.Printf("...cannot get container state: %s. Assuming the container is changed, restarting sync session.",
			err)

		return true
	}

	log.Debugf("...current synced container (%s) state is: %s.", container, containerState)
	log.Debugln("Getting container ID by Name...")

	containerID, err := c.Docker.ContainerIDByName(container)
	if err != nil {
		log.Printf("...cannot get container ID: %s. Assuming the container is changed, restarting sync session.",
			err)

		return true
	}

	log.Debugf("...current synced container (%s) ID is: %s.", container, containerID)

	if containerState != "running" {
		log.Println("...synced container is not running. Assuming the container is changed, restarting sync session.")

		return true
	}

	log.Debugln("Listing mutagen sync sessions...")

	syncListOut, err := c.RunCmdSyncList(shell.WithCatchOutput(), shell.WithSuppressOutput())
	if err != nil {
		log.Printf("...cannot list mutagen sync sessions: %s. Assuming the container is changed, restarting sync session.",
			err)

		return true
	}

	re := regexp.MustCompile(`URL: docker://[^/]*`)
	stripRe := regexp.MustCompile(`(URL: docker://)([^/\\]*)`)
	previousContainerID := stripRe.ReplaceAllString(re.FindString(syncListOut), "$2")

	log.Debugf("Previously synced container ID: %s", previousContainerID)

	if previousContainerID != containerID {
		log.Println("...synced container ID changed. Assuming the container changed, restarting sync session.")

		return true
	}

	log.Println("...synced container is not changed.")

	return false
}
