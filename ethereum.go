package ethereum

import (
	"context"
	"fmt"
	"time"

	"github.com/ethpandaops/ethereum-package-go/pkg/config"
	"github.com/ethpandaops/ethereum-package-go/pkg/discovery"
	"github.com/ethpandaops/ethereum-package-go/pkg/kurtosis"
	"github.com/ethpandaops/ethereum-package-go/pkg/network"
)

const (
	// DefaultPackageRepository is the default ethereum-package repository
	DefaultPackageRepository = "github.com/ethpandaops/ethereum-package"
	// DefaultPackageVersion is the pinned version of ethereum-package
	DefaultPackageVersion = "5.0.1"
)

// RunOption configures how the Ethereum network is started
type RunOption func(*RunConfig)

// RunConfig holds configuration for running an Ethereum network
type RunConfig struct {
	// Package configuration
	PackageID      string
	PackageVersion string
	EnclaveName    string
	ConfigSource   config.ConfigSource
	NetworkParams  *config.NetworkParams
	ChainID        uint64

	// MEV configuration
	MEV *config.MEVConfig

	// Additional services
	AdditionalServices []config.AdditionalService

	// Global settings
	GlobalLogLevel string

	// Runtime options
	DryRun         bool
	Parallelism    int
	VerboseMode    bool
	Timeout        time.Duration
	WaitForGenesis bool

	// Lifecycle management
	OrphanOnExit  bool // Don't cleanup enclave when process exits
	ReuseExisting bool // Try to reuse existing enclave

	// Dependencies (can be injected for testing)
	KurtosisClient kurtosis.Client
}

// defaultRunConfig returns a RunConfig with sensible defaults
func defaultRunConfig() *RunConfig {
	return &RunConfig{
		PackageID:      DefaultPackageRepository,
		PackageVersion: DefaultPackageVersion,
		EnclaveName:    generateEnclaveName(),
		ConfigSource:   config.NewPresetConfigSource(config.PresetMinimal),
		ChainID:        12345,
		DryRun:         false,
		Parallelism:    4,
		VerboseMode:    false,
		Timeout:        10 * time.Minute,
		GlobalLogLevel: "info",
		OrphanOnExit:   false, // Auto-cleanup by default (testcontainers style)
		ReuseExisting:  false,
	}
}

// generateEnclaveName creates a unique enclave name to avoid conflicts
func generateEnclaveName() string {
	// Use nanoseconds for more uniqueness and add a random component
	return fmt.Sprintf("ethereum-package-%d", time.Now().UnixNano())
}

