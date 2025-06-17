package config

import (
	"github.com/ethpandaops/ethereum-package-go/pkg/types"
)

// GetPresetConfig returns the configuration for a given preset
func GetPresetConfig(preset types.Preset) (*types.EthereumPackageConfig, error) {
	switch preset {
	case types.PresetAllELs:
		return getAllELsConfig(), nil
	case types.PresetAllCLs:
		return getAllCLsConfig(), nil
	case types.PresetAllClientsMatrix:
		return getAllClientsMatrixConfig(), nil
	case types.PresetMinimal:
		return getMinimalConfig(), nil
	default:
		return nil, types.ErrInvalidPreset
	}
}

// getAllELsConfig returns a configuration with all execution layer clients
func getAllELsConfig() *types.EthereumPackageConfig {
	return &types.EthereumPackageConfig{
		Participants: []types.ParticipantConfig{
			{
				ELType: types.ClientGeth,
				CLType: types.ClientLighthouse,
				Count:  1,
			},
			{
				ELType: types.ClientBesu,
				CLType: types.ClientLighthouse,
				Count:  1,
			},
			{
				ELType: types.ClientNethermind,
				CLType: types.ClientLighthouse,
				Count:  1,
			},
			{
				ELType: types.ClientErigon,
				CLType: types.ClientLighthouse,
				Count:  1,
			},
			{
				ELType: types.ClientReth,
				CLType: types.ClientLighthouse,
				Count:  1,
			},
		},
	}
}

// getAllCLsConfig returns a configuration with all consensus layer clients
func getAllCLsConfig() *types.EthereumPackageConfig {
	return &types.EthereumPackageConfig{
		Participants: []types.ParticipantConfig{
			{
				ELType: types.ClientGeth,
				CLType: types.ClientLighthouse,
				Count:  1,
			},
			{
				ELType: types.ClientGeth,
				CLType: types.ClientTeku,
				Count:  1,
			},
			{
				ELType: types.ClientGeth,
				CLType: types.ClientPrysm,
				Count:  1,
			},
			{
				ELType: types.ClientGeth,
				CLType: types.ClientNimbus,
				Count:  1,
			},
			{
				ELType: types.ClientGeth,
				CLType: types.ClientLodestar,
				Count:  1,
			},
			{
				ELType: types.ClientGeth,
				CLType: types.ClientGrandine,
				Count:  1,
			},
		},
	}
}

// getAllClientsMatrixConfig returns a configuration with all client combinations
func getAllClientsMatrixConfig() *types.EthereumPackageConfig {
	elClients := []types.ClientType{
		types.ClientGeth,
		types.ClientBesu,
		types.ClientNethermind,
		types.ClientErigon,
		types.ClientReth,
	}

	clClients := []types.ClientType{
		types.ClientLighthouse,
		types.ClientTeku,
		types.ClientPrysm,
		types.ClientNimbus,
		types.ClientLodestar,
		types.ClientGrandine,
	}

	var participants []types.ParticipantConfig

	// Create a matrix of all combinations
	for _, el := range elClients {
		for _, cl := range clClients {
			participants = append(participants, types.ParticipantConfig{
				ELType: el,
				CLType: cl,
				Count:  1,
			})
		}
	}

	return &types.EthereumPackageConfig{
		Participants: participants,
	}
}

// getMinimalConfig returns a minimal configuration with one node
func getMinimalConfig() *types.EthereumPackageConfig {
	return &types.EthereumPackageConfig{
		Participants: []types.ParticipantConfig{
			{
				ELType:         types.ClientGeth,
				CLType:         types.ClientLighthouse,
				Count:          1,
				ValidatorCount: 32,
			},
		},
	}
}

// PresetBuilder helps build configurations based on presets with customizations
type PresetBuilder struct {
	baseConfig *types.EthereumPackageConfig
	builder    *ConfigBuilder
}

// NewPresetBuilder creates a new preset builder
func NewPresetBuilder(preset types.Preset) (*PresetBuilder, error) {
	baseConfig, err := GetPresetConfig(preset)
	if err != nil {
		return nil, err
	}

	builder := NewConfigBuilder()
	builder.WithParticipants(baseConfig.Participants)

	return &PresetBuilder{
		baseConfig: baseConfig,
		builder:    builder,
	}, nil
}

// WithChainID sets a custom chain ID
func (p *PresetBuilder) WithChainID(chainID uint64) *PresetBuilder {
	p.builder.WithChainID(chainID)
	return p
}

// WithNetworkParams sets custom network parameters
func (p *PresetBuilder) WithNetworkParams(params *types.NetworkParams) *PresetBuilder {
	p.builder.WithNetworkParams(params)
	return p
}

// WithMEV enables MEV configuration
func (p *PresetBuilder) WithMEV(mevConfig *types.MEVConfig) *PresetBuilder {
	p.builder.WithMEV(mevConfig)
	return p
}

// WithAdditionalService adds an additional service
func (p *PresetBuilder) WithAdditionalService(service types.AdditionalService) *PresetBuilder {
	p.builder.WithAdditionalService(service)
	return p
}

// WithGlobalLogLevel sets the global client log level
func (p *PresetBuilder) WithGlobalLogLevel(level string) *PresetBuilder {
	p.builder.WithGlobalLogLevel(level)
	return p
}

// Build returns the built configuration
func (p *PresetBuilder) Build() (*types.EthereumPackageConfig, error) {
	return p.builder.Build()
}