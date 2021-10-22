package scylla

import (
	"context"
	"github.com/pkg/errors"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"time"
)

// Repair executes `nodetool repair --partitioner-range`
// see https://docs.scylladb.com/operating-scylla/nodetool-commands/repair/
func (c *Client) Repair(ctx context.Context, node *entity.Node) (entity.RepairResult, error) {
	logCtx := c.logger.With("host", node.Info.Host)
	timeStart := time.Now()
	command := cmd.Command(
		node.Info.Binaries.Nodetool,
		"repair",
		"--partitioner-range",
	)
	output, err := node.Cmd.Execute(ctx, command)
	duration := time.Now().Sub(timeStart)

	if err != nil {
		logCtx.Errorw(
			"could not execute nodetool repair",
			"error",
			err,
			"output",
			string(output),
		)
		err = errors.Wrapf(
			err,
			"could not execute nodetool repair on %s. output: %s",
			node.Info.Host,
			string(output),
		)
	} else {
		logCtx.Infow("repair completed", "duration", duration)
		logCtx.Debugw("repair output", "output", string(output))
	}

	return entity.RepairResult{
		Output:   string(output),
		Duration: duration,
	}, err
}
