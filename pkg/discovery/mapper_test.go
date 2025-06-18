package discovery

import (
	"context"
	"testing"
	"time"

	"github.com/ethpandaops/ethereum-package-go/pkg/client"
	"github.com/ethpandaops/ethereum-package-go/pkg/config"
	"github.com/ethpandaops/ethereum-package-go/pkg/kurtosis"
	"github.com/ethpandaops/ethereum-package-go/pkg/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockKurtosisClient is a mock implementation of kurtosis.Client
type MockKurtosisClient struct {
	mock.Mock
}

func (m *MockKurtosisClient) RunPackage(ctx context.Context, config kurtosis.RunPackageConfig) (*kurtosis.RunPackageResult, error) {
	args := m.Called(ctx, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kurtosis.RunPackageResult), args.Error(1)
}

func (m *MockKurtosisClient) GetServices(ctx context.Context, enclaveName string) (map[string]*kurtosis.ServiceInfo, error) {
	args := m.Called(ctx, enclaveName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*kurtosis.ServiceInfo), args.Error(1)
}

func (m *MockKurtosisClient) StopEnclave(ctx context.Context, enclaveName string) error {
	args := m.Called(ctx, enclaveName)
	return args.Error(0)
}

func (m *MockKurtosisClient) DestroyEnclave(ctx context.Context, enclaveName string) error {
	args := m.Called(ctx, enclaveName)
	return args.Error(0)
}

func (m *MockKurtosisClient) WaitForServices(ctx context.Context, enclaveName string, serviceNames []string, timeout time.Duration) error {
	args := m.Called(ctx, enclaveName, serviceNames, timeout)
	return args.Error(0)
}

func TestServiceMapper_MapToNetwork(t *testing.T) {
	ctx := context.Background()
	mockClient := new(MockKurtosisClient)
	mapper := NewServiceMapper(mockClient)

	services := map[string]*kurtosis.ServiceInfo{
		"el-1-geth-lighthouse": {
			Name:      "el-1-geth-lighthouse",
			UUID:      "uuid-1",
			Status:    "running",
			IPAddress: "10.0.0.1",
			Ports: map[string]kurtosis.PortInfo{
				"rpc": {
					Number:   8545,
					Protocol: "TCP",
					MaybeURL: "http://10.0.0.1:8545",
				},
				"ws": {
					Number:   8546,
					Protocol: "TCP",
					MaybeURL: "ws://10.0.0.1:8546",
				},
				"engine": {
					Number:   8551,
					Protocol: "TCP",
					MaybeURL: "http://10.0.0.1:8551",
				},
			},
		},
		"cl-1-lighthouse-geth": {
			Name:      "cl-1-lighthouse-geth",
			UUID:      "uuid-2",
			Status:    "running",
			IPAddress: "10.0.0.2",
			Ports: map[string]kurtosis.PortInfo{
				"http": {
					Number:   5052,
					Protocol: "TCP",
					MaybeURL: "http://10.0.0.2:5052",
				},
				"metrics": {
					Number:   5054,
					Protocol: "TCP",
					MaybeURL: "http://10.0.0.2:5054",
				},
			},
		},
		"apache": {
			Name:      "apache",
			UUID:      "uuid-3",
			Status:    "running",
			IPAddress: "10.0.0.3",
			Ports: map[string]kurtosis.PortInfo{
				"http": {
					Number:   80,
					Protocol: "TCP",
					MaybeURL: "http://10.0.0.3:80",
				},
			},
		},
		"prometheus": {
			Name:      "prometheus",
			UUID:      "uuid-4",
			Status:    "running",
			IPAddress: "10.0.0.4",
			Ports: map[string]kurtosis.PortInfo{
				"http": {
					Number:   9090,
					Protocol: "TCP",
					MaybeURL: "http://10.0.0.4:9090",
				},
			},
		},
	}

	mockClient.On("GetServices", ctx, "test-enclave").Return(services, nil)
	mockClient.On("DestroyEnclave", ctx, "test-enclave").Return(nil)

	ethConfig := &config.EthereumPackageConfig{
		NetworkParams: &config.NetworkParams{
			ChainID: 12345,
		},
	}

	network, err := mapper.MapToNetwork(ctx, "test-enclave", ethConfig)
	require.NoError(t, err)
	require.NotNil(t, network)

	// Verify network properties
	assert.Equal(t, uint64(12345), network.ChainID())
	assert.Equal(t, "test-enclave", network.EnclaveName())

	// Verify execution clients
	execClients := network.ExecutionClients()
	assert.Equal(t, 1, len(execClients.All()))
	gethClients := execClients.ByType(client.Geth)
	assert.Equal(t, 1, len(gethClients))
	geth := gethClients[0]
	assert.Equal(t, "http://10.0.0.1:8545", geth.RPCURL())
	assert.Equal(t, "ws://10.0.0.1:8546", geth.WSURL())
	assert.Equal(t, "http://10.0.0.1:8551", geth.EngineURL())

	// Verify consensus clients
	consClients := network.ConsensusClients()
	t.Logf("Consensus clients: %d", len(consClients.All()))
	for _, client := range consClients.All() {
		t.Logf("Consensus client: %s, type: %s", client.Name(), client.Type())
	}
	assert.Equal(t, 1, len(consClients.All()))
	lighthouseClients := consClients.ByType(client.Lighthouse)
	if len(lighthouseClients) > 0 {
		assert.Equal(t, 1, len(lighthouseClients))
		lighthouse := lighthouseClients[0]
		assert.Equal(t, "http://10.0.0.2:5052", lighthouse.BeaconAPIURL())
	} else {
		t.Error("No lighthouse clients found")
	}

	// Verify Apache config server
	apache := network.ApacheConfig()
	require.NotNil(t, apache)
	assert.Equal(t, "http://10.0.0.3:80", apache.URL())


	// Verify services list
	networkServices := network.Services()
	assert.Equal(t, 4, len(networkServices))

	// Test cleanup
	err = network.Cleanup(ctx)
	assert.NoError(t, err)
	mockClient.AssertCalled(t, "DestroyEnclave", ctx, "test-enclave")
}

func TestServiceMapper_DetectServiceTypes(t *testing.T) {
	tests := []struct {
		name         string
		serviceName  string
		expectedType network.ServiceType
	}{
		{"geth execution", "el-1-geth-lighthouse", network.ServiceTypeExecutionClient},
		{"besu execution", "el-2-besu-teku", network.ServiceTypeExecutionClient},
		{"nethermind execution", "el-3-nethermind", network.ServiceTypeExecutionClient},
		{"lighthouse consensus", "cl-1-lighthouse-geth", network.ServiceTypeConsensusClient},
		{"teku consensus", "cl-2-teku-besu", network.ServiceTypeConsensusClient},
		{"prysm consensus", "cl-3-prysm", network.ServiceTypeConsensusClient},
		{"validator", "validator-1", network.ServiceTypeValidator},
		{"prometheus", "prometheus", network.ServiceTypePrometheus},
		{"grafana", "grafana", network.ServiceTypeGrafana},
		{"blockscout", "blockscout", network.ServiceTypeBlockscout},
		{"apache", "apache", network.ServiceTypeApache},
		{"apache config", "apache-config-server", network.ServiceTypeApache},
		{"unknown", "random-service", network.ServiceTypeOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceType := detectServiceType(tt.serviceName)
			assert.Equal(t, tt.expectedType, serviceType)
		})
	}
}

