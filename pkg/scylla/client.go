package scylla

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

// Client provides abstractions over "nodetool" and "cqlsh" commands on a scylladb node.
type Client struct {
	credentials entity.Credentials
	logger      *zap.SugaredLogger
}

func NewClient(
	credentials entity.Credentials,
	logger *zap.SugaredLogger,
) (client *Client) {
	client = &Client{
		credentials: credentials,
		logger:      logger.Named("[scylla]"),
	}

	return client
}

// Healthcheck ensures a database node is operable.
// FIXME: as a side effect, this also sets node status and cluster name, which are used later elsewhere.
// So we can't really use a node before healthchecking it.
func (c *Client) Healthcheck(ctx context.Context, node *entity.Node) error {
	logCtx := c.logger.With("host", node.Info.Host)
	logCtx.Debug("[healthcheck] checking scylla binaries")
	err := c.validateBinaries(ctx, node)
	if err != nil {
		return err
	}

	logCtx.Debug("[healthcheck] checking scylla data directory")
	if !cmd.DirectoryExists(ctx, node.Cmd, node.Info.DataPath) {
		return fmt.Errorf(
			"data directory %s does not exist",
			node.Info.DataPath,
		)
	}

	logCtx.Debug("[healthcheck] checking scylla node status")
	err = c.updateNodeInfo(ctx, node)
	if err != nil {
		return err
	}

	logCtx.Debug("[healthcheck] checking cqlsh")
	_, err = c.describeCluster(ctx, node)
	if err != nil {
		return err
	}

	logCtx.Debug("[healthcheck] OK")

	return nil
}

// ExportSchema writes a database schema to a file
func (c *Client) ExportSchema(ctx context.Context, node *entity.Node, path string) (string, error) {
	filePath := strings.TrimRight(path, "/") + "/db_schema.cql"

	cqlshCmd := c.cqlshCmd(node.Info)
	cqlshCmd.Args = append(
		cqlshCmd.Args,
		"-e",
		`"DESC SCHEMA"`,
		">",
		filePath,
	)

	output, err := node.Cmd.Execute(ctx, cqlshCmd)
	if err != nil {
		return "", errors.Wrapf(
			err,
			"could not export database schema at %s to %s. output:\n%s",
			node.Info.Host,
			filePath,
			string(output),
		)
	}

	return filePath, nil
}

// Runs `describe cluster`. Used to check that `cqlsh` is working correctly.
func (c *Client) describeCluster(ctx context.Context, node *entity.Node) (string, error) {
	cqlshCmd := c.cqlshCmd(node.Info)
	cqlshCmd.Args = append(
		cqlshCmd.Args,
		"-e",
		`"describe cluster"`,
	)

	output, err := node.Cmd.Execute(ctx, cqlshCmd)
	if err != nil {
		return "", errors.Wrapf(
			err,
			"could not describe cluster. output:\n%s",
			string(output),
		)
	}

	return string(output), nil
}

// Returns a base `cqlsh` command
func (c *Client) cqlshCmd(nodeInfo entity.NodeInfo) *exec.Cmd {
	cqlshCmd := cmd.Command(
		nodeInfo.Binaries.Cqlsh,
	)

	if len(nodeInfo.Host) > 0 {
		cqlshCmd.Args = append(cqlshCmd.Args, nodeInfo.Host)
	}

	if len(c.credentials.User) > 0 {
		cqlshCmd.Args = append(cqlshCmd.Args, "-u", c.credentials.User)
	}

	if len(c.credentials.Password) > 0 {
		cqlshCmd.Args = append(cqlshCmd.Args, "-p", c.credentials.Password)
	}

	return cqlshCmd
}

// Checks that necessary executables exists on a node
func (c *Client) validateBinaries(ctx context.Context, node *entity.Node) error {
	err := cmd.ExecutableFileExists(ctx, node.Cmd, node.Info.Binaries.Cqlsh)
	if err != nil {
		return err
	}

	err = cmd.ExecutableFileExists(ctx, node.Cmd, node.Info.Binaries.Nodetool)
	if err != nil {
		return err
	}

	return nil
}
