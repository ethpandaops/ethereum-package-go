package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ethpandaops/ethereum-package-go"
)

// Example demonstrating network lifecycle management options
func main() {
	ctx := context.Background()

	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run main.go auto-cleanup    # Default behavior - auto cleanup on exit")
		fmt.Println("  go run main.go orphan          # Orphan network - no cleanup")
		fmt.Println("  go run main.go reuse <name>    # Reuse existing network")
		os.Exit(1)
	}

	mode := os.Args[1]

	switch mode {
	case "auto-cleanup":
		demonstrateAutoCleanup(ctx)
	case "orphan":
		demonstrateOrphan(ctx)
	case "reuse":
		if len(os.Args) < 3 {
			log.Fatal("Please provide a network name for reuse")
		}
		demonstrateReuse(ctx, os.Args[2])
	default:
		log.Fatalf("Unknown mode: %s", mode)
	}
}

// demonstrateAutoCleanup shows the default behavior (testcontainers-style)
func demonstrateAutoCleanup(ctx context.Context) {
	fmt.Println("🔄 Demonstrating AUTO-CLEANUP behavior...")
	fmt.Println("The network will be automatically destroyed when this process exits.")

	// Create network with default settings (auto-cleanup enabled)
	network, err := ethereum.Run(ctx,
		ethereum.Minimal(),
		ethereum.WithTimeout(3*time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to start network: %v", err)
	}

	fmt.Printf("✅ Network created: %s\n", network.EnclaveName())
	fmt.Println("💡 This network will be automatically cleaned up when the process exits.")
	fmt.Println("   Try killing this process and check 'kurtosis enclave ls' - the enclave will be gone.")

	// Simulate some work
	fmt.Println("⏳ Simulating work for 30 seconds...")
	time.Sleep(30 * time.Second)

	fmt.Println("🏁 Process ending - auto-cleanup will trigger")
	// No explicit cleanup needed - it happens automatically
}

// demonstrateOrphan shows how to create a persistent network
func demonstrateOrphan(ctx context.Context) {
	fmt.Println("🔄 Demonstrating ORPHAN behavior...")
	fmt.Println("The network will persist after this process exits.")

	// Create network with orphan option
	network, err := ethereum.Run(ctx,
		ethereum.Minimal(),
		ethereum.WithOrphanOnExit(), // This prevents auto-cleanup
		ethereum.WithTimeout(3*time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to start network: %v", err)
	}

	fmt.Printf("✅ Network created: %s\n", network.EnclaveName())
	fmt.Println("💡 This network is ORPHANED and will persist after process exit.")
	fmt.Printf("   To clean it up manually, run: kurtosis enclave rm %s\n", network.EnclaveName())

	// Simulate some work
	fmt.Println("⏳ Simulating work for 30 seconds...")
	time.Sleep(30 * time.Second)

	fmt.Println("🏁 Process ending - network will remain running")
	fmt.Printf("   Remember to run: kurtosis enclave rm %s\n", network.EnclaveName())
}

// demonstrateReuse shows how to reuse an existing network
func demonstrateReuse(ctx context.Context, networkName string) {
	fmt.Printf("🔄 Demonstrating REUSE behavior for network: %s\n", networkName)

	// Try to reuse existing network or create new one with that name
	network, err := ethereum.Run(ctx,
		ethereum.Minimal(),
		ethereum.WithReuse(networkName), // This enables reuse
		ethereum.WithTimeout(3*time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to start/reuse network: %v", err)
	}

	fmt.Printf("✅ Network ready: %s\n", network.EnclaveName())
	fmt.Printf("   Chain ID: %d\n", network.ChainID())
	fmt.Printf("   Execution clients: %d\n", len(network.ExecutionClients().All()))
	fmt.Printf("   Consensus clients: %d\n", len(network.ConsensusClients().All()))

	fmt.Println("💡 This network is marked for REUSE and will persist after process exit.")
	fmt.Printf("   To clean it up manually, run: kurtosis enclave rm %s\n", network.EnclaveName())

	// Simulate some work
	fmt.Println("⏳ Simulating work for 30 seconds...")
	time.Sleep(30 * time.Second)

	fmt.Println("🏁 Process ending - reused network will remain running")
}
