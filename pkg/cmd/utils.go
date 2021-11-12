package cmd

import (
	"context"
	"fmt"
	"strings"
)

// This keeps commonly used shell commands

// ExecutableFileExists checks if an executable file exists
func ExecutableFileExists(ctx context.Context, executor Executor, file string) error {
	err := executor.Run(ctx, Command("test", "-x", file))
	if err != nil {
		// if `test` failed, try `whereis` instead.
		// `whereis` always exits with code 0, so we have to parse its output.
		// it is successful if it prints "file: path-to-file",
		// and is not successful if it prints "file:".
		whereisOutput, _ := executor.Execute(ctx, Command("whereis", file))
		whereisLines := strings.Split(strings.TrimSpace(string(whereisOutput)), "\n")
		whereisParts := strings.SplitN(whereisLines[len(whereisLines)-1], ":", 2)

		if len(whereisParts) < 2 || len(whereisParts[1]) == 0 {
			return fmt.Errorf(
				"an executable %s is unavailable",
				file,
			)
		}
	}

	return nil
}

// DirectoryExists checks if a directory exists
func DirectoryExists(ctx context.Context, executor Executor, path string) bool {
	err := executor.Run(ctx, Command("test", "-d", path))
	if err != nil {
		return false
	}

	return true
}

// EnsureDirectoryIsEmpty checks if the directory exists and is empty.
// Removes directory contents, if it exists. Creates a directory, if it doesn't exist.
func EnsureDirectoryIsEmpty(ctx context.Context, executor Executor, path string) error {
	if DirectoryExists(ctx, executor, path) {
		return ClearDirectory(ctx, executor, path)
	} else {
		return CreateDirectory(ctx, executor, path)
	}
}

func CreateDirectory(ctx context.Context, executor Executor, path string) error {
	return executor.Run(ctx, Command("mkdir", "-p", path))
}

func RemoveDirectory(ctx context.Context, executor Executor, path string) error {
	return executor.Run(ctx, Command("rm", "-r", path))
}

func ClearDirectory(ctx context.Context, executor Executor, path string) error {
	return executor.Run(ctx, Command("rm", "-rf", path+"/*"))
}
