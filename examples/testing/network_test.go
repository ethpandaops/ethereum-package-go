package testing

import (
	"context"
	"testing"
	"time"

	"github.com/ethpandaops/ethereum-package-go"
	"github.com/ethpandaops/ethereum-package-go/pkg/client"
	"github.com/ethpandaops/ethereum-package-go/pkg/config"
	"github.com/ethpandaops/ethereum-package-go/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBasicNetwork demonstrates basic network testing
func TestBasicNetwork(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start a test network - automatically cleaned up
	net := testutil.StartNetwork(t)

	// Assert network properties
	testutil.Assert(t, net).
		HasExecutionClients(1).
		HasConsensusClients(1).
		HasChainID(12345)

	// Get clients
	execClient := net.GetExecutionClient()
	consClient := net.GetConsensusClient()

	assert.NotEmpty(t, execClient.RPCURL())
	assert.NotEmpty(t, consClient.BeaconAPIURL())

	// Wait for network to be ready
	net.WaitForSync(30 * time.Second)

	// Verify network is healthy
	net.RequireHealthy()
}

// TestCustomParticipants demonstrates testing with custom client configurations
func TestCustomParticipants(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Define participants - in practice, Kurtosis may not deploy exactly
	// the number requested, so we'll verify what we get
	participants := []config.ParticipantConfig{
		{ELType: client.Geth, CLType: client.Lighthouse, Count: 2},
		{ELType: client.Besu, CLType: client.Teku, Count: 1},
	}

	net := testutil.StartNetworkWithParticipants(t, participants)

	// Verify we have clients
	execClients := net.ExecutionClients()
	consClients := net.ConsensusClients()

	// We should have at least one execution and consensus client
	require.NotEmpty(t, execClients.All(), "Should have at least one execution client")
	require.NotEmpty(t, consClients.All(), "Should have at least one consensus client")

	// Verify the clients are accessible
	for _, client := range execClients.All() {
		assert.NotEmpty(t, client.RPCURL(), "Execution client should have RPC URL")
		assert.NotEmpty(t, client.Name(), "Execution client should have name")
		assert.NotEmpty(t, client.Type(), "Execution client should have type")
	}

	for _, client := range consClients.All() {
		assert.NotEmpty(t, client.BeaconAPIURL(), "Consensus client should have beacon API URL")
		assert.NotEmpty(t, client.Name(), "Consensus client should have name")
		assert.NotEmpty(t, client.Type(), "Consensus client should have type")
	}
}

// TestSharedNetwork demonstrates how to reuse a network across multiple tests
func TestSharedNetwork(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Create a network with a specific name that can be reused
	enclaveName := "test-shared-network"
	net, err := ethereum.FindOrCreateNetwork(ctx, enclaveName, ethereum.Minimal())
	require.NoError(t, err)

	// Log the enclave name for reference
	t.Logf("Using network: %s", net.EnclaveName())

	// Run multiple tests against the same network
	t.Run("TestClients", func(t *testing.T) {
		// This test runs against the existing network
		execClients := net.ExecutionClients()
		consClients := net.ConsensusClients()

		assert.NotEmpty(t, execClients.All())
		assert.NotEmpty(t, consClients.All())
	})

	t.Run("TestEndpoints", func(t *testing.T) {
		// This test also runs against the same network
		execClient := net.ExecutionClients().All()[0]
		assert.NotEmpty(t, execClient.RPCURL())
	})

	// Cleanup is handled by test cleanup, not here
	t.Cleanup(func() {
		// Only clean up if this is the main test
		if t.Name() == "TestSharedNetwork" {
			ctx := context.Background()
			err := net.Cleanup(ctx)
			if err != nil {
				t.Logf("Failed to cleanup network: %v", err)
			}
		}
	})
}

// TestFindExistingNetwork demonstrates finding an existing network
func TestFindExistingNetwork(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Try to find the network created in TestSharedNetwork
	enclaveName := "test-shared-network"
	network, err := ethereum.FindOrCreateNetwork(ctx, enclaveName)

	if err != nil {
		// Network doesn't exist, skip this test
		t.Skipf("Shared network not found: %v", err)
	}

	// Verify we can access the existing network
	assert.Equal(t, enclaveName, network.EnclaveName())
	assert.NotEmpty(t, network.ExecutionClients().All())
}

// TestRandomNetworkNames demonstrates how networks get unique names by default
func TestRandomNetworkNames(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Create two networks without specifying names
	network1, err := ethereum.Run(ctx, ethereum.Minimal())
	require.NoError(t, err)
	defer network1.Cleanup(ctx)

	network2, err := ethereum.Run(ctx, ethereum.Minimal())
	require.NoError(t, err)
	defer network2.Cleanup(ctx)

	// Verify they have different enclave names
	assert.NotEqual(t, network1.EnclaveName(), network2.EnclaveName())
	t.Logf("Network 1: %s", network1.EnclaveName())
	t.Logf("Network 2: %s", network2.EnclaveName())

	// Both should be functional
	assert.NotEmpty(t, network1.ExecutionClients().All())
	assert.NotEmpty(t, network2.ExecutionClients().All())
}

// TestNetworkFailure demonstrates testing network failure scenarios
func TestNetworkFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test network startup failure
	t.Run("InvalidConfiguration", func(t *testing.T) {
		ctx := context.Background()

		// This should fail due to invalid configuration
		network, err := ethereum.Run(ctx,
			ethereum.WithTimeout(0), // Invalid timeout
		)

		assert.Error(t, err)
		assert.Nil(t, network)

		// Ensure cleanup if network was partially created
		if network != nil {
			testutil.CleanupNetwork(t, network)
		}
	})
}

// BenchmarkNetworkStartup benchmarks network startup time
func BenchmarkNetworkStartup(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		network, err := ethereum.Run(ctx, ethereum.Minimal())
		require.NoError(b, err)

		b.StopTimer()
		err = network.Cleanup(ctx)
		require.NoError(b, err)
		b.StartTimer()
	}
}

// ExampleTestNetwork shows how to use testutil in tests
func ExampleTestNetwork() {
	// This would be in a test function
	var t *testing.T

	// Start a test network
	network := testutil.StartNetwork(t)

	// Get clients
	execClient := network.GetExecutionClient()
	consClient := network.GetConsensusClient()

	// Use the clients
	_ = execClient.RPCURL()
	_ = consClient.BeaconAPIURL()

	// Network is automatically cleaned up when test ends
}
