package entity

import (
	"fmt"
	"time"
)

// SnapshotTagDateFormat a date format used in snapshot tag
const SnapshotTagDateFormat = "01-02-2006-15-04"

// NewSnapshotTag creates a snapshot tag based on database domain name and current time
func NewSnapshotTag(domainName string, now time.Time) string {
	return fmt.Sprintf(
		"%s-%s",
		domainName,
		BackupDateToPath(now),
	)
}

// BackupDateToPath converts a date to a directory path, where the backup will be stored in s3
func BackupDateToPath(date time.Time) string {
	return date.Format(SnapshotTagDateFormat)
}
