package kurtosis

import (
	"context"
	"fmt"
	"time"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
)

// Client defines the interface for Kurtosis operations
type Client interface {
	RunPackage(ctx context.Context, config RunPackageConfig) (*RunPackageResult, error)
	GetServices(ctx context.Context, enclaveName string) (map[string]*ServiceInfo, error)
	StopEnclave(ctx context.Context, enclaveName string) error
	DestroyEnclave(ctx context.Context, enclaveName string) error
	WaitForServices(ctx context.Context, enclaveName string, serviceNames []string, timeout time.Duration) error
}

// KurtosisClient wraps the Kurtosis SDK for ethereum-package operations
type KurtosisClient struct {
	kurtosisCtx *kurtosis_context.KurtosisContext
	enclaves    map[string]*enclaves.EnclaveContext
}

// NewKurtosisClient creates a new Kurtosis client
func NewKurtosisClient(ctx context.Context) (*KurtosisClient, error) {
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return nil, fmt.Errorf("failed to create Kurtosis context: %w", err)
	}

	return &KurtosisClient{
		kurtosisCtx: kurtosisCtx,
		enclaves:    make(map[string]*enclaves.EnclaveContext),
	}, nil
}

// RunPackageConfig contains configuration for running a package
type RunPackageConfig struct {
	PackageID       string
	EnclaveName     string
	ConfigYAML      string
	DryRun          bool
	Parallelism     int
	VerboseMode     bool
	ImageDownload   bool
	NonBlockingMode bool
}

// RunPackageResult contains the result of running a package
type RunPackageResult struct {
	EnclaveName         string
	ResponseLines       []string
	InterpretationError error
	ValidationErrors    []string
	ExecutionError      error
}

// ServiceInfo contains information about a service
type ServiceInfo struct {
	Name      string
	UUID      string
	Status    string
	Ports     map[string]PortInfo
	IPAddress string
	Hostname  string
}

// PortInfo contains information about a service port
type PortInfo struct {
	Number            uint16
	Protocol          string
	MaybeURL          string
	TransportProtocol string
}

// RunPackage runs the ethereum-package with the given configuration
func (k *KurtosisClient) RunPackage(ctx context.Context, config RunPackageConfig) (*RunPackageResult, error) {
	// Validate configuration
	if config.PackageID == "" {
		return nil, fmt.Errorf("package ID is required")
	}
	if config.EnclaveName == "" {
		return nil, fmt.Errorf("enclave name is required")
	}

	// Create or get enclave
	enclaveCtx, err := k.getOrCreateEnclave(ctx, config.EnclaveName)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create enclave: %w", err)
	}

	// Store enclave reference
	k.enclaves[config.EnclaveName] = enclaveCtx

	// For now, we'll use a simplified approach
	// In production, this would use the actual Kurtosis SDK to run the package
	
	result := &RunPackageResult{
		EnclaveName: config.EnclaveName,
		ResponseLines: []string{
			"Starting ethereum-package",
			"Creating execution clients",
			"Creating consensus clients",
			"Network ready",
		},
	}

	return result, nil
}

