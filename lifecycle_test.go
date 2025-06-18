package ethereum

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ethpandaops/ethereum-package-go/pkg/kurtosis"
	"github.com/ethpandaops/ethereum-package-go/pkg/types"
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

func (m *MockKurtosisClient) DestroyEnclave(ctx context.Context, enclaveName string) error {
	args := m.Called(ctx, enclaveName)
	return args.Error(0)
}

func (m *MockKurtosisClient) StopEnclave(ctx context.Context, enclaveName string) error {
	args := m.Called(ctx, enclaveName)
	return args.Error(0)
}

func (m *MockKurtosisClient) WaitForServices(ctx context.Context, enclaveName string, serviceNames []string, timeout time.Duration) error {
	args := m.Called(ctx, enclaveName, serviceNames, timeout)
	return args.Error(0)
}

func TestRun_NetworkLifecycle(t *testing.T) {
	ctx := context.Background()
	mockClient := new(MockKurtosisClient)

	// Setup mock expectations
	runResult := &kurtosis.RunPackageResult{
		EnclaveName: "test-enclave",
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

	mockClient.On("RunPackage", ctx, mock.AnythingOfType("kurtosis.RunPackageConfig")).Return(runResult, nil)
	mockClient.On("WaitForServices", ctx, mock.Anything, []string{}, mock.AnythingOfType("time.Duration")).Return(nil)
	mockClient.On("GetServices", ctx, mock.Anything).Return(services, nil)
	mockClient.On("DestroyEnclave", ctx, mock.Anything).Return(nil)

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

	// Verify all expected calls were made
	mockClient.AssertExpectations(t)
}

func TestRun_FailureScenarios(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*MockKurtosisClient)
		options       []RunOption
		expectedError string
	}{
		{
			name: "invalid configuration",
			setupMock: func(m *MockKurtosisClient) {
				// No expectations - should fail before calling Kurtosis
			},
			options: []RunOption{
				WithTimeout(0), // Invalid timeout
			},
			expectedError: "invalid configuration: timeout must be positive",
		},
		{
			name: "kurtosis run package failure",
			setupMock: func(m *MockKurtosisClient) {
				m.On("RunPackage", mock.Anything, mock.AnythingOfType("kurtosis.RunPackageConfig")).
					Return(nil, errors.New("kurtosis error"))
			},
			options:       []RunOption{Minimal()},
			expectedError: "failed to run ethereum-package: kurtosis error",
		},
		{
			name: "service startup timeout",
			setupMock: func(m *MockKurtosisClient) {
				runResult := &kurtosis.RunPackageResult{
					EnclaveName: "test-enclave",
					ResponseLines: []string{"Network started"},
				}
				m.On("RunPackage", mock.Anything, mock.AnythingOfType("kurtosis.RunPackageConfig")).
					Return(runResult, nil)
				m.On("WaitForServices", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("timeout waiting for services"))
				m.On("DestroyEnclave", mock.Anything, mock.Anything).Return(nil)
			},
			options:       []RunOption{Minimal()},
			expectedError: "services failed to start: timeout waiting for services",
		},
		{
			name: "service discovery failure",
			setupMock: func(m *MockKurtosisClient) {
				runResult := &kurtosis.RunPackageResult{
					EnclaveName: "test-enclave",
					ResponseLines: []string{"Network started"},
				}
				m.On("RunPackage", mock.Anything, mock.AnythingOfType("kurtosis.RunPackageConfig")).
					Return(runResult, nil)
				m.On("WaitForServices", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				m.On("GetServices", mock.Anything, mock.Anything).
					Return(nil, errors.New("failed to get services"))
				m.On("DestroyEnclave", mock.Anything, mock.Anything).Return(nil)
			},
			options:       []RunOption{Minimal()},
			expectedError: "failed to discover services: failed to get services",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockClient := new(MockKurtosisClient)
			tt.setupMock(mockClient)

			// Add WithKurtosisClient to options
			opts := append(tt.options, WithKurtosisClient(mockClient))

			network, err := Run(ctx, opts...)
			assert.Nil(t, network)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)

			mockClient.AssertExpectations(t)
		})
	}
}

func TestRun_DryRunMode(t *testing.T) {
	ctx := context.Background()
	mockClient := new(MockKurtosisClient)

	runResult := &kurtosis.RunPackageResult{
		EnclaveName: "dry-run-enclave",
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

	// In dry run mode, WaitForServices should not be called
	mockClient.On("RunPackage", ctx, mock.MatchedBy(func(cfg kurtosis.RunPackageConfig) bool {
		return cfg.DryRun == true
	})).Return(runResult, nil)
	mockClient.On("GetServices", ctx, mock.Anything).Return(services, nil)

	network, err := Run(ctx,
		Minimal(),
		WithDryRun(true),
		WithKurtosisClient(mockClient),
	)
	require.NoError(t, err)
	require.NotNil(t, network)

	// Verify WaitForServices was NOT called
	mockClient.AssertNotCalled(t, "WaitForServices", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockClient.AssertExpectations(t)
}

func TestNetwork_Cleanup(t *testing.T) {
	ctx := context.Background()
	mockClient := new(MockKurtosisClient)

	// Setup successful cleanup
	mockClient.On("DestroyEnclave", ctx, "test-enclave").Return(nil).Once()

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

	mockClient.AssertExpectations(t)
}

func TestNetwork_CleanupFailure(t *testing.T) {
	ctx := context.Background()
	mockClient := new(MockKurtosisClient)

	// Setup failing cleanup
	mockClient.On("DestroyEnclave", ctx, "test-enclave").Return(errors.New("cleanup failed")).Once()

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

	mockClient.AssertExpectations(t)
}

func TestRun_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	mockClient := new(MockKurtosisClient)

	// Cancel context immediately
	cancel()

	// Setup mock to handle context cancellation
	mockClient.On("RunPackage", mock.Anything, mock.AnythingOfType("kurtosis.RunPackageConfig")).
		Return(nil, context.Canceled).Maybe()

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