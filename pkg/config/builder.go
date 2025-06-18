package config

import (
	"github.com/ethpandaops/ethereum-package-go/pkg/client"
)

// ConfigBuilder helps build ethereum-package configurations
type ConfigBuilder struct {
	config *EthereumPackageConfig
}

// NewConfigBuilder creates a new configuration builder
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: &EthereumPackageConfig{
			Participants:       []ParticipantConfig{},
			AdditionalServices: []AdditionalService{},
		},
	}
}

// WithParticipant adds a participant to the configuration
func (b *ConfigBuilder) WithParticipant(participant ParticipantConfig) *ConfigBuilder {
	b.config.Participants = append(b.config.Participants, participant)
	return b
}

// WithParticipants sets all participants at once
func (b *ConfigBuilder) WithParticipants(participants []ParticipantConfig) *ConfigBuilder {
	b.config.Participants = participants
	return b
}

// WithNetworkParams sets network parameters
func (b *ConfigBuilder) WithNetworkParams(params *NetworkParams) *ConfigBuilder {
	b.config.NetworkParams = params
	return b
}

// WithChainID sets the chain ID
func (b *ConfigBuilder) WithChainID(chainID uint64) *ConfigBuilder {
	if b.config.NetworkParams == nil {
		b.config.NetworkParams = &NetworkParams{}
	}
	b.config.NetworkParams.ChainID = chainID
	b.config.NetworkParams.NetworkID = chainID // Often the same
	return b
}

// WithMEV enables MEV configuration
func (b *ConfigBuilder) WithMEV(mevConfig *MEVConfig) *ConfigBuilder {
	b.config.MEV = mevConfig
	return b
}

// WithAdditionalService adds an additional service
func (b *ConfigBuilder) WithAdditionalService(service AdditionalService) *ConfigBuilder {
	b.config.AdditionalServices = append(b.config.AdditionalServices, service)
	return b
}

// WithGlobalLogLevel sets the global client log level
func (b *ConfigBuilder) WithGlobalLogLevel(level string) *ConfigBuilder {
	b.config.GlobalClientLogLevel = level
	return b
}

// Build returns the built configuration
func (b *ConfigBuilder) Build() (*EthereumPackageConfig, error) {
	// Apply defaults
	b.config.ApplyDefaults()

	// Validate configuration
	if err := b.config.Validate(); err != nil {
		return nil, err
	}

	// Return a copy to prevent further modifications
	config := *b.config
	return &config, nil
}

// SimpleParticipantBuilder helps build participant configurations
type SimpleParticipantBuilder struct {
	participant ParticipantConfig
}

// NewParticipantBuilder creates a new participant builder
func NewParticipantBuilder() *SimpleParticipantBuilder {
	return &SimpleParticipantBuilder{
		participant: ParticipantConfig{
			Count: 1, // Default
		},
	}
}

// WithEL sets the execution layer client
func (p *SimpleParticipantBuilder) WithEL(clientType client.Type) *SimpleParticipantBuilder {
	p.participant.ELType = clientType
	return p
}

// WithCL sets the consensus layer client
func (p *SimpleParticipantBuilder) WithCL(clientType client.Type) *SimpleParticipantBuilder {
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
func (p *SimpleParticipantBuilder) Build() ParticipantConfig {
	return p.participant
}
