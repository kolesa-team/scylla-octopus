package local

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// Executor is an implementation of executor on a local machine.
// Based on standard `exec` package.
// Expects a local machine to have `sh` executable.
type Executor struct {
	Debug bool
}

func (r Executor) Execute(ctx context.Context, cmd *exec.Cmd) ([]byte, error) {
	// wrap the command into "sh -c '...command...'",
	// so that special things like output redirection work well.
	wrapperCmd := exec.CommandContext(ctx, "sh", "-c", cmd.String())
	timeStarted := time.Now()

	if r.Debug {
		fmt.Printf(
			"\n---[CMD] executing command ---\n%s\n",
			wrapperCmd.String(),
		)
	}

	output, err := wrapperCmd.CombinedOutput()
	if r.Debug {
		fmt.Printf(
			"\n---[CMD] command done in %s, output:---\n%s\n",
			time.Now().Sub(timeStarted).String(),
			string(output),
		)
	}

	return output, err
}

func (r Executor) Run(ctx context.Context, cmd *exec.Cmd) error {
	_, err := r.Execute(ctx, cmd)

	return err
}

func (r Executor) ReadFile(ctx context.Context, path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (r Executor) WriteFile(ctx context.Context, path string, data []byte) error {
	return os.WriteFile(path, data, os.ModePerm)
}
