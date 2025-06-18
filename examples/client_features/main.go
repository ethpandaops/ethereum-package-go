package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ethpandaops/ethereum-package-go/pkg/client"
)

func main() {
	ctx := context.Background()

	// Example: Working with Consensus Client Peer IDs
	fmt.Println("=== Consensus Client Peer ID Examples ===")
	demonstratePeerIDFunctionality(ctx)

	fmt.Println("\n=== Log Filtering Examples ===")
	demonstrateLogsFunctionality(ctx)
}

func demonstratePeerIDFunctionality(ctx context.Context) {
	// Create some example consensus clients
	clients := client.NewConsensusClients()

	// Add lighthouse client
	lighthouseClient := client.NewConsensusClient(
		client.Lighthouse,
		"lighthouse-1",
		"v4.6.0",
		"http://lighthouse-beacon:5052", // beacon API URL
		"http://lighthouse-beacon:8080", // metrics URL
		"enr:-MS4QBU9k_cMlyFm7Tlj4bpMRdiq6bOvl3KGfJUrm3JWy5hUr8l_M1S3TI2AXKQ5z_wQbr_jzb_LIGEr7vDRWKwv4_MBgmlkgnY0gmlwhH8AAAGJc2VjcDI1NmsxoQKKFf5UZkP2-_CK9bM34hRzW5zR-ZPh1O6_bFf9NbCl1YN0Y3CCdl-DdWRwgnZf",
		"cached-peer-id", // cached peer ID (will be overridden by FetchPeerID)
		"lighthouse-beacon",
		"lighthouse-container",
		9000,
	)

	// Add teku client
	tekuClient := client.NewConsensusClient(
		client.Teku,
		"teku-1",
		"v24.1.1",
		"http://teku-beacon:5051",
		"http://teku-beacon:8080",
		"enr:-MS4QExample...",
		"cached-teku-peer-id",
		"teku-beacon",
		"teku-container",
		9000,
	)

	clients.Add(lighthouseClient)
	clients.Add(tekuClient)

	fmt.Printf("Added %d consensus clients to collection\n", clients.Count())

	// Example 1: Fetch peer ID for a single client
	fmt.Println("\n1. Fetching Peer ID for single client:")
	fmt.Printf("   Client: %s (%s)\n", lighthouseClient.Name(), lighthouseClient.Type())
	fmt.Printf("   Cached Peer ID: %s\n", lighthouseClient.PeerID())

	// Note: In a real environment, this would make an HTTP request to the beacon API
	// For this example, we'll show what the call would look like
	fmt.Printf("   To fetch live peer ID: client.FetchPeerID(ctx)\n")
	fmt.Printf("   This calls: GET %s/eth/v1/node/identity\n", lighthouseClient.BeaconAPIURL())

	// Example 2: Fetch peer IDs for all clients in collection
	fmt.Println("\n2. Fetching Peer IDs for all clients:")
	fmt.Printf("   Available client types: %v\n", clients.Types())

	// Note: In real usage, this would make HTTP requests to all beacon APIs
	fmt.Printf("   To fetch all peer IDs: clients.PeerIDs(ctx)\n")
	fmt.Printf("   This would return a map[string]string of client names to peer IDs\n")

	// Example 3: Fetch peer IDs by type
	fmt.Println("\n3. Fetching Peer IDs by type:")
	lighthouseClients := clients.ByType(client.Lighthouse)
	fmt.Printf("   Lighthouse clients: %d\n", len(lighthouseClients))
	for _, lc := range lighthouseClients {
		fmt.Printf("   - %s at %s\n", lc.Name(), lc.BeaconAPIURL())
	}

	tekuClients := clients.ByType(client.Teku)
	fmt.Printf("   Teku clients: %d\n", len(tekuClients))
	for _, tc := range tekuClients {
		fmt.Printf("   - %s at %s\n", tc.Name(), tc.BeaconAPIURL())
	}

	fmt.Printf("   To fetch peer IDs by type: clients.PeerIDsByType(ctx, client.Lighthouse)\n")
}

