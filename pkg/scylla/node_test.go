package scylla

import (
	"context"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd/test"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
)

func TestClient_NodeStatus_Ok(t *testing.T) {
	cmdExecutor := &test.Executor{
		// the output is copied from a real `nodetool status` execution
		Output: `
Using /etc/scylla/scylla.yaml as the config file
Datacenter: DC1
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address     Load       Tokens       Owns    GetByHost ID                               Rack
UN  172.20.0.2  205.89 KB  256          ?       ba29ac08-b405-4d26-8a8b-2b8c18c93b9c  Rack1
DN  172.20.0.3  205.89 KB  256          ?       ba29ac08-b405-4d26-8a8b-2b8c18c93b9c  Rack2
`,
	}
	client := Client{
		credentials: entity.Credentials{},
		logger:      zap.S(),
	}
	status, dc, err := client.getNodeStatus(context.Background(), entity.NewNode(
		entity.NodeInfo{
			IpAddress: "172.20.0.2",
		},
		cmdExecutor,
		nil,
	))
	require.NoError(t, err)
	require.Equal(t, entity.NodeStatusOk, status)
	require.Equal(t, "DC1", dc)
}

func TestClient_NodeStatus_Error(t *testing.T) {
	cmdExecutor := &test.Executor{
		Output: `
Using /etc/scylla/scylla.yaml as the config file
Datacenter: DC1
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address     Load       Tokens       Owns    GetByHost ID                               Rack
DN  172.20.0.2  205.89 KB  256          ?       ba29ac08-b405-4d26-8a8b-2b8c18c93b9c  Rack1
UN  172.20.0.3  205.89 KB  256          ?       ba29ac08-b405-4d26-8a8b-2b8c18c93b9c  Rack2
`,
	}
	client := Client{
		credentials: entity.Credentials{},
		logger:      zap.S(),
	}
	status, dc, err := client.getNodeStatus(context.Background(), entity.NewNode(
		entity.NodeInfo{
			IpAddress: "172.20.0.2",
		},
		cmdExecutor,
		nil,
	))
	require.NoError(t, err)
	require.Equal(t, "DC1", dc)
	require.Equal(t, "DN", status, "node status must be DN (down, normal)")
}

func TestClient_ClusterName_Ok(t *testing.T) {
	cmdExecutor := &test.Executor{
		Output: `Using /etc/scylla/scylla.yaml as the config file
Cluster Information:
        Name: Test Cluster
        Snitch: org.apache.cassandra.locator.GossipingPropertyFileSnitch
        DynamicEndPointSnitch: disabled
        Partitioner: org.apache.cassandra.dht.Murmur3Partitioner
        Schema versions:
                497cda5f-fdb2-3b57-8b68-611d4c6200f5: [172.20.0.3]`,
	}
	client := Client{}

	name, err := client.getClusterName(context.Background(), entity.NewNode(entity.NodeInfo{}, cmdExecutor, nil))
	require.NoError(t, err)
	require.Equal(t, "Test Cluster", name)
}

func TestClient_ClusterName_Error(t *testing.T) {
	cmdExecutor := &test.Executor{
		Output: `invalid output`,
	}
	client := Client{}

	_, err := client.getClusterName(context.Background(), entity.NewNode(entity.NodeInfo{}, cmdExecutor, nil))
	require.Error(t, err)
}
