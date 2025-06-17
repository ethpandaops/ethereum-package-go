package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGethClient(t *testing.T) {
	client := NewGethClient(
		"geth-1",
		"v1.13.0",
		"http://localhost:8545",
		"ws://localhost:8546",
		"http://localhost:8551",
		"http://localhost:9090",
		"enode://abc123@127.0.0.1:30303",
		"geth-service",
		"container-123",
		30303,
	)

	assert.Equal(t, "geth-1", client.Name())
	assert.Equal(t, ClientGeth, client.Type())
	assert.Equal(t, "v1.13.0", client.Version())
	assert.Equal(t, "http://localhost:8545", client.RPCURL())
	assert.Equal(t, "ws://localhost:8546", client.WSURL())
	assert.Equal(t, "http://localhost:8551", client.EngineURL())
	assert.Equal(t, "http://localhost:9090", client.MetricsURL())
	assert.Equal(t, "enode://abc123@127.0.0.1:30303", client.Enode())
	assert.Equal(t, 30303, client.P2PPort())
	assert.Equal(t, "geth-service", client.ServiceName())
	assert.Equal(t, "container-123", client.ContainerID())
}

func TestBesuClient(t *testing.T) {
	client := NewBesuClient(
		"besu-1",
		"v23.10.0",
		"http://localhost:8545",
		"ws://localhost:8546",
		"http://localhost:8551",
		"http://localhost:9090",
		"enode://def456@127.0.0.1:30303",
		"besu-service",
		"container-456",
		30303,
	)

	assert.Equal(t, "besu-1", client.Name())
	assert.Equal(t, ClientBesu, client.Type())
	assert.Equal(t, "v23.10.0", client.Version())
}

func TestNethermindClient(t *testing.T) {
	client := NewNethermindClient(
		"nethermind-1",
		"v1.21.0",
		"http://localhost:8545",
		"ws://localhost:8546",
		"http://localhost:8551",
		"http://localhost:9090",
		"enode://ghi789@127.0.0.1:30303",
		"nethermind-service",
		"container-789",
		30303,
	)

	assert.Equal(t, "nethermind-1", client.Name())
	assert.Equal(t, ClientNethermind, client.Type())
	assert.Equal(t, "v1.21.0", client.Version())
}

func TestErigonClient(t *testing.T) {
	client := NewErigonClient(
		"erigon-1",
		"v2.54.0",
		"http://localhost:8545",
		"ws://localhost:8546",
		"http://localhost:8551",
		"http://localhost:9090",
		"enode://jkl012@127.0.0.1:30303",
		"erigon-service",
		"container-012",
		30303,
	)

	assert.Equal(t, "erigon-1", client.Name())
	assert.Equal(t, ClientErigon, client.Type())
	assert.Equal(t, "v2.54.0", client.Version())
}

func TestRethClient(t *testing.T) {
	client := NewRethClient(
		"reth-1",
		"v0.1.0",
		"http://localhost:8545",
		"ws://localhost:8546",
		"http://localhost:8551",
		"http://localhost:9090",
		"enode://mno345@127.0.0.1:30303",
		"reth-service",
		"container-345",
		30303,
	)

	assert.Equal(t, "reth-1", client.Name())
	assert.Equal(t, ClientReth, client.Type())
	assert.Equal(t, "v0.1.0", client.Version())
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