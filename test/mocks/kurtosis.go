package mocks

import (
	"context"
	"fmt"
	"time"

	"github.com/ethpandaops/ethereum-package-go/pkg/kurtosis"
)

// MockKurtosisClient is a mock implementation of the Kurtosis client for testing
type MockKurtosisClient struct {
	// Control behavior
	RunPackageFunc      func(ctx context.Context, config kurtosis.RunPackageConfig) (*kurtosis.RunPackageResult, error)
	GetServicesFunc     func(ctx context.Context, enclaveName string) (map[string]*kurtosis.ServiceInfo, error)
	StopEnclaveFunc     func(ctx context.Context, enclaveName string) error
	DestroyEnclaveFunc  func(ctx context.Context, enclaveName string) error
	WaitForServicesFunc func(ctx context.Context, enclaveName string, serviceNames []string, timeout time.Duration) error

	// State tracking
	Enclaves       map[string]*EnclaveState
	CallCount      map[string]int
	LastRunConfig  *kurtosis.RunPackageConfig
}

// EnclaveState tracks the state of a mock enclave
type EnclaveState struct {
	Name     string
	Services map[string]*kurtosis.ServiceInfo
	Running  bool
}

// NewMockKurtosisClient creates a new mock Kurtosis client
func NewMockKurtosisClient() *MockKurtosisClient {
	return &MockKurtosisClient{
		Enclaves:  make(map[string]*EnclaveState),
		CallCount: make(map[string]int),
	}
}

// RunPackage mocks the RunPackage method
func (m *MockKurtosisClient) RunPackage(ctx context.Context, config kurtosis.RunPackageConfig) (*kurtosis.RunPackageResult, error) {
	m.CallCount["RunPackage"]++
	m.LastRunConfig = &config

	if m.RunPackageFunc != nil {
		return m.RunPackageFunc(ctx, config)
	}

	// Default behavior - create enclave with sample services
	enclave := &EnclaveState{
		Name:     config.EnclaveName,
		Services: m.createDefaultServices(),
		Running:  true,
	}
	m.Enclaves[config.EnclaveName] = enclave

	return &kurtosis.RunPackageResult{
		EnclaveName: config.EnclaveName,
		ResponseLines: []string{
			"Starting ethereum-package",
			"Creating execution clients",
			"Creating consensus clients",
			"Starting validators",
			"Network ready",
		},
	}, nil
}

// GetServices mocks the GetServices method
func (m *MockKurtosisClient) GetServices(ctx context.Context, enclaveName string) (map[string]*kurtosis.ServiceInfo, error) {
	m.CallCount["GetServices"]++

	if m.GetServicesFunc != nil {
		return m.GetServicesFunc(ctx, enclaveName)
	}

	enclave, exists := m.Enclaves[enclaveName]
	if !exists {
		return nil, fmt.Errorf("enclave not found: %s", enclaveName)
	}

	if !enclave.Running {
		return nil, fmt.Errorf("enclave not running: %s", enclaveName)
	}

	return enclave.Services, nil
}

// StopEnclave mocks the StopEnclave method
func (m *MockKurtosisClient) StopEnclave(ctx context.Context, enclaveName string) error {
	m.CallCount["StopEnclave"]++

	if m.StopEnclaveFunc != nil {
		return m.StopEnclaveFunc(ctx, enclaveName)
	}

	enclave, exists := m.Enclaves[enclaveName]
	if !exists {
		return fmt.Errorf("enclave not found: %s", enclaveName)
	}

	enclave.Running = false
	return nil
}

// DestroyEnclave mocks the DestroyEnclave method
func (m *MockKurtosisClient) DestroyEnclave(ctx context.Context, enclaveName string) error {
	m.CallCount["DestroyEnclave"]++

	if m.DestroyEnclaveFunc != nil {
		return m.DestroyEnclaveFunc(ctx, enclaveName)
	}

	if _, exists := m.Enclaves[enclaveName]; !exists {
		return fmt.Errorf("enclave not found: %s", enclaveName)
	}

	delete(m.Enclaves, enclaveName)
	return nil
}

// WaitForServices mocks the WaitForServices method
func (m *MockKurtosisClient) WaitForServices(ctx context.Context, enclaveName string, serviceNames []string, timeout time.Duration) error {
	m.CallCount["WaitForServices"]++

	if m.WaitForServicesFunc != nil {
		return m.WaitForServicesFunc(ctx, enclaveName, serviceNames, timeout)
	}

	// Default behavior - immediate success
	return nil
}

