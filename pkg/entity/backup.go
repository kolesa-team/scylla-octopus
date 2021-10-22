package entity

import (
	"errors"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"
)

// a regexp to parse backup filename
// from a path like .../host-prefix/MM-DD-YYYY-HH-mm/...
var remoteBackupFromPathRegexp = regexp.MustCompile(`(?P<Host>[^/]+)/(?P<date>\d\d-\d\d-\d\d\d\d-\d\d-\d\d)`)

// RemoteBackup a backup info in remote storage
type RemoteBackup struct {
	Path        string
	HostPrefix  string
	DateCreated time.Time
	Removed     bool
	RemoveError error
}

type RemoteBackupsByHost map[string][]RemoteBackup

// NewRemoteBackupFromPath parses a backup info from its path in remote storage
func NewRemoteBackupFromPath(path string) (RemoteBackup, error) {
	matches := remoteBackupFromPathRegexp.FindStringSubmatch(path)
	if len(matches) < 3 {
		return RemoteBackup{}, errors.New("path is not a backup")
	}

	date, err := time.Parse(SnapshotTagDateFormat, matches[2])
	if err != nil {
		return RemoteBackup{}, err
	}

	return RemoteBackup{
		Path:        path,
		HostPrefix:  matches[1],
		DateCreated: date,
	}, nil
}

// IsExpired whether a backup has expired
func (r RemoteBackup) IsExpired(now time.Time, retention time.Duration) bool {
	if retention.Seconds() < 1 {
		// sanity check
		return false
	}

	return r.DateCreated.Before(now.Add(-retention))
}

func (r RemoteBackup) String() string {
	return r.Path
}

// BackupResult a result of running a backup on a single database node
type BackupResult struct {
	Error         error
	DateStarted   time.Time
	Duration      time.Duration
	SnapshotTag   string
	Keyspaces     []string
	Uploaded      bool
	CleanupResult CleanupResult
}

// BackupResults a list of backup results on multiple database nodes
type BackupResults struct {
	TotalNodes    int
	BackedUpNodes int
	ByHost        map[string]BackupResult
	Error         error
}

// Report creates a human-readable report about backup results
func (b BackupResults) Report() string {
	lines := []string{
		fmt.Sprintf("Total nodes: %d", b.TotalNodes),
		fmt.Sprintf("Backed up nodes: %d", b.BackedUpNodes),
		"",
	}

	lines = append(lines, "Details:")

	for host, result := range b.ByHost {
		lines = append(lines, fmt.Sprintf("%s", host))

		if result.Error != nil {
			lines = append(
				lines,
				"Error:",
				fmt.Sprintf(
					"%s",
					html.EscapeString(result.Error.Error()),
				),
			)
		}

		lines = append(lines, fmt.Sprintf(
			"Backup uploaded: %t",
			result.Uploaded,
		))
		lines = append(lines, fmt.Sprintf(
			"Duration: %s",
			result.Duration.String(),
		))
		lines = append(lines, fmt.Sprintf(
			"Snapshot tag: %s",
			result.SnapshotTag,
		))
		lines = append(lines, fmt.Sprintf(
			"Expired backups removed: %d",
			len(result.CleanupResult.RemovedRemoteBackups),
		))

		if result.CleanupResult.RemoteError != nil {
			lines = append(lines, fmt.Sprintf(
				"error while removing expired backups: %s",
				html.EscapeString(result.CleanupResult.RemoteError.Error()),
			))
		}

		if result.CleanupResult.LocalError != nil {
			lines = append(lines, fmt.Sprintf(
				"error while removing a snapshot on database node: %s",
				html.EscapeString(result.CleanupResult.LocalError.Error()),
			))
		}

		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

type CleanupResult struct {
	// an error that occurred while cleaning up files on a database node, if any
	LocalError error
	// an error that occurred while cleaning up a backup in remote storage, if any
	RemoteError          error
	RemovedRemoteBackups []RemoteBackup
}
