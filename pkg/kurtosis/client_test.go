package kurtosis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockKurtosisClient is a mock implementation for testing
type MockKurtosisClient struct {
	services       map[string]map[string]*ServiceInfo
	runPackageFunc func(ctx context.Context, config RunPackageConfig) (*RunPackageResult, error)
	enclaveStatus  map[string]bool
}

func NewMockKurtosisClient() *MockKurtosisClient {
	return &MockKurtosisClient{
		services:      make(map[string]map[string]*ServiceInfo),
		enclaveStatus: make(map[string]bool),
	}
}

func (m *MockKurtosisClient) RunPackage(ctx context.Context, config RunPackageConfig) (*RunPackageResult, error) {
	if m.runPackageFunc != nil {
		return m.runPackageFunc(ctx, config)
	}

	// Default mock behavior
	m.enclaveStatus[config.EnclaveName] = true
	m.services[config.EnclaveName] = make(map[string]*ServiceInfo)

	return &RunPackageResult{
		EnclaveName: config.EnclaveName,
		ResponseLines: []string{
			"Starting ethereum-package",
			"Creating execution clients",
			"Creating consensus clients",
			"Network ready",
		},
	}, nil
}

func (m *MockKurtosisClient) GetServices(ctx context.Context, enclaveName string) (map[string]*ServiceInfo, error) {
	services, exists := m.services[enclaveName]
	if !exists {
		return nil, fmt.Errorf("enclave not found: %s", enclaveName)
	}
	return services, nil
}

func (m *MockKurtosisClient) StopEnclave(ctx context.Context, enclaveName string) error {
	if _, exists := m.enclaveStatus[enclaveName]; !exists {
		return fmt.Errorf("enclave not found: %s", enclaveName)
	}
	m.enclaveStatus[enclaveName] = false
	return nil
}

func (m *MockKurtosisClient) AddService(enclaveName, serviceName string, service *ServiceInfo) {
	if m.services[enclaveName] == nil {
		m.services[enclaveName] = make(map[string]*ServiceInfo)
	}
	m.services[enclaveName][serviceName] = service
}

func TestRunPackageConfig(t *testing.T) {
	config := RunPackageConfig{
		PackageID:       "github.com/ethpandaops/ethereum-package",
		EnclaveName:     "test-enclave",
		ConfigYAML:      "participants: []",
		DryRun:          false,
		Parallelism:     4,
		VerboseMode:     true,
		ImageDownload:   true,
		NonBlockingMode: false,
	}

	assert.Equal(t, "github.com/ethpandaops/ethereum-package", config.PackageID)
	assert.Equal(t, "test-enclave", config.EnclaveName)
	assert.Equal(t, "participants: []", config.ConfigYAML)
	assert.False(t, config.DryRun)
	assert.Equal(t, 4, config.Parallelism)
	assert.True(t, config.VerboseMode)
	assert.True(t, config.ImageDownload)
	assert.False(t, config.NonBlockingMode)
}

func TestServiceInfo(t *testing.T) {
	service := ServiceInfo{
		Name:      "geth-1",
		UUID:      "uuid-123",
		Status:    "RUNNING",
		IPAddress: "172.16.0.2",
		Hostname:  "geth-1.local",
		Ports: map[string]PortInfo{
			"rpc": {
				Number:            8545,
				Protocol:          "TCP",
				MaybeURL:          "http://172.16.0.2:8545",
				TransportProtocol: "TCP",
			},
			"ws": {
				Number:            8546,
				Protocol:          "TCP",
				MaybeURL:          "ws://172.16.0.2:8546",
				TransportProtocol: "TCP",
			},
		},
	}

	assert.Equal(t, "geth-1", service.Name)
	assert.Equal(t, "uuid-123", service.UUID)
	assert.Equal(t, "RUNNING", service.Status)
	assert.Equal(t, "172.16.0.2", service.IPAddress)
	assert.Equal(t, "geth-1.local", service.Hostname)
	assert.Len(t, service.Ports, 2)

	rpcPort, exists := service.Ports["rpc"]
	require.True(t, exists)
	assert.Equal(t, uint16(8545), rpcPort.Number)
	assert.Equal(t, "TCP", rpcPort.Protocol)
	assert.Equal(t, "http://172.16.0.2:8545", rpcPort.MaybeURL)
}

func TestMockRunPackage(t *testing.T) {
	mock := NewMockKurtosisClient()
	ctx := context.Background()

	config := RunPackageConfig{
		PackageID:   "github.com/ethpandaops/ethereum-package",
		EnclaveName: "test-enclave",
		ConfigYAML:  "participants: []",
	}

	result, err := mock.RunPackage(ctx, config)
	require.NoError(t, err)
	assert.Equal(t, "test-enclave", result.EnclaveName)
	assert.Len(t, result.ResponseLines, 4)
	assert.Contains(t, result.ResponseLines[0], "Starting ethereum-package")
}

func TestMockGetServices(t *testing.T) {
	mock := NewMockKurtosisClient()
	ctx := context.Background()

	// Add test service
	service := &ServiceInfo{
		Name:   "geth-1",
		Status: "RUNNING",
		Ports: map[string]PortInfo{
			"rpc": {Number: 8545, Protocol: "TCP"},
		},
	}
	mock.AddService("test-enclave", "geth-1", service)

	// Get services
	services, err := mock.GetServices(ctx, "test-enclave")
	require.NoError(t, err)
	assert.Len(t, services, 1)

	gethService, exists := services["geth-1"]
	require.True(t, exists)
	assert.Equal(t, "geth-1", gethService.Name)
	assert.Equal(t, "RUNNING", gethService.Status)
}

func TestMockGetServicesEnclaveNotFound(t *testing.T) {
	mock := NewMockKurtosisClient()
	ctx := context.Background()

	_, err := mock.GetServices(ctx, "non-existent-enclave")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "enclave not found")
}

func TestMockStopEnclave(t *testing.T) {
	mock := NewMockKurtosisClient()
	ctx := context.Background()

	// Create enclave first
	config := RunPackageConfig{
		EnclaveName: "test-enclave",
	}
	_, err := mock.RunPackage(ctx, config)
	require.NoError(t, err)

	// Stop enclave
	err = mock.StopEnclave(ctx, "test-enclave")
	assert.NoError(t, err)
	assert.False(t, mock.enclaveStatus["test-enclave"])
}

func TestMockStopEnclaveNotFound(t *testing.T) {
	mock := NewMockKurtosisClient()
	ctx := context.Background()

	err := mock.StopEnclave(ctx, "non-existent-enclave")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "enclave not found")
}

func TestWaitForServicesTimeout(t *testing.T) {
	client := &KurtosisClient{
		enclaves: make(map[string]bool),
	}
	client.enclaves["test-enclave"] = true

	ctx := context.Background()
	err := client.WaitForServices(ctx, "test-enclave", []string{"service1"}, 2*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

