package testutil

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ethpandaops/ethereum-package-go"
	"github.com/ethpandaops/ethereum-package-go/pkg/client"
	"github.com/ethpandaops/ethereum-package-go/pkg/config"
	"github.com/ethpandaops/ethereum-package-go/pkg/network"
)

// TestNetwork wraps a network with test-specific functionality
type TestNetwork struct {
	network.Network
	t       testing.TB
	cleanup []func()
	mu      sync.Mutex
}

// NewTestNetwork creates a new test network
func NewTestNetwork(t testing.TB, opts ...ethereum.RunOption) *TestNetwork {
	t.Helper()

	t.Log("Starting network setup...")
	ctx := context.Background()
	t.Log("Calling ethereum.Run() to create network")
	network, err := ethereum.Run(ctx, opts...)
	if err != nil {
		t.Fatalf("Failed to start test network: %v", err)
	}
	t.Log("Network created successfully")

	tn := &TestNetwork{
		Network: network,
		t:       t,
		cleanup: []func(){},
	}

	t.Logf("Network enclave: %s", network.EnclaveName())
	t.Logf("Found %d execution clients", len(network.ExecutionClients().All()))
	t.Logf("Found %d consensus clients", len(network.ConsensusClients().All()))
	t.Logf("Found %d total services", len(network.Services()))

	// Register cleanup
	t.Cleanup(func() {
		t.Log("Cleaning up test network...")
		ctx := context.Background()
		if err := tn.Cleanup(ctx); err != nil {
			t.Logf("Failed to cleanup network: %v", err)
		} else {
			t.Log("Network cleanup completed")
		}
	})

	return tn
}

// StartNetwork starts a test network with default options
func StartNetwork(t testing.TB) *TestNetwork {
	return NewTestNetwork(t, ethereum.Minimal())
}

// StartSharedNetwork starts or connects to a shared test network
// This is useful for running multiple tests against the same network
func StartSharedNetwork(t testing.TB, name string, opts ...ethereum.RunOption) *TestNetwork {
	t.Helper()

	t.Logf("Starting or finding shared network: %s", name)
	ctx := context.Background()
	network, err := ethereum.FindOrCreateNetwork(ctx, name, opts...)
	if err != nil {
		t.Fatalf("Failed to start/find shared network: %v", err)
	}
	t.Logf("Shared network ready: %s", name)

	tn := &TestNetwork{
		Network: network,
		t:       t,
		cleanup: []func(){},
	}

	// Don't automatically cleanup shared networks
	// The caller should decide when to clean up

	return tn
}

// StartNetworkWithChainID starts a test network with a specific chain ID
func StartNetworkWithChainID(t testing.TB, chainID uint64) *TestNetwork {
	return NewTestNetwork(t,
		ethereum.Minimal(),
		ethereum.WithChainID(chainID),
	)
}

// StartNetworkWithParticipants starts a test network with custom participants
func StartNetworkWithParticipants(t testing.TB, participants []config.ParticipantConfig) *TestNetwork {
	return NewTestNetwork(t,
		ethereum.WithParticipants(participants),
	)
}

// GetExecutionClient returns the first execution client or fails the test
func (tn *TestNetwork) GetExecutionClient() client.ExecutionClient {
	tn.t.Helper()

	clients := tn.ExecutionClients().All()
	if len(clients) == 0 {
		tn.t.Fatal("No execution clients available")
	}

	return clients[0]
}

// GetConsensusClient returns the first consensus client or fails the test
func (tn *TestNetwork) GetConsensusClient() client.ConsensusClient {
	tn.t.Helper()

	clients := tn.ConsensusClients().All()
	if len(clients) == 0 {
		tn.t.Fatal("No consensus clients available")
	}

	return clients[0]
}

