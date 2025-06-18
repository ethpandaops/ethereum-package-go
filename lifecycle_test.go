package ethereum

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ethpandaops/ethereum-package-go/pkg/kurtosis"
	"github.com/ethpandaops/ethereum-package-go/pkg/types"
	"github.com/ethpandaops/ethereum-package-go/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_NetworkLifecycle(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewMockKurtosisClient()

	// Setup mock expectations
	runResult := &kurtosis.RunPackageResult{
		EnclaveName:   "test-enclave",
		ResponseLines: []string{"Network started successfully"},
	}

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
			},
		},
	}

	mockClient.RunPackageFunc = func(ctx context.Context, config kurtosis.RunPackageConfig) (*kurtosis.RunPackageResult, error) {
		return runResult, nil
	}
	mockClient.WaitForServicesFunc = func(ctx context.Context, enclaveName string, serviceNames []string, timeout time.Duration) error {
		return nil
	}
	mockClient.GetServicesFunc = func(ctx context.Context, enclaveName string) (map[string]*kurtosis.ServiceInfo, error) {
		return services, nil
	}
	mockClient.DestroyEnclaveFunc = func(ctx context.Context, enclaveName string) error {
		return nil
	}

	// Run the network
	network, err := Run(ctx,
		Minimal(),
		WithKurtosisClient(mockClient),
		WithTimeout(1*time.Minute),
	)
	require.NoError(t, err)
	require.NotNil(t, network)

	// Verify network properties
	assert.Equal(t, uint64(12345), network.ChainID())
	assert.NotEmpty(t, network.EnclaveName())

	// Debug: log all services
	for _, svc := range network.Services() {
		t.Logf("Service: %s, Type: %s", svc.Name, svc.Type)
	}

	// Verify clients were discovered
	t.Logf("Execution clients: %d", len(network.ExecutionClients().All()))
	t.Logf("Consensus clients: %d", len(network.ConsensusClients().All()))
	for _, client := range network.ExecutionClients().All() {
		t.Logf("Execution client: %s", client.Name())
	}
	for _, client := range network.ConsensusClients().All() {
		t.Logf("Consensus client: %s", client.Name())
	}
	assert.NotEmpty(t, network.ExecutionClients().All())
	assert.NotEmpty(t, network.ConsensusClients().All())

	// Test cleanup
	err = network.Cleanup(ctx)
	assert.NoError(t, err)

	// Verify calls were made
	assert.Greater(t, mockClient.CallCount["RunPackage"], 0)
	assert.Greater(t, mockClient.CallCount["GetServices"], 0)
}

