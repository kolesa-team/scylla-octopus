package app

import (
	"context"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"time"
)

// Test implementations of the package dependencies

// testCluster operations always return whatever is given in structure properties
type testCluster struct {
	nodeCount       int
	callbackResults entity.NodeCallbackResults
}

func (t testCluster) Run(ctx context.Context, callback entity.NodeCallback) entity.NodeCallbackResults {
	return t.callbackResults
}

func (t testCluster) RunParallel(ctx context.Context, callback entity.NodeCallback) entity.NodeCallbackResults {
	return t.callbackResults
}

func (t testCluster) Connect(ctx context.Context) entity.NodeCallbackResults {
	return t.callbackResults
}

func (t testCluster) Size() int {
	return t.nodeCount
}

// testDb operations always return whatever is given in structure properties
type testDb struct {
	err          error
	snapshots    entity.Snapshots
	repairResult entity.RepairResult
}

func (t testDb) Healthcheck(ctx context.Context, node *entity.Node) error {
	return t.err
}

func (t testDb) ListSnapshots(ctx context.Context, node *entity.Node) (entity.Snapshots, error) {
	return t.snapshots, t.err
}

func (t testDb) Repair(ctx context.Context, node *entity.Node) (entity.RepairResult, error) {
	return t.repairResult, t.err
}

// testBackupService operations always return whatever is given in structure properties
type testBackupService struct {
	err                 error
	backupResultsByHost map[string]entity.BackupResult
	remoteBackups       []entity.RemoteBackup
}

func (t testBackupService) Healthcheck(ctx context.Context, node *entity.Node) error {
	return t.err
}

func (t testBackupService) Backup(ctx context.Context, node *entity.Node) entity.BackupResult {
	return t.backupResultsByHost[node.Info.Host]
}

func (t testBackupService) CleanupExpiredBackups(ctx context.Context, node *entity.Node, now time.Time) ([]entity.RemoteBackup, error) {
	return t.remoteBackups, t.err
}

func (t testBackupService) ListExpiredBackups(ctx context.Context, node *entity.Node, now time.Time) ([]entity.RemoteBackup, error) {
	return t.remoteBackups, t.err
}

// testStorage operations always return whatever is given in structure properties
type testStorage struct {
	err     error
	backups []entity.RemoteBackup
}

func (t testStorage) Healthcheck(ctx context.Context, cmdExecutor cmd.Executor) error {
	return t.err
}

func (t testStorage) ListBackups(ctx context.Context, cmdExecutor cmd.Executor, basePath string) ([]entity.RemoteBackup, error) {
	return t.backups, t.err
}
