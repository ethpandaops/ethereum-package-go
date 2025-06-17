package config

import (
	"fmt"

	"github.com/ethpandaops/ethereum-package-go/pkg/types"
)

// ConfigBuilder helps build ethereum-package configurations
type ConfigBuilder struct {
	config *types.EthereumPackageConfig
}

// NewConfigBuilder creates a new configuration builder
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: &types.EthereumPackageConfig{
			Participants:       []types.ParticipantConfig{},
			AdditionalServices: []types.AdditionalService{},
		},
	}
}

// WithParticipant adds a participant to the configuration
func (b *ConfigBuilder) WithParticipant(participant types.ParticipantConfig) *ConfigBuilder {
	b.config.Participants = append(b.config.Participants, participant)
	return b
}

// WithParticipants sets all participants at once
func (b *ConfigBuilder) WithParticipants(participants []types.ParticipantConfig) *ConfigBuilder {
	b.config.Participants = participants
	return b
}

// WithNetworkParams sets network parameters
func (b *ConfigBuilder) WithNetworkParams(params *types.NetworkParams) *ConfigBuilder {
	b.config.NetworkParams = params
	return b
}

// WithChainID sets the chain ID
func (b *ConfigBuilder) WithChainID(chainID uint64) *ConfigBuilder {
	if b.config.NetworkParams == nil {
		b.config.NetworkParams = &types.NetworkParams{}
	}
	b.config.NetworkParams.ChainID = chainID
	b.config.NetworkParams.NetworkID = chainID // Often the same
	return b
}

// WithMEV enables MEV configuration
func (b *ConfigBuilder) WithMEV(mevConfig *types.MEVConfig) *ConfigBuilder {
	b.config.MEV = mevConfig
	return b
}

// WithAdditionalService adds an additional service
func (b *ConfigBuilder) WithAdditionalService(service types.AdditionalService) *ConfigBuilder {
	b.config.AdditionalServices = append(b.config.AdditionalServices, service)
	return b
}

// WithGlobalLogLevel sets the global client log level
func (b *ConfigBuilder) WithGlobalLogLevel(level string) *ConfigBuilder {
	b.config.GlobalClientLogLevel = level
	return b
}

// Build returns the built configuration
func (b *ConfigBuilder) Build() (*types.EthereumPackageConfig, error) {
	// Validate configuration
	if err := b.validate(); err != nil {
		return nil, err
	}

	// Return a copy to prevent further modifications
	config := *b.config
	return &config, nil
}

// validate checks if the configuration is valid
func (b *ConfigBuilder) validate() error {
	if len(b.config.Participants) == 0 {
		return fmt.Errorf("at least one participant is required")
	}

	for i, p := range b.config.Participants {
		if p.ELType == "" {
			return fmt.Errorf("participant %d: execution layer type is required", i)
		}
		if p.CLType == "" {
			return fmt.Errorf("participant %d: consensus layer type is required", i)
		}
		if p.Count <= 0 {
			b.config.Participants[i].Count = 1 // Default to 1 if not specified
		}
	}

	return nil
}

// SimpleParticipantBuilder helps build participant configurations
type SimpleParticipantBuilder struct {
	participant types.ParticipantConfig
}

// NewParticipantBuilder creates a new participant builder
func NewParticipantBuilder() *SimpleParticipantBuilder {
	return &SimpleParticipantBuilder{
		participant: types.ParticipantConfig{
			Count: 1, // Default
		},
	}
}

// WithEL sets the execution layer client
func (p *SimpleParticipantBuilder) WithEL(clientType types.ClientType) *SimpleParticipantBuilder {
	p.participant.ELType = clientType
	return p
}

// WithCL sets the consensus layer client
func (p *SimpleParticipantBuilder) WithCL(clientType types.ClientType) *SimpleParticipantBuilder {
	p.participant.CLType = clientType
	return p
}

// WithELVersion sets the execution layer version
func (p *SimpleParticipantBuilder) WithELVersion(version string) *SimpleParticipantBuilder {
	p.participant.ELVersion = version
	return p
}

// WithCLVersion sets the consensus layer version
func (p *SimpleParticipantBuilder) WithCLVersion(version string) *SimpleParticipantBuilder {
	p.participant.CLVersion = version
	return p
}

// WithCount sets the number of nodes
func (p *SimpleParticipantBuilder) WithCount(count int) *SimpleParticipantBuilder {
	p.participant.Count = count
	return p
}

// WithValidatorCount sets the number of validators
func (p *SimpleParticipantBuilder) WithValidatorCount(count int) *SimpleParticipantBuilder {
	p.participant.ValidatorCount = count
	return p
}

// Build returns the built participant configuration
func (p *SimpleParticipantBuilder) Build() types.ParticipantConfig {
	return p.participant
}