// WaitForSync waits for all clients to sync or fails the test
func (tn *TestNetwork) WaitForSync(timeout time.Duration) {
	tn.t.Helper()

	tn.t.Logf("Waiting for network sync (timeout: %v)...", timeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Wait for execution clients
	execClients := tn.ExecutionClients().All()
	if len(execClients) > 0 {
		tn.t.Logf("Waiting for %d execution clients to sync...", len(execClients))
		for _, client := range execClients {
			tn.t.Logf("Syncing execution client: %s", client.Name())
			if waiter, ok := client.(interface{ WaitForSync(context.Context) error }); ok {
				if err := waiter.WaitForSync(ctx); err != nil {
					tn.t.Fatalf("Execution client %s failed to sync: %v", client.Name(), err)
				}
				tn.t.Logf("Execution client %s synced successfully", client.Name())
			}
		}
	}

	// Wait for consensus clients
	consClients := tn.ConsensusClients().All()
	if len(consClients) > 0 {
		tn.t.Logf("Waiting for %d consensus clients to sync...", len(consClients))
		for _, client := range consClients {
			tn.t.Logf("Syncing consensus client: %s", client.Name())
			if waiter, ok := client.(interface{ WaitForSync(context.Context) error }); ok {
				if err := waiter.WaitForSync(ctx); err != nil {
					tn.t.Fatalf("Consensus client %s failed to sync: %v", client.Name(), err)
				}
				tn.t.Logf("Consensus client %s synced successfully", client.Name())
			}
		}
	}

	tn.t.Log("All clients synced successfully")
}

// RequireHealthy checks that all services are healthy or fails the test
func (tn *TestNetwork) RequireHealthy() {
	tn.t.Helper()

	tn.t.Log("Checking network health...")
	ctx := context.Background()

	// Check execution clients
	execClients := tn.ExecutionClients().All()
	if len(execClients) > 0 {
		tn.t.Logf("Checking health of %d execution clients...", len(execClients))
		for _, client := range execClients {
			tn.t.Logf("Checking execution client health: %s", client.Name())
			if checker, ok := client.(interface{ IsHealthy(context.Context) bool }); ok {
				if !checker.IsHealthy(ctx) {
					tn.t.Fatalf("Execution client %s is not healthy", client.Name())
				}
				tn.t.Logf("Execution client %s is healthy", client.Name())
			}
		}
	}

	// Check consensus clients
	consClients := tn.ConsensusClients().All()
	if len(consClients) > 0 {
		tn.t.Logf("Checking health of %d consensus clients...", len(consClients))
		for _, client := range consClients {
			tn.t.Logf("Checking consensus client health: %s", client.Name())
			if checker, ok := client.(interface{ IsHealthy(context.Context) bool }); ok {
				if !checker.IsHealthy(ctx) {
					tn.t.Fatalf("Consensus client %s is not healthy", client.Name())
				}
				tn.t.Logf("Consensus client %s is healthy", client.Name())
			}
		}
	}

	tn.t.Log("All network components are healthy")
}

// AddCleanup adds a cleanup function to be called when the test ends
func (tn *TestNetwork) AddCleanup(fn func()) {
	tn.mu.Lock()
	defer tn.mu.Unlock()
	tn.cleanup = append(tn.cleanup, fn)
}

// Cleanup cleans up the test network
func (tn *TestNetwork) Cleanup(ctx context.Context) error {
	tn.mu.Lock()
	cleanupFuncs := tn.cleanup
	tn.mu.Unlock()

	// Run cleanup functions in reverse order
	for i := len(cleanupFuncs) - 1; i >= 0; i-- {
		cleanupFuncs[i]()
	}

	// Cleanup the network
	return tn.Network.Cleanup(ctx)
}

// CleanupNetwork is a helper to cleanup a network in tests
func CleanupNetwork(t testing.TB, network network.Network) {
	t.Helper()

	if network == nil {
		return
	}

	ctx := context.Background()
	if err := network.Cleanup(ctx); err != nil {
		t.Logf("Failed to cleanup network: %v", err)
	}
}

// ParallelNetworks manages multiple networks for parallel tests
type ParallelNetworks struct {
	networks map[string]*TestNetwork
	mu       sync.RWMutex
}

// NewParallelNetworks creates a new parallel networks manager
func NewParallelNetworks() *ParallelNetworks {
	return &ParallelNetworks{
		networks: make(map[string]*TestNetwork),
	}
}

// StartNetwork starts a network for a parallel test
func (pn *ParallelNetworks) StartNetwork(t *testing.T, opts ...ethereum.RunOption) *TestNetwork {
	t.Helper()

	network := NewTestNetwork(t, opts...)

	pn.mu.Lock()
	pn.networks[t.Name()] = network
	pn.mu.Unlock()

	t.Cleanup(func() {
		pn.mu.Lock()
		delete(pn.networks, t.Name())
		pn.mu.Unlock()
	})

	return network
}

// GetNetwork gets a network by test name
func (pn *ParallelNetworks) GetNetwork(testName string) *TestNetwork {
	pn.mu.RLock()
	defer pn.mu.RUnlock()
	return pn.networks[testName]
}

// GetAllNetworks returns all active networks
func (pn *ParallelNetworks) GetAllNetworks() []*TestNetwork {
	pn.mu.RLock()
	defer pn.mu.RUnlock()

	networks := make([]*TestNetwork, 0, len(pn.networks))
	for _, network := range pn.networks {
		networks = append(networks, network)
	}

	return networks
}

// NetworkAssertion provides assertion helpers for networks
type NetworkAssertion struct {
	t       testing.TB
	network network.Network
}

// Assert creates a new network assertion helper
func Assert(t testing.TB, net network.Network) *NetworkAssertion {
	return &NetworkAssertion{
		t:       t,
		network: net,
	}
}

// HasExecutionClients asserts that the network has execution clients
func (na *NetworkAssertion) HasExecutionClients(count int) *NetworkAssertion {
	na.t.Helper()

	actual := len(na.network.ExecutionClients().All())
	if actual != count {
		na.t.Errorf("Expected %d execution clients, got %d", count, actual)
	}

	return na
}

// HasConsensusClients asserts that the network has consensus clients
func (na *NetworkAssertion) HasConsensusClients(count int) *NetworkAssertion {
	na.t.Helper()

	actual := len(na.network.ConsensusClients().All())
	if actual != count {
		na.t.Errorf("Expected %d consensus clients, got %d", count, actual)
	}

	return na
}

// HasChainID asserts that the network has the expected chain ID
func (na *NetworkAssertion) HasChainID(chainID uint64) *NetworkAssertion {
	na.t.Helper()

	actual := na.network.ChainID()
	if actual != chainID {
		na.t.Errorf("Expected chain ID %d, got %d", chainID, actual)
	}

	return na
}

// HasService asserts that the network has a specific service
func (na *NetworkAssertion) HasService(serviceType network.ServiceType) *NetworkAssertion {
	na.t.Helper()

	found := false
	for _, service := range na.network.Services() {
		if service.Type == serviceType {
			found = true
			break
		}
	}

	if !found {
		na.t.Errorf("Expected to find service of type %s", serviceType)
	}

	return na
}