// Run starts an Ethereum network and returns a Network interface
func Run(ctx context.Context, opts ...RunOption) (network.Network, error) {
	// Apply configuration
	cfg := defaultRunConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	// Validate configuration
	if err := validateRunConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	fmt.Printf("[ethereum-package-go] Starting network deployment...\n")
	fmt.Printf("[ethereum-package-go] Package: %s\n", cfg.PackageID)
	if cfg.PackageVersion != "" {
		fmt.Printf("[ethereum-package-go] Version: %s\n", cfg.PackageVersion)
	}
	fmt.Printf("[ethereum-package-go] Enclave: %s\n", cfg.EnclaveName)

	// Initialize Kurtosis client if not provided
	if cfg.KurtosisClient == nil {
		fmt.Printf("[ethereum-package-go] Initializing Kurtosis client...\n")
		client, err := kurtosis.NewKurtosisClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create Kurtosis client: %w", err)
		}
		cfg.KurtosisClient = client
		fmt.Printf("[ethereum-package-go] Kurtosis client initialized\n")
	}

	// Build ethereum-package configuration
	fmt.Printf("[ethereum-package-go] Building ethereum-package configuration...\n")
	ethConfig, err := buildEthereumConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build configuration: %w", err)
	}

	// Log configuration details
	if ethConfig.Participants != nil {
		fmt.Printf("[ethereum-package-go] Participants: %d\n", len(ethConfig.Participants))
		for i, p := range ethConfig.Participants {
			fmt.Printf("[ethereum-package-go]   %d: %s/%s (count: %d, validators: %d)\n",
				i, p.ELType, p.CLType, p.Count, p.ValidatorCount)
		}
	}
	if ethConfig.NetworkParams != nil {
		fmt.Printf("[ethereum-package-go] Network ID: %s\n", ethConfig.NetworkParams.NetworkID)
		fmt.Printf("[ethereum-package-go] Validators per node: %d\n", ethConfig.NetworkParams.NumValidatorKeysPerNode)
	}

	// Convert to YAML
	fmt.Printf("[ethereum-package-go] Converting configuration to YAML...\n")
	yamlConfig, err := config.ToYAML(ethConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate YAML configuration: %w", err)
	}

	// Create Kurtosis run configuration
	packageID := cfg.PackageID
	if cfg.PackageVersion != "" {
		packageID = fmt.Sprintf("%s@%s", cfg.PackageID, cfg.PackageVersion)
	}

	runConfig := kurtosis.RunPackageConfig{
		PackageID:       packageID,
		EnclaveName:     cfg.EnclaveName,
		ConfigYAML:      yamlConfig,
		DryRun:          cfg.DryRun,
		Parallelism:     cfg.Parallelism,
		VerboseMode:     cfg.VerboseMode,
		ImageDownload:   true,
		NonBlockingMode: false,
	}

	// Run the package
	fmt.Printf("[ethereum-package-go] Starting ethereum-package deployment...\n")
	fmt.Printf("[ethereum-package-go] This may take several minutes...\n")
	result, err := cfg.KurtosisClient.RunPackage(ctx, runConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to run ethereum-package: %w", err)
	}
	fmt.Printf("[ethereum-package-go] Package deployment completed\n")

	// Check for Kurtosis execution errors even if err is nil
	fmt.Printf("[ethereum-package-go] Checking deployment result...\n")
	if result.ExecutionError != nil {
		fmt.Printf("[ethereum-package-go] ERROR: Execution failed: %v\n", result.ExecutionError)
		return nil, fmt.Errorf("ethereum-package execution error: %w", result.ExecutionError)
	}
	if result.InterpretationError != nil {
		fmt.Printf("[ethereum-package-go] ERROR: Interpretation failed: %v\n", result.InterpretationError)
		return nil, fmt.Errorf("ethereum-package interpretation error: %w", result.InterpretationError)
	}
	if len(result.ValidationErrors) > 0 {
		fmt.Printf("[ethereum-package-go] ERROR: Validation failed: %v\n", result.ValidationErrors)
		return nil, fmt.Errorf("ethereum-package validation errors: %v", result.ValidationErrors)
	}
	fmt.Printf("[ethereum-package-go] Deployment validation passed\n")

	// Wait for services to be ready
	if !cfg.DryRun {
		fmt.Printf("[ethereum-package-go] Waiting for services to be ready (timeout: %v)...\n", cfg.Timeout)
		err = cfg.KurtosisClient.WaitForServices(ctx, cfg.EnclaveName, []string{}, cfg.Timeout)
		if err != nil {
			fmt.Printf("[ethereum-package-go] ERROR: Services failed to start: %v\n", err)
			fmt.Printf("[ethereum-package-go] Cleaning up failed deployment...\n")
			// Cleanup on failure
			_ = cfg.KurtosisClient.DestroyEnclave(ctx, cfg.EnclaveName)
			return nil, fmt.Errorf("services failed to start: %w", err)
		}
		fmt.Printf("[ethereum-package-go] All services are ready\n")
	}

	// Discover and map services
	fmt.Printf("[ethereum-package-go] Discovering and mapping services...\n")
	mapper := discovery.NewServiceMapper(cfg.KurtosisClient)
	network, err := mapper.MapToNetwork(ctx, cfg.EnclaveName, ethConfig, cfg.OrphanOnExit)
	if err != nil {
		fmt.Printf("[ethereum-package-go] ERROR: Failed to discover services: %v\n", err)
		fmt.Printf("[ethereum-package-go] Cleaning up failed deployment...\n")
		// Cleanup on failure
		_ = cfg.KurtosisClient.DestroyEnclave(ctx, cfg.EnclaveName)
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}
	fmt.Printf("[ethereum-package-go] Service discovery completed\n")
	fmt.Printf("[ethereum-package-go] Found %d execution clients\n", len(network.ExecutionClients().All()))
	fmt.Printf("[ethereum-package-go] Found %d consensus clients\n", len(network.ConsensusClients().All()))
	fmt.Printf("[ethereum-package-go] Found %d total services\n", len(network.Services()))

	// Wait for genesis if requested
	if cfg.WaitForGenesis && !cfg.DryRun {
		fmt.Printf("[ethereum-package-go] Waiting for genesis block...\n")
		if err := WaitForGenesis(ctx, network); err != nil {
			fmt.Printf("[ethereum-package-go] WARNING: Failed to wait for genesis: %v\n", err)
			// Don't cleanup on genesis wait failure - network is already running
			return network, fmt.Errorf("failed to wait for genesis: %w", err)
		}
		fmt.Printf("[ethereum-package-go] Genesis block detected\n")
	}

	fmt.Printf("[ethereum-package-go] Network deployment completed successfully!\n")
	fmt.Printf("[ethereum-package-go] Network name: %s\n", network.Name())
	fmt.Printf("[ethereum-package-go] Enclave: %s\n", network.EnclaveName())
	fmt.Printf("[ethereum-package-go] Chain ID: %d\n", network.ChainID())

	if cfg.OrphanOnExit {
		fmt.Printf("[ethereum-package-go] Network will be ORPHANED - manual cleanup required\n")
		fmt.Printf("[ethereum-package-go] Run 'kurtosis enclave rm %s' to clean up manually\n", network.EnclaveName())
	} else {
		fmt.Printf("[ethereum-package-go] Network will auto-cleanup on process exit\n")
	}

	return network, nil
}

