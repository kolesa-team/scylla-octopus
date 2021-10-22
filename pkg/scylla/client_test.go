package scylla

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd/test"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"testing"
)

func TestClient_ExportSchema(t *testing.T) {
	cmdExecutor := &test.Executor{}
	client := Client{
		credentials: entity.Credentials{
			User:     "root",
			Password: "pass",
		},
	}

	node := entity.NewNode(entity.NewNodeInfo(
		"scylla.test",
		"/var/scylla/data",
		entity.NodeBinaries{
			Cqlsh: "cqlsh",
		},
		false,
	), cmdExecutor, nil)
	path, err := client.ExportSchema(
		context.Background(),
		node,
		"/var/lib/backup",
	)
	require.NoError(t, err)
	require.Equal(t, "/var/lib/backup/db_schema.cql", path)
	require.Equal(
		t,
		`cqlsh scylla.test -u root -p pass -e "DESC SCHEMA" > /var/lib/backup/db_schema.cql`,
		cmdExecutor.LastCmd.String(),
	)
}
