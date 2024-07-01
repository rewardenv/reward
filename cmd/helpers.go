package cmd

import (
	"fmt"
	logpkg "log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/pkg/util"
)

type Command struct {
	*cobra.Command
	Config *config.Config
}

func (c *Command) AddCommands(commands ...*Command) {
	for _, command := range commands {
		c.AddCommand(command.Command)
	}
}

func (c *Command) AddGroups(title string, cmds ...*Command) {
	g := &cobra.Group{
		Title: title,
		ID:    title,
	}

	c.AddGroup(g)

	for _, cmd := range cmds {
		cmd.GroupID = g.ID
		c.AddCommands(cmd)
	}
}

func (c *Command) AddPlugins() {
	plugins := make([]*Command, len(c.Config.Plugins()))

	for i, plugin := range c.Config.Plugins() {
		plugins[i] = NewCmdPlugin(filepath.Base(plugin.Path), plugin.Description)
	}

	c.AddGroups("Plugins:", plugins...)
}

// DefaultSubCommandRun prints a command's help string to the specified output if no
// arguments (sub-commands) are provided, or a usage error otherwise.
func DefaultSubCommandRun() func(c *cobra.Command, args []string) {
	return func(c *cobra.Command, args []string) {
		c.SetOut(logpkg.Writer())
		c.SetErr(logpkg.Writer())
		RequireNoArguments(c, args)
		_ = c.Help()
	}
}

// RequireNoArguments exits with a usage error if extra arguments are provided.
func RequireNoArguments(c *cobra.Command, args []string) {
	if len(args) > 0 {
		log.Println(UsageErrorf(c, "unknown command %q", strings.Join(args, " ")))
	}
}

func UsageErrorf(cmd *cobra.Command, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)

	return errors.Errorf("%s\nSee '%s -h' for help and examples", msg, cmd.CommandPath())
}

func Run(executablePath string, cmdArgs, environment []string) error {
	cmd := Cmnd(executablePath, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = environment

	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, "running command")
	}

	return nil
}

func Execute(executablePath string, cmdArgs, environment []string) error {
	// Windows does not support exec syscall.
	if runtime.GOOS == "windows" {
		cmd := Cmnd(executablePath, cmdArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = environment

		err := cmd.Run()
		if err == nil {
			os.Exit(0)
		}

		return errors.Wrap(err, "running command")
	}

	// invoke cmd binary relaying the environment and args given
	// append executablePath to cmdArgs, as execve will make first argument the "binary name".
	err := syscall.Exec(executablePath, append([]string{executablePath}, cmdArgs...), environment) //nolint:gosec
	if err != nil {
		return errors.Wrap(err, "executing command")
	}

	return nil
}

func Cmnd(name string, arg ...string) *exec.Cmd {
	cmd := &exec.Cmd{
		Path: name,
		Args: append([]string{name}, arg...),
	}

	if filepath.Base(name) == name {
		lp, _ := exec.LookPath(name)
		if lp != "" {
			// Update cmd.Path even if err is non-nil.
			// If err is ErrDot (especially on Windows), lp may include a resolved
			// extension (like .exe or .bat) that should be preserved.
			cmd.Path = lp
		}
	}

	return cmd
}

// HandlePluginCommand receives a pluginHandler and command-line arguments and attempts to find
// a plugin executable on the PATH that satisfies the given arguments.
func (c *Command) HandlePluginCommand(cmdArgs []string) error {
	//nolint:prealloc
	var remainingArgs []string // all "non-flag" arguments

	log.Tracef("cmdArgs: %s", cmdArgs)

	for i, arg := range cmdArgs {
		if i == 0 && strings.HasPrefix(arg, "-") {
			break
		}
		// remainingArgs = append(remainingArgs, strings.Replace(arg, "-", "_", -1))
		remainingArgs = append(remainingArgs, arg)
	}

	log.Tracef("remainingArgs: %s", remainingArgs)

	if len(remainingArgs) == 0 {
		// the length of cmdArgs is at least 1
		return errors.Errorf("flags cannot be placed before plugin name: %s", cmdArgs[0])
	}

	foundBinaryPath := ""

out:
	for i := range remainingArgs {
		plugins := c.Config.Plugins()

		log.Tracef("evaluating args: %s", remainingArgs[i])

		log.Tracef("plugins: %s", plugins)

		for _, plugin := range plugins {
			if filepath.Base(plugin.Name) == remainingArgs[i] {
				remainingArgs = remainingArgs[i+1:]

				foundBinaryPath = filepath.Join(c.Config.PluginsDir(),
					fmt.Sprintf("%s-%s", c.Config.AppName(), plugin.Name))

				break out
			}
		}

		lookupFileInPath := fmt.Sprintf("%s-%s", c.Config.AppName(), remainingArgs[i])
		if util.OSDistro() == "windows" {
			lookupFileInPath += ".exe"
		}

		log.Tracef("looking up %s in PATH", lookupFileInPath)

		path, _ := exec.LookPath(lookupFileInPath)
		if path != "" {
			remainingArgs = remainingArgs[i+1:]
			foundBinaryPath = path

			break out
		}
	}

	if len(foundBinaryPath) == 0 {
		return nil
	}

	log.Tracef("found binary path: %s", foundBinaryPath)

	// invoke cmd binary relaying the current environment and args given
	if err := Execute(foundBinaryPath, remainingArgs, os.Environ()); err != nil {
		return errors.Errorf("executing plugin %q: %w", foundBinaryPath, err)
	}

	return nil
}

func NewCmdPlugin(name, description string) *Command {
	return &Command{
		Command: &cobra.Command{
			Use:   name,
			Short: description,
			Long:  description,
			Run:   DefaultSubCommandRun(),
		},
	}
}
