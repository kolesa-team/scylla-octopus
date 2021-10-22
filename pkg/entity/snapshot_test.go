package entity

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseSnapshots(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected Snapshots
	}{
		{
			name: "a non-empty output of `nodetool listsnapshots`",
			output: `
Using /etc/scylla/scylla.yaml as the config file
Snapshot Details:
Snapshot name Keyspace name      Column family name              True size Size on disk
snapshot-1 system             compaction_history              18.04 KB  18.04 KB
snapshot-1 system             clients                         6.56 KB   6.56 KB
snapshot-1 system_schema      scylla_tables                   13.85 KB  13.85 KB
snapshot-2       system             large_cells                     0 bytes   0 bytes

Total TrueDiskSpaceUsed: 401.44 KiB
`,
			expected: Snapshots{
				"snapshot-1": {
					Tag: "snapshot-1",
					Items: []SnapshotItem{
						{
							Keyspace:     "system",
							ColumnFamily: "compaction_history",
						},
						{
							Keyspace:     "system",
							ColumnFamily: "clients",
						},
						{
							Keyspace:     "system_schema",
							ColumnFamily: "scylla_tables",
						},
					},
				},
				"snapshot-2": {
					Tag: "snapshot-2",
					Items: []SnapshotItem{
						{
							Keyspace:     "system",
							ColumnFamily: "large_cells",
						},
					},
				},
			},
		},
		{
			name: "an empty output of `nodetool listsnapshots`",
			output: `
Using /etc/scylla/scylla.yaml as the config file
There are no snapshots`,
			expected: Snapshots{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ParseSnapshots(tt.output)
			require.Equal(t, tt.expected, actual)
		})
	}
}
