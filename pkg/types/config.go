package types

// Preset represents a predefined configuration preset
type Preset string

const (
	// PresetAllELs runs all execution layer clients
	PresetAllELs Preset = "all-els"
	// PresetAllCLs runs all consensus layer clients
	PresetAllCLs Preset = "all-cls"
	// PresetAllClientsMatrix runs all client combinations
	PresetAllClientsMatrix Preset = "all-clients-matrix"
	// PresetMinimal runs a minimal setup with one EL and one CL
	PresetMinimal Preset = "minimal"
)

// ParticipantConfig represents configuration for a network participant
type ParticipantConfig struct {
	// Client names
	ELType ClientType `yaml:"el_type,omitempty"`
	CLType ClientType `yaml:"cl_type,omitempty"`

	// Version overrides
	ELVersion string `yaml:"el_version,omitempty"`
	CLVersion string `yaml:"cl_version,omitempty"`

	// Node count
	Count int `yaml:"count,omitempty"`

	// Validator configuration
	ValidatorCount int `yaml:"validator_count,omitempty"`
}

// NetworkParams represents network-wide parameters
type NetworkParams struct {
	ChainID                  uint64 `yaml:"chain_id,omitempty"`
	NetworkID                uint64 `yaml:"network_id,omitempty"`
	SecondsPerSlot           int    `yaml:"seconds_per_slot,omitempty"`
	SlotsPerEpoch            int    `yaml:"slots_per_epoch,omitempty"`
	CapellaForkEpoch         int    `yaml:"capella_fork_epoch,omitempty"`
	DenebForkEpoch           int    `yaml:"deneb_fork_epoch,omitempty"`
	ElectraForkEpoch         int    `yaml:"electra_fork_epoch,omitempty"`
	MinValidatorWithdrawability int `yaml:"min_validator_withdrawability,omitempty"`
}

// MEVConfig represents MEV-boost configuration
type MEVConfig struct {
	Type            string `yaml:"type,omitempty"`
	RelayURL        string `yaml:"relay_url,omitempty"`
	MinBidEth       string `yaml:"min_bid_eth,omitempty"`
	MaxBundleLength int    `yaml:"max_bundle_length,omitempty"`
}

// AdditionalService represents an additional service to run
type AdditionalService struct {
	Name   string                 `yaml:"name"`
	Config map[string]interface{} `yaml:"config,omitempty"`
}

// EthereumPackageConfig represents the full configuration for ethereum-package
type EthereumPackageConfig struct {
	// Participants in the network
	Participants []ParticipantConfig `yaml:"participants,omitempty"`

	// Network parameters
	NetworkParams *NetworkParams `yaml:"network_params,omitempty"`

	// MEV configuration
	MEV *MEVConfig `yaml:"mev_params,omitempty"`

	// Additional services
	AdditionalServices []AdditionalService `yaml:"additional_services,omitempty"`

	// Global client settings
	GlobalClientLogLevel string `yaml:"global_client_log_level,omitempty"`
}

// ConfigSource represents the source of configuration
type ConfigSource interface {
	// Type returns the type of config source
	Type() string
	// Validate checks if the source is valid
	Validate() error
}

// PresetConfigSource uses a predefined preset
type PresetConfigSource struct {
	preset Preset
}

// NewPresetConfigSource creates a config source from a preset
func NewPresetConfigSource(preset Preset) ConfigSource {
	return &PresetConfigSource{preset: preset}
}

func (p *PresetConfigSource) Type() string {
	return "preset"
}

func (p *PresetConfigSource) Validate() error {
	switch p.preset {
	case PresetAllELs, PresetAllCLs, PresetAllClientsMatrix, PresetMinimal:
		return nil
	default:
		return ErrInvalidPreset
	}
}

// GetPreset returns the preset
func (p *PresetConfigSource) GetPreset() Preset {
	return p.preset
}

// FileConfigSource uses a configuration file
type FileConfigSource struct {
	path string
}

// NewFileConfigSource creates a config source from a file path
func NewFileConfigSource(path string) ConfigSource {
	return &FileConfigSource{path: path}
}

func (f *FileConfigSource) Type() string {
	return "file"
}

func (f *FileConfigSource) Validate() error {
	if f.path == "" {
		return ErrEmptyConfigPath
	}
	return nil
}

// GetPath returns the file path
func (f *FileConfigSource) GetPath() string {
	return f.path
}

// InlineConfigSource uses inline configuration
type InlineConfigSource struct {
	config *EthereumPackageConfig
}

// NewInlineConfigSource creates a config source from inline config
func NewInlineConfigSource(config *EthereumPackageConfig) ConfigSource {
	return &InlineConfigSource{config: config}
}

func (i *InlineConfigSource) Type() string {
	return "inline"
}

func (i *InlineConfigSource) Validate() error {
	if i.config == nil {
		return ErrNilConfig
	}
	return nil
}

// GetConfig returns the inline configuration
func (i *InlineConfigSource) GetConfig() *EthereumPackageConfig {
	return i.config
}