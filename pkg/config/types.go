package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ethpandaops/ethereum-package-go/pkg/client"
)

// Config errors
var (
	ErrInvalidPreset   = errors.New("invalid preset")
	ErrEmptyConfigPath = errors.New("config path is empty")
	ErrNilConfig       = errors.New("config is nil")
)

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
	ELType client.Type `yaml:"el_type,omitempty"`
	CLType client.Type `yaml:"cl_type,omitempty"`

	// Version overrides
	ELVersion string `yaml:"el_version,omitempty"`
	CLVersion string `yaml:"cl_version,omitempty"`

	// Node count
	Count int `yaml:"count,omitempty"`

	// Validator configuration
	ValidatorCount int `yaml:"validator_count,omitempty"`
}

// Validate validates the participant configuration
func (p *ParticipantConfig) Validate(index int) error {
	if p.ELType == "" {
		return fmt.Errorf("participant %d: execution layer type is required", index)
	}
	if p.CLType == "" {
		return fmt.Errorf("participant %d: consensus layer type is required", index)
	}

	if !p.ELType.IsExecution() {
		return fmt.Errorf("participant %d: invalid execution client type: %s", index, p.ELType)
	}

	if !p.CLType.IsConsensus() {
		return fmt.Errorf("participant %d: invalid consensus client type: %s", index, p.CLType)
	}

	if p.Count < 0 {
		return fmt.Errorf("participant %d: count cannot be negative", index)
	}
	if p.Count > 100 {
		return fmt.Errorf("participant %d: count %d exceeds maximum of 100", index, p.Count)
	}

	if p.ValidatorCount < 0 {
		return fmt.Errorf("participant %d: validator count cannot be negative", index)
	}
	if p.ValidatorCount > 1000000 {
		return fmt.Errorf("participant %d: validator count cannot exceed 1000000", index)
	}

	return nil
}

// ApplyDefaults applies default values to the participant configuration
func (p *ParticipantConfig) ApplyDefaults() {
	if p.Count == 0 {
		p.Count = 1
	}
}

// NetworkParams represents network-wide parameters
type NetworkParams struct {
	Network                     string `yaml:"network,omitempty"`
	NetworkID                   string `yaml:"network_id,omitempty"`
	DepositContractAddress      string `yaml:"deposit_contract_address,omitempty"`
	SecondsPerSlot              int    `yaml:"seconds_per_slot,omitempty"`
	NumValidatorKeysPerNode     int    `yaml:"num_validator_keys_per_node,omitempty"`
	PreregisteredValidatorCount int    `yaml:"preregistered_validator_count,omitempty"`
	GenesisDelay                int    `yaml:"genesis_delay,omitempty"`
	GenesisGasLimit             uint64 `yaml:"genesis_gaslimit,omitempty"`
	AltairForkEpoch             int    `yaml:"altair_fork_epoch,omitempty"`
	BellatrixForkEpoch          int    `yaml:"bellatrix_fork_epoch,omitempty"`
	CapellaForkEpoch            int    `yaml:"capella_fork_epoch,omitempty"`
	DenebForkEpoch              int    `yaml:"deneb_fork_epoch,omitempty"`
	ElectraForkEpoch            int    `yaml:"electra_fork_epoch,omitempty"`
	FuluForkEpoch               int    `yaml:"fulu_fork_epoch,omitempty"`
}

// Validate validates the network parameters
func (n *NetworkParams) Validate() error {
	if n.SecondsPerSlot < 1 || n.SecondsPerSlot > 60 {
		return fmt.Errorf("seconds per slot must be between 1 and 60, got %d", n.SecondsPerSlot)
	}

	if n.NumValidatorKeysPerNode < 0 || n.NumValidatorKeysPerNode > 1000000 {
		return fmt.Errorf("num validator keys per node must be between 0 and 1000000, got %d", n.NumValidatorKeysPerNode)
	}

	if n.GenesisDelay < 0 {
		return fmt.Errorf("genesis delay cannot be negative")
	}

	// Validate fork epochs ordering
	if n.AltairForkEpoch < 0 || n.BellatrixForkEpoch < 0 || n.CapellaForkEpoch < 0 || 
		n.DenebForkEpoch < 0 || n.ElectraForkEpoch < 0 || n.FuluForkEpoch < 0 {
		return fmt.Errorf("fork epochs cannot be negative")
	}

	// Fork epochs should be in order
	forkEpochs := []int{n.AltairForkEpoch, n.BellatrixForkEpoch, n.CapellaForkEpoch, 
		n.DenebForkEpoch, n.ElectraForkEpoch, n.FuluForkEpoch}
	for i := 1; i < len(forkEpochs); i++ {
		if forkEpochs[i] != 0 && forkEpochs[i] < forkEpochs[i-1] {
			return fmt.Errorf("fork epochs must be in chronological order")
		}
	}

	return nil
}

