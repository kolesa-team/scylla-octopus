package entity

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd"
	"net"
	"strings"
)

// NodeStatusOk a normal database node status (Up, Normal)
const NodeStatusOk = "UN"

// NodeInfo database node information
type NodeInfo struct {
	// ip or domain name
	Host        string
	IpAddress   string
	DomainName  string
	DataPath    string
	ClusterName string
	// a status according to `nodetool status`
	Status   string
	Binaries NodeBinaries
}

type Node struct {
	Info NodeInfo
	// a shell command executor on this node
	Cmd cmd.Executor
	// an SSH connection error, if any
	ConnectionErr error
}

func NewNode(info NodeInfo, cmdExecutor cmd.Executor, connectionErr error) *Node {
	return &Node{
		Info:          info,
		Cmd:           cmdExecutor,
		ConnectionErr: connectionErr,
	}
}

// NodeCallback a callback function to be applied on database node
type NodeCallback func(ctx context.Context, node *Node) NodeCallbackResult

// NodeCallbackResult a result of running a callback on database node
type NodeCallbackResult struct {
	Host  string
	Value interface{}
	Err   error
}

// CallbackOk creates a callback result with a given value
func CallbackOk(value interface{}) NodeCallbackResult {
	return NodeCallbackResult{Value: value}
}

// CallbackError creates a callback result with a given error
func CallbackError(err error) NodeCallbackResult {
	return NodeCallbackResult{Err: err}
}

// CallbackErrorWithValue creates a callback result with a given error and value
func CallbackErrorWithValue(err error, value interface{}) NodeCallbackResult {
	return NodeCallbackResult{Err: err, Value: value}
}

// NodeCallbackResults callback results by hosts
type NodeCallbackResults map[string]NodeCallbackResult

// Returns errors from callback results (if any) as a single error
func (results NodeCallbackResults) Error() error {
	var err *multierror.Error

	for _, result := range results {
		if result.Err != nil {
			err = multierror.Append(err, result.Err)
		}
	}

	return err.ErrorOrNil()
}

func (n NodeInfo) IsStatusOk() bool {
	return n.Status == NodeStatusOk
}

// RemoteStoragePath returns a path where the backup from a node should be stored in s3.
// The format is "cluster-name/short-domain-name"
func (n NodeInfo) RemoteStoragePath() string {
	return fmt.Sprintf(
		"%s/%s",
		n.ClusterName,
		n.ShortDomainName(),
	)
}

// ShortDomainName returns a short domain name of a node (up to the first "."), if it is known.
// Otherwise, returns the "host" value from configuration.
// Short name is used as a part of a backup path when storing it in s3.
func (n NodeInfo) ShortDomainName() string {
	if len(n.DomainName) == 0 {
		return n.Host
	} else {
		return strings.Split(n.DomainName, ".")[0]
	}
}

// NewNodeInfo initializes a node information.
// If resolveDns=true, then it will attempt to resolve the domain name and/or IP address.
func NewNodeInfo(host, dataPath string, binaries NodeBinaries, resolveDns bool) NodeInfo {
	info := NodeInfo{
		Host:     host,
		DataPath: dataPath,
		Binaries: binaries,
	}
	hostIsIp := net.ParseIP(info.Host) != nil

	if len(info.IpAddress) == 0 && hostIsIp {
		// if Host is an ip address, then just use that
		info.IpAddress = info.Host
	}

	if len(info.DomainName) == 0 {
		if hostIsIp && resolveDns {
			// if we have an IP address, then we'll try to find its domain name
			addr, err := net.LookupAddr(info.Host)

			if err == nil && len(addr) > 0 {
				info.DomainName = addr[0]
			}
		} else {
			// otherwise use Host as domain name
			info.DomainName = info.Host
		}
	}

	if len(info.IpAddress) == 0 && resolveDns {
		// if the IP address is still unknown, try to resolve it by domain name
		ipAddr, err := net.LookupIP(info.DomainName)
		if err == nil && len(ipAddr) > 0 {
			info.IpAddress = ipAddr[0].String()
		}
	}

	return info
}