func TestServiceMapper_EmptyServices(t *testing.T) {
	ctx := context.Background()
	mockClient := new(MockKurtosisClient)
	mapper := NewServiceMapper(mockClient)

	// Return empty services
	emptyServices := make(map[string]*kurtosis.ServiceInfo)
	mockClient.On("GetServices", ctx, "empty-enclave").Return(emptyServices, nil)
	mockClient.On("DestroyEnclave", ctx, "empty-enclave").Return(nil)

	config := &config.EthereumPackageConfig{}

	network, err := mapper.MapToNetwork(ctx, "empty-enclave", config)
	require.NoError(t, err)
	require.NotNil(t, network)

	// Verify empty collections
	assert.Empty(t, network.ExecutionClients().All())
	assert.Empty(t, network.ConsensusClients().All())
	assert.Empty(t, network.Services())
	assert.Nil(t, network.ApacheConfig())
}

func TestServiceMapper_MixedClients(t *testing.T) {
	ctx := context.Background()
	mockClient := new(MockKurtosisClient)
	mapper := NewServiceMapper(mockClient)

	services := map[string]*kurtosis.ServiceInfo{
		"el-1-geth-lighthouse": {
			Name:      "el-1-geth-lighthouse",
			UUID:      "uuid-1",
			Status:    "running",
			IPAddress: "10.0.0.1",
			Ports: map[string]kurtosis.PortInfo{
				"rpc": {Number: 8545, Protocol: "TCP", MaybeURL: "http://10.0.0.1:8545"},
			},
		},
		"el-2-besu-teku": {
			Name:      "el-2-besu-teku",
			UUID:      "uuid-2",
			Status:    "running",
			IPAddress: "10.0.0.2",
			Ports: map[string]kurtosis.PortInfo{
				"rpc": {Number: 8545, Protocol: "TCP", MaybeURL: "http://10.0.0.2:8545"},
			},
		},
		"cl-1-lighthouse-geth": {
			Name:      "cl-1-lighthouse-geth",
			UUID:      "uuid-3",
			Status:    "running",
			IPAddress: "10.0.0.3",
			Ports: map[string]kurtosis.PortInfo{
				"http": {Number: 5052, Protocol: "TCP", MaybeURL: "http://10.0.0.3:5052"},
			},
		},
		"cl-2-teku-besu": {
			Name:      "cl-2-teku-besu",
			UUID:      "uuid-4",
			Status:    "running",
			IPAddress: "10.0.0.4",
			Ports: map[string]kurtosis.PortInfo{
				"http": {Number: 5052, Protocol: "TCP", MaybeURL: "http://10.0.0.4:5052"},
			},
		},
	}

	mockClient.On("GetServices", ctx, "mixed-enclave").Return(services, nil)
	mockClient.On("DestroyEnclave", ctx, "mixed-enclave").Return(nil)

	config := &config.EthereumPackageConfig{}

	network, err := mapper.MapToNetwork(ctx, "mixed-enclave", config)
	require.NoError(t, err)

	// Verify multiple client types
	execClients := network.ExecutionClients()
	assert.Equal(t, 2, len(execClients.All()))
	assert.Equal(t, 1, len(execClients.ByType(client.Geth)))
	assert.Equal(t, 1, len(execClients.ByType(client.Besu)))

	consClients := network.ConsensusClients()
	assert.Equal(t, 2, len(consClients.All()))
	assert.Equal(t, 1, len(consClients.ByType(client.Lighthouse)))
	assert.Equal(t, 1, len(consClients.ByType(client.Teku)))
}

func TestServiceMapper_ServiceWithoutPorts(t *testing.T) {
	ctx := context.Background()
	mockClient := new(MockKurtosisClient)
	mapper := NewServiceMapper(mockClient)

	services := map[string]*kurtosis.ServiceInfo{
		"el-1-geth": {
			Name:      "el-1-geth",
			UUID:      "uuid-1",
			Status:    "running",
			IPAddress: "10.0.0.1",
			Ports:     map[string]kurtosis.PortInfo{}, // No ports
		},
	}

	mockClient.On("GetServices", ctx, "no-ports-enclave").Return(services, nil)
	mockClient.On("DestroyEnclave", ctx, "no-ports-enclave").Return(nil)

	config := &config.EthereumPackageConfig{}

	network, err := mapper.MapToNetwork(ctx, "no-ports-enclave", config)
	require.NoError(t, err)

	// Service should still be mapped even without ports
	assert.Equal(t, 1, len(network.Services()))
	
	// But execution client won't have URLs
	execClients := network.ExecutionClients()
	assert.Equal(t, 1, len(execClients.All()))
	client := execClients.All()[0]
	assert.Empty(t, client.RPCURL())
	assert.Empty(t, client.WSURL())
}