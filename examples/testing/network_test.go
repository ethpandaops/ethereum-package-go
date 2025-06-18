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

	t.Log("=== Starting TestBasicNetwork ===")
	t.Log("Creating minimal test network...")
	// Start a test network - automatically cleaned up
	net := testutil.StartNetwork(t)
	t.Log("Test network startup completed")

	// Debug: Print network information
	t.Logf("Network created: %s", net.EnclaveName())
	t.Logf("Found %d execution clients", len(net.ExecutionClients().All()))
	t.Logf("Found %d consensus clients", len(net.ConsensusClients().All()))
	t.Logf("Found %d total services", len(net.Services()))

	// Debug: Print all services
	for i, service := range net.Services() {
		t.Logf("Service %d: %s (type: %s, status: %s)", i, service.Name, service.Type, service.Status)
	}

	// Assert network properties
	// ethereum-package v5.0.1 creates:
	// - 1 execution client (el-1-geth-lighthouse)
	// - 1 consensus client (cl-1-lighthouse-geth)
	// - validator services should not be counted as execution/consensus clients
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
	t.Log("Waiting for network to sync...")
	net.WaitForSync(30 * time.Second)

	// Verify network is healthy
	t.Log("Verifying network health...")
	net.RequireHealthy()

	t.Log("=== TestBasicNetwork completed successfully ===")
}

// TestCustomParticipants demonstrates testing with custom client configurations
func TestCustomParticipants(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Log("=== Starting TestCustomParticipants ===")
	// Define participants - in practice, Kurtosis may not deploy exactly
	// the number requested, so we'll verify what we get
	participants := []config.ParticipantConfig{
		{ELType: client.Geth, CLType: client.Lighthouse, Count: 2},
		{ELType: client.Besu, CLType: client.Teku, Count: 1},
	}

	t.Logf("Creating network with %d participant configurations...", len(participants))
	for i, p := range participants {
		t.Logf("Participant %d: %s/%s (count: %d)", i, p.ELType, p.CLType, p.Count)
	}
	net := testutil.StartNetworkWithParticipants(t, participants)
	t.Log("Custom participants network created")

	// Verify we have clients
	execClients := net.ExecutionClients()
	consClients := net.ConsensusClients()

	// We should have at least one execution and consensus client
	require.NotEmpty(t, execClients.All(), "Should have at least one execution client")
	require.NotEmpty(t, consClients.All(), "Should have at least one consensus client")

	// Verify the clients are accessible
	t.Log("Verifying execution client accessibility...")
	for i, client := range execClients.All() {
		t.Logf("Execution client %d: %s (%s) - RPC: %s", i, client.Name(), client.Type(), client.RPCURL())
		assert.NotEmpty(t, client.RPCURL(), "Execution client should have RPC URL")
		assert.NotEmpty(t, client.Name(), "Execution client should have name")
		assert.NotEmpty(t, client.Type(), "Execution client should have type")
	}

	t.Log("Verifying consensus client accessibility...")
	for i, client := range consClients.All() {
		t.Logf("Consensus client %d: %s (%s) - Beacon API: %s", i, client.Name(), client.Type(), client.BeaconAPIURL())
		assert.NotEmpty(t, client.BeaconAPIURL(), "Consensus client should have beacon API URL")
		assert.NotEmpty(t, client.Name(), "Consensus client should have name")
		assert.NotEmpty(t, client.Type(), "Consensus client should have type")
	}

	t.Log("=== TestCustomParticipants completed successfully ===")
}

