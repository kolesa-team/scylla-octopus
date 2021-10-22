package entity

import (
	"strings"
)

// Snapshot holds an information about database snapshot
type Snapshot struct {
	Tag string
	// keeps the snapshots of specific database tables and indices
	Items []SnapshotItem
}

// Snapshots a set of snapshots, where the map key is snapshot tag
type Snapshots map[string]Snapshot

// SnapshotsByNode a set of snapshots on multiple nodes, where the map key is a hostname
type SnapshotsByNode map[string]Snapshots

// SnapshotItem a table or index, part of a snapshot
type SnapshotItem struct {
	Keyspace     string
	ColumnFamily string
}

// ParseSnapshots returns a list of snapshots based on `nodetool listsnapshot` output.
func ParseSnapshots(output string) Snapshots {
	snapshots := map[string]Snapshot{}

	if strings.Contains(output, "There are no snapshots") {
		return snapshots
	}

	lines := strings.Split(output, "\n")
	snapshotListFound := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if !snapshotListFound {
			if strings.HasPrefix(line, "Snapshot Details:") {
				snapshotListFound = true
				i++
			}

			continue
		}

		if strings.HasPrefix(line, "Total") {
			break
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		snapshotTag := parts[0]
		keyspaceName := parts[1]
		columnFamilyName := parts[2]

		snapshot, ok := snapshots[snapshotTag]
		if !ok {
			snapshot = Snapshot{
				Tag:   snapshotTag,
				Items: []SnapshotItem{},
			}
		}
		snapshot.Items = append(snapshot.Items, SnapshotItem{
			Keyspace:     keyspaceName,
			ColumnFamily: columnFamilyName,
		})
		snapshots[snapshotTag] = snapshot
	}

	return snapshots
}