// createDefaultServices creates a default set of services for testing
func (m *MockKurtosisClient) createDefaultServices() map[string]*kurtosis.ServiceInfo {
	return map[string]*kurtosis.ServiceInfo{
		"cl-1-geth-lighthouse": {
			Name:      "cl-1-geth-lighthouse",
			UUID:      "uuid-el-1",
			Status:    "RUNNING",
			IPAddress: "172.16.0.2",
			Hostname:  "cl-1-geth-lighthouse.local",
			Ports: map[string]kurtosis.PortInfo{
				"rpc":     {Number: 8545, Protocol: "TCP", MaybeURL: "http://172.16.0.2:8545"},
				"ws":      {Number: 8546, Protocol: "TCP", MaybeURL: "ws://172.16.0.2:8546"},
				"engine":  {Number: 8551, Protocol: "TCP"},
				"metrics": {Number: 9090, Protocol: "TCP"},
				"p2p":     {Number: 30303, Protocol: "TCP"},
			},
		},
		"cl-1-lighthouse-geth": {
			Name:      "cl-1-lighthouse-geth",
			UUID:      "uuid-cl-1",
			Status:    "RUNNING",
			IPAddress: "172.16.0.3",
			Hostname:  "cl-1-lighthouse-geth.local",
			Ports: map[string]kurtosis.PortInfo{
				"beacon":  {Number: 5052, Protocol: "TCP", MaybeURL: "http://172.16.0.3:5052"},
				"metrics": {Number: 5054, Protocol: "TCP"},
				"p2p":     {Number: 9000, Protocol: "TCP"},
			},
		},
		"apache": {
			Name:      "apache",
			UUID:      "uuid-apache",
			Status:    "RUNNING",
			IPAddress: "172.16.0.4",
			Hostname:  "apache.local",
			Ports: map[string]kurtosis.PortInfo{
				"http": {Number: 80, Protocol: "TCP", MaybeURL: "http://172.16.0.4:80"},
			},
		},
		"prometheus": {
			Name:      "prometheus",
			UUID:      "uuid-prometheus",
			Status:    "RUNNING",
			IPAddress: "172.16.0.5",
			Hostname:  "prometheus.local",
			Ports: map[string]kurtosis.PortInfo{
				"http": {Number: 9090, Protocol: "TCP", MaybeURL: "http://172.16.0.5:9090"},
			},
		},
		"grafana": {
			Name:      "grafana",
			UUID:      "uuid-grafana",
			Status:    "RUNNING",
			IPAddress: "172.16.0.6",
			Hostname:  "grafana.local",
			Ports: map[string]kurtosis.PortInfo{
				"http": {Number: 3000, Protocol: "TCP", MaybeURL: "http://172.16.0.6:3000"},
			},
		},
	}
}

// AddService adds a service to an enclave
func (m *MockKurtosisClient) AddService(enclaveName string, service *kurtosis.ServiceInfo) error {
	enclave, exists := m.Enclaves[enclaveName]
	if !exists {
		return fmt.Errorf("enclave not found: %s", enclaveName)
	}

	if enclave.Services == nil {
		enclave.Services = make(map[string]*kurtosis.ServiceInfo)
	}

	enclave.Services[service.Name] = service
	return nil
}

// SetServiceStatus updates the status of a service
func (m *MockKurtosisClient) SetServiceStatus(enclaveName, serviceName, status string) error {
	enclave, exists := m.Enclaves[enclaveName]
	if !exists {
		return fmt.Errorf("enclave not found: %s", enclaveName)
	}

	service, exists := enclave.Services[serviceName]
	if !exists {
		return fmt.Errorf("service not found: %s", serviceName)
	}

	service.Status = status
	return nil
}

// Reset resets the mock state
func (m *MockKurtosisClient) Reset() {
	m.Enclaves = make(map[string]*EnclaveState)
	m.CallCount = make(map[string]int)
	m.LastRunConfig = nil
	m.RunPackageFunc = nil
	m.GetServicesFunc = nil
	m.StopEnclaveFunc = nil
	m.DestroyEnclaveFunc = nil
	m.WaitForServicesFunc = nil
}

// Verify interface compliance
var _ kurtosis.Client = (*MockKurtosisClient)(nil)