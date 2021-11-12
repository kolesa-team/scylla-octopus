package backup

import (
	"context"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd/test"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestService_writeMetadata(t *testing.T) {
	service := &Service{
		options: Options{LocalPath: "/backup"},
		logger:  zap.S(),
	}
	cmdExecutor := test.Executor{}
	metadata := entity.BackupMetadata{
		DateCreated: time.Date(2021, 10, 22, 15, 1, 0, 0, time.UTC),
		Host:        "test-host",
		Keyspaces:   []string{"a", "b", "c"},
		SnapshotTag: "snapshot-tag",
		BuildInfo: entity.BuildInfo{
			Version: "1.0.0",
			Commit:  "abcdef",
			Date:    "2021-10-22",
		},
		Archive: entity.Archive{
			Method:         "pigz",
			ArchiveOptions: entity.ArchiveOptions{
				Compression: "1",
				Threads:     "2",
			},
		},
	}

	err := service.writeMetadata(context.Background(), &cmdExecutor, "test-host", metadata)
	require.NoError(t, err)

	require.Equal(t, "/backup/metadata.yml", cmdExecutor.WrittenFilePath)
	require.Equal(
		t,
		`dateCreated: 2021-10-22T15:01:00Z
host: test-host
keyspaces:
    - a
    - b
    - c
snapshotTag: snapshot-tag
buildInfo:
    version: 1.0.0
    commit: abcdef
    date: "2021-10-22"
archive:
    method: pigz
    options:
        compression: "1"
        threads: "2"
`,
		string(cmdExecutor.WrittenFileBytes),
	)
}
