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
			"/data/config": "static_files/nginx-config",
		}).
		WithCLExtraParams([]string{
			"--graffiti=ethereum-package-go-example",
		}).
		Build()

	// Build the ethereum network configuration
	networkConfig, err := config.NewConfigBuilder().
		WithParticipant(mountParticipant).
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
