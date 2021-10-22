package backup

import (
	"context"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"time"
)

// Cleanup removes temporary files on database node, as well as expired backups from this node in remote storage.
func (s *Service) Cleanup(ctx context.Context, node *entity.Node, snapshotTag string) entity.CleanupResult {
	result := entity.CleanupResult{
		RemovedRemoteBackups: []entity.RemoteBackup{},
	}

	if s.options.CleanupLocal {
		result.LocalError = s.cleanupLocal(ctx, node, snapshotTag)
	}

	if s.options.CleanupRemote {
		result.RemovedRemoteBackups, result.RemoteError = s.CleanupExpiredBackups(ctx, node, time.Now())
	}

	return result
}

// CleanupExpiredBackups removes expired backups from a node in remote storage.
func (s *Service) CleanupExpiredBackups(
	ctx context.Context,
	node *entity.Node,
	now time.Time,
) ([]entity.RemoteBackup, error) {
	expiredBackups, err := s.ListExpiredBackups(ctx, node, now)
	if err != nil {
		return []entity.RemoteBackup{}, err
	}

	for i, backup := range expiredBackups {
		backup.RemoveError = s.remoteStorage.RemoveBackup(ctx, node.Cmd, backup.Path)
		if backup.RemoveError == nil {
			backup.Removed = true
		}

		expiredBackups[i] = backup
	}

	return expiredBackups, err
}

// removes given snapshot and temporary files on database node
func (s *Service) cleanupLocal(ctx context.Context, node *entity.Node, snapshotTag string) error {
	logCtx := s.logger.With("host", node.Info.Host)
	err := s.scylla.RemoveSnapshot(ctx, node, snapshotTag)
	if err == nil {
		logCtx.Infow("database snapshot removed", "tag", snapshotTag)
	}

	err = cmd.EnsureDirectoryIsEmpty(ctx, node.Cmd, s.options.LocalPath)
	if err != nil {
		logCtx.Errorw("could not remove local backup directory", "error", err)
	} else {
		logCtx.Info("local backup directory removed")
	}

	return err
}
