package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ethpandaops/ethereum-package-go"
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

	// Ensure cleanup on exit
	defer func() {
		fmt.Println("🧹 Cleaning up network...")
		if err := network.Cleanup(ctx); err != nil {
			log.Printf("Failed to cleanup network: %v", err)
		}
	}()

	fmt.Printf("✅ Network started successfully!\n")
	fmt.Printf("   Chain ID: %d\n", network.ChainID())
	fmt.Printf("   Enclave: %s\n", network.EnclaveName())

	// Get execution clients
	executionClients := network.ExecutionClients()
	fmt.Printf("📦 Execution clients: %d\n", len(executionClients.All()))

	if len(executionClients.Geth()) > 0 {
		geth := executionClients.Geth()[0]
		fmt.Printf("   Geth RPC: %s\n", geth.RPCURL())
		fmt.Printf("   Geth WS:  %s\n", geth.WSURL())
	}

	if len(executionClients.Besu()) > 0 {
		besu := executionClients.Besu()[0]
		fmt.Printf("   Besu RPC: %s\n", besu.RPCURL())
		fmt.Printf("   Besu WS:  %s\n", besu.WSURL())
	}

	// Get consensus clients
	consensusClients := network.ConsensusClients()
	fmt.Printf("🔗 Consensus clients: %d\n", len(consensusClients.All()))

	if len(consensusClients.Lighthouse()) > 0 {
		lighthouse := consensusClients.Lighthouse()[0]
		fmt.Printf("   Lighthouse Beacon: %s\n", lighthouse.BeaconAPIURL())
	}

	if len(consensusClients.Teku()) > 0 {
		teku := consensusClients.Teku()[0]
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

	// Get monitoring services
	if prometheusURL := network.PrometheusURL(); prometheusURL != "" {
		fmt.Printf("📊 Prometheus: %s\n", prometheusURL)
	}

	if grafanaURL := network.GrafanaURL(); grafanaURL != "" {
		fmt.Printf("📈 Grafana: %s\n", grafanaURL)
	}

	if blockscoutURL := network.BlockscoutURL(); blockscoutURL != "" {
		fmt.Printf("🔍 Blockscout: %s\n", blockscoutURL)
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