package test

import (
	"context"
	"os/exec"
)

// Executor is a test implementation of command executor.
// Always returns the pre-configured results.
// Remembers the last executed command.
type Executor struct {
	// an error that must be returned (if Func is not set)
	Err error
	// an output that must be returned (if Func is not set)
	Output string
	// a callback that will be executed instead of an actual shell command.
	Func func(*exec.Cmd, int) (string, error)
	// a last executed command
	LastCmd *exec.Cmd
	// a number of executed commands
	ExecutedCount int
	// what to return when calling ReadFile
	FileToRead []byte
	// last file that was written with WriteFile
	WrittenFileBytes []byte
	WrittenFilePath  string
}

func (c *Executor) Execute(ctx context.Context, cmd *exec.Cmd) ([]byte, error) {
	c.LastCmd = cmd

	var output string
	var err error

	if c.Func != nil {
		output, err = c.Func(cmd, c.ExecutedCount)
	} else {
		output = c.Output
		err = c.Err
	}

	c.ExecutedCount++

	return []byte(output), err
}

func (c *Executor) Run(ctx context.Context, cmd *exec.Cmd) error {
	c.LastCmd = cmd

	return c.Err
}

func (c *Executor) ReadFile(ctx context.Context, path string) ([]byte, error) {
	return c.FileToRead, c.Err
}

func (c *Executor) WriteFile(ctx context.Context, path string, data []byte) error {
	c.WrittenFilePath = path
	c.WrittenFileBytes = data

	return c.Err
}
