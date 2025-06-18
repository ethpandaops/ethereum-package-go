package config

import (
	"github.com/ethpandaops/ethereum-package-go/pkg/client"
)

// GetPresetConfig returns the configuration for a given preset
func GetPresetConfig(preset Preset) (*EthereumPackageConfig, error) {
	switch preset {
	case PresetAllELs:
		return getAllELsConfig(), nil
	case PresetAllCLs:
		return getAllCLsConfig(), nil
	case PresetAllClientsMatrix:
		return getAllClientsMatrixConfig(), nil
	case PresetMinimal:
		return getMinimalConfig(), nil
	default:
		return nil, ErrInvalidPreset
	}
}

// getAllELsConfig returns a configuration with all execution layer clients
func getAllELsConfig() *EthereumPackageConfig {
	return &EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{
				ELType: client.Geth,
				CLType: client.Lighthouse,
				Count:  1,
			},
			{
				ELType: client.Besu,
				CLType: client.Lighthouse,
				Count:  1,
			},
			{
				ELType: client.Nethermind,
				CLType: client.Lighthouse,
				Count:  1,
			},
			{
				ELType: client.Erigon,
				CLType: client.Lighthouse,
				Count:  1,
			},
			{
				ELType: client.Reth,
				CLType: client.Lighthouse,
				Count:  1,
			},
		},
	}
}

// getAllCLsConfig returns a configuration with all consensus layer clients
func getAllCLsConfig() *EthereumPackageConfig {
	return &EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{
				ELType: client.Geth,
				CLType: client.Lighthouse,
				Count:  1,
			},
			{
				ELType: client.Geth,
				CLType: client.Teku,
				Count:  1,
			},
			{
				ELType: client.Geth,
				CLType: client.Prysm,
				Count:  1,
			},
			{
				ELType: client.Geth,
				CLType: client.Nimbus,
				Count:  1,
			},
			{
				ELType: client.Geth,
				CLType: client.Lodestar,
				Count:  1,
			},
			{
				ELType: client.Geth,
				CLType: client.Grandine,
				Count:  1,
			},
		},
	}
}

// getAllClientsMatrixConfig returns a configuration with all client combinations
func getAllClientsMatrixConfig() *EthereumPackageConfig {
	elClients := []client.Type{
		client.Geth,
		client.Besu,
		client.Nethermind,
		client.Erigon,
		client.Reth,
	}

	clClients := []client.Type{
		client.Lighthouse,
		client.Teku,
		client.Prysm,
		client.Nimbus,
		client.Lodestar,
		client.Grandine,
	}

	var participants []ParticipantConfig

	for _, el := range elClients {
		for _, cl := range clClients {
			participants = append(participants, ParticipantConfig{
				ELType: el,
				CLType: cl,
				Count:  1,
			})
		}
	}

	return &EthereumPackageConfig{
		Participants: participants,
	}
}

// getMinimalConfig returns a minimal configuration
func getMinimalConfig() *EthereumPackageConfig {
	return &EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{
				ELType:         client.Geth,
				CLType:         client.Lighthouse,
				Count:          1,
				ValidatorCount: 64,
			},
		},
	}
}

// PresetBuilder helps build configurations from presets
type PresetBuilder struct {
	preset Preset
	config *EthereumPackageConfig
}

// NewPresetBuilder creates a new preset builder
func NewPresetBuilder(preset Preset) (*PresetBuilder, error) {
	config, err := GetPresetConfig(preset)
	if err != nil {
		return nil, err
	}
	return &PresetBuilder{
		preset: preset,
		config: config,
	}, nil
}

// WithNetworkID sets the network ID
func (p *PresetBuilder) WithNetworkID(networkID string) *PresetBuilder {
	if p.config.NetworkParams == nil {
		p.config.NetworkParams = &NetworkParams{}
	}
	p.config.NetworkParams.NetworkID = networkID
	return p
}

// WithNetworkParams sets the network parameters
func (p *PresetBuilder) WithNetworkParams(params *NetworkParams) *PresetBuilder {
	p.config.NetworkParams = params
	return p
}

// WithMEV sets the MEV configuration
func (p *PresetBuilder) WithMEV(mev *MEVConfig) *PresetBuilder {
	p.config.MEV = mev
	return p
}

// WithAdditionalService adds an additional service
func (p *PresetBuilder) WithAdditionalService(service AdditionalService) *PresetBuilder {
	p.config.AdditionalServices = append(p.config.AdditionalServices, service)
	return p
}

// WithGlobalLogLevel sets the global log level
func (p *PresetBuilder) WithGlobalLogLevel(level string) *PresetBuilder {
	p.config.GlobalLogLevel = level
	return p
}

// Build returns the configuration for the preset
func (p *PresetBuilder) Build() (*EthereumPackageConfig, error) {
	// Apply defaults
	p.config.ApplyDefaults()

	// Validate
	if err := p.config.Validate(); err != nil {
		return nil, err
	}

	return p.config, nil
}
