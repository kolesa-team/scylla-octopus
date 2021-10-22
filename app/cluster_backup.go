package app

import (
	"context"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"time"
)

// Backup backs up every cluster node
func (m *Octopus) Backup(ctx context.Context) entity.BackupResults {
	results := m.cluster.RunParallel(ctx, func(ctx context.Context, node *entity.Node) entity.NodeCallbackResult {
		backupResult := m.backup.Backup(ctx, node)
		if backupResult.Error != nil {
			return entity.CallbackErrorWithValue(backupResult.Error, backupResult)
		}

		return entity.CallbackOk(backupResult)
	})

	backupResults := entity.BackupResults{
		TotalNodes:    m.cluster.Size(),
		BackedUpNodes: 0,
		ByHost:        map[string]entity.BackupResult{},
		Error:         results.Error(),
	}
	for _, result := range results {
		backupResults.ByHost[result.Host] = result.Value.(entity.BackupResult)

		if backupResults.ByHost[result.Host].Error == nil {
			backupResults.BackedUpNodes++
		}
	}

	if backupResults.Error != nil {
		m.notifier.Error(
			"Could not back up cluster nodes",
			backupResults.Report(),
			backupResults.Error,
			nil,
		)
	} else {
		m.notifier.Info(
			"Backup completed successfully",
			backupResults.Report(),
			nil,
		)
	}

	return backupResults
}

// ListSnapshots lists snapshots on every node in the cluser
func (m *Octopus) ListSnapshots(ctx context.Context) (entity.SnapshotsByNode, error) {
	results := m.cluster.RunParallel(ctx, func(ctx context.Context, node *entity.Node) entity.NodeCallbackResult {
		snapshots, err := m.scylla.ListSnapshots(ctx, node)
		return entity.NodeCallbackResult{
			Value: snapshots,
			Err:   err,
		}
	})

	snapshotsByNode := entity.SnapshotsByNode{}

	for _, result := range results {
		if result.Err == nil {
			snapshotsByNode[result.Host] = result.Value.(entity.Snapshots)
		}
	}

	return snapshotsByNode, results.Error()
}

// CleanupExpiredBackups removes expired backups in remote storage
func (m *Octopus) CleanupExpiredBackups(ctx context.Context) (entity.RemoteBackupsByHost, error) {
	now := time.Now()
	results := m.cluster.RunParallel(ctx, func(ctx context.Context, node *entity.Node) entity.NodeCallbackResult {
		expiredBackups, err := m.backup.CleanupExpiredBackups(ctx, node, now)
		if err != nil {
			return entity.CallbackError(err)
		}

		return entity.CallbackOk(expiredBackups)
	})

	cleanupResults := entity.RemoteBackupsByHost{}
	for _, result := range results {
		if result.Err == nil {
			cleanupResults[result.Host] = result.Value.([]entity.RemoteBackup)
		}
	}

	return cleanupResults, results.Error()
}

// ListExpiredBackups returns a list of expired backups in remote storage
func (m *Octopus) ListExpiredBackups(ctx context.Context) (entity.RemoteBackupsByHost, error) {
	now := time.Now()
	results := m.cluster.RunParallel(ctx, func(ctx context.Context, node *entity.Node) entity.NodeCallbackResult {
		expiredBackups, err := m.backup.ListExpiredBackups(ctx, node, now)
		if err != nil {
			return entity.CallbackError(err)
		}

		return entity.CallbackOk(expiredBackups)
	})

	expiredBackups := entity.RemoteBackupsByHost{}
	for _, result := range results {
		if result.Err == nil {
			expiredBackups[result.Host] = result.Value.([]entity.RemoteBackup)
		}
	}

	return expiredBackups, results.Error()
}

// ListBackups returns a list of all backups in remote storage
func (m *Octopus) ListBackups(ctx context.Context) (entity.RemoteBackupsByHost, error) {
	results := m.cluster.RunParallel(ctx, func(ctx context.Context, node *entity.Node) entity.NodeCallbackResult {
		backups, err := m.storage.ListBackups(ctx, node.Cmd, node.Info.RemoteStoragePath())
		if err != nil {
			return entity.CallbackError(err)
		}

		return entity.CallbackOk(backups)
	})

	expiredBackups := entity.RemoteBackupsByHost{}
	for _, result := range results {
		if result.Err == nil {
			expiredBackups[result.Host] = result.Value.([]entity.RemoteBackup)
		}
	}

	return expiredBackups, results.Error()
}
