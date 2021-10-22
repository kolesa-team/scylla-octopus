package app

import (
	"context"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"github.com/kolesa-team/scylla-octopus/pkg/notifier"
	"go.uber.org/zap"
)

// Octopus operates a cluster of scylladb nodes
type Octopus struct {
	cluster  cluster
	scylla   dbClient
	backup   backupService
	storage  remoteStorageClient
	notifier notifier.Notifier
	logger   *zap.SugaredLogger
}

func NewOctopus(
	cluster cluster,
	scylla dbClient,
	backup backupService,
	storage remoteStorageClient,
	notifier notifier.Notifier,
	logger *zap.SugaredLogger,
) *Octopus {
	return &Octopus{
		cluster:  cluster,
		scylla:   scylla,
		backup:   backup,
		storage:  storage,
		notifier: notifier,
		logger:   logger,
	}
}

// Healthcheck checks is each node in cluster is OK.
// Returns a map of hostnames to strings, that contain either "OK" or an error.
func (m *Octopus) Healthcheck(ctx context.Context) (map[string]string, error) {
	results := m.cluster.Connect(ctx)

	if results.Error() == nil {
		results = m.cluster.RunParallel(ctx, func(ctx context.Context, node *entity.Node) entity.NodeCallbackResult {
			// check the scylladb
			err := m.scylla.Healthcheck(ctx, node)
			if err != nil {
				return entity.CallbackError(err)
			}

			// check the awscli
			err = m.storage.Healthcheck(ctx, node.Cmd)
			if err != nil {
				return entity.CallbackError(err)
			}

			// check backup directories
			err = m.backup.Healthcheck(ctx, node)
			if err != nil {
				return entity.CallbackError(err)
			}

			return entity.CallbackOk(nil)
		})
	}

	report := map[string]string{}

	for _, result := range results {
		if result.Err != nil {
			report[result.Host] = result.Err.Error()
		} else {
			report[result.Host] = "OK"
		}
	}

	return report, results.Error()
}