// TestSharedNetwork demonstrates how to reuse a network across multiple tests
func TestSharedNetwork(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Log("=== Starting TestSharedNetwork ===")
	ctx := context.Background()

	// Create a network with a specific name that can be reused
	enclaveName := "test-shared-network"
	t.Logf("Finding or creating shared network: %s", enclaveName)
	net, err := ethereum.FindOrCreateNetwork(ctx, enclaveName, ethereum.Minimal())
	require.NoError(t, err)
	t.Logf("Shared network ready: %s", enclaveName)

	// Log the enclave name for reference
	t.Logf("Using network: %s", net.EnclaveName())

	// Run multiple tests against the same network
	t.Run("TestClients", func(t *testing.T) {
		t.Log("Running subtest: TestClients")
		// This test runs against the existing network
		execClients := net.ExecutionClients()
		consClients := net.ConsensusClients()

		t.Logf("Found %d execution clients, %d consensus clients", len(execClients.All()), len(consClients.All()))
		assert.NotEmpty(t, execClients.All())
		assert.NotEmpty(t, consClients.All())
		t.Log("TestClients subtest completed")
	})

	t.Run("TestEndpoints", func(t *testing.T) {
		t.Log("Running subtest: TestEndpoints")
		// This test also runs against the same network
		execClient := net.ExecutionClients().All()[0]
		t.Logf("Testing execution client endpoint: %s", execClient.RPCURL())
		assert.NotEmpty(t, execClient.RPCURL())
		t.Log("TestEndpoints subtest completed")
	})

	// Cleanup is handled by test cleanup, not here
	t.Cleanup(func() {
		// Only clean up if this is the main test
		if t.Name() == "TestSharedNetwork" {
			t.Log("Cleaning up shared network...")
			ctx := context.Background()
			err := net.Cleanup(ctx)
			if err != nil {
				t.Logf("Failed to cleanup network: %v", err)
			} else {
				t.Log("Shared network cleanup completed")
			}
		}
	})

	t.Log("=== TestSharedNetwork completed successfully ===")
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

	t.Log("=== Starting TestRandomNetworkNames ===")
	ctx := context.Background()

	// Create two networks without specifying names
	t.Log("Creating first network...")
	network1, err := ethereum.Run(ctx, ethereum.Minimal())
	require.NoError(t, err)
	defer func() {
		t.Log("Cleaning up network 1...")
		network1.Cleanup(ctx)
	}()

	t.Log("Creating second network...")
	network2, err := ethereum.Run(ctx, ethereum.Minimal())
	require.NoError(t, err)
	defer func() {
		t.Log("Cleaning up network 2...")
		network2.Cleanup(ctx)
	}()

	// Verify they have different enclave names
	assert.NotEqual(t, network1.EnclaveName(), network2.EnclaveName())
	t.Logf("Network 1: %s", network1.EnclaveName())
	t.Logf("Network 2: %s", network2.EnclaveName())

	// Both should be functional
	t.Log("Verifying both networks are functional...")
	assert.NotEmpty(t, network1.ExecutionClients().All())
	assert.NotEmpty(t, network2.ExecutionClients().All())
	t.Log("Both networks verified as functional")

	t.Log("=== TestRandomNetworkNames completed successfully ===")
}

// TestNetworkFailure demonstrates testing network failure scenarios
func TestNetworkFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Log("=== Starting TestNetworkFailure ===")
	// Test network startup failure
	t.Run("InvalidConfiguration", func(t *testing.T) {
		t.Log("Testing invalid configuration scenario...")
		ctx := context.Background()

		// This should fail due to invalid configuration
		t.Log("Attempting to create network with invalid timeout...")
		network, err := ethereum.Run(ctx,
			ethereum.WithTimeout(0), // Invalid timeout
		)

		if err != nil {
			t.Logf("Expected error occurred: %v", err)
		} else {
			t.Log("Unexpected: no error occurred")
		}

		assert.Error(t, err)
		assert.Nil(t, network)

		// Ensure cleanup if network was partially created
		if network != nil {
			t.Log("Cleaning up partially created network...")
			testutil.CleanupNetwork(t, network)
		}
		t.Log("Invalid configuration test completed")
	})

	t.Log("=== TestNetworkFailure completed successfully ===")
}

// BenchmarkNetworkStartup benchmarks network startup time
func BenchmarkNetworkStartup(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	b.Log("=== Starting BenchmarkNetworkStartup ===")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.Logf("Benchmark iteration %d: creating network...", i+1)
		network, err := ethereum.Run(ctx, ethereum.Minimal())
		require.NoError(b, err)

		b.StopTimer()
		b.Logf("Benchmark iteration %d: cleaning up network...", i+1)
		err = network.Cleanup(ctx)
		require.NoError(b, err)
		b.StartTimer()
	}
	b.Log("=== BenchmarkNetworkStartup completed ===")
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
