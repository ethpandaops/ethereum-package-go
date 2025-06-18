package kurtosis

import (
	"testing"

	"github.com/ethpandaops/ethereum-package-go/pkg/client"
	"github.com/stretchr/testify/assert"
)

func TestConvertServiceInfoToExecutionClient(t *testing.T) {
	service := &ServiceInfo{
		Name:      "geth-1",
		UUID:      "uuid-123",
		IPAddress: "172.16.0.2",
		Ports: map[string]PortInfo{
			"rpc":     {Number: 8545, Protocol: "TCP"},
			"ws":      {Number: 8546, Protocol: "TCP"},
			"engine":  {Number: 8551, Protocol: "TCP"},
			"metrics": {Number: 9090, Protocol: "TCP"},
			"p2p":     {Number: 30303, Protocol: "TCP"},
		},
	}

	execClient := ConvertServiceInfoToExecutionClient(service, client.Geth)
	assert.NotNil(t, execClient)
	assert.Equal(t, "geth-1", execClient.Name())
	assert.Equal(t, client.Geth, execClient.Type())
	assert.Equal(t, "http://172.16.0.2:8545", execClient.RPCURL())
	assert.Equal(t, "ws://172.16.0.2:8546", execClient.WSURL())
	assert.Equal(t, "http://172.16.0.2:8551", execClient.EngineURL())
	assert.Equal(t, "http://172.16.0.2:9090", execClient.MetricsURL())
	assert.Equal(t, 30303, execClient.P2PPort())
}

func TestConvertServiceInfoToConsensusClient(t *testing.T) {
	service := &ServiceInfo{
		Name:      "lighthouse-1",
		UUID:      "uuid-456",
		IPAddress: "172.16.0.3",
		Ports: map[string]PortInfo{
			"beacon":  {Number: 5052, Protocol: "TCP"},
			"metrics": {Number: 5054, Protocol: "TCP"},
			"p2p":     {Number: 9000, Protocol: "TCP"},
		},
	}

	consClient := ConvertServiceInfoToConsensusClient(service, client.Lighthouse)
	assert.NotNil(t, consClient)
	assert.Equal(t, "lighthouse-1", consClient.Name())
	assert.Equal(t, client.Lighthouse, consClient.Type())
	assert.Equal(t, "http://172.16.0.3:5052", consClient.BeaconAPIURL())
	assert.Equal(t, "http://172.16.0.3:5054", consClient.MetricsURL())
	assert.Equal(t, 9000, consClient.P2PPort())
}

func TestDetectClientType(t *testing.T) {
	tests := []struct {
		name         string
		serviceName  string
		expectedType client.Type
	}{
		// Execution clients
		{"detect geth", "el-1-geth-lighthouse", client.Geth},
		{"detect besu", "el-2-besu", client.Besu},
		{"detect nethermind", "nethermind-node-1", client.Nethermind},
		{"detect erigon", "erigon-archive", client.Erigon},
		{"detect reth", "reth-full-node", client.Reth},

		// Consensus clients
		{"detect lighthouse", "cl-1-lighthouse-geth", client.Lighthouse},
		{"detect teku", "teku-validator", client.Teku},
		{"detect prysm", "prysm-beacon", client.Prysm},
		{"detect nimbus", "nimbus-eth2", client.Nimbus},
		{"detect lodestar", "lodestar-beacon", client.Lodestar},
		{"detect grandine", "grandine-full", client.Grandine},

		// Unknown
		{"unknown client", "random-service", client.Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detected := DetectClientType(tt.serviceName)
			assert.Equal(t, tt.expectedType, detected)
		})
	}
}

func TestConvertWithMaybeURLs(t *testing.T) {
	service := &ServiceInfo{
		Name:      "geth-1",
		UUID:      "uuid-123",
		IPAddress: "172.16.0.2",
		Ports: map[string]PortInfo{
			"rpc": {
				Number:   8545,
				Protocol: "TCP",
				MaybeURL: "http://custom-domain:8545",
			},
		},
	}

	execClient := ConvertServiceInfoToExecutionClient(service, client.Geth)
	assert.Equal(t, "http://custom-domain:8545", execClient.RPCURL())
}

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"geth-lighthouse", "geth", true},
		{"GETH-lighthouse", "geth", true},
		{"geth-LIGHTHOUSE", "lighthouse", true},
		{"random-name", "geth", false},
		{"", "test", false},
		{"test", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			result := containsIgnoreCase(tt.s, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEqualIgnoreCase(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected bool
	}{
		{"test", "test", true},
		{"Test", "test", true},
		{"TEST", "test", true},
		{"test", "TEST", true},
		{"test", "different", false},
		{"", "", true},
		{"test", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			result := equalIgnoreCase(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		input    byte
		expected byte
	}{
		{'A', 'a'},
		{'Z', 'z'},
		{'a', 'a'},
		{'z', 'z'},
		{'0', '0'},
		{'!', '!'},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			result := toLower(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
