package app

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"github.com/kolesa-team/scylla-octopus/pkg/notifier"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestOctopus_Repair(t *testing.T) {
	// a test implementation of database cluster returning 2 repair results:
	// one is successful and one is not.
	cluster := testCluster{
		nodeCount: 2,
		callbackResults: map[string]entity.NodeCallbackResult{
			"host-1": {
				Host: "host-1",
				Value: entity.RepairResult{
					Duration: time.Second,
				},
				Err: nil,
			},
			"host-2": {
				Host: "host-2",
				Value: entity.RepairResult{
					Duration: time.Second * 2,
				},
				Err: errors.New("test error"),
			},
		},
	}
	app := NewOctopus(
		cluster,
		testDb{},
		testBackupService{},
		testStorage{},
		notifier.Disabled{},
		zap.S(),
	)

	result := app.Repair(context.Background())

	require.Equal(t, 2, result.TotalNodes, "a cluster must contain 2 nodes")
	require.Equal(t, 1, result.RepairedNodes, "only  1 node should be repaired")

	require.Equal(t, `Total nodes: 2
Repaired nodes: 1

Error:
1 error occurred:
	* test error


Duration:
host-1: 1s`,
		result.Report())
}
