package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ethpandaops/ethereum-package-go"
	"github.com/ethpandaops/ethereum-package-go/pkg/client"
)

// Example showing complete usage of ethereum-package-go
func main() {
	ctx := context.Background()

	fmt.Println("🚀 Starting Ethereum network with ethereum-package-go...")

	// Start a simple network with default settings
	network, err := ethereum.Run(ctx,
		ethereum.Minimal(),
		ethereum.WithChainID(12345),
		ethereum.WithAdditionalServices("prometheus", "grafana"),
		ethereum.WithTimeout(5*time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to start network: %v", err)
	}

	// Network will auto-cleanup when process exits (testcontainers-style)
	// If you want to prevent this, use ethereum.WithOrphanOnExit()
	// If you want explicit control, you can still call network.Cleanup(ctx)

	fmt.Printf("✅ Network started successfully!\n")
	fmt.Printf("   Chain ID: %d\n", network.ChainID())
	fmt.Printf("   Enclave: %s\n", network.EnclaveName())

	// Get execution clients
	executionClients := network.ExecutionClients()
	fmt.Printf("📦 Execution clients: %d\n", len(executionClients.All()))

	if gethClients := executionClients.ByType(client.Geth); len(gethClients) > 0 {
		geth := gethClients[0]
		fmt.Printf("   Geth RPC: %s\n", geth.RPCURL())
		fmt.Printf("   Geth WS:  %s\n", geth.WSURL())
	}

	if besuClients := executionClients.ByType(client.Besu); len(besuClients) > 0 {
		besu := besuClients[0]
		fmt.Printf("   Besu RPC: %s\n", besu.RPCURL())
		fmt.Printf("   Besu WS:  %s\n", besu.WSURL())
	}

	// Get consensus clients
	consensusClients := network.ConsensusClients()
	fmt.Printf("🔗 Consensus clients: %d\n", len(consensusClients.All()))

	if lighthouseClients := consensusClients.ByType(client.Lighthouse); len(lighthouseClients) > 0 {
		lighthouse := lighthouseClients[0]
		fmt.Printf("   Lighthouse Beacon: %s\n", lighthouse.BeaconAPIURL())
	}

	if tekuClients := consensusClients.ByType(client.Teku); len(tekuClients) > 0 {
		teku := tekuClients[0]
		fmt.Printf("   Teku Beacon: %s\n", teku.BeaconAPIURL())
	}

	// Get Apache config server
	if apache := network.ApacheConfig(); apache != nil {
		fmt.Printf("📄 Apache Config Server: %s\n", apache.URL())
		fmt.Printf("   Genesis SSZ: %s\n", apache.GenesisSSZURL())
		fmt.Printf("   Config YAML: %s\n", apache.ConfigYAMLURL())
		fmt.Printf("   Boot ENR: %s\n", apache.BootnodesYAMLURL())
		fmt.Printf("   Deposit Contract: %s\n", apache.DepositContractBlockURL())
	}

	// Show all services
	fmt.Println("\n📊 Additional Services:")
	for _, service := range network.Services() {
		if service.Type != "execution" &&
			service.Type != "consensus" &&
			service.Type != "apache" {
			fmt.Printf("   %s (%s): %s\n", service.Name, service.Type, service.Status)
		}
	}

	fmt.Println("\n💡 Network is ready for testing!")
	fmt.Println("   You can now connect your applications to the endpoints above.")
	fmt.Println("   Press Ctrl+C to shutdown the network.")

	// Keep running until interrupted
	select {
	case <-ctx.Done():
		fmt.Println("Context cancelled, shutting down...")
	}
}
