package discovery

import (
	"context"
	"testing"

	"github.com/ethpandaops/ethereum-package-go/pkg/client"
	"github.com/ethpandaops/ethereum-package-go/pkg/config"
	"github.com/ethpandaops/ethereum-package-go/pkg/kurtosis"
	"github.com/ethpandaops/ethereum-package-go/test/helpers"
	"github.com/ethpandaops/ethereum-package-go/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceMapper_MapToNetwork(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewMockKurtosisClient()
	mapper := NewServiceMapper(mockClient)

	// Use helper to create test services
	serviceBuilder := helpers.NewTestServiceBuilder()
	services := serviceBuilder.CreateDefaultServices()

	mockClient.GetServicesFunc = func(ctx context.Context, enclaveName string) (map[string]*kurtosis.ServiceInfo, error) {
		assert.Equal(t, "test-enclave", enclaveName)
		return services, nil
	}
	mockClient.DestroyEnclaveFunc = func(ctx context.Context, enclaveName string) error {
		return nil
	}

	ethConfig := &config.EthereumPackageConfig{
		NetworkParams: &config.NetworkParams{
			NetworkID: "12345",
		},
	}

	networkObj, err := mapper.MapToNetwork(ctx, "test-enclave", ethConfig, false)
	require.NoError(t, err)
	require.NotNil(t, networkObj)

	// Verify network properties
	assert.Equal(t, uint64(12345), networkObj.ChainID())
	assert.Equal(t, "test-enclave", networkObj.EnclaveName())

	// Verify clients were discovered
	execClients := networkObj.ExecutionClients().All()
	consClients := networkObj.ConsensusClients().All()

	assert.NotEmpty(t, execClients, "should have execution clients")
	assert.NotEmpty(t, consClients, "should have consensus clients")

	// Verify Apache config
	apache := networkObj.ApacheConfig()
	require.NotNil(t, apache)
	assert.Contains(t, apache.URL(), "10.0.0.3")

	// Verify service mapping was called
	assert.Greater(t, mockClient.CallCount["GetServices"], 0)
}

func TestServiceMapper_MapToNetworkWithConfiguredServices(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewMockKurtosisClient()
	mapper := NewServiceMapper(mockClient)

	services := map[string]*kurtosis.ServiceInfo{
		"geth-1": {
			Name:      "geth-1",
			UUID:      "uuid-geth",
			Status:    "running",
			IPAddress: "192.168.1.10",
			Ports: map[string]kurtosis.PortInfo{
				"rpc": {Number: 8545, Protocol: "TCP", MaybeURL: "http://192.168.1.10:8545"},
				"ws":  {Number: 8546, Protocol: "TCP", MaybeURL: "ws://192.168.1.10:8546"},
			},
		},
		"lighthouse-1": {
			Name:      "lighthouse-1",
			UUID:      "uuid-lighthouse",
			Status:    "running",
			IPAddress: "192.168.1.11",
			Ports: map[string]kurtosis.PortInfo{
				"http": {Number: 5052, Protocol: "TCP", MaybeURL: "http://192.168.1.11:5052"},
			},
		},
	}

	mockClient.GetServicesFunc = func(ctx context.Context, enclaveName string) (map[string]*kurtosis.ServiceInfo, error) {
		return services, nil
	}

	ethConfig := &config.EthereumPackageConfig{
		Participants: []config.ParticipantConfig{
			{ELType: client.Geth, CLType: client.Lighthouse, Count: 1},
		},
		NetworkParams: &config.NetworkParams{
			NetworkID: "54321",
		},
	}

	networkObj, err := mapper.MapToNetwork(ctx, "custom-enclave", ethConfig, false)
	require.NoError(t, err)
	require.NotNil(t, networkObj)

	// Verify network was configured correctly
	assert.Equal(t, uint64(54321), networkObj.ChainID())
	assert.Equal(t, "custom-enclave", networkObj.EnclaveName())

	// Verify we got the expected clients
	execClients := networkObj.ExecutionClients().All()
	consClients := networkObj.ConsensusClients().All()

	assert.Len(t, execClients, 1)
	assert.Len(t, consClients, 1)

	assert.Equal(t, "geth-1", execClients[0].Name())
	assert.Equal(t, "lighthouse-1", consClients[0].Name())
}

func TestServiceMapper_MapToNetworkEmpty(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewMockKurtosisClient()
	mapper := NewServiceMapper(mockClient)

	// Empty services
	services := map[string]*kurtosis.ServiceInfo{}

	mockClient.GetServicesFunc = func(ctx context.Context, enclaveName string) (map[string]*kurtosis.ServiceInfo, error) {
		return services, nil
	}

	ethConfig := &config.EthereumPackageConfig{
		NetworkParams: &config.NetworkParams{
			NetworkID: "9999",
		},
	}

	networkObj, err := mapper.MapToNetwork(ctx, "empty-enclave", ethConfig, false)
	require.NoError(t, err)
	require.NotNil(t, networkObj)

	// Should still have basic network properties
	assert.Equal(t, uint64(9999), networkObj.ChainID())
	assert.Equal(t, "empty-enclave", networkObj.EnclaveName())

	// But no clients should be discovered
	assert.Empty(t, networkObj.ExecutionClients().All())
	assert.Empty(t, networkObj.ConsensusClients().All())
}

