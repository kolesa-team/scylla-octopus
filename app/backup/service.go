package backup

import (
	"context"
	"github.com/kolesa-team/scylla-octopus/pkg/archive"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"github.com/kolesa-team/scylla-octopus/pkg/notifier"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"strings"
	"time"
)

// Service is a backup management service.
// It operates on a single node.
type Service struct {
	options       Options
	logger        *zap.SugaredLogger
	scylla        dbClient
	remoteStorage remoteStorageClient
	notifier      notifier.Notifier
	buildInfo     entity.BuildInfo
}

type Options struct {
	// Scylladb snapshot tag.
	// If empty, the snapshot tag is generated using current date and a hostname.
	SnapshotTag string `yaml:"snapshotTag"`
	// Where to store a backup on a database host before uploading to s3
	LocalPath string `yaml:"localPath"`
	// A list of keyspaces to back up.
	// Empty means all keyspaces.
	Keyspaces []string
	// Turns off upload to a remote storage
	DisableUpload bool `yaml:"disableUpload"`
	// Should local directories be cleaned up after uploading to remote storage
	CleanupLocal bool `yaml:"cleanupLocal"`
	// Should the expired backups be removed from s3 when creating a new backup
	CleanupRemote bool `yaml:"cleanupRemote"`
	// How long should the backups live in remote storage
	Retention time.Duration
	// Settings for compress backup
	Archive entity.Archive
}

// database client interface (implemented by `pkg/scylla`)
type dbClient interface {
	ExportSchema(ctx context.Context, node *entity.Node, path string) (string, error)
	CreateSnapshot(ctx context.Context, node *entity.Node, tag, path string, keyspaces []string) error
	RemoveSnapshot(ctx context.Context, node *entity.Node, tag string) error
}

// remote storage interface ( (implemented by `pkg/awscli`)
type remoteStorageClient interface {
	Upload(ctx context.Context, cmdExecutor cmd.Executor, source, dest string) (string, error)
	ListBackups(ctx context.Context, cmdExecutor cmd.Executor, basePath string) ([]entity.RemoteBackup, error)
	RemoveBackup(ctx context.Context, cmdExecutor cmd.Executor, string string) error
}

func NewService(
	options Options,
	buildInfo entity.BuildInfo,
	db dbClient,
	remoteStorage remoteStorageClient,
	notifier notifier.Notifier,
	logger *zap.SugaredLogger,
) *Service {
	options.LocalPath = strings.TrimRight(options.LocalPath, "/")

	return &Service{
		options:       options,
		scylla:        db,
		remoteStorage: remoteStorage,
		notifier:      notifier,
		logger:        logger,
		buildInfo:     buildInfo,
	}
}

// Healthcheck ensures the directories for backup exist on a node
func (s *Service) Healthcheck(ctx context.Context, node *entity.Node) error {
	s.logger.Debugw("[healthcheck] checking backup directory", "host", node.Info.Host)

	if !cmd.DirectoryExists(ctx, node.Cmd, s.options.LocalPath) {
		err := cmd.CreateDirectory(ctx, node.Cmd, s.options.LocalPath)
		if err != nil {
			return errors.Wrapf(
				err,
				"directory %s does not exist and could not be created",
				s.options.LocalPath,
			)
		}
	}

	return nil
}