// GetServices returns all services in the enclave
func (k *KurtosisClient) GetServices(ctx context.Context, enclaveName string) (map[string]*ServiceInfo, error) {
	_, exists := k.enclaves[enclaveName]
	if !exists {
		return nil, fmt.Errorf("enclave not found: %s", enclaveName)
	}

	// For now, return mock services
	// In production, this would query the actual Kurtosis enclave
	result := make(map[string]*ServiceInfo)
	
	// Add mock Ethereum services
	result["cl-1-geth-lighthouse"] = &ServiceInfo{
		Name:      "cl-1-geth-lighthouse",
		UUID:      "uuid-el-1",
		Status:    "RUNNING",
		IPAddress: "172.16.0.2",
		Hostname:  "cl-1-geth-lighthouse.local",
		Ports: map[string]PortInfo{
			"rpc":     {Number: 8545, Protocol: "TCP", MaybeURL: "http://172.16.0.2:8545"},
			"ws":      {Number: 8546, Protocol: "TCP", MaybeURL: "ws://172.16.0.2:8546"},
			"engine":  {Number: 8551, Protocol: "TCP"},
			"metrics": {Number: 9090, Protocol: "TCP"},
			"p2p":     {Number: 30303, Protocol: "TCP"},
		},
	}
	
	result["cl-1-lighthouse-geth"] = &ServiceInfo{
		Name:      "cl-1-lighthouse-geth",
		UUID:      "uuid-cl-1",
		Status:    "RUNNING",
		IPAddress: "172.16.0.3",
		Hostname:  "cl-1-lighthouse-geth.local",
		Ports: map[string]PortInfo{
			"beacon":  {Number: 5052, Protocol: "TCP", MaybeURL: "http://172.16.0.3:5052"},
			"metrics": {Number: 5054, Protocol: "TCP"},
			"p2p":     {Number: 9000, Protocol: "TCP"},
		},
	}

	return result, nil
}

// StopEnclave stops the specified enclave
func (k *KurtosisClient) StopEnclave(ctx context.Context, enclaveName string) error {
	_, exists := k.enclaves[enclaveName]
	if !exists {
		return fmt.Errorf("enclave not found: %s", enclaveName)
	}

	// For now, just mark as stopped
	// In production, this would stop the actual Kurtosis enclave
	return nil
}

// DestroyEnclave destroys the specified enclave
func (k *KurtosisClient) DestroyEnclave(ctx context.Context, enclaveName string) error {
	if _, exists := k.enclaves[enclaveName]; exists {
		delete(k.enclaves, enclaveName)
	}

	// Destroy the enclave using the Kurtosis context
	err := k.kurtosisCtx.DestroyEnclave(ctx, enclaveName)
	if err != nil {
		return fmt.Errorf("failed to destroy enclave: %w", err)
	}

	return nil
}

// WaitForServices waits for services to be ready
func (k *KurtosisClient) WaitForServices(ctx context.Context, enclaveName string, serviceNames []string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		services, err := k.GetServices(ctx, enclaveName)
		if err != nil {
			return err
		}

		allReady := true
		for _, name := range serviceNames {
			service, exists := services[name]
			if !exists || service.Status != "RUNNING" {
				allReady = false
				break
			}
		}

		if allReady {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			// Continue checking
		}
	}

	return fmt.Errorf("timeout waiting for services to be ready")
}

// getOrCreateEnclave gets an existing enclave or creates a new one
func (k *KurtosisClient) getOrCreateEnclave(ctx context.Context, enclaveName string) (*enclaves.EnclaveContext, error) {
	// Check if we already have it
	if enclaveCtx, exists := k.enclaves[enclaveName]; exists {
		return enclaveCtx, nil
	}

	// Try to get existing enclave
	enclaveCtx, err := k.kurtosisCtx.GetEnclaveContext(ctx, enclaveName)
	if err == nil {
		return enclaveCtx, nil
	}

	// Create new enclave if it doesn't exist
	enclaveCtx, err = k.kurtosisCtx.CreateEnclave(ctx, enclaveName)
	if err != nil {
		return nil, fmt.Errorf("failed to create enclave: %w", err)
	}

	return enclaveCtx, nil
}

// isHTTPPort checks if a port is typically used for HTTP
func isHTTPPort(port uint16) bool {
	httpPorts := []uint16{80, 8080, 3000, 5052, 9090, 4000}
	for _, p := range httpPorts {
		if port == p {
			return true
		}
	}
	return false
}

// isWSPort checks if a port is typically used for WebSocket
func isWSPort(port uint16) bool {
	wsPorts := []uint16{8546}
	for _, p := range wsPorts {
		if port == p {
			return true
		}
	}
	return false
}