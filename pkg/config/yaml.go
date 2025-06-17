package config

import (
	"bytes"
	"fmt"

	"github.com/ethpandaops/ethereum-package-go/pkg/types"
	"gopkg.in/yaml.v3"
)

// ToYAML converts the configuration to YAML format
func ToYAML(config *types.EthereumPackageConfig) (string, error) {
	// Create a clean structure for YAML marshaling
	yamlConfig := make(map[string]interface{})

	// Add participants
	if len(config.Participants) > 0 {
		participants := make([]map[string]interface{}, 0, len(config.Participants))
		for _, p := range config.Participants {
			participant := make(map[string]interface{})
			
			if p.ELType != "" {
				participant["el_type"] = string(p.ELType)
			}
			if p.CLType != "" {
				participant["cl_type"] = string(p.CLType)
			}
			if p.ELVersion != "" {
				participant["el_version"] = p.ELVersion
			}
			if p.CLVersion != "" {
				participant["cl_version"] = p.CLVersion
			}
			if p.Count > 0 {
				participant["count"] = p.Count
			}
			if p.ValidatorCount > 0 {
				participant["validator_count"] = p.ValidatorCount
			}
			
			participants = append(participants, participant)
		}
		yamlConfig["participants"] = participants
	}

	// Add network params
	if config.NetworkParams != nil {
		networkParams := make(map[string]interface{})
		
		if config.NetworkParams.ChainID > 0 {
			networkParams["chain_id"] = config.NetworkParams.ChainID
		}
		if config.NetworkParams.NetworkID > 0 {
			networkParams["network_id"] = config.NetworkParams.NetworkID
		}
		if config.NetworkParams.SecondsPerSlot > 0 {
			networkParams["seconds_per_slot"] = config.NetworkParams.SecondsPerSlot
		}
		if config.NetworkParams.SlotsPerEpoch > 0 {
			networkParams["slots_per_epoch"] = config.NetworkParams.SlotsPerEpoch
		}
		if config.NetworkParams.CapellaForkEpoch > 0 {
			networkParams["capella_fork_epoch"] = config.NetworkParams.CapellaForkEpoch
		}
		if config.NetworkParams.DenebForkEpoch > 0 {
			networkParams["deneb_fork_epoch"] = config.NetworkParams.DenebForkEpoch
		}
		if config.NetworkParams.ElectraForkEpoch > 0 {
			networkParams["electra_fork_epoch"] = config.NetworkParams.ElectraForkEpoch
		}
		if config.NetworkParams.MinValidatorWithdrawability > 0 {
			networkParams["min_validator_withdrawability"] = config.NetworkParams.MinValidatorWithdrawability
		}
		
		if len(networkParams) > 0 {
			yamlConfig["network_params"] = networkParams
		}
	}

	// Add MEV params
	if config.MEV != nil {
		mevParams := make(map[string]interface{})
		
		if config.MEV.Type != "" {
			mevParams["type"] = config.MEV.Type
		}
		if config.MEV.RelayURL != "" {
			mevParams["relay_url"] = config.MEV.RelayURL
		}
		if config.MEV.MinBidEth != "" {
			mevParams["min_bid_eth"] = config.MEV.MinBidEth
		}
		if config.MEV.MaxBundleLength > 0 {
			mevParams["max_bundle_length"] = config.MEV.MaxBundleLength
		}
		
		if len(mevParams) > 0 {
			yamlConfig["mev_params"] = mevParams
		}
	}

	// Add additional services
	if len(config.AdditionalServices) > 0 {
		services := make([]interface{}, 0, len(config.AdditionalServices))
		for _, svc := range config.AdditionalServices {
			service := make(map[string]interface{})
			service["name"] = svc.Name
			
			if len(svc.Config) > 0 {
				service["config"] = svc.Config
			}
			
			services = append(services, service)
		}
		yamlConfig["additional_services"] = services
	}

	// Add global client log level
	if config.GlobalClientLogLevel != "" {
		yamlConfig["global_client_log_level"] = config.GlobalClientLogLevel
	}

	// Marshal to YAML
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	
	if err := encoder.Encode(yamlConfig); err != nil {
		return "", fmt.Errorf("failed to encode YAML: %w", err)
	}

	return buf.String(), nil
}

