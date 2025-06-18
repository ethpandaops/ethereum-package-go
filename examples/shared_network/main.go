package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethpandaops/ethereum-package-go"
)

func main() {
	ctx := context.Background()

	// Example 1: Create a network with a random name (avoids conflicts)
	fmt.Println("Creating network with random name...")
	network1, err := ethereum.Run(ctx, ethereum.Minimal())
	if err != nil {
		log.Fatalf("Failed to create network: %v", err)
	}
	fmt.Printf("Created network: %s\n", network1.EnclaveName())

	// Example 2: Create a network with a specific name
	fmt.Println("\nCreating network with specific name...")
	network2, err := ethereum.Run(ctx,
		ethereum.WithEnclaveName("my-test-network"),
		ethereum.Minimal(),
	)
	if err != nil {
		log.Fatalf("Failed to create named network: %v", err)
	}
	fmt.Printf("Created network: %s\n", network2.EnclaveName())

	// Example 3: Find or create a network (useful for shared test environments)
	fmt.Println("\nFinding or creating shared network...")
	network3, err := ethereum.FindOrCreateNetwork(ctx, "my-test-network", ethereum.Minimal())
	if err != nil {
		log.Fatalf("Failed to find/create network: %v", err)
	}
	fmt.Printf("Found/created network: %s\n", network3.EnclaveName())

	// network3 should be the same as network2
	fmt.Printf("Network2 and Network3 are the same: %v\n",
		network2.EnclaveName() == network3.EnclaveName())

	// Example 4: Create another random network (no name = unique network)
	fmt.Println("\nCreating another network with random name...")
	network4, err := ethereum.FindOrCreateNetwork(ctx, "", ethereum.Minimal())
	if err != nil {
		log.Fatalf("Failed to create network: %v", err)
	}
	fmt.Printf("Created network: %s\n", network4.EnclaveName())

	// Clean up
	fmt.Println("\nCleaning up networks...")
	if err := network1.Cleanup(ctx); err != nil {
		log.Printf("Failed to cleanup network1: %v", err)
	}
	if err := network2.Cleanup(ctx); err != nil {
		log.Printf("Failed to cleanup network2: %v", err)
	}
	// network3 is the same as network2, so already cleaned up
	if err := network4.Cleanup(ctx); err != nil {
		log.Printf("Failed to cleanup network4: %v", err)
	}

	fmt.Println("Done!")
}
