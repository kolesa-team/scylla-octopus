package entity

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNewSnapshotTag(t *testing.T) {
	now := time.Date(2020, 9, 2, 12, 33, 0, 0, time.Local)

	tests := []struct {
		name       string
		domainName string
		now        time.Time
		expected   string
	}{
		{
			domainName: "scylla-node1",
			now:        now,
			expected:   "scylla-node1-09-02-2020-12-33",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, NewSnapshotTag(tt.domainName, tt.now))
		})
	}
}
