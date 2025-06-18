package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// executionClientTestCase represents a test case for execution client constructors
type executionClientTestCase struct {
	name        string
	clientType  ClientType
	constructor func(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID string, p2pPort int) ExecutionClient
	clientName  string
	version     string
	rpcURL      string
	wsURL       string
	engineURL   string
	metricsURL  string
	enode       string
	serviceName string
	containerID string
	p2pPort     int
}

func TestExecutionClientConstructors(t *testing.T) {
	tests := []executionClientTestCase{
		{
			name:       "Geth",
			clientType: ClientGeth,
			constructor: func(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID string, p2pPort int) ExecutionClient {
				return NewGethClient(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID, p2pPort)
			},
			clientName:  "geth-1",
			version:     "v1.13.0",
			rpcURL:      "http://localhost:8545",
			wsURL:       "ws://localhost:8546",
			engineURL:   "http://localhost:8551",
			metricsURL:  "http://localhost:9090",
			enode:       "enode://abc123@127.0.0.1:30303",
			serviceName: "geth-service",
			containerID: "container-123",
			p2pPort:     30303,
		},
		{
			name:       "Besu",
			clientType: ClientBesu,
			constructor: func(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID string, p2pPort int) ExecutionClient {
				return NewBesuClient(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID, p2pPort)
			},
			clientName:  "besu-1",
			version:     "v23.10.0",
			rpcURL:      "http://localhost:8545",
			wsURL:       "ws://localhost:8546",
			engineURL:   "http://localhost:8551",
			metricsURL:  "http://localhost:9090",
			enode:       "enode://def456@127.0.0.1:30303",
			serviceName: "besu-service",
			containerID: "container-456",
			p2pPort:     30303,
		},
		{
			name:       "Nethermind",
			clientType: ClientNethermind,
			constructor: func(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID string, p2pPort int) ExecutionClient {
				return NewNethermindClient(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID, p2pPort)
			},
			clientName:  "nethermind-1",
			version:     "v1.21.0",
			rpcURL:      "http://localhost:8545",
			wsURL:       "ws://localhost:8546",
			engineURL:   "http://localhost:8551",
			metricsURL:  "http://localhost:9090",
			enode:       "enode://ghi789@127.0.0.1:30303",
			serviceName: "nethermind-service",
			containerID: "container-789",
			p2pPort:     30303,
		},
		{
			name:       "Erigon",
			clientType: ClientErigon,
			constructor: func(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID string, p2pPort int) ExecutionClient {
				return NewErigonClient(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID, p2pPort)
			},
			clientName:  "erigon-1",
			version:     "v2.54.0",
			rpcURL:      "http://localhost:8545",
			wsURL:       "ws://localhost:8546",
			engineURL:   "http://localhost:8551",
			metricsURL:  "http://localhost:9090",
			enode:       "enode://jkl012@127.0.0.1:30303",
			serviceName: "erigon-service",
			containerID: "container-012",
			p2pPort:     30303,
		},
		{
			name:       "Reth",
			clientType: ClientReth,
			constructor: func(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID string, p2pPort int) ExecutionClient {
				return NewRethClient(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID, p2pPort)
			},
			clientName:  "reth-1",
			version:     "v0.1.0",
			rpcURL:      "http://localhost:8545",
			wsURL:       "ws://localhost:8546",
			engineURL:   "http://localhost:8551",
			metricsURL:  "http://localhost:9090",
			enode:       "enode://mno345@127.0.0.1:30303",
			serviceName: "reth-service",
			containerID: "container-345",
			p2pPort:     30303,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.constructor(
				tt.clientName,
				tt.version,
				tt.rpcURL,
				tt.wsURL,
				tt.engineURL,
				tt.metricsURL,
				tt.enode,
				tt.serviceName,
				tt.containerID,
				tt.p2pPort,
			)

			// Common assertions for all execution clients
			assert.Equal(t, tt.clientName, client.Name())
			assert.Equal(t, tt.clientType, client.Type())
			assert.Equal(t, tt.version, client.Version())
			assert.Equal(t, tt.rpcURL, client.RPCURL())
			assert.Equal(t, tt.wsURL, client.WSURL())
			assert.Equal(t, tt.engineURL, client.EngineURL())
			assert.Equal(t, tt.metricsURL, client.MetricsURL())
			assert.Equal(t, tt.enode, client.Enode())
			assert.Equal(t, tt.p2pPort, client.P2PPort())
			assert.Equal(t, tt.serviceName, client.ServiceName())
			assert.Equal(t, tt.containerID, client.ContainerID())
		})
	}
}

func TestExecutionClients(t *testing.T) {
	clients := NewExecutionClients()

	// Add various clients
	geth1 := NewGethClient("geth-1", "v1.13.0", "http://localhost:8545", "ws://localhost:8546", "http://localhost:8551", "http://localhost:9090", "enode://abc@127.0.0.1:30303", "geth-1", "container-1", 30303)
	geth2 := NewGethClient("geth-2", "v1.13.0", "http://localhost:8555", "ws://localhost:8556", "http://localhost:8561", "http://localhost:9091", "enode://def@127.0.0.1:30304", "geth-2", "container-2", 30304)
	besu := NewBesuClient("besu-1", "v23.10.0", "http://localhost:8565", "ws://localhost:8566", "http://localhost:8571", "http://localhost:9092", "enode://ghi@127.0.0.1:30305", "besu-1", "container-3", 30305)
	nethermind := NewNethermindClient("nethermind-1", "v1.21.0", "http://localhost:8575", "ws://localhost:8576", "http://localhost:8581", "http://localhost:9093", "enode://jkl@127.0.0.1:30306", "nethermind-1", "container-4", 30306)

	clients.Add(geth1)
	clients.Add(geth2)
	clients.Add(besu)
	clients.Add(nethermind)

	// Test All()
	all := clients.All()
	assert.Len(t, all, 4)

	// Test type-specific getters
	gethClients := clients.Geth()
	require.Len(t, gethClients, 2)
	assert.Equal(t, "geth-1", gethClients[0].Name())
	assert.Equal(t, "geth-2", gethClients[1].Name())

	besuClients := clients.Besu()
	require.Len(t, besuClients, 1)
	assert.Equal(t, "besu-1", besuClients[0].Name())

	nethermindClients := clients.Nethermind()
	require.Len(t, nethermindClients, 1)
	assert.Equal(t, "nethermind-1", nethermindClients[0].Name())

	// Test empty client types
	erigonClients := clients.Erigon()
	assert.Len(t, erigonClients, 0)

	rethClients := clients.Reth()
	assert.Len(t, rethClients, 0)
}

func TestExecutionClientInterface(t *testing.T) {
	// Ensure all client types implement ExecutionClient interface
	var _ ExecutionClient = &GethClient{}
	var _ ExecutionClient = &BesuClient{}
	var _ ExecutionClient = &NethermindClient{}
	var _ ExecutionClient = &ErigonClient{}
	var _ ExecutionClient = &RethClient{}
}
