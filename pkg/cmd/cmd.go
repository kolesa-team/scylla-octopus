package cmd

import (
	"context"
	"os/exec"
)

// Executor is a shell command executor interface
type Executor interface {
	Execute(ctx context.Context, cmd *exec.Cmd) ([]byte, error)
	Run(ctx context.Context, cmd *exec.Cmd) error
	ReadFile(ctx context.Context, path string) ([]byte, error)
	WriteFile(ctx context.Context, path string, data []byte) error
}

// Command creates a shell command, almost like a standard `exec.Command()`,
// except it doesn't try to resolve an absolute path of `name`.
// This allows to work with local commands as well as with SSH.
func Command(name string, arg ...string) *exec.Cmd {
	return &exec.Cmd{
		Path: name,
		Args: append([]string{name}, arg...),
	}
}
