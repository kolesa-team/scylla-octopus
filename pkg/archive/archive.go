package archive

import (
	"context"
	"fmt"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"github.com/pkg/errors"
)

// Compression compression backup before upload to s3
func Compression(ctx context.Context, node *entity.Node, localPath string, archive entity.Archive) error {
	options := fmt.Sprintf("-%s -p%s", archive.Options.Compression, archive.Options.Threads)
	archiveName := "backup.tar." + archive.Method

	_, err := node.Cmd.Execute(ctx, cmd.Command(
		"sh",
		"-c",
		fmt.Sprintf(
			`'cd %s && /usr/bin/tar cf - ./ --exclude='./%s' | %s %s > %s'`,
			localPath,
			archiveName,
			archive.Method,
			options,
			archiveName,
		),
	))

	if err != nil {
		return errors.Wrapf(
			err,
			"failed to compress backup. Method: %s. Options: %s.",
			archive.Method,
			options,
		)
	}

	return clearDirectory(ctx, node, localPath, archiveName)
}

// clearDirectory cleaning the directory except archive and metadata for uploading to s3
func clearDirectory(ctx context.Context, node *entity.Node, localPath string, archiveName string) error {
	_, err := node.Cmd.Execute(ctx, cmd.Command(
		"sh",
		"-c",
		fmt.Sprintf(
			`'cd %s && find . ! -name "%s" -exec rm -rf {} +'`,
			localPath,
			archiveName,
		),
	))

	if err != nil {
		return errors.Wrapf(
			err,
			"failed to clear directory. Path: %s. Excluded file: %s.",
			localPath,
			archiveName,
		)
	}

	return nil
}
