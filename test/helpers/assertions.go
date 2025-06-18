package helpers

import (
	"testing"

	"github.com/ethpandaops/ethereum-package-go/pkg/config"
	"github.com/ethpandaops/ethereum-package-go/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertValidNetwork checks that a network is properly configured
func AssertValidNetwork(t *testing.T, network types.Network) {
	t.Helper()

	require.NotNil(t, network, "network should not be nil")
	assert.NotEmpty(t, network.Name(), "network should have a name")
	assert.NotEmpty(t, network.EnclaveName(), "network should have an enclave name")
	assert.NotZero(t, network.ChainID(), "network should have a chain ID")
}

// AssertExecutionClient checks that an execution client is properly configured
func AssertExecutionClient(t *testing.T, client types.ExecutionClient) {
	t.Helper()

	require.NotNil(t, client, "execution client should not be nil")
	assert.NotEmpty(t, client.Name(), "client should have a name")
	assert.NotEmpty(t, client.Type(), "client should have a type")
	assert.NotEmpty(t, client.RPCURL(), "client should have an RPC URL")
	assert.NotEmpty(t, client.ServiceName(), "client should have a service name")
}

// AssertConsensusClient checks that a consensus client is properly configured
func AssertConsensusClient(t *testing.T, client types.ConsensusClient) {
	t.Helper()

	require.NotNil(t, client, "consensus client should not be nil")
	assert.NotEmpty(t, client.Name(), "client should have a name")
	assert.NotEmpty(t, client.Type(), "client should have a type")
	assert.NotEmpty(t, client.BeaconAPIURL(), "client should have a beacon API URL")
	assert.NotEmpty(t, client.ServiceName(), "client should have a service name")
}

// AssertClientCounts checks that the network has the expected number of clients
func AssertClientCounts(t *testing.T, network types.Network, expectedEL, expectedCL int) {
	t.Helper()

	require.NotNil(t, network.ExecutionClients(), "network should have execution clients")
	require.NotNil(t, network.ConsensusClients(), "network should have consensus clients")

	actualEL := len(network.ExecutionClients().All())
	actualCL := len(network.ConsensusClients().All())

	assert.Equal(t, expectedEL, actualEL, "unexpected number of execution clients")
	assert.Equal(t, expectedCL, actualCL, "unexpected number of consensus clients")
}

// AssertServiceExists checks that a service exists in the network
func AssertServiceExists(t *testing.T, network types.Network, serviceName string) {
	t.Helper()

	services := network.Services()
	found := false
	
	for _, service := range services {
		if service.Name == serviceName {
			found = true
			break
		}
	}

	assert.True(t, found, "service %s should exist in the network", serviceName)
}

// AssertApacheConfig checks that Apache config server is properly configured
func AssertApacheConfig(t *testing.T, network types.Network) {
	t.Helper()

	apache := network.ApacheConfig()
	require.NotNil(t, apache, "network should have Apache config server")
	
	assert.NotEmpty(t, apache.URL(), "Apache should have a base URL")
	assert.Contains(t, apache.GenesisSSZURL(), "genesis.ssz", "Genesis URL should contain genesis.ssz")
	assert.Contains(t, apache.ConfigYAMLURL(), "config.yaml", "Config URL should contain config.yaml")
	assert.Contains(t, apache.BootnodesYAMLURL(), "boot_enr.yaml", "Bootnodes URL should contain boot_enr.yaml")
	assert.Contains(t, apache.DepositContractBlockURL(), "deposit_contract_block.txt", "Deposit contract URL should contain deposit_contract_block.txt")
}

// AssertValidConfig checks that a configuration is valid
func AssertValidConfig(t *testing.T, config *config.EthereumPackageConfig) {
	t.Helper()

	require.NotNil(t, config, "config should not be nil")
	assert.NotEmpty(t, config.Participants, "config should have participants")

	for i, p := range config.Participants {
		assert.NotEmpty(t, p.ELType, "participant %d should have EL type", i)
		assert.NotEmpty(t, p.CLType, "participant %d should have CL type", i)
		if p.Count == 0 {
			assert.Equal(t, 1, p.Count, "participant %d count should default to 1", i)
		}
	}
}

// AssertEqualConfigs checks that two configurations are equivalent
func AssertEqualConfigs(t *testing.T, expected, actual *config.EthereumPackageConfig) {
	t.Helper()

	require.NotNil(t, expected, "expected config should not be nil")
	require.NotNil(t, actual, "actual config should not be nil")

	// Compare participants
	assert.Equal(t, len(expected.Participants), len(actual.Participants), "should have same number of participants")
	for i := range expected.Participants {
		assert.Equal(t, expected.Participants[i].ELType, actual.Participants[i].ELType, "participant %d EL type mismatch", i)
		assert.Equal(t, expected.Participants[i].CLType, actual.Participants[i].CLType, "participant %d CL type mismatch", i)
		assert.Equal(t, expected.Participants[i].Count, actual.Participants[i].Count, "participant %d count mismatch", i)
	}

	// Compare network params
	if expected.NetworkParams != nil {
		require.NotNil(t, actual.NetworkParams, "actual should have network params")
		assert.Equal(t, expected.NetworkParams.ChainID, actual.NetworkParams.ChainID, "chain ID mismatch")
	}

	// Compare MEV
	if expected.MEV != nil {
		require.NotNil(t, actual.MEV, "actual should have MEV config")
		assert.Equal(t, expected.MEV.Type, actual.MEV.Type, "MEV type mismatch")
	}

	// Compare global settings
	assert.Equal(t, expected.GlobalClientLogLevel, actual.GlobalClientLogLevel, "global log level mismatch")
}