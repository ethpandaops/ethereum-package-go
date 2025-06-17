package ethereum

import (
	"context"
	"fmt"
	"time"

	"github.com/ethpandaops/ethereum-package-go/pkg/config"
	"github.com/ethpandaops/ethereum-package-go/pkg/discovery"
	"github.com/ethpandaops/ethereum-package-go/pkg/kurtosis"
	"github.com/ethpandaops/ethereum-package-go/pkg/types"
)

// RunOption configures how the Ethereum network is started
type RunOption func(*RunConfig)

// RunConfig holds configuration for running an Ethereum network
type RunConfig struct {
	// Package configuration
	PackageID     string
	EnclaveName   string
	ConfigSource  types.ConfigSource
	NetworkParams *types.NetworkParams
	ChainID       uint64

	// MEV configuration
	MEV *types.MEVConfig

	// Additional services
	AdditionalServices []types.AdditionalService

	// Global settings
	GlobalLogLevel string

	// Runtime options
	DryRun       bool
	Parallelism  int
	VerboseMode  bool
	Timeout      time.Duration

	// Dependencies (can be injected for testing)
	KurtosisClient kurtosis.Client
}

// defaultRunConfig returns a RunConfig with sensible defaults
func defaultRunConfig() *RunConfig {
	return &RunConfig{
		PackageID:      "github.com/ethpandaops/ethereum-package",
		EnclaveName:    fmt.Sprintf("ethereum-package-%d", time.Now().Unix()),
		ConfigSource:   types.NewPresetConfigSource(types.PresetMinimal),
		ChainID:        12345,
		DryRun:         false,
		Parallelism:    4,
		VerboseMode:    false,
		Timeout:        10 * time.Minute,
		GlobalLogLevel: "info",
	}
}

// Run starts an Ethereum network and returns a Network interface
func Run(ctx context.Context, opts ...RunOption) (types.Network, error) {
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
	runConfig := kurtosis.RunPackageConfig{
		PackageID:       cfg.PackageID,
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

	return network, nil
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
func buildEthereumConfig(cfg *RunConfig) (*types.EthereumPackageConfig, error) {
	// Get base configuration from source
	var baseConfig *types.EthereumPackageConfig
	var err error

	switch cfg.ConfigSource.Type() {
	case "preset":
		preset := cfg.ConfigSource.(*types.PresetConfigSource)
		baseConfig, err = config.GetPresetConfig(preset.GetPreset())
	case "file":
		file := cfg.ConfigSource.(*types.FileConfigSource)
		yamlContent, readErr := readFile(file.GetPath())
		if readErr != nil {
			return nil, fmt.Errorf("failed to read config file: %w", readErr)
		}
		baseConfig, err = config.FromYAML(yamlContent)
	case "inline":
		inline := cfg.ConfigSource.(*types.InlineConfigSource)
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

// readFile reads a file and returns its contents as a string
func readFile(path string) (string, error) {
	// In a real implementation, this would read from filesystem
	// For now, we'll return an error since we don't implement file reading
	return "", fmt.Errorf("file reading not implemented")
}