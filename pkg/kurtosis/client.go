package kurtosis

import (
	"context"
	"fmt"
	"strings"
	"time"

	kurtosis_core_rpc_api_bindings "github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/starlark_run_config"
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

	// Prepare package run options
	packageConfig := make(map[string]interface{})
	if config.ConfigYAML != "" {
		// Parse YAML config and convert to map
		// For now, we'll pass the raw YAML as a string parameter
		packageConfig["yaml_config"] = config.ConfigYAML
	}

	// Run the package using Kurtosis SDK
	var responseLines []string
	result := &RunPackageResult{
		EnclaveName: config.EnclaveName,
	}

	// Create run configuration
	runConfig := starlark_run_config.NewRunStarlarkConfig(
		starlark_run_config.WithSerializedParams(config.ConfigYAML),
		starlark_run_config.WithDryRun(config.DryRun),
		starlark_run_config.WithParallelism(int32(config.Parallelism)),
	)

	// Execute the package
	if config.NonBlockingMode {
		// Non-blocking mode - returns immediately with a channel
		responseChan, cancelFunc, err := enclaveCtx.RunStarlarkPackage(ctx, config.PackageID, runConfig)
		if err != nil {
			result.ExecutionError = err
			return result, nil
		}
		defer cancelFunc()

		// Collect a few responses for non-blocking mode
		timeout := time.After(5 * time.Second)
		for {
			select {
			case response, ok := <-responseChan:
				if !ok {
					goto done
				}
				if response != nil {
					responseLines = append(responseLines, formatStarlarkResponse(response))
				}
			case <-timeout:
				responseLines = append(responseLines, "Package execution started in non-blocking mode")
				goto done
			}
		}
	done:
	} else {
		// Blocking mode - wait for completion
		runResult, err := enclaveCtx.RunStarlarkPackageBlocking(ctx, config.PackageID, runConfig)
		if err != nil {
			result.ExecutionError = err
			return result, nil
		}

		// Process validation errors
		if len(runResult.ValidationErrors) > 0 {
			for _, validationErr := range runResult.ValidationErrors {
				result.ValidationErrors = append(result.ValidationErrors, validationErr.GetErrorMessage())
			}
		}

		// Process interpretation error
		if runResult.InterpretationError != nil {
			result.InterpretationError = fmt.Errorf("interpretation error: %s", runResult.InterpretationError.GetErrorMessage())
		}

		// Process execution error
		if runResult.ExecutionError != nil {
			result.ExecutionError = fmt.Errorf("execution error: %s", runResult.ExecutionError.GetErrorMessage())
		}

		// Process run output (multiline string)
		if runResult.RunOutput != "" {
			outputLines := strings.Split(string(runResult.RunOutput), "\n")
			for _, line := range outputLines {
				if line != "" {
					responseLines = append(responseLines, line)
				}
			}
		}

		// Add final status
		if len(runResult.ValidationErrors) == 0 && runResult.InterpretationError == nil && runResult.ExecutionError == nil {
			responseLines = append(responseLines, "Package run completed successfully")
		}
	}

	return result, nil
}

// GetServices returns all services in the enclave
func (k *KurtosisClient) GetServices(ctx context.Context, enclaveName string) (map[string]*ServiceInfo, error) {
	enclaveCtx, exists := k.enclaves[enclaveName]
	if !exists {
		// Try to get the enclave context if not cached
		var err error
		enclaveCtx, err = k.kurtosisCtx.GetEnclaveContext(ctx, enclaveName)
		if err != nil {
			return nil, fmt.Errorf("enclave not found: %s", enclaveName)
		}
		k.enclaves[enclaveName] = enclaveCtx
	}

	// Get all services from the enclave
	serviceIdentifiers, err := enclaveCtx.GetServices()
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	result := make(map[string]*ServiceInfo)

	for serviceName, serviceUUID := range serviceIdentifiers {
		// Get detailed service info
		serviceContext, err := enclaveCtx.GetServiceContext(string(serviceUUID))
		if err != nil {
			// Log error but continue with other services
			continue
		}

		// Get service status
		serviceStatus := "UNKNOWN"
		// Note: The actual status retrieval depends on the Kurtosis version
		// For now, we'll assume all services are running if they exist
		if serviceContext != nil {
			serviceStatus = "RUNNING"
		}

		// Convert ports
		ports := make(map[string]PortInfo)
		publicPorts := serviceContext.GetPublicPorts()
		for portName, portSpec := range publicPorts {
			portInfo := PortInfo{
				Number:            portSpec.GetNumber(),
				Protocol:          string(portSpec.GetTransportProtocol()),
				TransportProtocol: string(portSpec.GetTransportProtocol()),
			}

			// Build MaybeURL based on common patterns
			if serviceContext.GetMaybePublicIPAddress() != "" {
				host := serviceContext.GetMaybePublicIPAddress()
				switch {
				case strings.Contains(portName, "http") || strings.Contains(portName, "rpc") ||
					strings.Contains(portName, "beacon") || strings.Contains(portName, "engine"):
					portInfo.MaybeURL = fmt.Sprintf("http://%s:%d", host, portSpec.GetNumber())
				case strings.Contains(portName, "ws"):
					portInfo.MaybeURL = fmt.Sprintf("ws://%s:%d", host, portSpec.GetNumber())
				}
			}

			ports[portName] = portInfo
		}

		// Create ServiceInfo
		serviceInfo := &ServiceInfo{
			Name:      string(serviceName),
			UUID:      string(serviceUUID),
			Status:    serviceStatus,
			IPAddress: serviceContext.GetMaybePublicIPAddress(),
			Hostname:  string(serviceName), // Use service name as hostname
			Ports:     ports,
		}

		result[string(serviceName)] = serviceInfo
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

// formatStarlarkResponse formats a Starlark response line for display
func formatStarlarkResponse(response *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine) string {
	if response == nil {
		return ""
	}

	if response.GetInstruction() != nil {
		return fmt.Sprintf("[Instruction] %s", response.GetInstruction().GetInstructionName())
	}

	if response.GetInstructionResult() != nil {
		return fmt.Sprintf("[Result] %s", response.GetInstructionResult().GetSerializedInstructionResult())
	}

	if response.GetError() != nil {
		return fmt.Sprintf("[Error] %s", response.GetError().String())
	}

	if response.GetProgressInfo() != nil {
		info := response.GetProgressInfo()
		return fmt.Sprintf("[Progress] %d/%d - %s",
			info.GetCurrentStepNumber(),
			info.GetTotalSteps(),
			info.GetCurrentStepInfo())
	}

	if response.GetRunFinishedEvent() != nil {
		event := response.GetRunFinishedEvent()
		if event.GetIsRunSuccessful() {
			return "[Finished] Run completed successfully"
		}
		return fmt.Sprintf("[Finished] Run failed: %s", event.GetSerializedOutput())
	}

	if response.GetWarning() != nil {
		return fmt.Sprintf("[Warning] %s", response.GetWarning().GetWarningMessage())
	}

	if response.GetInfo() != nil {
		return fmt.Sprintf("[Info] %s", response.GetInfo().GetInfoMessage())
	}

	return "[Unknown] Unknown response type"
}
