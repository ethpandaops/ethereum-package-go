package config

import (
	"strings"
)

// Validator validates ethereum-package configurations
type Validator struct {
	config *EthereumPackageConfig
}

// NewValidator creates a new configuration validator
func NewValidator(config *EthereumPackageConfig) *Validator {
	return &Validator{config: config}
}

// Validate performs full validation of the configuration
func (v *Validator) Validate() error {
	// First use the config's own validation
	if err := v.config.Validate(); err != nil {
		return err
	}

	// Then apply additional, more stringent validations
	// Validate participants with additional constraints
	if err := v.validateParticipants(); err != nil {
		return err
	}

	// Validate network parameters with additional constraints
	if err := v.validateNetworkParams(); err != nil {
		return err
	}

	// Validate MEV configuration with additional constraints
	if err := v.validateMEV(); err != nil {
		return err
	}

	// Validate additional services
	if err := v.validateAdditionalServices(); err != nil {
		return err
	}

	return nil
}

func (v *Validator) validateParticipants() error {
	// Additional constraints beyond basic validation
	// These are already checked in ParticipantConfig.Validate()
	// but we keep this for any additional future checks
	return nil
}

func (v *Validator) validateNetworkParams() error {
	if v.config.NetworkParams == nil {
		return nil // Network params are optional
	}

	// Additional network param validations beyond the basic ones
	// These are already checked in NetworkParams.Validate()
	// but we keep this for any additional future checks
	return nil
}

func (v *Validator) validateMEV() error {
	if v.config.MEV == nil {
		return nil // MEV is optional
	}

	// Additional MEV validations beyond the basic ones
	// These are already checked in MEVConfig.Validate()
	// but we keep this for any additional future checks
	return nil
}

func (v *Validator) validateAdditionalServices() error {
	// Additional service validations beyond the basic ones
	// These are already checked in EthereumPackageConfig.Validate()
	// but we can add extra checks here if needed
	return nil
}

// Helper functions

func isValidMEVType(mevType string) bool {
	validTypes := []string{"none", "mock", "full", "relay", ""}
	for _, valid := range validTypes {
		if mevType == valid {
			return true
		}
	}
	return false
}

func isValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

func isValidServiceName(name string) bool {
	validServices := []string{
		"prometheus",
		"grafana",
		"blockscout",
		"dora",
		"spamoor",
		"el_forkmon",
		"beacon_metrics_gazer",
		"ethereum_metrics_exporter",
		"explorer",
		"forkmon",
	}
	for _, valid := range validServices {
		if name == valid {
			return true
		}
	}
	return false
}