func TestServiceMapper_MapToNetworkError(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewMockKurtosisClient()
	mapper := NewServiceMapper(mockClient)

	// Mock GetServices to return an error
	mockClient.GetServicesFunc = func(ctx context.Context, enclaveName string) (map[string]*kurtosis.ServiceInfo, error) {
		return nil, assert.AnError
	}

	ethConfig := &config.EthereumPackageConfig{
		NetworkParams: &config.NetworkParams{
			NetworkID: "7777",
		},
	}

	networkObj, err := mapper.MapToNetwork(ctx, "error-enclave", ethConfig, false)
	assert.Nil(t, networkObj)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get services")
}

func TestServiceMapper_DiscoverApacheConfig(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewMockKurtosisClient()
	mapper := NewServiceMapper(mockClient)

	services := map[string]*kurtosis.ServiceInfo{
		"apache": {
			Name:      "apache",
			UUID:      "uuid-apache",
			Status:    "running",
			IPAddress: "172.16.0.100",
			Ports: map[string]kurtosis.PortInfo{
				"http": {Number: 80, Protocol: "TCP", MaybeURL: "http://172.16.0.100:80"},
			},
		},
	}

	mockClient.GetServicesFunc = func(ctx context.Context, enclaveName string) (map[string]*kurtosis.ServiceInfo, error) {
		return services, nil
	}

	ethConfig := &config.EthereumPackageConfig{
		NetworkParams: &config.NetworkParams{
			NetworkID: "8888",
		},
	}

	networkObj, err := mapper.MapToNetwork(ctx, "apache-test", ethConfig, false)
	require.NoError(t, err)
	require.NotNil(t, networkObj)

	// Verify Apache config was discovered
	apache := networkObj.ApacheConfig()
	require.NotNil(t, apache)
	assert.Equal(t, "http://172.16.0.100:80", apache.URL())
	assert.Contains(t, apache.GenesisSSZURL(), "genesis.ssz")
	assert.Contains(t, apache.ConfigYAMLURL(), "config.yaml")
}

func TestServiceMapper_MultipleClientTypes(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewMockKurtosisClient()
	mapper := NewServiceMapper(mockClient)

	services := map[string]*kurtosis.ServiceInfo{
		"el-1-geth-lighthouse": {
			Name: "el-1-geth-lighthouse", UUID: "uuid-1", Status: "running", IPAddress: "10.0.1.1",
			Ports: map[string]kurtosis.PortInfo{
				"rpc": {Number: 8545, Protocol: "TCP", MaybeURL: "http://10.0.1.1:8545"},
			},
		},
		"el-2-besu-teku": {
			Name: "el-2-besu-teku", UUID: "uuid-2", Status: "running", IPAddress: "10.0.1.2",
			Ports: map[string]kurtosis.PortInfo{
				"rpc": {Number: 8545, Protocol: "TCP", MaybeURL: "http://10.0.1.2:8545"},
			},
		},
		"cl-1-lighthouse-geth": {
			Name: "cl-1-lighthouse-geth", UUID: "uuid-3", Status: "running", IPAddress: "10.0.2.1",
			Ports: map[string]kurtosis.PortInfo{
				"http": {Number: 5052, Protocol: "TCP", MaybeURL: "http://10.0.2.1:5052"},
			},
		},
		"cl-2-teku-besu": {
			Name: "cl-2-teku-besu", UUID: "uuid-4", Status: "running", IPAddress: "10.0.2.2",
			Ports: map[string]kurtosis.PortInfo{
				"http": {Number: 5052, Protocol: "TCP", MaybeURL: "http://10.0.2.2:5052"},
			},
		},
	}

	mockClient.GetServicesFunc = func(ctx context.Context, enclaveName string) (map[string]*kurtosis.ServiceInfo, error) {
		return services, nil
	}

	ethConfig := &config.EthereumPackageConfig{
		NetworkParams: &config.NetworkParams{
			NetworkID: "1111",
		},
	}

	networkObj, err := mapper.MapToNetwork(ctx, "multi-client", ethConfig, false)
	require.NoError(t, err)
	require.NotNil(t, networkObj)

	// Should discover multiple execution and consensus clients
	execClients := networkObj.ExecutionClients().All()
	consClients := networkObj.ConsensusClients().All()

	assert.Len(t, execClients, 2, "should have 2 execution clients")
	assert.Len(t, consClients, 2, "should have 2 consensus clients")

	// Verify client names were parsed correctly
	execNames := make([]string, len(execClients))
	for i, client := range execClients {
		execNames[i] = client.Name()
	}
	assert.Contains(t, execNames, "el-1-geth-lighthouse")
	assert.Contains(t, execNames, "el-2-besu-teku")

	consNames := make([]string, len(consClients))
	for i, client := range consClients {
		consNames[i] = client.Name()
	}
	assert.Contains(t, consNames, "cl-1-lighthouse-geth")
	assert.Contains(t, consNames, "cl-2-teku-besu")
}
