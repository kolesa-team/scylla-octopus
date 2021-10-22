package awscli

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"go.uber.org/zap"
	"os/exec"
	"strings"
)

type Client struct {
	options Options
	logger  *zap.SugaredLogger
}

type Options struct {
	Disabled    bool
	Binary      string
	Bucket      string
	EndpointUrl string `yaml:"endpointUrl"`
	Profile     string
}

func NewClient(opts Options, logger *zap.SugaredLogger) *Client {
	if len(opts.Binary) == 0 {
		opts.Binary = "aws"
	}

	opts.Bucket = strings.Trim(opts.Bucket, "/")

	return &Client{
		options: opts,
		logger:  logger.Named("awscli").With("bucket", opts.Bucket),
	}
}

// Healthcheck ensures awscli executable exists
func (c *Client) Healthcheck(ctx context.Context, cmdExecutor cmd.Executor) error {
	if c.options.Disabled {
		return nil
	}

	if len(c.options.Bucket) == 0 {
		return errors.New("[healthcheck] bucket is required")
	}

	c.logger.Debugw(
		"[healthcheck] checking aws cli binary",
		"path",
		c.options.Binary,
	)

	return cmd.ExecutableFileExists(ctx, cmdExecutor, c.options.Binary)
}

// Upload uploads a given source directory to s3
func (c *Client) Upload(ctx context.Context, cmdExecutor cmd.Executor, source, dest string) (string, error) {
	destUrl := c.getDestinationUrl(dest)
	err := c.sync(ctx, cmdExecutor, source, destUrl)

	return destUrl, err
}

// ListBackups returns backups from a given directory
func (c *Client) ListBackups(ctx context.Context, cmdExecutor cmd.Executor, basePath string) ([]entity.RemoteBackup, error) {
	backups := []entity.RemoteBackup{}
	// TODO the backups are kept at a 2nd leven of hierarchy, e.g. /basePath/scylla-node1/09-07-2021-10-29
	// this probably should not be hardcoded
	paths, err := c.listDirectoriesRecursive(ctx, cmdExecutor, basePath, 2)
	if err != nil {
		c.logger.Errorw(
			"could not list backups",
			"error", err,
		)

		return backups, err
	}

	for _, path := range paths {
		backup, err := entity.NewRemoteBackupFromPath(path)
		if err == nil {
			backups = append(backups, backup)
		}
	}

	return backups, nil
}

// RemoveBackup removes given backup directory
func (c *Client) RemoveBackup(ctx context.Context, cmdExecutor cmd.Executor, path string) error {
	command := cmd.Command(
		c.options.Binary,
		"s3",
		"rm",
		fmt.Sprintf("'%s'", c.getDestinationUrl(path)),
		"--recursive",
	)
	c.addCommandFlags(command)
	output, err := cmdExecutor.Execute(ctx, command)
	if err != nil {
		c.logger.Errorw(
			"could not remove a backup",
			"error", err,
			"output", string(output),
			"path", path,
		)

		return errors.Wrapf(
			err,
			"could not remove a backup at %s. output: %s",
			path,
			string(output),
		)
	}

	c.logger.Infow("backup removed", "path", path)

	return nil
}

// Returns a complete url to a destination directory in s3 format
func (c *Client) getDestinationUrl(dest string) string {
	url := fmt.Sprintf(
		"s3://%s",
		c.options.Bucket,
	)

	if len(dest) > 0 {
		url += "/" + strings.TrimLeft(dest, "/")
	}

	return url
}

// Runs "aws s3 sync" to sync a directory with a bucket.
// see https://docs.aws.amazon.com/cli/latest/userguide/cli-services-s3-commands.html#using-s3-commands-managing-objects-sync
func (c *Client) sync(ctx context.Context, cmdExecutor cmd.Executor, source, dest string) error {
	command := cmd.Command(
		c.options.Binary,
		"s3",
		"sync",
		source,
		fmt.Sprintf("'%s'", dest),
	)
	c.addCommandFlags(command)
	output, err := cmdExecutor.Execute(ctx, command)
	if err != nil {
		return errors.Wrapf(
			err,
			"could not sync files to %s. output: %s",
			dest,
			string(output),
		)
	}

	return nil
}

// Runs "aws s3 ls" to list directories recursively until it reaches a given depth.
func (c *Client) listDirectoriesRecursive(ctx context.Context, cmdExecutor cmd.Executor, path string, depth int) ([]string, error) {
	result := []string{}

	if ctx.Err() != nil {
		return result, ctx.Err()
	}

	dirs, err := c.listDirectories(ctx, cmdExecutor, path)
	if err != nil || depth == 0 {
		return dirs, err
	}

	for _, dir := range dirs {
		tmp, err := c.listDirectoriesRecursive(ctx, cmdExecutor, dir, depth-1)
		if err != nil {
			return result, err
		}
		if len(tmp) > 0 {
			result = append(result, tmp...)
		} else {
			result = append(result, dir)
		}
	}

	return result, nil
}

// Runs "aws s3 ls" to get directories at given path
func (c *Client) listDirectories(ctx context.Context, cmdExecutor cmd.Executor, path string) ([]string, error) {
	command := cmd.Command(
		c.options.Binary,
		"s3",
		"ls",
		fmt.Sprintf("'%s/'", c.getDestinationUrl(path)),
	)
	c.addCommandFlags(command)
	output, err := cmdExecutor.Execute(ctx, command)
	if err != nil {
		return []string{}, errors.Wrapf(
			err,
			"could not list files at %s. command: %s\noutput: %s",
			path,
			command.String(),
			string(output),
		)
	}

	return c.parseDirectoryList(path, string(output)), nil
}

// Returns a list of directories from "aws s3 ls" output
func (c *Client) parseDirectoryList(basePath string, output string) []string {
	var result []string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// the directories have a "PRE" prefix
		// see https://docs.aws.amazon.com/cli/latest/reference/s3/ls.html#examples
		if !strings.HasPrefix(line, "PRE ") {
			continue
		}

		line = strings.TrimPrefix(line, "PRE ")

		result = append(
			result,
			basePath+"/"+strings.TrimRight(line, "/"),
		)
	}

	return result
}

func (c *Client) addCommandFlags(command *exec.Cmd) {
	if len(c.options.EndpointUrl) > 0 {
		command.Args = append(
			command.Args,
			"--endpoint-url",
			c.options.EndpointUrl,
		)
	}

	if len(c.options.Profile) > 0 {
		command.Args = append(
			command.Args,
			"--profile",
			c.options.Profile,
		)
	}
}
