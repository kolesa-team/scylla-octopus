package cluster

// A "cluster" is an abstraction over a set of database nodes.
// Allows running callbacks on each node, consecutively (Run) and in parallel (RunParallel).

import (
	"context"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"go.uber.org/zap"
	"sync"
)

type Cluster struct {
	options Options
	// a "factory" that creates shell-command executors on database nodes
	cmdFactory cmdFactory
	nodes      map[string]*entity.Node
	logger     *zap.SugaredLogger
}

type Options struct {
	// a list of hosts where the tool should run
	Hosts          []string
	Binaries       entity.NodeBinaries
	DataPath       string `yaml:"dataPath"`
	ClusterName    string `yaml:"clusterName"`
	SkipDnsResolve bool
}

// A factory interface that creates shell-command executors.
// Implemented in pkg/cmd. It either returns a local executor or an SSH client.
type cmdFactory interface {
	GetByHost(host string) (cmd.Executor, error)
}

func NewCluster(
	opts Options,
	cmdFactory cmdFactory,
	logger *zap.SugaredLogger,
) *Cluster {
	if opts.DataPath == "" {
		opts.DataPath = "/var/lib/scylla/data"
	}

	cluster := Cluster{
		options:    opts,
		cmdFactory: cmdFactory,
		logger:     logger,
		nodes:      map[string]*entity.Node{},
	}

	for _, host := range opts.Hosts {
		if _, alreadyExists := cluster.nodes[host]; alreadyExists {
			cluster.logger.Warnw("duplicate host", "host", host)
			continue
		}

		node := entity.Node{
			Info: entity.NewNodeInfo(
				host,
				opts.DataPath,
				opts.Binaries,
				// resolve domain names except when running tests
				opts.SkipDnsResolve == false,
			),
		}

		if len(opts.ClusterName) > 0 {
			node.Info.ClusterName = opts.ClusterName
		}

		cluster.nodes[host] = &node
	}

	return &cluster
}

// Size returns the number of nodes in cluster
func (c *Cluster) Size() int {
	return len(c.nodes)
}

// Run executes a given callback on each node consecutively.
// Stops the execution on error.
func (c *Cluster) Run(
	ctx context.Context,
	callback entity.NodeCallback,
) entity.NodeCallbackResults {
	results := entity.NodeCallbackResults{}

	for _, host := range c.options.Hosts {
		node := c.nodes[host]
		// do not execute a callback if SSH connection couldn't be established
		if node.ConnectionErr != nil {
			results[node.Info.Host] = entity.NodeCallbackResult{
				Host: node.Info.Host,
				Err:  node.ConnectionErr,
			}
			continue
		}

		result := callback(ctx, node)
		result.Host = node.Info.Host
		results[node.Info.Host] = result

		if ctx.Err() != nil {
			break
		}

		if result.Err != nil {
			break
		}
	}

	return results
}

// RunParallel executes a given callback on each node in parallel.
// The execution does not stop even if there's an error on one of the nodes.
func (c *Cluster) RunParallel(
	ctx context.Context,
	callback entity.NodeCallback,
) entity.NodeCallbackResults {
	var wg sync.WaitGroup
	results := entity.NodeCallbackResults{}
	wg.Add(len(c.nodes))

	resultsChan := make(chan entity.NodeCallbackResult, len(c.nodes))

	for _, host := range c.options.Hosts {
		node := c.nodes[host]

		// do not execute a callback if SSH connection couldn't be established
		if node.ConnectionErr != nil {
			resultsChan <- entity.NodeCallbackResult{
				Host: node.Info.Host,
				Err:  node.ConnectionErr,
			}
			wg.Done()
			continue
		}

		go func(node *entity.Node) {
			result := callback(ctx, node)
			result.Host = node.Info.Host
			resultsChan <- result
			wg.Done()
		}(node)
	}

	wg.Wait()
	close(resultsChan)

	for nodeResult := range resultsChan {
		results[nodeResult.Host] = nodeResult
	}

	return results
}

// Connect creates shell-command executors for each node.
// If the tool is running over SSH, the SSH connections are established here and the errors are reported.
func (c *Cluster) Connect(ctx context.Context) entity.NodeCallbackResults {
	return c.RunParallel(ctx, func(ctx context.Context, node *entity.Node) entity.NodeCallbackResult {
		node.Cmd, node.ConnectionErr = c.cmdFactory.GetByHost(node.Info.Host)
		if node.ConnectionErr != nil {
			return entity.CallbackError(node.ConnectionErr)
		}

		return entity.CallbackOk(nil)
	})
}
