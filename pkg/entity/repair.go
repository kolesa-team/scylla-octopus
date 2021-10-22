package entity

import (
	"fmt"
	"html"
	"strings"
	"time"
)

// RepairResult a result of "nodetool repair" command
type RepairResult struct {
	Output   string
	Duration time.Duration
}

// RepairResults results of "nodetool repair" on multiple database nodes
type RepairResults struct {
	TotalNodes    int
	RepairedNodes int
	ByHost        map[string]RepairResult
	Error         error
}

// Report creates a human-readable report string about repair results to be used in a notification
func (r RepairResults) Report() string {
	lines := []string{
		fmt.Sprintf("Total nodes: %d", r.TotalNodes),
		fmt.Sprintf("Repaired nodes: %d", r.RepairedNodes),
		"",
	}

	if r.Error != nil {
		lines = append(
			lines,
			"Error:",
			fmt.Sprintf("%s", html.EscapeString(r.Error.Error())),
		)
	}

	lines = append(lines, "Duration:")

	for host, result := range r.ByHost {
		lines = append(lines, fmt.Sprintf(
			"%s: %s",
			host,
			result.Duration.String(),
		))
	}

	return strings.Join(lines, "\n")
}
