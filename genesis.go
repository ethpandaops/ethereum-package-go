package ethereum

import (
	"context"
	"fmt"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	"github.com/ethpandaops/ethereum-package-go/pkg/client"
	"github.com/ethpandaops/ethereum-package-go/pkg/network"
)

// WaitForGenesis waits until the genesis time of the network
// It gets the first consensus client, retrieves genesis details via the attestant client,
// and sleeps until the genesis timestamp
func WaitForGenesis(ctx context.Context, net network.Network) error {
	// Get all consensus clients
	consensusClients := net.ConsensusClients().All()
	if len(consensusClients) == 0 {
		return fmt.Errorf("no consensus clients available")
	}

	// Get the first consensus client
	firstClient := consensusClients[0]

	// Get attestant client for the consensus client
	attestantClient, err := client.GetAttestantClient(ctx, firstClient)
	if err != nil {
		return fmt.Errorf("failed to get attestant client: %w", err)
	}

	// Type assert to GenesisProvider
	genesisProvider, ok := attestantClient.(eth2client.GenesisProvider)
	if !ok {
		return fmt.Errorf("client does not implement GenesisProvider")
	}

	// Get genesis details
	genesisResponse, err := genesisProvider.Genesis(ctx, &api.GenesisOpts{})
	if err != nil {
		return fmt.Errorf("failed to get genesis details: %w", err)
	}

	// Calculate time until genesis
	genesisTime := genesisResponse.Data.GenesisTime
	now := time.Now()
	
	if genesisTime.Before(now) {
		// Genesis already happened
		return nil
	}

	// Calculate sleep duration
	sleepDuration := genesisTime.Sub(now)
	
	// Log the wait time
	fmt.Printf("Waiting for genesis at %s (in %s)\n", genesisTime.Format(time.RFC3339), sleepDuration)

	// Sleep until genesis
	timer := time.NewTimer(sleepDuration)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}