package config

import (
	"fmt"

	"gopkg.in/yaml.v3"

)

// ToYAML converts the configuration to a YAML string
func ToYAML(config *EthereumPackageConfig) (string, error) {
	if config == nil {
		return "", fmt.Errorf("config cannot be nil")
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	return string(data), nil
}

// FromYAML parses a YAML string into an EthereumPackageConfig
func FromYAML(yamlStr string) (*EthereumPackageConfig, error) {
	var config EthereumPackageConfig

	if err := yaml.Unmarshal([]byte(yamlStr), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Apply defaults after unmarshaling
	config.ApplyDefaults()

	return &config, nil
}