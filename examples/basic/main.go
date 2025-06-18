package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethpandaops/ethereum-package-go"
)

func main() {
	// Create a context
	ctx := context.Background()

	// Start a minimal Ethereum network
	fmt.Println("Starting Ethereum network...")
	network, err := ethereum.Run(ctx, ethereum.Minimal())
	if err != nil {
		log.Fatalf("Failed to start network: %v", err)
	}
	defer network.Cleanup(ctx)

	fmt.Println("Network started successfully!")
	fmt.Printf("Chain ID: %d\n", network.ChainID())

	// Get execution clients
	execClients := network.ExecutionClients()
	if len(execClients.All()) > 0 {
		client := execClients.All()[0]
		fmt.Printf("\nExecution Client:\n")
		fmt.Printf("  Name: %s\n", client.Name())
		fmt.Printf("  Type: %s\n", client.Type())
		fmt.Printf("  RPC URL: %s\n", client.RPCURL())
		fmt.Printf("  WS URL: %s\n", client.WSURL())
	}

	// Get consensus clients
	consClients := network.ConsensusClients()
	if len(consClients.All()) > 0 {
		client := consClients.All()[0]
		fmt.Printf("\nConsensus Client:\n")
		fmt.Printf("  Name: %s\n", client.Name())
		fmt.Printf("  Type: %s\n", client.Type())
		fmt.Printf("  Beacon API URL: %s\n", client.BeaconAPIURL())
	}

	// Get Apache config server
	if apache := network.ApacheConfig(); apache != nil {
		fmt.Printf("\nApache Config Server:\n")
		fmt.Printf("  Genesis SSZ: %s\n", apache.GenesisSSZURL())
		fmt.Printf("  Config YAML: %s\n", apache.ConfigYAMLURL())
	}

	fmt.Println("\nNetwork is ready! Press Enter to cleanup and exit...")
	fmt.Scanln()
}