func TestRun_FailureScenarios(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*mocks.MockKurtosisClient)
		options       []RunOption
		expectedError string
	}{
		{
			name: "invalid configuration",
			setupMock: func(m *mocks.MockKurtosisClient) {
				// No expectations - should fail before calling Kurtosis
			},
			options: []RunOption{
				WithTimeout(0), // Invalid timeout
			},
			expectedError: "timeout must be positive",
		},
		{
			name: "kurtosis run package failure",
			setupMock: func(m *mocks.MockKurtosisClient) {
				m.RunPackageFunc = func(ctx context.Context, config kurtosis.RunPackageConfig) (*kurtosis.RunPackageResult, error) {
					return nil, errors.New("kurtosis error")
				}
			},
			options:       []RunOption{Minimal()},
			expectedError: "kurtosis error",
		},
		{
			name: "service startup timeout",
			setupMock: func(m *mocks.MockKurtosisClient) {
				runResult := &kurtosis.RunPackageResult{
					EnclaveName:   "test-enclave",
					ResponseLines: []string{"Network started"},
				}
				m.RunPackageFunc = func(ctx context.Context, config kurtosis.RunPackageConfig) (*kurtosis.RunPackageResult, error) {
					return runResult, nil
				}
				m.WaitForServicesFunc = func(ctx context.Context, enclaveName string, serviceNames []string, timeout time.Duration) error {
					return errors.New("timeout waiting for services")
				}
				m.DestroyEnclaveFunc = func(ctx context.Context, enclaveName string) error {
					return nil
				}
			},
			options:       []RunOption{Minimal()},
			expectedError: "timeout waiting for services",
		},
		{
			name: "service discovery failure",
			setupMock: func(m *mocks.MockKurtosisClient) {
				runResult := &kurtosis.RunPackageResult{
					EnclaveName:   "test-enclave",
					ResponseLines: []string{"Network started"},
				}
				m.RunPackageFunc = func(ctx context.Context, config kurtosis.RunPackageConfig) (*kurtosis.RunPackageResult, error) {
					return runResult, nil
				}
				m.WaitForServicesFunc = func(ctx context.Context, enclaveName string, serviceNames []string, timeout time.Duration) error {
					return nil
				}
				m.GetServicesFunc = func(ctx context.Context, enclaveName string) (map[string]*kurtosis.ServiceInfo, error) {
					return nil, errors.New("failed to get services")
				}
				m.DestroyEnclaveFunc = func(ctx context.Context, enclaveName string) error {
					return nil
				}
			},
			options:       []RunOption{Minimal()},
			expectedError: "failed to get services",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockClient := mocks.NewMockKurtosisClient()
			tt.setupMock(mockClient)

			// Add WithKurtosisClient to options
			opts := append(tt.options, WithKurtosisClient(mockClient))

			network, err := Run(ctx, opts...)
			assert.Nil(t, network)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestRun_DryRunMode(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewMockKurtosisClient()

	runResult := &kurtosis.RunPackageResult{
		EnclaveName:   "dry-run-enclave",
		ResponseLines: []string{"Dry run completed"},
	}

	services := map[string]*kurtosis.ServiceInfo{
		"el-1-geth-lighthouse": {
			Name:      "el-1-geth-lighthouse",
			UUID:      "uuid-1",
			Status:    "running",
			IPAddress: "10.0.0.1",
			Ports:     map[string]kurtosis.PortInfo{},
		},
	}

	// In dry run mode, setup functions
	mockClient.RunPackageFunc = func(ctx context.Context, config kurtosis.RunPackageConfig) (*kurtosis.RunPackageResult, error) {
		assert.True(t, config.DryRun)
		return runResult, nil
	}
	mockClient.GetServicesFunc = func(ctx context.Context, enclaveName string) (map[string]*kurtosis.ServiceInfo, error) {
		return services, nil
	}

	network, err := Run(ctx,
		Minimal(),
		WithDryRun(true),
		WithKurtosisClient(mockClient),
	)
	require.NoError(t, err)
	require.NotNil(t, network)

	// Verify WaitForServices was NOT called (CallCount should be 0)
	assert.Equal(t, 0, mockClient.CallCount["WaitForServices"])
}

func TestNetwork_Cleanup(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewMockKurtosisClient()

	// Setup successful cleanup
	mockClient.DestroyEnclaveFunc = func(ctx context.Context, enclaveName string) error {
		assert.Equal(t, "test-enclave", enclaveName)
		return nil
	}

	config := types.NetworkConfig{
		Name:        "test-network",
		ChainID:     12345,
		EnclaveName: "test-enclave",
		CleanupFunc: func(ctx context.Context) error {
			return mockClient.DestroyEnclave(ctx, "test-enclave")
		},
	}

	network := types.NewNetwork(config)
	err := network.Cleanup(ctx)
	assert.NoError(t, err)

	// Verify the cleanup function was called
	assert.Greater(t, mockClient.CallCount["DestroyEnclave"], 0)
}

func TestNetwork_CleanupFailure(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewMockKurtosisClient()

	// Setup failing cleanup
	mockClient.DestroyEnclaveFunc = func(ctx context.Context, enclaveName string) error {
		return errors.New("cleanup failed")
	}

	config := types.NetworkConfig{
		Name:        "test-network",
		ChainID:     12345,
		EnclaveName: "test-enclave",
		CleanupFunc: func(ctx context.Context) error {
			return mockClient.DestroyEnclave(ctx, "test-enclave")
		},
	}

	network := types.NewNetwork(config)
	err := network.Cleanup(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cleanup failed")
}

func TestRun_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	mockClient := mocks.NewMockKurtosisClient()

	// Cancel context immediately
	cancel()

	// Setup mock to handle context cancellation
	mockClient.RunPackageFunc = func(ctx context.Context, config kurtosis.RunPackageConfig) (*kurtosis.RunPackageResult, error) {
		return nil, context.Canceled
	}

	network, err := Run(ctx,
		Minimal(),
		WithKurtosisClient(mockClient),
	)

	assert.Nil(t, network)
	assert.Error(t, err)
}

func TestNetwork_Stop(t *testing.T) {
	ctx := context.Background()

	config := types.NetworkConfig{
		Name:        "test-network",
		ChainID:     12345,
		EnclaveName: "test-enclave",
	}

	network := types.NewNetwork(config)

	// Stop is currently a no-op, but we test it works
	err := network.Stop(ctx)
	assert.NoError(t, err)
}
