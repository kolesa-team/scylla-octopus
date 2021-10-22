package cluster

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd/local"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"go.uber.org/zap"
	"testing"
	"time"
)

// a test implementation of shell executor factory,
// that always returns a local command executor.
type localCmdFactory struct{}

func (l localCmdFactory) GetByHost(host string) (cmd.Executor, error) {
	return local.Executor{}, nil
}

// A successful, parallel execution of a callback on a cluster of 2 nodes.
func TestCluster_RunParallel_Ok(t *testing.T) {
	cluster := NewCluster(
		Options{
			Hosts: []string{"host-1", "host-2"},
		},
		localCmdFactory{},
		zap.S(),
	)
	require.Equal(t, 2, cluster.Size())

	result := cluster.RunParallel(
		context.Background(),
		func(ctx context.Context, node *entity.Node) entity.NodeCallbackResult {
			return entity.CallbackOk("result from " + node.Info.Host)
		},
	)

	require.NoError(t, result.Error(), "no errors are expected")
	require.Equal(
		t,
		"result from host-1",
		result["host-1"].Value,
		"running a callback on host-1 must produce a given string",
	)
	require.Equal(
		t,
		"result from host-2",
		result["host-2"].Value,
		"running a callback on host-2 must produce a given string",
	)
}

// A parallel execution of a callback on a cluster of 2 nodes with context timeout.
func TestCluster_RunParallel_WithContextTimeout(t *testing.T) {
	cluster := NewCluster(
		Options{
			Hosts: []string{"host-1", "host-2"},
		},
		localCmdFactory{},
		zap.S(),
	)
	require.Equal(t, 2, cluster.Size())

	require.NoError(t, cluster.Connect(context.Background()).Error())

	// a context will be cancelled after 10ms
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()

	result := cluster.RunParallel(
		ctx,
		func(ctx context.Context, node *entity.Node) entity.NodeCallbackResult {
			// execute a shell command "sleep 3" which should be cancelled
			err := node.Cmd.Run(ctx, cmd.Command("sleep", "3"))
			if err != nil {
				return entity.CallbackError(err)
			}

			return entity.CallbackOk(nil)
		},
	)

	require.Error(t, result.Error(), "an error is expected because of context timeout")
	require.EqualError(t, result.Error(),
		`2 errors occurred:
	* signal: killed
	* signal: killed

`,
		"result.Error() must contain a error info for every node",
	)
}

// A parallel execution of a callback on a cluster of 2 nodes,
// where there's an error on 1 of them.
func TestCluster_RunParallel_WithError(t *testing.T) {
	cluster := NewCluster(
		Options{
			Hosts: []string{"host-1", "host-2"},
		},
		localCmdFactory{},
		zap.S(),
	)

	result := cluster.RunParallel(
		context.Background(),
		func(ctx context.Context, node *entity.Node) entity.NodeCallbackResult {
			// return a error for host-1
			if node.Info.Host == "host-1" {
				return entity.CallbackError(errors.New("test error"))
			}

			return entity.CallbackOk("result from " + node.Info.Host)
		},
	)

	require.Error(t, result.Error(), "a callback result must have an error")

	require.Error(
		t,
		result["host-1"].Err,
		"a callback for host-1 must return an error",
	)
	require.NoError(
		t,
		result["host-2"].Err,
		"a callback for host-2 must not return an error",
	)
	require.Equal(
		t,
		"result from host-2",
		result["host-2"].Value,
		"a callback for host-2 must return an given string",
	)
}

// A successful, consecutive execution of a callback on a cluster of 2 nodes.
func TestCluster_Run_Ok(t *testing.T) {
	cluster := NewCluster(
		Options{
			Hosts: []string{"host-1", "host-2"},
		},
		localCmdFactory{},
		zap.S(),
	)

	result := cluster.Run(
		context.Background(),
		func(ctx context.Context, node *entity.Node) entity.NodeCallbackResult {
			return entity.CallbackOk("result from " + node.Info.Host)
		},
	)

	require.NoError(t, result.Error())

	require.Equal(
		t,
		"result from host-1",
		result["host-1"].Value,
		"a callback for host-1 must return an given string",
	)
	require.Equal(
		t,
		"result from host-2",
		result["host-2"].Value,
		"a callback for host-2 must return an given string",
	)
}

// A consecutive execution of a callback on a cluster of 2 nodes with an error.
// We expect that execution will be stopped after an error.
func TestCluster_Run_WithError(t *testing.T) {
	cluster := NewCluster(
		Options{
			Hosts: []string{"host-1", "host-2"},
		},
		localCmdFactory{},
		zap.S(),
	)
	require.Equal(t, 2, cluster.Size())

	result := cluster.Run(
		context.Background(),
		func(ctx context.Context, node *entity.Node) entity.NodeCallbackResult {
			// return an error for host-1
			if node.Info.Host == "host-1" {
				return entity.CallbackError(errors.New("test error"))
			}

			// there's no error for host-2,
			// but we expect this will not be returned at all,
			// because the execution should be stopped.
			return entity.CallbackOk("result from " + node.Info.Host)
		},
	)

	require.Error(t, result.Error())

	require.Len(t, result, 1, "only one execution result must be returned because of the error")
	require.Error(
		t,
		result["host-1"].Err,
		"a callback for host-1 must return an error",
	)
}
