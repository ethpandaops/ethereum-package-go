package main

import (
	"fmt"
	"log"

	"github.com/ethpandaops/ethereum-package-go/pkg/client"
	"github.com/ethpandaops/ethereum-package-go/pkg/config"
	"gopkg.in/yaml.v3"
)

// Simple test to verify extra configuration is properly marshaled to YAML
func main() {
	// Create a participant with all extra configurations
	participant := config.NewParticipantBuilder().
		WithCL(client.Prysm).
		WithEL(client.Geth).
		WithCLExtraMounts(map[string]string{
			"/etc/tysm": "static_files/tysm",
		}).
		WithCLExtraParams([]string{
			"--xatu-config-file=/etc/tysm/xatu-config.yaml",
		}).
		WithCLExtraEnvVars(map[string]string{
			"LOG_LEVEL": "debug",
		}).
		WithCLExtraLabels(map[string]string{
			"environment": "test",
		}).
		Build()

	// Build the configuration
	cfg, err := config.NewConfigBuilder().
		WithParticipant(participant).
		Build()
	if err != nil {
		log.Fatalf("Failed to build config: %v", err)
	}

	// Marshal to YAML
	yamlBytes, err := yaml.Marshal(cfg)
	if err != nil {
		log.Fatalf("Failed to marshal to YAML: %v", err)
	}

	fmt.Println("Generated YAML configuration:")
	fmt.Println(string(yamlBytes))

	// Verify the fields are present
	yamlStr := string(yamlBytes)
	fields := []string{
		"cl_extra_mounts",
		"cl_extra_params",
		"cl_extra_env_vars",
		"cl_extra_labels",
	}

	fmt.Println("\nVerifying fields:")
	for _, field := range fields {
		if contains(yamlStr, field) {
			fmt.Printf("✓ %s is present\n", field)
		} else {
			fmt.Printf("✗ %s is missing\n", field)
		}
	}
}

func contains(str, substr string) bool {
	return len(substr) > 0 && len(str) >= len(substr) && str != substr && (str == substr || str[0:len(substr)] == substr || str[len(str)-len(substr):] == substr || containsMiddle(str, substr))
}

func containsMiddle(str, substr string) bool {
	for i := 1; i < len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
