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
	DefaultPackageVersion = "3.0.1"
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

	// Initialize Kurtosis client if not provided
	if cfg.KurtosisClient == nil {
		client, err := kurtosis.NewKurtosisClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create Kurtosis client: %w", err)
		}
		cfg.KurtosisClient = client
	}

	// Build ethereum-package configuration
	ethConfig, err := buildEthereumConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build configuration: %w", err)
	}

	// Convert to YAML
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
	_, err = cfg.KurtosisClient.RunPackage(ctx, runConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to run ethereum-package: %w", err)
	}

	// Wait for services to be ready
	if !cfg.DryRun {
		err = cfg.KurtosisClient.WaitForServices(ctx, cfg.EnclaveName, []string{}, cfg.Timeout)
		if err != nil {
			// Cleanup on failure
			_ = cfg.KurtosisClient.DestroyEnclave(ctx, cfg.EnclaveName)
			return nil, fmt.Errorf("services failed to start: %w", err)
		}
	}

	// Discover and map services
	mapper := discovery.NewServiceMapper(cfg.KurtosisClient)
	network, err := mapper.MapToNetwork(ctx, cfg.EnclaveName, ethConfig)
	if err != nil {
		// Cleanup on failure
		_ = cfg.KurtosisClient.DestroyEnclave(ctx, cfg.EnclaveName)
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

	// Wait for genesis if requested
	if cfg.WaitForGenesis && !cfg.DryRun {
		if err := WaitForGenesis(ctx, network); err != nil {
			// Don't cleanup on genesis wait failure - network is already running
			return network, fmt.Errorf("failed to wait for genesis: %w", err)
		}
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
		network, err := mapper.MapToNetwork(ctx, enclaveName, ethConfig)
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
		builder.WithChainID(cfg.ChainID)
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
