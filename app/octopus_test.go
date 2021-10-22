package app

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"github.com/kolesa-team/scylla-octopus/pkg/notifier"
	"go.uber.org/zap"
	"testing"
)

// A successful healthcheck
func TestOctopus_Healthcheck_Ok(t *testing.T) {
	// a test implementation of database cluster that always returns successful results
	cluster := testCluster{
		nodeCount: 2,
		callbackResults: map[string]entity.NodeCallbackResult{
			"host-1": {
				Host: "host-1",
			},
			"host-2": {
				Host: "host-2",
			},
		},
	}
	app := NewOctopus(
		cluster,
		testDb{},
		testBackupService{},
		testStorage{},
		notifier.Disabled{},
		zap.S(),
	)

	result, err := app.Healthcheck(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 2, "a result must contain information about 2 nodes")
	require.Equal(t, map[string]string{
		"host-1": "OK",
		"host-2": "OK",
	}, result)
}

// A healthcheck with an error
func TestOctopus_Healthcheck_WithError(t *testing.T) {
	// a test implementation of a cluster returning 2 healthcheck results:
	// 1 is successful and 1 is not
	cluster := testCluster{
		nodeCount: 2,
		callbackResults: map[string]entity.NodeCallbackResult{
			"host-1": {
				Host: "host-1",
			},
			"host-2": {
				Host: "host-2",
				Err:  errors.New("test error"),
			},
		},
	}
	app := NewOctopus(
		cluster,
		testDb{},
		testBackupService{},
		testStorage{},
		notifier.Disabled{},
		zap.S(),
	)

	result, err := app.Healthcheck(context.Background())

	require.Error(t, err)
	require.Len(t, result, 2, "a result must contain information about 2 nodes")
	require.Equal(t, map[string]string{
		"host-1": "OK",
		"host-2": "test error",
	}, result)
}
