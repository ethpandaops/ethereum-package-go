package ethereum

import (
	"time"

	"github.com/ethpandaops/ethereum-package-go/pkg/config"
	"github.com/ethpandaops/ethereum-package-go/pkg/kurtosis"
)

// WithPreset sets a predefined configuration preset
func WithPreset(preset config.Preset) RunOption {
	return func(cfg *RunConfig) {
		cfg.ConfigSource = config.NewPresetConfigSource(preset)
	}
}

// WithConfigFile loads configuration from a YAML file
func WithConfigFile(path string) RunOption {
	return func(cfg *RunConfig) {
		cfg.ConfigSource = config.NewFileConfigSource(path)
	}
}

// WithConfig uses an inline configuration
func WithConfig(cfg *config.EthereumPackageConfig) RunOption {
	return func(rc *RunConfig) {
		rc.ConfigSource = config.NewInlineConfigSource(cfg)
	}
}

// WithChainID sets the chain ID for the network
func WithChainID(chainID uint64) RunOption {
	return func(cfg *RunConfig) {
		cfg.ChainID = chainID
	}
}

// WithNetworkParams sets custom network parameters
func WithNetworkParams(params *config.NetworkParams) RunOption {
	return func(cfg *RunConfig) {
		cfg.NetworkParams = params
	}
}

// WithMEV enables MEV configuration
func WithMEV(mevConfig *config.MEVConfig) RunOption {
	return func(cfg *RunConfig) {
		cfg.MEV = mevConfig
	}
}

// WithAdditionalServices adds additional services to the network
func WithAdditionalServices(services ...string) RunOption {
	return func(cfg *RunConfig) {
		for _, service := range services {
			cfg.AdditionalServices = append(cfg.AdditionalServices, config.AdditionalService{
				Name: service,
			})
		}
	}
}

// WithAdditionalService adds a single additional service with configuration
func WithAdditionalService(service config.AdditionalService) RunOption {
	return func(cfg *RunConfig) {
		cfg.AdditionalServices = append(cfg.AdditionalServices, service)
	}
}

// WithGlobalLogLevel sets the global client log level
func WithGlobalLogLevel(level string) RunOption {
	return func(cfg *RunConfig) {
		cfg.GlobalLogLevel = level
	}
}

// WithEnclaveName sets a custom enclave name
func WithEnclaveName(name string) RunOption {
	return func(cfg *RunConfig) {
		cfg.EnclaveName = name
	}
}

// WithPackageID sets a custom ethereum-package ID
func WithPackageID(packageID string) RunOption {
	return func(cfg *RunConfig) {
		cfg.PackageID = packageID
	}
}

// WithPackageVersion sets a custom ethereum-package version
func WithPackageVersion(version string) RunOption {
	return func(cfg *RunConfig) {
		cfg.PackageVersion = version
	}
}

// WithPackageRepo sets both repository and version for the ethereum-package
func WithPackageRepo(repo, version string) RunOption {
	return func(cfg *RunConfig) {
		cfg.PackageID = repo
		cfg.PackageVersion = version
	}
}

// WithDryRun enables dry run mode (validation only, no actual deployment)
func WithDryRun(dryRun bool) RunOption {
	return func(cfg *RunConfig) {
		cfg.DryRun = dryRun
	}
}

// WithParallelism sets the parallelism level for Kurtosis operations
func WithParallelism(parallelism int) RunOption {
	return func(cfg *RunConfig) {
		cfg.Parallelism = parallelism
	}
}

// WithVerbose enables verbose output
func WithVerbose(verbose bool) RunOption {
	return func(cfg *RunConfig) {
		cfg.VerboseMode = verbose
	}
}

// WithTimeout sets the timeout for network startup
func WithTimeout(timeout time.Duration) RunOption {
	return func(cfg *RunConfig) {
		cfg.Timeout = timeout
	}
}

// WithKurtosisClient injects a custom Kurtosis client (mainly for testing)
func WithKurtosisClient(client kurtosis.Client) RunOption {
	return func(cfg *RunConfig) {
		cfg.KurtosisClient = client
	}
}

// Convenience functions for common configurations

// AllELs returns a preset with all execution layer clients
func AllELs() RunOption {
	return WithPreset(config.PresetAllELs)
}

// AllCLs returns a preset with all consensus layer clients
func AllCLs() RunOption {
	return WithPreset(config.PresetAllCLs)
}

// AllClientsMatrix returns a preset with all client combinations
func AllClientsMatrix() RunOption {
	return WithPreset(config.PresetAllClientsMatrix)
}

// Minimal returns a minimal preset with one EL and one CL
func Minimal() RunOption {
	return WithPreset(config.PresetMinimal)
}

// WithExplorer adds Dora block explorer
func WithExplorer() RunOption {
	return WithAdditionalServices("dora")
}

// WithFullObservability adds all observability tools
func WithFullObservability() RunOption {
	return WithAdditionalServices("prometheus", "grafana", "dora")
}

// WithParticipants sets custom participant configurations
func WithParticipants(participants []config.ParticipantConfig) RunOption {
	return func(cfg *RunConfig) {
		// Create inline config with participants
		ethConfig := &config.EthereumPackageConfig{
			Participants: participants,
		}
		cfg.ConfigSource = config.NewInlineConfigSource(ethConfig)
	}
}

// WithCustomChain creates a custom chain configuration
func WithCustomChain(chainID uint64, secondsPerSlot, slotsPerEpoch int) RunOption {
	return func(cfg *RunConfig) {
		cfg.NetworkParams = &config.NetworkParams{
			ChainID:        chainID,
			NetworkID:      chainID,
			SecondsPerSlot: secondsPerSlot,
			SlotsPerEpoch:  slotsPerEpoch,
		}
	}
}

// WithMEVBoost enables MEV-boost with default configuration
func WithMEVBoost() RunOption {
	return WithMEV(&config.MEVConfig{
		Type: "full",
	})
}

// WithMEVBoostRelay enables MEV-boost with a custom relay
func WithMEVBoostRelay(relayURL string) RunOption {
	return WithMEV(&config.MEVConfig{
		Type:     "full",
		RelayURL: relayURL,
	})
}

// WithWaitForGenesis waits for the network genesis time before returning
func WithWaitForGenesis() RunOption {
	return func(cfg *RunConfig) {
		cfg.WaitForGenesis = true
	}
}

// WithSpamoor adds the spamoor service to the network
func WithSpamoor() RunOption {
	return WithAdditionalServices("spamoor")
}