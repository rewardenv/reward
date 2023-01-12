package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"reward/internal/config"
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
	var plugins []*Command

	for _, plugin := range c.Config.Plugins() {
		plugins = append(plugins, NewCmdPlugin(filepath.Base(plugin)))
	}

	c.AddGroups("Plugins:", plugins...)
}

// DefaultSubCommandRun prints a command's help string to the specified output if no
// arguments (sub-commands) are provided, or a usage error otherwise.
func DefaultSubCommandRun() func(c *cobra.Command, args []string) {
	return func(c *cobra.Command, args []string) {
		c.SetOut(log.Writer())
		c.SetErr(log.Writer())
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

	return fmt.Errorf("%s\nSee '%s -h' for help and examples", msg, cmd.CommandPath())
}

func Run(executablePath string, cmdArgs, environment []string) error {
	cmd := Cmnd(executablePath, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = environment

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run command: %w", err)
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

		return fmt.Errorf("failed to run command: %w", err)
	}

	// invoke cmd binary relaying the environment and args given
	// append executablePath to cmdArgs, as execve will make first argument the "binary name".
	return syscall.Exec(executablePath, append([]string{executablePath}, cmdArgs...), environment) // nolint: gosec
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
	var remainingArgs []string // all "non-flag" arguments

	for i, arg := range cmdArgs {
		if i == 0 && strings.HasPrefix(arg, "-") {
			break
		}
		// remainingArgs = append(remainingArgs, strings.Replace(arg, "-", "_", -1))
		remainingArgs = append(remainingArgs, arg)
	}

	if len(remainingArgs) == 0 {
		// the length of cmdArgs is at least 1
		return fmt.Errorf("flags cannot be placed before plugin name: %s", cmdArgs[0])
	}

	foundBinaryPath := ""

out:
	for i := range remainingArgs {
		for _, plugin := range c.Config.Plugins() {
			if filepath.Base(plugin) == remainingArgs[i] {
				remainingArgs = remainingArgs[i+1:]
				foundBinaryPath = plugin

				break out
			}
		}
	}

	if len(foundBinaryPath) == 0 {
		return nil
	}

	// invoke cmd binary relaying the current environment and args given
	if err := Execute(foundBinaryPath, remainingArgs, os.Environ()); err != nil {
		return fmt.Errorf("failed to execute plugin %q: %w", foundBinaryPath, err)
	}

	return nil
}

func NewCmdPlugin(s string) *Command {
	return &Command{
		Command: &cobra.Command{
			Use: s,
			Run: func(cmd *cobra.Command, args []string) {
				// TODO: implement
			},
		},
	}
}
