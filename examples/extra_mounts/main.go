package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethpandaops/ethereum-package-go"
	"github.com/ethpandaops/ethereum-package-go/pkg/client"
	"github.com/ethpandaops/ethereum-package-go/pkg/config"
)

func main() {
	// Create a context for the ethereum network
	ctx := context.Background()

	// Example 1: Participant with CL extra mounts for configuration files
	mountParticipant := config.NewParticipantBuilder().
		WithCL(client.Prysm).
		WithEL(client.Geth).
		// NOTE: Mount paths must exist in the ethereum-package repository
		WithCLExtraMounts(map[string]string{
			"/data/config": "static_files/jwt",
		}).
		WithCLExtraParams([]string{
			"--graffiti=ethereum-package-go-example",
		}).
		Build()

	// Example 2: Using extra environment variables and labels
	advancedParticipant := config.NewParticipantBuilder().
		WithCL(client.Lighthouse).
		WithEL(client.Nethermind).
		WithCount(2).
		WithValidatorCount(32).
		// NOTE: Mount paths must exist in the ethereum-package repository
		WithCLExtraMounts(map[string]string{
			"/data/config": "static_files/jwt",
		}).
		// Add extra command line parameters
		WithCLExtraParams([]string{
			"--graffiti=ethereum-package-go-test",
			"--metrics-address=0.0.0.0",
			"--metrics-port=5054",
		}).
		WithELExtraParams([]string{
			"--Metrics.Enabled",
			"--Metrics.PushGatewayUrl=http://prometheus-pushgateway:9091",
		}).
		// Set environment variables
		WithCLExtraEnvVars(map[string]string{
			"RUST_LOG":           "debug",
			"LIGHTHOUSE_PROFILE": "minimal",
		}).
		WithELExtraEnvVars(map[string]string{
			"NETHERMIND_CLI_SWITCH_SYNC": "debug",
		}).
		// Add labels for container organization
		WithCLExtraLabels(map[string]string{
			"team":        "devops",
			"environment": "testing",
			"version":     "v1.0.0",
		}).
		Build()

	// Example 3: Validator client configuration
	validatorParticipant := config.NewParticipantBuilder().
		WithCL(client.Teku).
		WithEL(client.Besu).
		WithValidatorCount(100).
		// NOTE: Mount paths must exist in the ethereum-package repository
		WithCLExtraMounts(map[string]string{
			"/data/config": "static_files/jwt",
		}).
		WithVCExtraParams([]string{
			"--validators-proposer-default-fee-recipient=0x0000000000000000000000000000000000000000",
		}).
		WithVCExtraEnvVars(map[string]string{
			"JAVA_OPTS": "-Xmx2g -Xms1g",
		}).
		WithVCExtraLabels(map[string]string{
			"role":     "validator",
			"security": "high",
		}).
		Build()

	// Build the ethereum network configuration
	networkConfig, err := config.NewConfigBuilder().
		WithParticipant(mountParticipant).
		WithParticipant(advancedParticipant).
		WithParticipant(validatorParticipant).
		WithNetworkParams(&config.NetworkParams{
			Network:                 "kurtosis",
			SecondsPerSlot:          12,
			NumValidatorKeysPerNode: 64,
		}).
		Build()

	if err != nil {
		log.Fatalf("Failed to build network configuration: %v", err)
	}

	// Start the ethereum network with the configuration
	fmt.Println("Starting ethereum network with extra mounts configuration...")
	network, err := ethereum.Run(ctx, ethereum.WithConfig(networkConfig))
	if err != nil {
		log.Fatalf("Failed to start ethereum network: %v", err)
	}

	fmt.Println("Network started successfully!")
	fmt.Println("The following extra configurations have been applied:")
	fmt.Println("- Participant 1: CL extra mounts and parameters")
	fmt.Println("- Participant 2: Advanced configuration with extra params, env vars, and labels")
	fmt.Println("- Participant 3: Validator client configuration with extra params, env vars, and labels")

	// Get execution clients
	execClients := network.ExecutionClients()
	fmt.Printf("\nExecution clients running: %d\n", len(execClients.All()))

	// Get consensus clients
	consClients := network.ConsensusClients()
	fmt.Printf("Consensus clients running: %d\n", len(consClients.All()))

	// Display configuration information
	fmt.Println("\nExtra configurations include:")
	fmt.Println("- Custom command-line parameters for clients")
	fmt.Println("- Environment variables for runtime configuration")
	fmt.Println("- Labels for container organization")
	fmt.Println("- Support for mounting files (when paths exist in ethereum-package)")

	// Clean up
	fmt.Println("\nPress Enter to stop the network...")
	var input string
	fmt.Scanln(&input)

	fmt.Println("Stopping network...")
	if err := network.Cleanup(ctx); err != nil {
		log.Fatalf("Failed to cleanup network: %v", err)
	}

	fmt.Println("Network stopped successfully!")
}
