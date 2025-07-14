package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ethpandaops/ethereum-package-go"
	"github.com/ethpandaops/ethereum-package-go/pkg/client"
	"github.com/sirupsen/logrus"
)

// ExecutionSyncStatus holds sync status information
type ExecutionSyncStatus struct {
	BlockNumber  uint64
	IsSyncing    bool
	PeerCount    int
	SyncProgress *client.SyncProgress
}

// getExecutionSyncStatus returns all sync-related information from an execution client
func getExecutionSyncStatus(ctx context.Context, client *client.BaseExecutionClient) (*ExecutionSyncStatus, error) {
	// Get block number
	blockNumber, err := client.GetBlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get block number: %w", err)
	}

	// Check if syncing
	isSyncing, err := client.IsSyncing(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check sync status: %w", err)
	}

	// Get peer count
	peerCount, err := client.GetPeerCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get peer count: %w", err)
	}

	// Get sync progress
	syncProgress, err := client.SyncProgress(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync progress: %w", err)
	}

	return &ExecutionSyncStatus{
		BlockNumber:  blockNumber,
		IsSyncing:    isSyncing,
		PeerCount:    peerCount,
		SyncProgress: syncProgress,
	}, nil
}

// Example showing complete usage of ethereum-package-go
func main() {
	ctx := context.Background()

	// Start a simple network with default settings
	network, err := ethereum.Run(ctx,
		//ethereum.Minimal(),
		//ethereum.WithNetworkParams(&config.NetworkParams{
		//	Network: "hoodi",
		//}),
		ethereum.WithTimeout(5*time.Minute),
		ethereum.WithOrphanOnExit(),
		ethereum.WithReuse("sync-test"),
		ethereum.WithEnclaveName("sync-test"),
		ethereum.WithConfigFile("hoodi.yaml"),
	)
	if err != nil {
		log.Fatalf("Failed to start network: %v", err)
	}

	// Network will auto-cleanup when process exits (testcontainers-style)
	// If you want to prevent this, use ethereum.WithOrphanOnExit()
	// If you want explicit control, you can still call network.Cleanup(ctx)

	// Iterate over all execution clients
	for _, executionClient := range network.ExecutionClients().All() {
		// Create a BaseExecutionClient from the execution client info
		baseClient := client.NewBaseExecutionClient(client.ClientConfig{
			Name:       executionClient.Name(),
			RPCURL:     executionClient.RPCURL(),
			WSURL:      executionClient.WSURL(),
			EngineURL:  executionClient.EngineURL(),
			MetricsURL: executionClient.MetricsURL(),
			Enode:      executionClient.Enode(),
		})

		logrus.WithFields(logrus.Fields{
			"client":     executionClient.Name(),
			"rpc_url":    baseClient.RPCURL(),
			"ws_url":     executionClient.WSURL(),
			"engine_url": executionClient.EngineURL(),
			"enode":      executionClient.Enode(),
			"type":       executionClient.Type(),
			"image":      executionClient.Image(),
			"entrypoint": executionClient.Entrypoint(),
			"cmd":        executionClient.Cmd(),
		}).Info("Client info")

		// Start goroutine for this execution client
		go func(baseClient *client.BaseExecutionClient) {
			for {
				syncStatus, err := getExecutionSyncStatus(ctx, baseClient)
				if err != nil {
					log.Printf("Failed to get sync status for %s: %v", baseClient.Name(), err)
				} else {
					logrus.WithFields(logrus.Fields{
						"client":        baseClient.Name(),
						"current_block": syncStatus.BlockNumber,
						"is_syncing":    syncStatus.IsSyncing,
						"peer_count":    syncStatus.PeerCount,
					}).Info("Execution client sync status")

					if syncStatus.SyncProgress != nil && syncStatus.SyncProgress.CurrentBlock > 0 {
						percent := float64(syncStatus.SyncProgress.CurrentBlock) / float64(syncStatus.SyncProgress.HighestBlock) * 100
						logrus.WithFields(logrus.Fields{
							"client":         baseClient.Name(),
							"current_block":  syncStatus.SyncProgress.CurrentBlock,
							"highest_block":  syncStatus.SyncProgress.HighestBlock,
							"starting_block": syncStatus.SyncProgress.StartingBlock,
							"progress":       fmt.Sprintf("%.2f%%", percent),
						}).Info("Execution client sync progress")
					}
				}
				time.Sleep(10 * time.Second)
			}
		}(baseClient)
	}

	// Iterate over all consensus clients
	for _, consensusClient := range network.ConsensusClients().All() {
		logrus.WithFields(logrus.Fields{
			"client":         consensusClient.Name(),
			"type":           consensusClient.Type(),
			"version":        consensusClient.Version(),
			"beacon_api_url": consensusClient.BeaconAPIURL(),
			"metrics_url":    consensusClient.MetricsURL(),
			"enr":            consensusClient.ENR(),
		}).Info("Consensus client info")

		// Start goroutine for this consensus client
		go func(consensusClient client.ConsensusClient) {
			for {
				syncStatus, err := getConsensusSyncStatus(ctx, consensusClient)
				if err != nil {
					log.Printf("Failed to get sync status for consensus client %s: %v", consensusClient.Name(), err)
				} else {
					logrus.WithFields(logrus.Fields{
						"client":        consensusClient.Name(),
						"head_slot":     syncStatus.HeadSlot,
						"sync_distance": syncStatus.SyncDistance,
						"is_syncing":    syncStatus.IsSyncing,
						"is_optimistic": syncStatus.IsOptimistic,
						"el_offline":    syncStatus.ElOffline,
					}).Info("Consensus client sync status")
				}
				time.Sleep(10 * time.Second)
			}
		}(consensusClient)
	}

	logrus.WithFields(logrus.Fields{
		"enclave":           network.EnclaveName(),
		"execution_clients": len(network.ExecutionClients().All()),
		"consensus_clients": len(network.ConsensusClients().All()),
	}).Info("Network info")

	// Keep running until interrupted
	select {
	case <-ctx.Done():
		fmt.Println("Context cancelled, shutting down...")
	}
}
