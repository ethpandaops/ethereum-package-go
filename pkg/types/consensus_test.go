package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// consensusClientTestCase represents a test case for consensus client constructors
type consensusClientTestCase struct {
	name        string
	clientType  ClientType
	constructor func(name, version, beaconURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) ConsensusClient
	clientName  string
	version     string
	beaconURL   string
	metricsURL  string
	enr         string
	peerID      string
	serviceName string
	containerID string
	p2pPort     int
}

func TestConsensusClientConstructors(t *testing.T) {
	tests := []consensusClientTestCase{
		{
			name:       "Lighthouse",
			clientType: ClientLighthouse,
			constructor: func(name, version, beaconURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) ConsensusClient {
				return NewLighthouseClient(name, version, beaconURL, metricsURL, enr, peerID, serviceName, containerID, p2pPort)
			},
			clientName:  "lighthouse-1",
			version:     "v4.5.0",
			beaconURL:   "http://localhost:5052",
			metricsURL:  "http://localhost:5054",
			enr:         "enr:-abc123",
			peerID:      "16Uiu2HAm123",
			serviceName: "lighthouse-service",
			containerID: "container-123",
			p2pPort:     9000,
		},
		{
			name:       "Teku",
			clientType: ClientTeku,
			constructor: func(name, version, beaconURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) ConsensusClient {
				return NewTekuClient(name, version, beaconURL, metricsURL, enr, peerID, serviceName, containerID, p2pPort)
			},
			clientName:  "teku-1",
			version:     "v23.10.0",
			beaconURL:   "http://localhost:5052",
			metricsURL:  "http://localhost:8008",
			enr:         "enr:-def456",
			peerID:      "16Uiu2HAm456",
			serviceName: "teku-service",
			containerID: "container-456",
			p2pPort:     9000,
		},
		{
			name:       "Prysm",
			clientType: ClientPrysm,
			constructor: func(name, version, beaconURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) ConsensusClient {
				return NewPrysmClient(name, version, beaconURL, metricsURL, enr, peerID, serviceName, containerID, p2pPort)
			},
			clientName:  "prysm-1",
			version:     "v4.1.0",
			beaconURL:   "http://localhost:3500",
			metricsURL:  "http://localhost:8080",
			enr:         "enr:-ghi789",
			peerID:      "16Uiu2HAm789",
			serviceName: "prysm-service",
			containerID: "container-789",
			p2pPort:     13000,
		},
		{
			name:       "Nimbus",
			clientType: ClientNimbus,
			constructor: func(name, version, beaconURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) ConsensusClient {
				return NewNimbusClient(name, version, beaconURL, metricsURL, enr, peerID, serviceName, containerID, p2pPort)
			},
			clientName:  "nimbus-1",
			version:     "v23.10.0",
			beaconURL:   "http://localhost:5052",
			metricsURL:  "http://localhost:8008",
			enr:         "enr:-jkl012",
			peerID:      "16Uiu2HAm012",
			serviceName: "nimbus-service",
			containerID: "container-012",
			p2pPort:     9000,
		},
		{
			name:       "Lodestar",
			clientType: ClientLodestar,
			constructor: func(name, version, beaconURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) ConsensusClient {
				return NewLodestarClient(name, version, beaconURL, metricsURL, enr, peerID, serviceName, containerID, p2pPort)
			},
			clientName:  "lodestar-1",
			version:     "v1.12.0",
			beaconURL:   "http://localhost:9596",
			metricsURL:  "http://localhost:8008",
			enr:         "enr:-mno345",
			peerID:      "16Uiu2HAm345",
			serviceName: "lodestar-service",
			containerID: "container-345",
			p2pPort:     9000,
		},
		{
			name:       "Grandine",
			clientType: ClientGrandine,
			constructor: func(name, version, beaconURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) ConsensusClient {
				return NewGrandineClient(name, version, beaconURL, metricsURL, enr, peerID, serviceName, containerID, p2pPort)
			},
			clientName:  "grandine-1",
			version:     "v0.3.0",
			beaconURL:   "http://localhost:5052",
			metricsURL:  "http://localhost:8008",
			enr:         "enr:-pqr678",
			peerID:      "16Uiu2HAm678",
			serviceName: "grandine-service",
			containerID: "container-678",
			p2pPort:     9000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.constructor(
				tt.clientName,
				tt.version,
				tt.beaconURL,
				tt.metricsURL,
				tt.enr,
				tt.peerID,
				tt.serviceName,
				tt.containerID,
				tt.p2pPort,
			)

			// Common assertions for all consensus clients
			assert.Equal(t, tt.clientName, client.Name())
			assert.Equal(t, tt.clientType, client.Type())
			assert.Equal(t, tt.version, client.Version())
			assert.Equal(t, tt.beaconURL, client.BeaconAPIURL())
			assert.Equal(t, tt.metricsURL, client.MetricsURL())
			assert.Equal(t, tt.enr, client.ENR())
			assert.Equal(t, tt.peerID, client.PeerID())
			assert.Equal(t, tt.p2pPort, client.P2PPort())
			assert.Equal(t, tt.serviceName, client.ServiceName())
			assert.Equal(t, tt.containerID, client.ContainerID())
		})
	}
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
