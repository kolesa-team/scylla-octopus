package app

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	clusterPkg "github.com/kolesa-team/scylla-octopus/pkg/cluster"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd/factory"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"github.com/kolesa-team/scylla-octopus/pkg/notifier"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestOctopus_Backup(t *testing.T) {
	logger := zap.S()
	// a test cluster with 2 hosts: 127.0.0.1, 127.0.0.2
	clusterInstance := clusterPkg.NewCluster(
		clusterPkg.Options{
			Hosts: []string{"127.0.0.1", "127.0.0.2"},
		},
		factory.NewTestFactory(),
		logger,
	)
	// a test backup service implementation that returns
	// - a successful backup for 127.0.0.1
	// - an error for 127.0.0.2
	backupService := testBackupService{
		backupResultsByHost: map[string]entity.BackupResult{
			"127.0.0.1": {
				SnapshotTag: "host-1-snapshot",
				Duration:    time.Second,
			},
			"127.0.0.2": {
				SnapshotTag: "host-2-snapshot",
				Duration:    time.Second * 2,
				Error:       errors.New("test error"),
			},
		},
	}
	app := NewOctopus(
		clusterInstance,
		testDb{},
		backupService,
		testStorage{},
		notifier.Disabled{},
		zap.S(),
	)

	result := app.Backup(context.Background())

	require.Error(t, result.Error)
	require.Equal(t, 2, result.TotalNodes, "a cluster must contain 2 nodes")
	require.Equal(t, 1, result.BackedUpNodes, "only 1 node must be backed up")

	report := result.Report()
	require.Contains(
		t,
		report,
		`Total nodes: 2
Backed up nodes: 1`,
	)
}
