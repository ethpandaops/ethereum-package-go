package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	eth2http "github.com/attestantio/go-eth2-client/http"
)

// GetAttestantClient returns an attestant go-eth2 client for the consensus client
func GetAttestantClient(ctx context.Context, client ConsensusClient) (eth2client.Service, error) {
	if client == nil {
		return nil, fmt.Errorf("consensus client is nil")
	}

	beaconURL := client.BeaconAPIURL()
	if beaconURL == "" {
		return nil, fmt.Errorf("beacon API URL is empty")
	}

	// Create HTTP client with reasonable timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create attestant client
	attestantClient, err := eth2http.New(ctx,
		eth2http.WithAddress(beaconURL),
		eth2http.WithHTTPClient(httpClient),
		eth2http.WithTimeout(30*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create attestant client: %w", err)
	}

	return attestantClient, nil
}

// GetAttestantClientWithCustomHTTP returns an attestant go-eth2 client with a custom HTTP client
func GetAttestantClientWithCustomHTTP(ctx context.Context, client ConsensusClient, httpClient *http.Client) (eth2client.Service, error) {
	if client == nil {
		return nil, fmt.Errorf("consensus client is nil")
	}

	beaconURL := client.BeaconAPIURL()
	if beaconURL == "" {
		return nil, fmt.Errorf("beacon API URL is empty")
	}

	// Create attestant client with custom HTTP client
	attestantClient, err := eth2http.New(ctx,
		eth2http.WithAddress(beaconURL),
		eth2http.WithHTTPClient(httpClient),
		eth2http.WithTimeout(httpClient.Timeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create attestant client: %w", err)
	}

	return attestantClient, nil
}