package app

import (
	"context"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"time"
)

// Here we keep the interfaces that the app depends on

// Database client (implemented in `pkg/scylla`)
type dbClient interface {
	Healthcheck(ctx context.Context, node *entity.Node) error
	ListSnapshots(ctx context.Context, node *entity.Node) (entity.Snapshots, error)
	Repair(ctx context.Context, node *entity.Node) (entity.RepairResult, error)
}

// Remote storage (implemented in `pkg/awscli`)
type remoteStorageClient interface {
	Healthcheck(ctx context.Context, cmdExecutor cmd.Executor) error
	ListBackups(ctx context.Context, cmdExecutor cmd.Executor, basePath string) ([]entity.RemoteBackup, error)
}

// Backup service (implemented in `pkg/backup`)
type backupService interface {
	Healthcheck(ctx context.Context, node *entity.Node) error
	Backup(ctx context.Context, node *entity.Node) entity.BackupResult
	CleanupExpiredBackups(ctx context.Context, node *entity.Node, now time.Time) ([]entity.RemoteBackup, error)
	ListExpiredBackups(ctx context.Context, node *entity.Node, now time.Time) ([]entity.RemoteBackup, error)
}

// A cluster of database nodes (implemented in `pkg/cluster`)
type cluster interface {
	Run(ctx context.Context, callback entity.NodeCallback) entity.NodeCallbackResults
	RunParallel(ctx context.Context, callback entity.NodeCallback) entity.NodeCallbackResults
	Connect(ctx context.Context) entity.NodeCallbackResults
	Size() int
}
