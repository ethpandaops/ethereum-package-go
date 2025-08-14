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
	ctx := context.Background()

	// Method 1: Using ExtraFilesHelper for convenience
	helper := config.NewExtraFilesHelper()

	// Add inline content
	helper.AddFile("beacon-config.yaml", `
# Beacon node configuration
metrics:
  enabled: true
  port: 8080
logging:
  level: info
`)

	// Add from file (if it exists)
	if err := helper.AddFileFromPath("validator-config.yaml", "./configs/sample.yaml"); err != nil {
		log.Printf("Note: sample.yaml not found, using inline example: %v", err)
		helper.AddFile("validator-config.yaml", "graffiti: 'ethereum-package-go example'")
	}

	// Add JSON config
	jsonConfig := map[string]interface{}{
		"version": "1.0",
		"features": map[string]bool{
			"metrics": true,
			"tracing": false,
		},
	}
	helper.AddJSON("features.json", jsonConfig)

	// Method 2: Direct inline definition
	participant := config.NewParticipantBuilder().
		WithEL(client.Geth).
		WithCL(client.Lighthouse).
		//WithCLExtraMounts(map[string]string{
		//	"/configs/beacon.yaml":   "beacon-config.yaml",
		//	"/configs/features.json": "features.json",
		//}).
		//WithCLExtraParams([]string{
		//	"--config-file=/configs/beacon.yaml",
		//}).
		//WithVCExtraMounts(map[string]string{
		//	"/configs/validator.yaml": "validator-config.yaml",
		//}).
		Build()

	// Build network configuration with extra files at root level
	networkConfig, err := config.NewConfigBuilder().
		WithParticipant(participant).
		WithNetworkParams(&config.NetworkParams{
			Network:                 "kurtosis",
			SecondsPerSlot:          12,
			NumValidatorKeysPerNode: 32,
		}).
		WithExtraFiles(helper.Build()).
		Build()

	if err != nil {
		log.Fatalf("Failed to build config: %v", err)
	}

	fmt.Printf("Starting network with %d extra files...\n", helper.Count())

	// Start the network
	network, err := ethereum.Run(ctx, ethereum.WithConfig(networkConfig))
	if err != nil {
		log.Fatalf("Failed to start network: %v", err)
	}

	fmt.Println("Network started successfully!")
	fmt.Println("Extra files have been mounted into the containers.")

	// Display network info
	execClients := network.ExecutionClients()
	consClients := network.ConsensusClients()
	fmt.Printf("\nExecution clients: %d\n", len(execClients.All()))
	fmt.Printf("Consensus clients: %d\n", len(consClients.All()))

	// Wait for user input
	fmt.Println("\nPress Enter to stop the network...")
	var input string
	fmt.Scanln(&input)

	// Cleanup
	fmt.Println("Stopping network...")
	if err := network.Cleanup(ctx); err != nil {
		log.Fatalf("Failed to cleanup: %v", err)
	}

	fmt.Println("Network stopped successfully!")
}