func demonstrateLogsFunctionality(ctx context.Context) {
	// Example: Working with Log Filtering

	// Note: In a real scenario, you'd create the LogsClient with actual Kurtosis context
	fmt.Println("Log filtering examples (requires running Kurtosis environment):")

	fmt.Println("\n1. Basic log filtering options:")

	// Example filter options
	tailOptions := client.TailLogs(50, "ERROR")
	fmt.Printf("   TailLogs(50, \"ERROR\") creates options for: last 50 lines containing 'ERROR'\n")
	_ = tailOptions

	followOptions := client.FollowLogs("WARN")
	fmt.Printf("   FollowLogs(\"WARN\") creates options for: streaming logs containing 'WARN'\n")
	_ = followOptions

	fmt.Println("\n2. Advanced filtering with functional options:")

	// Chainable options example
	advancedOptions := []client.LogOption{
		client.WithLines(100),             // Last 100 lines
		client.WithGrep("peer"),           // Lines containing "peer"
		client.WithSince(5 * time.Minute), // From last 5 minutes
		client.WithCaseSensitive(false),   // Case insensitive
		client.WithExcludeRegex("debug"),  // Exclude debug messages
	}
	fmt.Printf("   Advanced options: last 100 lines, containing 'peer', from last 5 minutes, excluding 'debug'\n")
	_ = advancedOptions

	fmt.Println("\n3. Usage examples:")
	fmt.Println("   // Create logs client (requires Kurtosis context)")
	fmt.Println("   kurtosisCtx, _ := kurtosis_context.NewKurtosisContextFromLocalEngine()")
	fmt.Println("   logsClient := client.NewLogsClient(kurtosisCtx, \"my-enclave\")")
	fmt.Println("")
	fmt.Println("   // Get logs for a consensus client")
	fmt.Println("   logs, err := logsClient.ConsensusClientLogs(ctx, consensusClient,")
	fmt.Println("       client.WithLines(50),")
	fmt.Println("       client.WithGrep(\"ERROR\"),")
	fmt.Println("   )")
	fmt.Println("")
	fmt.Println("   // Stream logs for real-time monitoring")
	fmt.Println("   logChan, errChan := logsClient.LogsStream(ctx, consensusClient,")
	fmt.Println("       client.WithFollow(true),")
	fmt.Println("       client.WithGrep(\"peer connected\"),")
	fmt.Println("   )")
	fmt.Println("")
	fmt.Println("   // Get logs for all consensus clients")
	fmt.Println("   allLogs, err := logsClient.AllConsensusClientLogs(ctx, clients,")
	fmt.Println("       client.WithLines(20),")
	fmt.Println("   )")

	fmt.Println("\n4. Convenience functions:")
	fmt.Printf("   client.TailLogs(n, pattern) - Get last n lines matching pattern\n")
	fmt.Printf("   client.FollowLogs(pattern) - Stream logs matching pattern\n")

	// Real world example structure (commented out since it requires actual Kurtosis)
	fmt.Println("\n5. Real-world integration example:")
	fmt.Println("   ```go")
	fmt.Println("   // In a real scenario with running services:")
	fmt.Println("   kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()")
	fmt.Println("   if err != nil {")
	fmt.Println("       log.Fatal(err)")
	fmt.Println("   }")
	fmt.Println("   ")
	fmt.Println("   logsClient := client.NewLogsClient(kurtosisCtx, \"ethereum-network\")")
	fmt.Println("   ")
	fmt.Println("   // Monitor for sync issues across all clients")
	fmt.Println("   for _, consensusClient := range clients.All() {")
	fmt.Println("       go func(c client.ConsensusClient) {")
	fmt.Println("           logChan, errChan := logsClient.LogsStream(ctx, c,")
	fmt.Println("               client.WithFollow(true),")
	fmt.Println("               client.WithGrep(\"sync\"),")
	fmt.Println("           )")
	fmt.Println("           for {")
	fmt.Println("               select {")
	fmt.Println("               case line := <-logChan:")
	fmt.Printf("%s\n", "                   fmt.Printf(\"[%s] %s\\n\", c.Name(), line)")
	fmt.Println("               case err := <-errChan:")
	fmt.Printf("%s\n", "                   log.Printf(\"Error streaming logs for %s: %v\", c.Name(), err)")
	fmt.Println("                   return")
	fmt.Println("               }")
	fmt.Println("           }")
	fmt.Println("       }(consensusClient)")
	fmt.Println("   }")
	fmt.Println("   ```")
}

// Example of how to integrate with actual network
func integrateWithNetwork() {
	// This would be in a real application that uses the ethereum-package-go
	fmt.Println("\nIntegration with ethereum-package-go:")
	fmt.Println("```go")
	fmt.Println("// Deploy network")
	fmt.Println("network, err := ethereum.DeployNetwork(ctx, cfg)")
	fmt.Println("if err != nil {")
	fmt.Println("    log.Fatal(err)")
	fmt.Println("}")
	fmt.Println("")
	fmt.Println("// Get consensus clients from deployed network")
	fmt.Println("consensusClients := network.ConsensusClients()")
	fmt.Println("")
	fmt.Println("// Fetch all peer IDs")
	fmt.Println("peerIDs, err := consensusClients.PeerIDs(ctx)")
	fmt.Println("if err != nil {")
	fmt.Println("    log.Fatal(err)")
	fmt.Println("}")
	fmt.Println("")
	fmt.Println("for clientName, peerID := range peerIDs {")
	fmt.Printf("%s\n", "    fmt.Printf(\"Client %s has peer ID: %s\\n\", clientName, peerID)")
	fmt.Println("}")
	fmt.Println("")
	fmt.Println("// Monitor logs")
	fmt.Println("logsClient := client.NewLogsClient(network.KurtosisContext(), network.EnclaveName())")
	fmt.Println("logs, err := logsClient.AllConsensusClientLogs(ctx, consensusClients,")
	fmt.Println("    client.WithLines(100),")
	fmt.Println("    client.WithGrep(\"ERROR\"),")
	fmt.Println(")")
	fmt.Println("```")
}
