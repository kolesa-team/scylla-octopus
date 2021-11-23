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
Datacenter: DC1
=================
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address      Load       Tokens       Owns    Host ID                               Rack
UN  172.20.0.1  180.47 GB  256          ?       a36d1408-f32b-469f-b22e-df314becc200  R1
UN  172.20.0.2  169.46 GB  256          ?       985284bd-6373-480d-9780-05cbf82954de  R1
UN  172.20.0.3  182.25 GB  256          ?       8028683c-366c-49d3-b8e4-442dfdf7c4af  R1
UN  172.20.0.4  175.18 GB  256          ?       7aa8f49f-3855-4d34-a4b2-96226e86006e  R1
Datacenter: DC2
=================
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address      Load       Tokens       Owns    Host ID                               Rack
UN  172.20.1.1  175.68 GB  256          ?       aa355bca-2a86-4530-8671-564342d1eee2  R1
UN  172.20.1.2  169.21 GB  256          ?       485f0ef9-9269-427a-b974-926175017bd5  R1
UN  172.20.1.3  156.25 GB  256          ?       c71a980d-dcea-47e4-b0a9-0696618224de  R1
UN  172.20.1.4  186.78 GB  256          ?       56e0d871-91ea-4d7f-92e9-58fa99dac8ad  R1
`,
	}
	client := Client{
		credentials: entity.Credentials{},
		logger:      zap.S(),
	}
	status, dc, err := client.getNodeStatus(context.Background(), entity.NewNode(
		entity.NodeInfo{
			IpAddress: "172.20.1.2",
		},
		cmdExecutor,
		nil,
	))
	require.NoError(t, err)
	require.Equal(t, entity.NodeStatusOk, status)
	require.Equal(t, "DC2", dc)
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
