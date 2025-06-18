package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertConsensusClientProperties asserts common properties for consensus clients
func AssertConsensusClientProperties(t *testing.T, client ConsensusClient, expected struct {
	Name         string
	Type         ClientType
	Version      string
	BeaconAPIURL string
	MetricsURL   string
	ENR          string
	PeerID       string
	P2PPort      int
	ServiceName  string
	ContainerID  string
}) {
	t.Helper()
	assert.Equal(t, expected.Name, client.Name())
	assert.Equal(t, expected.Type, client.Type())
	assert.Equal(t, expected.Version, client.Version())
	assert.Equal(t, expected.BeaconAPIURL, client.BeaconAPIURL())
	assert.Equal(t, expected.MetricsURL, client.MetricsURL())
	assert.Equal(t, expected.ENR, client.ENR())
	assert.Equal(t, expected.PeerID, client.PeerID())
	assert.Equal(t, expected.P2PPort, client.P2PPort())
	assert.Equal(t, expected.ServiceName, client.ServiceName())
	assert.Equal(t, expected.ContainerID, client.ContainerID())
}

// AssertExecutionClientProperties asserts common properties for execution clients
func AssertExecutionClientProperties(t *testing.T, client ExecutionClient, expected struct {
	Name        string
	Type        ClientType
	Version     string
	RPCURL      string
	WSURL       string
	EngineURL   string
	MetricsURL  string
	Enode       string
	P2PPort     int
	ServiceName string
	ContainerID string
}) {
	t.Helper()
	assert.Equal(t, expected.Name, client.Name())
	assert.Equal(t, expected.Type, client.Type())
	assert.Equal(t, expected.Version, client.Version())
	assert.Equal(t, expected.RPCURL, client.RPCURL())
	assert.Equal(t, expected.WSURL, client.WSURL())
	assert.Equal(t, expected.EngineURL, client.EngineURL())
	assert.Equal(t, expected.MetricsURL, client.MetricsURL())
	assert.Equal(t, expected.Enode, client.Enode())
	assert.Equal(t, expected.P2PPort, client.P2PPort())
	assert.Equal(t, expected.ServiceName, client.ServiceName())
	assert.Equal(t, expected.ContainerID, client.ContainerID())
}

// TestConsensusClientDefaults provides test data for default consensus clients
var TestConsensusClientDefaults = struct {
	Lighthouse struct {
		Name         string
		Version      string
		BeaconAPIURL string
		MetricsURL   string
		ENR          string
		PeerID       string
		ServiceName  string
		ContainerID  string
		P2PPort      int
	}
	Teku struct {
		Name         string
		Version      string
		BeaconAPIURL string
		MetricsURL   string
		ENR          string
		PeerID       string
		ServiceName  string
		ContainerID  string
		P2PPort      int
	}
}{
	Lighthouse: struct {
		Name         string
		Version      string
		BeaconAPIURL string
		MetricsURL   string
		ENR          string
		PeerID       string
		ServiceName  string
		ContainerID  string
		P2PPort      int
	}{
		Name:         "lighthouse-1",
		Version:      "v4.5.0",
		BeaconAPIURL: "http://localhost:5052",
		MetricsURL:   "http://localhost:5054",
		ENR:          "enr:-abc123",
		PeerID:       "16Uiu2HAm123",
		ServiceName:  "lighthouse-service",
		ContainerID:  "container-123",
		P2PPort:      9000,
	},
	Teku: struct {
		Name         string
		Version      string
		BeaconAPIURL string
		MetricsURL   string
		ENR          string
		PeerID       string
		ServiceName  string
		ContainerID  string
		P2PPort      int
	}{
		Name:         "teku-1",
		Version:      "v23.10.0",
		BeaconAPIURL: "http://localhost:5052",
		MetricsURL:   "http://localhost:8008",
		ENR:          "enr:-def456",
		PeerID:       "16Uiu2HAm456",
		ServiceName:  "teku-service",
		ContainerID:  "container-456",
		P2PPort:      9000,
	},
}

// TestExecutionClientDefaults provides test data for default execution clients
var TestExecutionClientDefaults = struct {
	Geth struct {
		Name        string
		Version     string
		RPCURL      string
		WSURL       string
		EngineURL   string
		MetricsURL  string
		Enode       string
		ServiceName string
		ContainerID string
		P2PPort     int
	}
	Besu struct {
		Name        string
		Version     string
		RPCURL      string
		WSURL       string
		EngineURL   string
		MetricsURL  string
		Enode       string
		ServiceName string
		ContainerID string
		P2PPort     int
	}
}{
	Geth: struct {
		Name        string
		Version     string
		RPCURL      string
		WSURL       string
		EngineURL   string
		MetricsURL  string
		Enode       string
		ServiceName string
		ContainerID string
		P2PPort     int
	}{
		Name:        "geth-1",
		Version:     "v1.13.0",
		RPCURL:      "http://localhost:8545",
		WSURL:       "ws://localhost:8546",
		EngineURL:   "http://localhost:8551",
		MetricsURL:  "http://localhost:9090",
		Enode:       "enode://abc123@127.0.0.1:30303",
		ServiceName: "geth-service",
		ContainerID: "container-123",
		P2PPort:     30303,
	},
	Besu: struct {
		Name        string
		Version     string
		RPCURL      string
		WSURL       string
		EngineURL   string
		MetricsURL  string
		Enode       string
		ServiceName string
		ContainerID string
		P2PPort     int
	}{
		Name:        "besu-1",
		Version:     "v23.10.0",
		RPCURL:      "http://localhost:8545",
		WSURL:       "ws://localhost:8546",
		EngineURL:   "http://localhost:8551",
		MetricsURL:  "http://localhost:9090",
		Enode:       "enode://def456@127.0.0.1:30303",
		ServiceName: "besu-service",
		ContainerID: "container-456",
		P2PPort:     30303,
	},
}