package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLighthouseClient(t *testing.T) {
	client := NewLighthouseClient(
		"lighthouse-1",
		"v4.5.0",
		"http://localhost:5052",
		"http://localhost:5054",
		"enr:-abc123",
		"16Uiu2HAm123",
		"lighthouse-service",
		"container-123",
		9000,
	)

	assert.Equal(t, "lighthouse-1", client.Name())
	assert.Equal(t, ClientLighthouse, client.Type())
	assert.Equal(t, "v4.5.0", client.Version())
	assert.Equal(t, "http://localhost:5052", client.BeaconAPIURL())
	assert.Equal(t, "http://localhost:5054", client.MetricsURL())
	assert.Equal(t, "enr:-abc123", client.ENR())
	assert.Equal(t, "16Uiu2HAm123", client.PeerID())
	assert.Equal(t, 9000, client.P2PPort())
	assert.Equal(t, "lighthouse-service", client.ServiceName())
	assert.Equal(t, "container-123", client.ContainerID())
}

func TestTekuClient(t *testing.T) {
	client := NewTekuClient(
		"teku-1",
		"v23.10.0",
		"http://localhost:5052",
		"http://localhost:8008",
		"enr:-def456",
		"16Uiu2HAm456",
		"teku-service",
		"container-456",
		9000,
	)

	assert.Equal(t, "teku-1", client.Name())
	assert.Equal(t, ClientTeku, client.Type())
	assert.Equal(t, "v23.10.0", client.Version())
}

func TestPrysmClient(t *testing.T) {
	client := NewPrysmClient(
		"prysm-1",
		"v4.1.0",
		"http://localhost:3500",
		"http://localhost:8080",
		"enr:-ghi789",
		"16Uiu2HAm789",
		"prysm-service",
		"container-789",
		13000,
	)

	assert.Equal(t, "prysm-1", client.Name())
	assert.Equal(t, ClientPrysm, client.Type())
	assert.Equal(t, "v4.1.0", client.Version())
}

func TestNimbusClient(t *testing.T) {
	client := NewNimbusClient(
		"nimbus-1",
		"v23.10.0",
		"http://localhost:5052",
		"http://localhost:8008",
		"enr:-jkl012",
		"16Uiu2HAm012",
		"nimbus-service",
		"container-012",
		9000,
	)

	assert.Equal(t, "nimbus-1", client.Name())
	assert.Equal(t, ClientNimbus, client.Type())
	assert.Equal(t, "v23.10.0", client.Version())
}

func TestLodestarClient(t *testing.T) {
	client := NewLodestarClient(
		"lodestar-1",
		"v1.12.0",
		"http://localhost:9596",
		"http://localhost:8008",
		"enr:-mno345",
		"16Uiu2HAm345",
		"lodestar-service",
		"container-345",
		9000,
	)

	assert.Equal(t, "lodestar-1", client.Name())
	assert.Equal(t, ClientLodestar, client.Type())
	assert.Equal(t, "v1.12.0", client.Version())
}

func TestGrandineClient(t *testing.T) {
	client := NewGrandineClient(
		"grandine-1",
		"v0.3.0",
		"http://localhost:5052",
		"http://localhost:8008",
		"enr:-pqr678",
		"16Uiu2HAm678",
		"grandine-service",
		"container-678",
		9000,
	)

	assert.Equal(t, "grandine-1", client.Name())
	assert.Equal(t, ClientGrandine, client.Type())
	assert.Equal(t, "v0.3.0", client.Version())
}

func TestConsensusClients(t *testing.T) {
	clients := NewConsensusClients()

	// Add various clients
	lighthouse1 := NewLighthouseClient("lighthouse-1", "v4.5.0", "http://localhost:5052", "http://localhost:5054", "enr:-abc", "peer1", "lh-1", "container-1", 9000)
	lighthouse2 := NewLighthouseClient("lighthouse-2", "v4.5.0", "http://localhost:5062", "http://localhost:5064", "enr:-def", "peer2", "lh-2", "container-2", 9001)
	teku := NewTekuClient("teku-1", "v23.10.0", "http://localhost:5072", "http://localhost:8008", "enr:-ghi", "peer3", "teku-1", "container-3", 9002)
	prysm := NewPrysmClient("prysm-1", "v4.1.0", "http://localhost:3500", "http://localhost:8080", "enr:-jkl", "peer4", "prysm-1", "container-4", 13000)

	clients.Add(lighthouse1)
	clients.Add(lighthouse2)
	clients.Add(teku)
	clients.Add(prysm)

	// Test All()
	all := clients.All()
	assert.Len(t, all, 4)

	// Test type-specific getters
	lighthouseClients := clients.Lighthouse()
	require.Len(t, lighthouseClients, 2)
	assert.Equal(t, "lighthouse-1", lighthouseClients[0].Name())
	assert.Equal(t, "lighthouse-2", lighthouseClients[1].Name())

	tekuClients := clients.Teku()
	require.Len(t, tekuClients, 1)
	assert.Equal(t, "teku-1", tekuClients[0].Name())

	prysmClients := clients.Prysm()
	require.Len(t, prysmClients, 1)
	assert.Equal(t, "prysm-1", prysmClients[0].Name())

	// Test empty client types
	nimbusClients := clients.Nimbus()
	assert.Len(t, nimbusClients, 0)

	lodestarClients := clients.Lodestar()
	assert.Len(t, lodestarClients, 0)

	grandineClients := clients.Grandine()
	assert.Len(t, grandineClients, 0)
}

func TestConsensusClientInterface(t *testing.T) {
	// Ensure all client types implement ConsensusClient interface
	var _ ConsensusClient = &LighthouseClient{}
	var _ ConsensusClient = &TekuClient{}
	var _ ConsensusClient = &PrysmClient{}
	var _ ConsensusClient = &NimbusClient{}
	var _ ConsensusClient = &LodestarClient{}
	var _ ConsensusClient = &GrandineClient{}
}