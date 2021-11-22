package scylla

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

// Updates a cluster name, status, and datacenter for a given node.
// Returns error if the status is not "UN" or if it cannot be updated.
func (c *Client) updateNodeInfo(ctx context.Context, node *entity.Node) error {
	var err error
	var result *multierror.Error

	if len(node.Info.ClusterName)  == 0 {
		node.Info.ClusterName, err = c.getClusterName(ctx, node)
		if err != nil {
			result = multierror.Append(result, err)
		}
	}

	node.Info.Status, node.Info.Datacenter, err = c.getNodeStatus(ctx, node)
	if err != nil {
		result = multierror.Append(result, err)
	} else if !node.Info.IsStatusOk() {
		result = multierror.Append(result, fmt.Errorf(
			"invalid status on node %s: %s",
			node.Info.Host,
			node.Info.Status,
		))
	}

	return result.ErrorOrNil()
}

// executes `nodetool status`; returns node status and datacenter
func (c *Client) getNodeStatus(
	ctx context.Context,
	node *entity.Node,
) (status, datacenter string, err error) {
	output, err := node.Cmd.Execute(ctx, cmd.Command(
		node.Info.Binaries.Nodetool,
		"status",
	))
	if err != nil {
		return "", "", err
	}

	possibleNodeAddresses := []string{
		node.Info.IpAddress,
		node.Info.DomainName,
	}
	status, datacenter = parseNodeStatus(
		possibleNodeAddresses,
		string(output),
	)

	if status == "" {
		return status, datacenter, fmt.Errorf(
			"could not parse node status from output for %+v:\n%s",
			possibleNodeAddresses,
			output,
		)
	}

	return status, datacenter, nil
}

// a regexp to retrieve a datacenter name from `nodetool status` command
var datacenterRegexp = regexp.MustCompile("Datacenter: (.+)")

// Parses an output of `nodetool status`
func parseNodeStatus(possibleNodeAddresses []string, output string) (status, datacenter string) {

	datacenterMatches := datacenterRegexp.FindStringSubmatch(output)
	if len(datacenterMatches) >= 2 {
		datacenter = datacenterMatches[1]
	}

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		isCurrentNode := false

		for _, nodeAddr := range possibleNodeAddresses {
			if len(nodeAddr) > 0 && strings.Contains(line, nodeAddr) {
				isCurrentNode = true
			}
		}

		if !isCurrentNode {
			continue
		}

		status = strings.Split(line, " ")[0]
		break
	}

	return status, datacenter
}

// a regexp to retrieve a cluster name from `nodetool describecluster` command
var clusterNameRegexp = regexp.MustCompile("Name: (.+)")

// Returns a cluster name from `nodetool describecluster` command output
func (c *Client) getClusterName(
	ctx context.Context,
	node *entity.Node,
) (string, error) {
	output, err := node.Cmd.Execute(ctx, cmd.Command(
		node.Info.Binaries.Nodetool,
		"describecluster",
	))
	if err != nil {
		return "", errors.Wrap(err, "could not execute `nodetool describecluster`")
	}

	matches := clusterNameRegexp.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return "", errors.New("could not get cluster name from `nodetool describecluster` output")
	}

	return matches[1], nil
}
