package shell

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Shell interface {
	Execute(name string, args ...string) (output []byte, err error)
	ExecuteWithOptions(name string, args []string, opts ...Opt) (output []byte, err error)
	Pipeline(cmds ...*exec.Cmd) (pipeLineOutput, collectedStandardError []byte, pipeLineError error)
	RunCommand(args []string, opts ...Opt) (output []byte, err error)
	ExitCodeOfCommand(command string) int
}

type Opt func(shell *LocalShell)

func NewLocalShellWithOpts(opts ...Opt) *LocalShell {
	c := &LocalShell{}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func NewMockShell(cmd string, output []byte, err error) *MockShell {
	return &MockShell{
		LastCommand: cmd,
		Output:      output,
		Err:         err,
	}
}

func WithCatchOutput(b bool) Opt {
	return func(c *LocalShell) {
		c.CatchStdout = &b
	}
}

func WithSuppressOutput(b bool) Opt {
	return func(c *LocalShell) {
		c.SuppressStdout = &b
	}
}

type LocalShell struct {
	CatchStdout    *bool
	SuppressStdout *bool
}

func (c *LocalShell) Reset() {
	c.CatchStdout = nil
	c.SuppressStdout = nil
}

func (c *LocalShell) ExecuteWithOptions(name string, args []string, opts ...Opt) ([]byte, error) {
	for _, opt := range opts {
		opt(c)
	}

	return c.Execute(name, args...)
}

func (c *LocalShell) CatchOutput() bool {
	if c.CatchStdout == nil {
		return false
	}

	return *c.CatchStdout
}

func (c *LocalShell) SuppressOutput() bool {
	if c.SuppressStdout == nil {
		return false
	}

	return *c.SuppressStdout
}

func (c *LocalShell) Execute(name string, arg ...string) ([]byte, error) {
	log.Debugf("Executing command: %s %s", name, strings.Join(arg, " "))
	log.Debugf("Catch stdout: %t", c.CatchOutput())
	log.Debugf("Suppress stdout: %t", c.SuppressOutput())

	defer c.Reset()

	cmd := exec.Command(name)
	cmd.Args = append(cmd.Args, arg...)
	cmd.Stdin = os.Stdin

	var combinedOutBuf bytes.Buffer

	switch {
	case c.CatchOutput() && c.SuppressOutput():
		cmd.Stdout = io.MultiWriter(io.Discard, &combinedOutBuf)
		cmd.Stderr = io.MultiWriter(io.Discard, &combinedOutBuf)
	case c.CatchOutput():
		cmd.Stdout = io.MultiWriter(os.Stdout, &combinedOutBuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &combinedOutBuf)
	default:
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	err := cmd.Run()
	outStr := combinedOutBuf.Bytes()

	log.Debugf("Command output: %s", outStr)

	if err != nil {
		return outStr, fmt.Errorf("error running command: %s: %w", name, err)
	}

	return outStr, nil
}

type MockShell struct {
	Output      []byte
	Err         error
	LastCommand string
}

func (c *MockShell) ExecuteWithOptions(name string, args []string, opts ...Opt) ([]byte, error) {
	return c.Execute(name, args...)
}

func (c *MockShell) Execute(name string, args ...string) ([]byte, error) {
	c.LastCommand = name

	return c.Output, c.Err
}

// Pipeline runs cmds piped after each other.
func (c *LocalShell) Pipeline(cmds ...*exec.Cmd) (pipeLineOutput, collectedStandardError []byte, pipeLineError error) {
	// Require at least one command
	if len(cmds) < 1 {
		return nil, nil, nil
	}

	// Collect the output from the command(s)
	var output, stderr bytes.Buffer

	last := len(cmds) - 1
	for i, cmd := range cmds[:last] {
		var err error
		// Connect each command's stdin to the previous command's stdout
		if cmds[i+1].Stdin, err = cmd.StdoutPipe(); err != nil {
			return nil, nil, fmt.Errorf("%w", err)
		}
		// Connect each command's stderr to a buffer
		cmd.Stderr = &stderr
	}

	// Connect the output and error for the last command
	cmds[last].Stdout, cmds[last].Stderr = &output, &stderr

	// Start each command
	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			return output.Bytes(), stderr.Bytes(), fmt.Errorf("%w", err)
		}
	}

	// Wait for each command to complete
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			return output.Bytes(), stderr.Bytes(), fmt.Errorf("%w", err)
		}
	}

	// Return the pipeline output and the collected standard error
	return output.Bytes(), stderr.Bytes(), nil
}

// Pipeline runs cmds piped after each other.
func (c *MockShell) Pipeline(cmds ...*exec.Cmd) (pipeLineOutput, collectedStandardError []byte, pipeLineError error) {
	// Require at least one command
	if len(cmds) < 1 {
		return nil, nil, nil
	}

	c.LastCommand = cmds[len(cmds)-1].String()

	return c.Output, []byte(c.Err.Error()), nil
}

// RunCommand is going to run a command depending on the caller's operating system.
func (c *LocalShell) RunCommand(args []string, opts ...Opt) ([]byte, error) {
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "cmd"

		args = append([]string{"/c"}, args...)
	} else {
		cmd = "sh"

		args = append([]string{"-c"}, strings.Join(args, " "))
	}

	return c.ExecuteWithOptions(cmd, args, opts...)
}

// RunCommand is going to run a command in a shell depending on the caller's operating system.
func (c *MockShell) RunCommand(args []string, opts ...Opt) ([]byte, error) {
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "cmd"

		args = append([]string{"/c"}, args...)
	} else {
		cmd = "sh"

		args = append([]string{"-c"}, strings.Join(args, " "))
	}

	return c.ExecuteWithOptions(cmd, args, opts...)
}

// ExitCodeOfCommand runs a command and returns its exit code.
func (c *LocalShell) ExitCodeOfCommand(command string) int {
	var status int

	_, err := c.RunCommand([]string{command})
	if err != nil {
		var exitError *exec.ExitError
		if ok := errors.As(err, &exitError); ok {
			status = exitError.ExitCode()
		}
	}

	return status
}

// ExitCodeOfCommand runs a command and returns its exit code.
func (c *MockShell) ExitCodeOfCommand(command string) int {
	var status int

	_, err := c.RunCommand([]string{command})
	if err != nil {
		var exitError *exec.ExitError
		if ok := errors.As(err, &exitError); ok {
			status = exitError.ExitCode()
		}
	}

	return status
}

// Interface guards.
var (
	_ Shell = &LocalShell{}
	_ Shell = &MockShell{}
)
