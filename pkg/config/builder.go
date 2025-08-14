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

func NewConfigBuilderFromConfig(config *EthereumPackageConfig) *ConfigBuilder {
	return &ConfigBuilder{
		config: config,
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

// WithNetworkID sets the network ID
func (b *ConfigBuilder) WithNetworkID(networkID string) *ConfigBuilder {
	if b.config.NetworkParams == nil {
		b.config.NetworkParams = &NetworkParams{}
	}
	b.config.NetworkParams.NetworkID = networkID
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

// WithGlobalLogLevel sets the global log level
func (b *ConfigBuilder) WithGlobalLogLevel(level string) *ConfigBuilder {
	b.config.GlobalLogLevel = level
	return b
}

// WithPortPublisher sets the port publisher configuration.
func (b *ConfigBuilder) WithPortPublisher(portPublisher *PortPublisherConfig) *ConfigBuilder {
	b.config.PortPublisher = portPublisher

	return b
}

// WithDockerCacheParams sets the Docker cache configuration.
func (b *ConfigBuilder) WithDockerCacheParams(dockerCache *DockerCacheParams) *ConfigBuilder {
	b.config.DockerCacheParams = dockerCache

	return b
}

// WithExtraFile adds a single extra file to the configuration
func (b *ConfigBuilder) WithExtraFile(name, content string) *ConfigBuilder {
	if b.config.ExtraFiles == nil {
		b.config.ExtraFiles = make(map[string]string)
	}
	b.config.ExtraFiles[name] = content
	return b
}

// WithExtraFiles sets all extra files at once
func (b *ConfigBuilder) WithExtraFiles(files map[string]string) *ConfigBuilder {
	b.config.ExtraFiles = files
	return b
}

// WithEthereumMetricsExporterEnabled sets the ethereum metrics exporter enabled
func (b *ConfigBuilder) WithEthereumMetricsExporterEnabled(enabled bool) *ConfigBuilder {
	b.config.EthereumMetricsExporterEnabled = &enabled
	return b
}

// WithPersistent sets the persistent configuration
func (b *ConfigBuilder) WithPersistent(persistent bool) *ConfigBuilder {
	b.config.Persistent = persistent
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
func (p *SimpleParticipantBuilder) WithELImage(image string) *SimpleParticipantBuilder {
	p.participant.ELImage = &image
	return p
}

// WithCLVersion sets the consensus layer version
func (p *SimpleParticipantBuilder) WithCLImage(image string) *SimpleParticipantBuilder {
	p.participant.CLImage = &image
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

// WithELExtraParams sets the execution layer extra parameters
func (p *SimpleParticipantBuilder) WithELExtraParams(params []string) *SimpleParticipantBuilder {
	p.participant.ELExtraParams = params
	return p
}

// WithELExtraMounts sets the execution layer extra mounts
func (p *SimpleParticipantBuilder) WithELExtraMounts(mounts map[string]string) *SimpleParticipantBuilder {
	p.participant.ELExtraMounts = mounts
	return p
}

// WithELExtraEnvVars sets the execution layer extra environment variables
func (p *SimpleParticipantBuilder) WithELExtraEnvVars(envVars map[string]string) *SimpleParticipantBuilder {
	p.participant.ELExtraEnvVars = envVars
	return p
}

// WithELExtraLabels sets the execution layer extra labels
func (p *SimpleParticipantBuilder) WithELExtraLabels(labels map[string]string) *SimpleParticipantBuilder {
	p.participant.ELExtraLabels = labels
	return p
}

// WithCLExtraParams sets the consensus layer extra parameters
func (p *SimpleParticipantBuilder) WithCLExtraParams(params []string) *SimpleParticipantBuilder {
	p.participant.CLExtraParams = params
	return p
}

// WithCLExtraMounts sets the consensus layer extra mounts
func (p *SimpleParticipantBuilder) WithCLExtraMounts(mounts map[string]string) *SimpleParticipantBuilder {
	p.participant.CLExtraMounts = mounts
	return p
}

// WithCLExtraEnvVars sets the consensus layer extra environment variables
func (p *SimpleParticipantBuilder) WithCLExtraEnvVars(envVars map[string]string) *SimpleParticipantBuilder {
	p.participant.CLExtraEnvVars = envVars
	return p
}

// WithCLExtraLabels sets the consensus layer extra labels
func (p *SimpleParticipantBuilder) WithCLExtraLabels(labels map[string]string) *SimpleParticipantBuilder {
	p.participant.CLExtraLabels = labels
	return p
}

// WithVCExtraParams sets the validator client extra parameters
func (p *SimpleParticipantBuilder) WithVCExtraParams(params []string) *SimpleParticipantBuilder {
	p.participant.VCExtraParams = params
	return p
}

// WithVCExtraMounts sets the validator client extra mounts
func (p *SimpleParticipantBuilder) WithVCExtraMounts(mounts map[string]string) *SimpleParticipantBuilder {
	p.participant.VCExtraMounts = mounts
	return p
}

// WithVCExtraEnvVars sets the validator client extra environment variables
func (p *SimpleParticipantBuilder) WithVCExtraEnvVars(envVars map[string]string) *SimpleParticipantBuilder {
	p.participant.VCExtraEnvVars = envVars
	return p
}

// WithVCExtraLabels sets the validator client extra labels
func (p *SimpleParticipantBuilder) WithVCExtraLabels(labels map[string]string) *SimpleParticipantBuilder {
	p.participant.VCExtraLabels = labels
	return p
}

// Build returns the built participant configuration
func (p *SimpleParticipantBuilder) Build() ParticipantConfig {
	return p.participant
}
