package entity

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestBackupResults_Report(t *testing.T) {
	results := BackupResults{
		TotalNodes:    3,
		BackedUpNodes: 2,
		ByHost: map[string]BackupResult{
			// a successful backup
			"127.0.0.1": {
				Duration:    time.Second,
				SnapshotTag: "snapshot-tag-1",
				Uploaded:    true,
				CleanupResult: CleanupResult{
					// there are 2 backups from this node in remote storage
					// (specific backup properties are unimportant)
					RemovedRemoteBackups: []RemoteBackup{
						{}, {},
					},
				},
			},
			// a failed backup
			"127.0.0.2": {
				Duration:    time.Second * 2,
				Uploaded:    false,
				SnapshotTag: "snapshot-tag-2",
				Error:       errors.New("could not backup a node"),
			},
			// a successful backup,
			// but with an error while removing expired backups
			"127.0.0.3": {
				Duration:    time.Second * 3,
				Uploaded:    true,
				SnapshotTag: "snapshot-tag-3",
				CleanupResult: CleanupResult{
					RemoteError: errors.New("could not remove expired backups"),
				},
			},
		},
	}

	report := results.Report()

	require.Contains(t, report, `Total nodes: 3
Backed up nodes: 2`)

	require.Contains(t, report, `127.0.0.1
Backup uploaded: true
Duration: 1s
Snapshot tag: snapshot-tag-1
Expired backups removed: 2`)

	require.Contains(t, report, `127.0.0.2
Error:
could not backup a node
Backup uploaded: false
Duration: 2s
Snapshot tag: snapshot-tag-2
Expired backups removed: 0`)

	require.Contains(t, report, `127.0.0.3
Backup uploaded: true
Duration: 3s
Snapshot tag: snapshot-tag-3
Expired backups removed: 0
error while removing expired backups: could not remove expired backups`)
}

func TestNewRemoteBackupFromPath(t *testing.T) {
	tests := []struct {
		name    string
		want    RemoteBackup
		path    string
		wantErr bool
	}{
		{
			name:    "empty line (invalid path)",
			path:    "",
			wantErr: true,
		},
		{
			name: "valid backup path",
			path: "/backup-test/common-scylla1-dc1/09-07-2021-10-29",
			want: RemoteBackup{
				Path:        "/backup-test/common-scylla1-dc1/09-07-2021-10-29",
				HostPrefix:  "common-scylla1-dc1",
				DateCreated: time.Date(2021, 9, 7, 10, 29, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name:    "a path with invalid date format",
			path:    "/backup-test/common-scylla1-dc1/99-99-2021-10-29",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRemoteBackupFromPath(tt.path)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestRemoteBackup_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		now       time.Time
		retention time.Duration
		date      time.Time
		want      bool
	}{
		{
			name:      "a backup should be expired in 2 hours if retention=1h",
			now:       time.Now(),
			retention: time.Hour,
			date:      time.Now().Add(-time.Hour * 2),
			want:      true,
		},
		{
			name:      "a backup should not be expired in 5 minutes if retention=1h",
			now:       time.Now(),
			retention: time.Hour,
			date:      time.Now().Add(-time.Minute * 5),
			want:      false,
		},
		{
			name:      "a backup should never be expired if retention=0",
			now:       time.Now(),
			retention: 0,
			date:      time.Now().AddDate(-1, 0, 0),
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RemoteBackup{
				DateCreated: tt.date,
			}
			require.Equal(t, tt.want, r.IsExpired(tt.now, tt.retention))
		})
	}
}
