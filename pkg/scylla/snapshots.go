package scylla

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
)

// CreateSnapshot creates a snapshot with `nodetool snapshot` and moves it into given path.
func (c *Client) CreateSnapshot(ctx context.Context, node *entity.Node, tag, path string, keyspaces []string) error {
	err := c.removeSnapshotIfExists(ctx, node, tag)
	if err != nil {
		return err
	}

	output, err := c.runSnapshotCmd(ctx, node, tag, keyspaces)
	if err != nil {
		c.logger.Errorw(
			"could not create snapshot",
			"error", err,
			"tag", tag,
			"keyspaces", keyspaces,
			"output", string(output),
		)

		return errors.Wrapf(
			err,
			"could not create snapshot %s. output: %s",
			tag,
			string(output),
		)
	}

	return c.moveSnapshot(ctx, node, tag, path)
}

// ListSnapshots returns all snapshots on a given node
func (c *Client) ListSnapshots(ctx context.Context, node *entity.Node) (entity.Snapshots, error) {
	output, err := c.listSnapshots(ctx, node)
	if err != nil {
		return nil, err
	}

	return entity.ParseSnapshots(output), nil
}

// RemoveSnapshot removes a snapshot by tag
func (c *Client) RemoveSnapshot(ctx context.Context, node *entity.Node, tag string) error {
	output, err := c.runClearSnapshotCmd(ctx, node, tag)
	if err != nil {
		c.logger.Errorw(
			"could not remove snapshot",
			"tag", tag,
			"error", err,
			"output", string(output),
		)

		return errors.Wrapf(err, "could not remove snapshot %s", tag)
	}

	return nil
}

// MoveSnapshot moves a snapshots from scylla directory to another temporary directory.
// TODO replace cp with mv to avoid unnecessary disk space usage
func (c *Client) moveSnapshot(ctx context.Context, node *entity.Node, tag string, targetPath string) error {
	output, err := node.Cmd.Execute(ctx, cmd.Command(
		"sh",
		"-c",
		fmt.Sprintf(
			`'cd %s && find . -type d | grep -i snapshots/%s | xargs -i cp --parents -r {} %s'`,
			node.Info.DataPath,
			tag,
			targetPath,
		),
	))
	if err != nil {
		c.logger.Errorw(
			"could not move snapshot",
			"error", err,
			"output", string(output),
			"tag", tag,
			"targetPath", targetPath,
		)

		return errors.Wrapf(err, "could not move snapshot %s to %s", tag, targetPath)
	}

	return nil
}

func (c *Client) removeSnapshotIfExists(ctx context.Context, node *entity.Node, tag string) error {
	snapshotExists, err := c.snapshotExists(ctx, node, tag)
	if err != nil {
		return err
	}

	if !snapshotExists {
		return nil
	}

	logCtx := c.logger.With("tag", tag)
	logCtx.Debug("removing existing snapshot")

	err = c.RemoveSnapshot(ctx, node, tag)
	if err != nil {
		return err
	}

	logCtx.Debug("existing snapshot removed")

	return nil
}

func (c *Client) snapshotExists(ctx context.Context, node *entity.Node, tag string) (bool, error) {
	existingSnapshots, err := c.ListSnapshots(ctx, node)
	if err != nil {
		return false, errors.Wrapf(err, "could not check if snapshot %s exists", tag)
	}

	_, exists := existingSnapshots[tag]
	return exists, nil
}

func (c *Client) runClearSnapshotCmd(ctx context.Context, node *entity.Node, tag string) ([]byte, error) {
	return node.Cmd.Execute(ctx, cmd.Command(
		node.Info.Binaries.Nodetool,
		"clearsnapshot",
		"-t",
		tag,
	))
}

func (c *Client) runSnapshotCmd(ctx context.Context, node *entity.Node, tag string, keyspaces []string) ([]byte, error) {
	command := cmd.Command(
		node.Info.Binaries.Nodetool,
		"snapshot",
		"-t",
		tag,
	)

	if len(keyspaces) > 0 {
		command.Args = append(command.Args, keyspaces...)
	}

	return node.Cmd.Execute(ctx, command)
}

func (c *Client) listSnapshots(ctx context.Context, node *entity.Node) (string, error) {
	output, err := node.Cmd.Execute(ctx, cmd.Command(
		node.Info.Binaries.Nodetool,
		"listsnapshots",
	))
	if err != nil {
		return "", errors.Wrapf(err, "could not list snapshots. output:\n%s", string(output))
	}

	return string(output), nil
}