// Backup creates a database backup, uploads it to remote storage, and cleans up temporary files
func (s *Service) Backup(ctx context.Context, node *entity.Node) entity.BackupResult {
	result := entity.BackupResult{
		SnapshotTag: s.options.SnapshotTag,
		Keyspaces:   s.options.Keyspaces,
		DateStarted: time.Now(),
	}

	if len(result.SnapshotTag) == 0 {
		result.SnapshotTag = entity.NewSnapshotTag(
			node.Info.ShortDomainName(),
			result.DateStarted,
		)
	}

	result.Error = s.exportSnapshot(ctx, node, result.SnapshotTag)
	if result.Error != nil {
		return result
	}

	metadata := entity.BackupMetadata{
		DateCreated: time.Now(),
		Host:        node.Info.Host,
		Keyspaces:   s.options.Keyspaces,
		SnapshotTag: result.SnapshotTag,
		BuildInfo:   s.buildInfo,
	}

	if s.options.Archive.Method != "" {
		metadata.Archive = s.options.Archive
	}

	result.Error = s.writeMetadata(ctx, node.Cmd, node.Info.Host, metadata)
	if result.Error != nil {
		return result
	}

	if ctx.Err() != nil {
		result.Error = ctx.Err()
		return result
	}

	if !s.options.DisableUpload {
		remotePath := node.Info.RemoteStoragePath() + "/" + entity.BackupDateToPath(result.DateStarted)

		if result.Error = s.upload(ctx, node, remotePath); result.Error != nil {
			return result
		}

		if ctx.Err() != nil {
			result.Error = ctx.Err()
			return result
		}

		result.Uploaded = true
		result.CleanupResult = s.Cleanup(ctx, node, result.SnapshotTag)
	}

	result.Duration = time.Now().Sub(result.DateStarted)

	return result
}

// ListExpiredBackups returns expired backups from a node.
func (s *Service) ListExpiredBackups(
	ctx context.Context,
	node *entity.Node,
	now time.Time,
) ([]entity.RemoteBackup, error) {
	backups, err := s.remoteStorage.ListBackups(ctx, node.Cmd, node.Info.RemoteStoragePath())
	if err != nil {
		return []entity.RemoteBackup{}, err
	}

	expiredBackups := []entity.RemoteBackup{}

	for _, backup := range backups {
		if backup.IsExpired(now, s.options.Retention) {
			expiredBackups = append(expiredBackups, backup)
		}
	}

	return expiredBackups, err
}

// creates a snapshot with a given tag and exports a database schema
func (s *Service) exportSnapshot(ctx context.Context, node *entity.Node, snapshotTag string) error {
	logCtx := s.logger.With("host", node.Info.Host)
	targetDir := s.options.LocalPath
	dataDir := targetDir + "/data"
	err := cmd.EnsureDirectoryIsEmpty(ctx, node.Cmd, targetDir)
	if err != nil {
		return errors.Wrapf(
			err,
			"directory %s does not exist or not empty",
			targetDir,
		)
	}

	err = cmd.CreateDirectory(ctx, node.Cmd, dataDir)
	if err != nil {
		return errors.Wrapf(err, "could not create data directory")
	}

	logCtx.Info("exporting schema")
	_, err = s.scylla.ExportSchema(ctx, node, targetDir)
	if err != nil {
		return err
	}

	logCtx.Infow("creating snapshot",
		"tag", snapshotTag,
		"target", dataDir,
	)
	err = s.scylla.CreateSnapshot(ctx, node, snapshotTag, dataDir, s.options.Keyspaces)
	if err != nil {
		return err
	}

	if s.options.Archive.Method != "" {
		err = archive.Compression(ctx, node, s.options.LocalPath, s.options.Archive)

		if err != nil {
			return err
		}
	}

	logCtx.Infow("snapshot created", "tag", snapshotTag)

	return nil
}

// checks whether a local backup with a given tag exists
func (s *Service) localBackupExists(ctx context.Context, cmdExecutor cmd.Executor, snapshotTag string) bool {
	return cmd.DirectoryExists(ctx, cmdExecutor, s.options.LocalPath+"/"+snapshotTag)
}

// Uploads the local directory to a remote storage.
// Creates the following directory hierarchy:
//├── cluster_name
//│   ├── node_1_short_domain_name
//│   │   ├── date(dd-mm-yyy-hh-mm)
//│   │   │   ├── db_schema.cql
//│   │   │   ├── data
//│   │   │   │   ├── keyspaces
//│   │   │   │   │   ├── tables
//│   ├── node_2_short_domain_name...
func (s *Service) upload(ctx context.Context, node *entity.Node, remotePath string) error {
	logCtx := s.logger.With("host", node.Info.Host, "remotePath", remotePath)
	logCtx.Infow("uploading backup", "remotePath", remotePath)
	url, err := s.remoteStorage.Upload(
		ctx,
		node.Cmd,
		s.options.LocalPath,
		remotePath,
	)
	if err != nil {
		return err
	}

	logCtx.Infow("backup uploaded", "url", url)

	return nil
}
