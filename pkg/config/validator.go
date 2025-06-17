package config

import (
	"fmt"
	"strings"

	"github.com/ethpandaops/ethereum-package-go/pkg/types"
)

// Validator validates ethereum-package configurations
type Validator struct {
	config *types.EthereumPackageConfig
}

// NewValidator creates a new configuration validator
func NewValidator(config *types.EthereumPackageConfig) *Validator {
	return &Validator{config: config}
}

// Validate performs full validation of the configuration
func (v *Validator) Validate() error {
	if v.config == nil {
		return fmt.Errorf("configuration is nil")
	}

	// Validate participants
	if err := v.validateParticipants(); err != nil {
		return err
	}

	// Validate network parameters
	if err := v.validateNetworkParams(); err != nil {
		return err
	}

	// Validate MEV configuration
	if err := v.validateMEV(); err != nil {
		return err
	}

	// Validate additional services
	if err := v.validateAdditionalServices(); err != nil {
		return err
	}

	// Validate global settings
	if err := v.validateGlobalSettings(); err != nil {
		return err
	}

	return nil
}

func (v *Validator) validateParticipants() error {
	if len(v.config.Participants) == 0 {
		return fmt.Errorf("at least one participant is required")
	}

	for i, p := range v.config.Participants {
		// Validate EL type
		if p.ELType == "" {
			return fmt.Errorf("participant %d: execution layer type is required", i)
		}
		if !isValidELClient(p.ELType) {
			return fmt.Errorf("participant %d: invalid execution layer type: %s", i, p.ELType)
		}

		// Validate CL type
		if p.CLType == "" {
			return fmt.Errorf("participant %d: consensus layer type is required", i)
		}
		if !isValidCLClient(p.CLType) {
			return fmt.Errorf("participant %d: invalid consensus layer type: %s", i, p.CLType)
		}

		// Validate count
		if p.Count < 0 {
			return fmt.Errorf("participant %d: count cannot be negative", i)
		}
		if p.Count > 100 {
			return fmt.Errorf("participant %d: count cannot exceed 100", i)
		}

		// Validate validator count
		if p.ValidatorCount < 0 {
			return fmt.Errorf("participant %d: validator count cannot be negative", i)
		}
		if p.ValidatorCount > 1000000 {
			return fmt.Errorf("participant %d: validator count cannot exceed 1000000", i)
		}
	}

	return nil
}

func (v *Validator) validateNetworkParams() error {
	if v.config.NetworkParams == nil {
		return nil // Network params are optional
	}

	np := v.config.NetworkParams

	// Validate chain ID
	if np.ChainID > 0 && np.ChainID < 1 {
		return fmt.Errorf("chain ID must be positive")
	}

	// Validate network ID
	if np.NetworkID > 0 && np.NetworkID < 1 {
		return fmt.Errorf("network ID must be positive")
	}

	// Validate slot times
	if np.SecondsPerSlot != 0 && (np.SecondsPerSlot < 1 || np.SecondsPerSlot > 120) {
		return fmt.Errorf("seconds per slot must be between 1 and 120")
	}

	if np.SlotsPerEpoch != 0 && (np.SlotsPerEpoch < 1 || np.SlotsPerEpoch > 64) {
		return fmt.Errorf("slots per epoch must be between 1 and 64")
	}

	// Validate fork epochs
	if np.CapellaForkEpoch < 0 || np.DenebForkEpoch < 0 || np.ElectraForkEpoch < 0 {
		return fmt.Errorf("fork epochs cannot be negative")
	}

	// Ensure fork ordering
	if np.CapellaForkEpoch > 0 && np.DenebForkEpoch > 0 && np.CapellaForkEpoch > np.DenebForkEpoch {
		return fmt.Errorf("capella fork epoch must be before deneb fork epoch")
	}
	if np.DenebForkEpoch > 0 && np.ElectraForkEpoch > 0 && np.DenebForkEpoch > np.ElectraForkEpoch {
		return fmt.Errorf("deneb fork epoch must be before electra fork epoch")
	}

	return nil
}

func (v *Validator) validateMEV() error {
	if v.config.MEV == nil {
		return nil // MEV is optional
	}

	mev := v.config.MEV

	// Validate MEV type
	if mev.Type != "" && !isValidMEVType(mev.Type) {
		return fmt.Errorf("invalid MEV type: %s", mev.Type)
	}

	// Validate relay URL
	if mev.RelayURL != "" && !isValidURL(mev.RelayURL) {
		return fmt.Errorf("invalid MEV relay URL: %s", mev.RelayURL)
	}

	// Validate max bundle length
	if mev.MaxBundleLength < 0 {
		return fmt.Errorf("max bundle length cannot be negative")
	}
	if mev.MaxBundleLength > 100 {
		return fmt.Errorf("max bundle length cannot exceed 100")
	}

	return nil
}

func (v *Validator) validateAdditionalServices() error {
	if len(v.config.AdditionalServices) == 0 {
		return nil // Additional services are optional
	}

	serviceNames := make(map[string]bool)
	for i, svc := range v.config.AdditionalServices {
		if svc.Name == "" {
			return fmt.Errorf("additional service %d: name is required", i)
		}

		if serviceNames[svc.Name] {
			return fmt.Errorf("duplicate additional service: %s", svc.Name)
		}
		serviceNames[svc.Name] = true

		if !isValidServiceName(svc.Name) {
			return fmt.Errorf("invalid additional service name: %s", svc.Name)
		}
	}

	return nil
}

func (v *Validator) validateGlobalSettings() error {
	if v.config.GlobalClientLogLevel != "" {
		if !isValidLogLevel(v.config.GlobalClientLogLevel) {
			return fmt.Errorf("invalid global client log level: %s", v.config.GlobalClientLogLevel)
		}
	}

	return nil
}

// Helper functions

func isValidELClient(clientType types.ClientType) bool {
	switch clientType {
	case types.ClientGeth, types.ClientBesu, types.ClientNethermind, types.ClientErigon, types.ClientReth:
		return true
	default:
		return false
	}
}

func isValidCLClient(clientType types.ClientType) bool {
	switch clientType {
	case types.ClientLighthouse, types.ClientTeku, types.ClientPrysm, types.ClientNimbus, types.ClientLodestar, types.ClientGrandine:
		return true
	default:
		return false
	}
}

func isValidMEVType(mevType string) bool {
	validTypes := []string{"none", "mock", "full", "relay"}
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

func isValidLogLevel(level string) bool {
	validLevels := []string{"trace", "debug", "info", "warn", "error", "fatal"}
	for _, valid := range validLevels {
		if strings.ToLower(level) == valid {
			return true
		}
	}
	return false
}