// FindOrCreateNetwork finds an existing network by enclave name or creates a new one
// If enclaveName is empty, a new network with a random name will be created
func FindOrCreateNetwork(ctx context.Context, enclaveName string, opts ...RunOption) (network.Network, error) {
	// If no enclave name provided, just create a new network
	if enclaveName == "" {
		return Run(ctx, opts...)
	}

	// Apply configuration with the specified enclave name
	allOpts := append([]RunOption{WithEnclaveName(enclaveName)}, opts...)
	cfg := defaultRunConfig()
	for _, opt := range allOpts {
		opt(cfg)
	}

	// Initialize Kurtosis client if not provided
	if cfg.KurtosisClient == nil {
		client, err := kurtosis.NewKurtosisClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create Kurtosis client: %w", err)
		}
		cfg.KurtosisClient = client
	}

	// Try to get existing services first
	services, err := cfg.KurtosisClient.GetServices(ctx, enclaveName)
	if err == nil && len(services) > 0 {
		// Enclave exists with services, map it to a network
		ethConfig, err := buildEthereumConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to build configuration: %w", err)
		}

		mapper := discovery.NewServiceMapper(cfg.KurtosisClient)
		network, err := mapper.MapToNetwork(ctx, enclaveName, ethConfig, cfg.OrphanOnExit)
		if err != nil {
			return nil, fmt.Errorf("failed to map existing network: %w", err)
		}

		return network, nil
	}

	// Enclave doesn't exist or has no services, create a new network
	return Run(ctx, allOpts...)
}

// validateRunConfig validates the run configuration
func validateRunConfig(cfg *RunConfig) error {
	if cfg.PackageID == "" {
		return fmt.Errorf("package ID is required")
	}
	if cfg.EnclaveName == "" {
		return fmt.Errorf("enclave name is required")
	}
	if cfg.ConfigSource == nil {
		return fmt.Errorf("config source is required")
	}
	if err := cfg.ConfigSource.Validate(); err != nil {
		return fmt.Errorf("invalid config source: %w", err)
	}
	if cfg.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	return nil
}

// buildEthereumConfig builds the ethereum-package configuration from RunConfig
func buildEthereumConfig(cfg *RunConfig) (*config.EthereumPackageConfig, error) {
	// Get base configuration from source
	var baseConfig *config.EthereumPackageConfig
	var err error

	switch cfg.ConfigSource.Type() {
	case "preset":
		preset := cfg.ConfigSource.(*config.PresetConfigSource)
		baseConfig, err = config.GetPresetConfig(preset.GetPreset())
	case "inline":
		inline := cfg.ConfigSource.(*config.InlineConfigSource)
		baseConfig = inline.GetConfig()
	default:
		return nil, fmt.Errorf("unsupported config source type: %s", cfg.ConfigSource.Type())
	}

	if err != nil {
		return nil, err
	}

	// Apply overrides using ConfigBuilder
	builder := config.NewConfigBuilder().WithParticipants(baseConfig.Participants)

	// Apply network parameters
	if cfg.NetworkParams != nil {
		builder.WithNetworkParams(cfg.NetworkParams)
	} else if cfg.ChainID != 0 {
		builder.WithNetworkID(fmt.Sprintf("%d", cfg.ChainID))
	}

	// Apply MEV configuration
	if cfg.MEV != nil {
		builder.WithMEV(cfg.MEV)
	}

	// Apply additional services
	for _, service := range cfg.AdditionalServices {
		builder.WithAdditionalService(service)
	}

	// Apply global log level
	if cfg.GlobalLogLevel != "" {
		builder.WithGlobalLogLevel(cfg.GlobalLogLevel)
	}

	return builder.Build()
}