// FromYAML parses YAML configuration into EthereumPackageConfig
func FromYAML(yamlContent string) (*types.EthereumPackageConfig, error) {
	var rawConfig map[string]interface{}
	
	if err := yaml.Unmarshal([]byte(yamlContent), &rawConfig); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	config := &types.EthereumPackageConfig{}

	// Parse participants
	if participants, ok := rawConfig["participants"].([]interface{}); ok {
		for _, p := range participants {
			if participant, ok := p.(map[string]interface{}); ok {
				pc := types.ParticipantConfig{}
				
				if elType, ok := participant["el_type"].(string); ok {
					pc.ELType = types.ClientType(elType)
				}
				if clType, ok := participant["cl_type"].(string); ok {
					pc.CLType = types.ClientType(clType)
				}
				if elVersion, ok := participant["el_version"].(string); ok {
					pc.ELVersion = elVersion
				}
				if clVersion, ok := participant["cl_version"].(string); ok {
					pc.CLVersion = clVersion
				}
				if count, ok := participant["count"].(int); ok {
					pc.Count = count
				}
				if validatorCount, ok := participant["validator_count"].(int); ok {
					pc.ValidatorCount = validatorCount
				}
				
				config.Participants = append(config.Participants, pc)
			}
		}
	}

	// Parse network params
	if networkParams, ok := rawConfig["network_params"].(map[string]interface{}); ok {
		np := &types.NetworkParams{}
		
		if chainID, ok := networkParams["chain_id"].(int); ok {
			np.ChainID = uint64(chainID)
		}
		if networkID, ok := networkParams["network_id"].(int); ok {
			np.NetworkID = uint64(networkID)
		}
		if secondsPerSlot, ok := networkParams["seconds_per_slot"].(int); ok {
			np.SecondsPerSlot = secondsPerSlot
		}
		if slotsPerEpoch, ok := networkParams["slots_per_epoch"].(int); ok {
			np.SlotsPerEpoch = slotsPerEpoch
		}
		if capellaForkEpoch, ok := networkParams["capella_fork_epoch"].(int); ok {
			np.CapellaForkEpoch = capellaForkEpoch
		}
		if denebForkEpoch, ok := networkParams["deneb_fork_epoch"].(int); ok {
			np.DenebForkEpoch = denebForkEpoch
		}
		if electraForkEpoch, ok := networkParams["electra_fork_epoch"].(int); ok {
			np.ElectraForkEpoch = electraForkEpoch
		}
		if minValidatorWithdrawability, ok := networkParams["min_validator_withdrawability"].(int); ok {
			np.MinValidatorWithdrawability = minValidatorWithdrawability
		}
		
		config.NetworkParams = np
	}

	// Parse MEV params
	if mevParams, ok := rawConfig["mev_params"].(map[string]interface{}); ok {
		mev := &types.MEVConfig{}
		
		if mevType, ok := mevParams["type"].(string); ok {
			mev.Type = mevType
		}
		if relayURL, ok := mevParams["relay_url"].(string); ok {
			mev.RelayURL = relayURL
		}
		if minBidEth, ok := mevParams["min_bid_eth"].(string); ok {
			mev.MinBidEth = minBidEth
		}
		if maxBundleLength, ok := mevParams["max_bundle_length"].(int); ok {
			mev.MaxBundleLength = maxBundleLength
		}
		
		config.MEV = mev
	}

	// Parse additional services
	if services, ok := rawConfig["additional_services"].([]interface{}); ok {
		for _, s := range services {
			if service, ok := s.(map[string]interface{}); ok {
				svc := types.AdditionalService{}
				
				if name, ok := service["name"].(string); ok {
					svc.Name = name
				}
				if cfg, ok := service["config"].(map[string]interface{}); ok {
					svc.Config = cfg
				}
				
				config.AdditionalServices = append(config.AdditionalServices, svc)
			}
		}
	}

	// Parse global client log level
	if logLevel, ok := rawConfig["global_client_log_level"].(string); ok {
		config.GlobalClientLogLevel = logLevel
	}

	return config, nil
}