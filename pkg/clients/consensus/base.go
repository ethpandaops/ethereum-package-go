package consensus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ethpandaops/ethereum-package-go/pkg/types"
)

// ClientConfig holds configuration for creating a consensus client
type ClientConfig struct {
	Name       string
	BeaconURL  string
	P2PURL     string
	MetricsURL string
	ENR        string
	PeerID     string
}

// BaseConsensusClient provides common functionality for all consensus clients
type BaseConsensusClient struct {
	name       string
	clientType types.ClientType
	beaconURL  string
	p2pURL     string
	metricsURL string
	enr        string
	peerID     string
	httpClient *http.Client
}

// NewBaseConsensusClient creates a new base consensus client
func NewBaseConsensusClient(config ClientConfig) *BaseConsensusClient {
	return &BaseConsensusClient{
		name:       config.Name,
		beaconURL:  config.BeaconURL,
		p2pURL:     config.P2PURL,
		metricsURL: config.MetricsURL,
		enr:        config.ENR,
		peerID:     config.PeerID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the client name
func (b *BaseConsensusClient) Name() string {
	return b.name
}

// Type returns the client type
func (b *BaseConsensusClient) Type() types.ClientType {
	return b.clientType
}

// BeaconURL returns the Beacon API URL
func (b *BaseConsensusClient) BeaconURL() string {
	return b.beaconURL
}

// P2PAddr returns the P2P address
func (b *BaseConsensusClient) P2PAddr() string {
	return b.p2pURL
}

// ENR returns the node's ENR
func (b *BaseConsensusClient) ENR() string {
	return b.enr
}

// PeerID returns the node's peer ID
func (b *BaseConsensusClient) PeerID() string {
	return b.peerID
}

// Metrics returns the metrics URL
func (b *BaseConsensusClient) Metrics() string {
	return b.metricsURL
}

// makeAPIRequest makes a REST API request
func (b *BaseConsensusClient) makeAPIRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	if b.beaconURL == "" {
		return nil, fmt.Errorf("beacon URL not configured")
	}

	url := fmt.Sprintf("%s%s", b.beaconURL, path)
	
	var reqBody io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		reqBody = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}

// GetNodeVersion gets the node version
func (b *BaseConsensusClient) GetNodeVersion(ctx context.Context) (*NodeVersion, error) {
	resp, err := b.makeAPIRequest(ctx, "GET", "/eth/v1/node/version", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get node version: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data NodeVersion `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode version response: %w", err)
	}

	return &result.Data, nil
}

// GetNodeSyncing gets the node sync status
func (b *BaseConsensusClient) GetNodeSyncing(ctx context.Context) (*SyncStatus, error) {
	resp, err := b.makeAPIRequest(ctx, "GET", "/eth/v1/node/syncing", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync status: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data SyncStatus `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode sync response: %w", err)
	}

	return &result.Data, nil
}

// GetPeers gets connected peers
func (b *BaseConsensusClient) GetPeers(ctx context.Context) ([]Peer, error) {
	resp, err := b.makeAPIRequest(ctx, "GET", "/eth/v1/node/peers", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get peers: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []Peer `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode peers response: %w", err)
	}

	return result.Data, nil
}

// GetHead gets the chain head
func (b *BaseConsensusClient) GetHead(ctx context.Context) (*BeaconBlock, error) {
	resp, err := b.makeAPIRequest(ctx, "GET", "/eth/v1/beacon/blocks/head", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get head block: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data BeaconBlock `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode head response: %w", err)
	}

	return &result.Data, nil
}

// WaitForSync waits for the client to finish syncing
func (b *BaseConsensusClient) WaitForSync(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			syncStatus, err := b.GetNodeSyncing(ctx)
			if err != nil {
				return fmt.Errorf("failed to check sync status: %w", err)
			}
			if !syncStatus.IsSyncing {
				return nil
			}
		}
	}
}

// NodeVersion represents node version information
type NodeVersion struct {
	Version string `json:"version"`
}

// SyncStatus represents sync status information
type SyncStatus struct {
	HeadSlot     string `json:"head_slot"`
	SyncDistance string `json:"sync_distance"`
	IsSyncing    bool   `json:"is_syncing"`
	IsOptimistic bool   `json:"is_optimistic"`
	ElOffline    bool   `json:"el_offline"`
}

// Peer represents peer information
type Peer struct {
	PeerID    string `json:"peer_id"`
	ENR       string `json:"enr"`
	LastSeen  string `json:"last_seen"`
	State     string `json:"state"`
	Direction string `json:"direction"`
}

// BeaconBlock represents a beacon block
type BeaconBlock struct {
	Message struct {
		Slot          string `json:"slot"`
		ProposerIndex string `json:"proposer_index"`
		ParentRoot    string `json:"parent_root"`
		StateRoot     string `json:"state_root"`
		Body          struct {
			ExecutionPayload struct {
				BlockNumber string `json:"block_number"`
				BlockHash   string `json:"block_hash"`
			} `json:"execution_payload"`
		} `json:"body"`
	} `json:"message"`
}