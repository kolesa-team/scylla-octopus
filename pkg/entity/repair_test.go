package entity

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRepairResults_Report(t *testing.T) {
	results := RepairResults{
		TotalNodes:    3,
		RepairedNodes: 2,
		ByHost: map[string]RepairResult{
			"localhost": {
				Duration: time.Second,
			},
			"google.com": {
				Duration: time.Second * 2,
			},
		},
		Error: errors.New("test error <test-tag>"),
	}

	report := results.Report()

	require.Contains(
		t,
		report,
		`Total nodes: 3
Repaired nodes: 2

Error:
test error &lt;test-tag&gt;`,
	)

	require.Contains(t, report, `localhost: 1s`)
	require.Contains(t, report, `google.com: 2s`)
}
