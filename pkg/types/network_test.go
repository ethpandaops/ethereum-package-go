package types

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPort(t *testing.T) {
	port := Port{
		Name:          "rpc",
		InternalPort:  8545,
		ExternalPort:  8545,
		Protocol:      "tcp",
		ExposedToHost: true,
	}

	assert.Equal(t, "rpc", port.Name)
	assert.Equal(t, 8545, port.InternalPort)
	assert.Equal(t, 8545, port.ExternalPort)
	assert.Equal(t, "tcp", port.Protocol)
	assert.True(t, port.ExposedToHost)
}

func TestService(t *testing.T) {
	service := Service{
		Name:        "geth-1",
		Type:        ServiceTypeExecutionClient,
		ContainerID: "container-123",
		Ports: []Port{
			{Name: "rpc", InternalPort: 8545, ExternalPort: 8545, Protocol: "tcp", ExposedToHost: true},
			{Name: "ws", InternalPort: 8546, ExternalPort: 8546, Protocol: "tcp", ExposedToHost: true},
		},
		Status: "running",
	}

	assert.Equal(t, "geth-1", service.Name)
	assert.Equal(t, ServiceTypeExecutionClient, service.Type)
	assert.Equal(t, "container-123", service.ContainerID)
	assert.Len(t, service.Ports, 2)
	assert.Equal(t, "running", service.Status)
}

func TestApacheConfigServer(t *testing.T) {
	apache := NewApacheConfigServer("http://127.0.0.1:32966")

	assert.Equal(t, "http://127.0.0.1:32966", apache.URL())
	assert.Equal(t, "http://127.0.0.1:32966/network-configs/genesis.ssz", apache.GenesisSSZURL())
	assert.Equal(t, "http://127.0.0.1:32966/network-configs/config.yaml", apache.ConfigYAMLURL())
	assert.Equal(t, "http://127.0.0.1:32966/network-configs/boot_enr.yaml", apache.BootnodesYAMLURL())
	assert.Equal(t, "http://127.0.0.1:32966/network-configs/deposit_contract_block.txt", apache.DepositContractBlockURL())
}

func TestNetwork(t *testing.T) {
	// Setup execution clients
	execClients := NewExecutionClients()
	geth := NewGethClient("geth-1", "v1.13.0", "http://localhost:8545", "ws://localhost:8546", "http://localhost:8551", "http://localhost:9090", "enode://abc@127.0.0.1:30303", "geth-1", "container-1", 30303)
	execClients.Add(geth)

	// Setup consensus clients
	consClients := NewConsensusClients()
	lighthouse := NewLighthouseClient("lighthouse-1", "v4.5.0", "http://localhost:5052", "http://localhost:5054", "enr:-abc", "peer1", "lh-1", "container-2", 9000)
	consClients.Add(lighthouse)

	// Setup services
	services := []Service{
		{Name: "geth-1", Type: ServiceTypeExecutionClient, ContainerID: "container-1", Status: "running"},
		{Name: "lighthouse-1", Type: ServiceTypeConsensusClient, ContainerID: "container-2", Status: "running"},
		{Name: "prometheus", Type: ServiceTypePrometheus, ContainerID: "container-3", Status: "running"},
	}

	// Setup Apache config server
	apache := NewApacheConfigServer("http://127.0.0.1:32966")

	// Create network
	config := NetworkConfig{
		Name:             "test-network",
		ChainID:          12345,
		EnclaveName:      "test-enclave",
		ExecutionClients: execClients,
		ConsensusClients: consClients,
		Services:         services,
		ApacheConfig:     apache,
		PrometheusURL:    "http://localhost:9090",
		GrafanaURL:       "http://localhost:3000",
		BlockscoutURL:    "http://localhost:4000",
	}

	network := NewNetwork(config)

	// Test basic properties
	assert.Equal(t, "test-network", network.Name())
	assert.Equal(t, uint64(12345), network.ChainID())
	assert.Equal(t, "test-enclave", network.EnclaveName())

	// Test client accessors
	require.NotNil(t, network.ExecutionClients())
	assert.Len(t, network.ExecutionClients().All(), 1)
	require.NotNil(t, network.ConsensusClients())
	assert.Len(t, network.ConsensusClients().All(), 1)

	// Test service accessors
	assert.Len(t, network.Services(), 3)
	assert.Equal(t, "http://localhost:9090", network.PrometheusURL())
	assert.Equal(t, "http://localhost:3000", network.GrafanaURL())
	assert.Equal(t, "http://localhost:4000", network.BlockscoutURL())

	// Test Apache config server
	require.NotNil(t, network.ApacheConfig())
	assert.Equal(t, "http://127.0.0.1:32966", network.ApacheConfig().URL())

	// Test lifecycle methods
	ctx := context.Background()
	assert.NoError(t, network.Stop(ctx))
	assert.NoError(t, network.Cleanup(ctx))
}

func TestNetworkWithCleanupFunc(t *testing.T) {
	cleanupCalled := false
	cleanupFunc := func(ctx context.Context) error {
		cleanupCalled = true
		return nil
	}

	config := NetworkConfig{
		Name:             "test-network",
		ChainID:          12345,
		EnclaveName:      "test-enclave",
		ExecutionClients: NewExecutionClients(),
		ConsensusClients: NewConsensusClients(),
		CleanupFunc:      cleanupFunc,
	}

	network := NewNetwork(config)
	ctx := context.Background()

	assert.NoError(t, network.Cleanup(ctx))
	assert.True(t, cleanupCalled)
}

func TestNetworkWithCleanupError(t *testing.T) {
	expectedErr := errors.New("cleanup failed")
	cleanupFunc := func(ctx context.Context) error {
		return expectedErr
	}

	config := NetworkConfig{
		Name:             "test-network",
		ChainID:          12345,
		EnclaveName:      "test-enclave",
		ExecutionClients: NewExecutionClients(),
		ConsensusClients: NewConsensusClients(),
		CleanupFunc:      cleanupFunc,
	}

	network := NewNetwork(config)
	ctx := context.Background()

	err := network.Cleanup(ctx)
	assert.Equal(t, expectedErr, err)
}

func TestServiceTypes(t *testing.T) {
	// Test that all service type constants are defined
	assert.Equal(t, ServiceType("execution"), ServiceTypeExecutionClient)
	assert.Equal(t, ServiceType("consensus"), ServiceTypeConsensusClient)
	assert.Equal(t, ServiceType("validator"), ServiceTypeValidator)
	assert.Equal(t, ServiceType("prometheus"), ServiceTypePrometheus)
	assert.Equal(t, ServiceType("grafana"), ServiceTypeGrafana)
	assert.Equal(t, ServiceType("blockscout"), ServiceTypeBlockscout)
	assert.Equal(t, ServiceType("apache"), ServiceTypeApache)
	assert.Equal(t, ServiceType("other"), ServiceTypeOther)
}