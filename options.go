package ethereum

import (
	"time"

	"github.com/ethpandaops/ethereum-package-go/pkg/kurtosis"
	"github.com/ethpandaops/ethereum-package-go/pkg/types"
)

// WithPreset sets a predefined configuration preset
func WithPreset(preset types.Preset) RunOption {
	return func(cfg *RunConfig) {
		cfg.ConfigSource = types.NewPresetConfigSource(preset)
	}
}

// WithConfigFile loads configuration from a YAML file
func WithConfigFile(path string) RunOption {
	return func(cfg *RunConfig) {
		cfg.ConfigSource = types.NewFileConfigSource(path)
	}
}

// WithConfig uses an inline configuration
func WithConfig(config *types.EthereumPackageConfig) RunOption {
	return func(cfg *RunConfig) {
		cfg.ConfigSource = types.NewInlineConfigSource(config)
	}
}

// WithChainID sets the chain ID for the network
func WithChainID(chainID uint64) RunOption {
	return func(cfg *RunConfig) {
		cfg.ChainID = chainID
	}
}

// WithNetworkParams sets custom network parameters
func WithNetworkParams(params *types.NetworkParams) RunOption {
	return func(cfg *RunConfig) {
		cfg.NetworkParams = params
	}
}

// WithMEV enables MEV configuration
func WithMEV(mevConfig *types.MEVConfig) RunOption {
	return func(cfg *RunConfig) {
		cfg.MEV = mevConfig
	}
}

// WithAdditionalServices adds additional services to the network
func WithAdditionalServices(services ...string) RunOption {
	return func(cfg *RunConfig) {
		for _, service := range services {
			cfg.AdditionalServices = append(cfg.AdditionalServices, types.AdditionalService{
				Name: service,
			})
		}
	}
}

// WithAdditionalService adds a single additional service with configuration
func WithAdditionalService(service types.AdditionalService) RunOption {
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
	return WithPreset(types.PresetAllELs)
}

// AllCLs returns a preset with all consensus layer clients
func AllCLs() RunOption {
	return WithPreset(types.PresetAllCLs)
}

// AllClientsMatrix returns a preset with all client combinations
func AllClientsMatrix() RunOption {
	return WithPreset(types.PresetAllClientsMatrix)
}

// Minimal returns a minimal preset with one EL and one CL
func Minimal() RunOption {
	return WithPreset(types.PresetMinimal)
}

// WithMonitoring adds Prometheus and Grafana monitoring
func WithMonitoring() RunOption {
	return WithAdditionalServices("prometheus", "grafana")
}

// WithExplorer adds Blockscout block explorer
func WithExplorer() RunOption {
	return WithAdditionalServices("blockscout")
}

// WithFullObservability adds all observability tools
func WithFullObservability() RunOption {
	return WithAdditionalServices("prometheus", "grafana", "blockscout")
}

// WithParticipants sets custom participant configurations
func WithParticipants(participants []types.ParticipantConfig) RunOption {
	return func(cfg *RunConfig) {
		// Create inline config with participants
		ethConfig := &types.EthereumPackageConfig{
			Participants: participants,
		}
		cfg.ConfigSource = types.NewInlineConfigSource(ethConfig)
	}
}

// WithCustomChain creates a custom chain configuration
func WithCustomChain(chainID uint64, secondsPerSlot, slotsPerEpoch int) RunOption {
	return func(cfg *RunConfig) {
		cfg.NetworkParams = &types.NetworkParams{
			ChainID:        chainID,
			NetworkID:      chainID,
			SecondsPerSlot: secondsPerSlot,
			SlotsPerEpoch:  slotsPerEpoch,
		}
	}
}

// WithMEVBoost enables MEV-boost with default configuration
func WithMEVBoost() RunOption {
	return WithMEV(&types.MEVConfig{
		Type: "full",
	})
}

// WithMEVBoostRelay enables MEV-boost with a custom relay
func WithMEVBoostRelay(relayURL string) RunOption {
	return WithMEV(&types.MEVConfig{
		Type:     "full",
		RelayURL: relayURL,
	})
}