// ApplyDefaults applies default values to network parameters
func (n *NetworkParams) ApplyDefaults() {
	if n.Network == "" {
		n.Network = "kurtosis"
	}
	if n.NetworkID == "" {
		n.NetworkID = "3151908"
	}
	if n.DepositContractAddress == "" {
		n.DepositContractAddress = "0x00000000219ab540356cBB839Cbe05303d7705Fa"
	}
	if n.SecondsPerSlot == 0 {
		n.SecondsPerSlot = 12
	}
	if n.NumValidatorKeysPerNode == 0 {
		n.NumValidatorKeysPerNode = 64
	}
	if n.GenesisDelay == 0 {
		n.GenesisDelay = 20
	}
	if n.GenesisGasLimit == 0 {
		n.GenesisGasLimit = 60000000
	}
}

// MEVConfig represents MEV-boost configuration
type MEVConfig struct {
	Type            string `yaml:"type,omitempty"`
	RelayURL        string `yaml:"relay_url,omitempty"`
	MinBidEth       string `yaml:"min_bid_eth,omitempty"`
	MaxBundleLength int    `yaml:"max_bundle_length,omitempty"`
}

// Validate validates the MEV configuration
func (m *MEVConfig) Validate() error {
	validTypes := map[string]bool{
		"full": true,
		"mock": true,
		"none": true,
		"":     true, // Empty is valid
	}

	if !validTypes[m.Type] {
		return fmt.Errorf("invalid MEV type: %s, must be one of: full, mock, none", m.Type)
	}

	if m.RelayURL != "" && !strings.HasPrefix(m.RelayURL, "http://") && !strings.HasPrefix(m.RelayURL, "https://") {
		return fmt.Errorf("invalid relay URL: %s, must start with http:// or https://", m.RelayURL)
	}

	if m.MaxBundleLength < 0 {
		return fmt.Errorf("max bundle length cannot be negative")
	}

	if m.MaxBundleLength > 10000 {
		return fmt.Errorf("max bundle length %d exceeds maximum of 10000", m.MaxBundleLength)
	}

	return nil
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
	GlobalLogLevel string `yaml:"global_log_level,omitempty"`
}

// Validate validates the EthereumPackageConfig
func (c *EthereumPackageConfig) Validate() error {
	if c == nil {
		return fmt.Errorf("configuration is nil")
	}

	if len(c.Participants) == 0 {
		return fmt.Errorf("at least one participant is required")
	}

	// Validate each participant
	for i, p := range c.Participants {
		if err := p.Validate(i); err != nil {
			return err
		}
	}

	// Validate network params
	if c.NetworkParams != nil {
		if err := c.NetworkParams.Validate(); err != nil {
			return err
		}
	}

	// Validate MEV config
	if c.MEV != nil {
		if err := c.MEV.Validate(); err != nil {
			return err
		}
	}

	// Validate additional services
	serviceNames := make(map[string]bool)
	for i, service := range c.AdditionalServices {
		if service.Name == "" {
			return fmt.Errorf("additional service %d: name is required", i)
		}
		if serviceNames[service.Name] {
			return fmt.Errorf("duplicate additional service: %s", service.Name)
		}
		serviceNames[service.Name] = true

		// Validate known service names
		validServices := map[string]bool{
			"prometheus": true,
			"grafana":    true,
			"dora":       true,
			"spamoor":    true,
			"blockscout": true,
		}
		if !validServices[service.Name] {
			return fmt.Errorf("invalid additional service name: %s", service.Name)
		}
	}

	// Validate global log level
	if c.GlobalLogLevel != "" && !isValidLogLevel(c.GlobalLogLevel) {
		return fmt.Errorf("invalid global log level: %s, must be one of: debug, info, warn, error, fatal", c.GlobalLogLevel)
	}

	return nil
}

// ApplyDefaults applies default values to the configuration
func (c *EthereumPackageConfig) ApplyDefaults() {
	if c == nil {
		return
	}

	// Apply defaults to participants
	for i := range c.Participants {
		c.Participants[i].ApplyDefaults()
	}

	// Apply defaults to network params
	if c.NetworkParams != nil {
		c.NetworkParams.ApplyDefaults()
	}
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

// Helper functions for validation

func isValidLogLevel(level string) bool {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}
	return validLevels[strings.ToLower(level)]
}
