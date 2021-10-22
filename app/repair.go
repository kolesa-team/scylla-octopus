package app

import (
	"context"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
)

// Repair executes `nodetool repair` on every cluster node consecutively.
func (m *Octopus) Repair(ctx context.Context) entity.RepairResults {
	callbackResults := m.cluster.Run(ctx, func(ctx context.Context, node *entity.Node) entity.NodeCallbackResult {
		result, err := m.scylla.Repair(ctx, node)
		if err != nil {
			return entity.CallbackError(err)
		}

		return entity.CallbackOk(result)
	})

	repairResults := entity.RepairResults{
		TotalNodes: m.cluster.Size(),
		ByHost:     map[string]entity.RepairResult{},
		Error:      callbackResults.Error(),
	}

	for _, callbackResult := range callbackResults {
		if callbackResult.Err != nil {
			continue
		}

		repairResults.ByHost[callbackResult.Host] = callbackResult.Value.(entity.RepairResult)
		repairResults.RepairedNodes++
	}

	if repairResults.Error != nil {
		m.notifier.Error(
			"Could not execute nodetool repair",
			repairResults.Report(),
			repairResults.Error,
			nil,
		)
	} else {
		m.notifier.Info(
			"nodetool repair executed successfully",
			repairResults.Report(),
			nil,
		)
	}

	return repairResults
}
