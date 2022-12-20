package shell

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type Shell interface {
	Execute(name string, args ...string) (output []byte, err error)
	ExecuteWithOptions(name string, args []string, opts ...Opt) (output []byte, err error)
}

type Opt func(shell *LocalShell)

func NewLocalShellWithOpts(opts ...Opt) *LocalShell {
	c := &LocalShell{}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func NewMockShell(output []byte, err error, cmd string) *MockShell {
	return &MockShell{
		Output:      output,
		Err:         err,
		LastCommand: cmd,
	}
}

func WithSuppressOutput(b bool) Opt {
	return func(c *LocalShell) {
		c.SuppressOutput = b
	}
}

type LocalShell struct {
	SuppressOutput bool
}

func (c *LocalShell) ExecuteWithOptions(name string, args []string, opts ...Opt) ([]byte, error) {
	for _, opt := range opts {
		opt(c)
	}

	return c.Execute(name, args...)
}

func (c *LocalShell) Execute(name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name)
	cmd.Args = append(cmd.Args, arg...)
	cmd.Stdin = os.Stdin

	var combinedOutBuf bytes.Buffer
	if c.SuppressOutput {
		cmd.Stdout = io.MultiWriter(io.Discard, &combinedOutBuf)
		cmd.Stderr = io.MultiWriter(io.Discard, &combinedOutBuf)
	} else {
		cmd.Stdout = io.MultiWriter(os.Stdout, &combinedOutBuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &combinedOutBuf)
	}

	err := cmd.Run()
	outStr := combinedOutBuf.Bytes()
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

// interface guards
var _ Shell = &LocalShell{}
var _ Shell = &MockShell{}
