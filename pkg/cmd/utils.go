package cmd

import (
	"context"
	"fmt"
)

// This keeps commonly used shell commands

// ExecutableFileExists checks if an executable file exists
func ExecutableFileExists(ctx context.Context, executor Executor, file string) error {
	err := executor.Run(ctx, Command("test", "-x", file))
	if err != nil {
		return fmt.Errorf(
			"an executable %s is unavailable",
			file,
		)
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
