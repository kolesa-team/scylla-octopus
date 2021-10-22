package backup

import (
	"context"
	"github.com/pkg/errors"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
)

const metadataFilename = "metadata.yml"

// adds metadata to a backup, so it helps us with restoration in future versions
func (s *Service) writeMetadata(ctx context.Context, cmd cmd.Executor, host string, metadata entity.BackupMetadata) error {
	targetPath := s.options.LocalPath + "/" + metadataFilename
	err := cmd.WriteFile(ctx, targetPath, metadata.Bytes())

	if err != nil {
		return errors.Wrapf(
			err,
			"could not write metadata on %s to %s",
			host,
			targetPath,
		)
	} else {
		s.logger.Debugw(
			"backup metadata added",
			"host", host,
			"path", targetPath,
		)
	}

	return nil
}
