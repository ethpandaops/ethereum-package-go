package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ConsensusClient represents a common interface for all consensus layer clients
type ConsensusClient interface {
	// Basic information
	Name() string
	Type() Type
	Version() string

	// Network endpoints
	BeaconAPIURL() string
	MetricsURL() string

	// P2P information
	P2PPort() int
	ENR() string
	PeerID() string

	// Service information
	ServiceName() string
	ContainerID() string

	// Live peer ID fetching
	FetchPeerID(ctx context.Context) (string, error)
}

// ConsensusClientImpl is a generic implementation of the ConsensusClient interface
type ConsensusClientImpl struct {
	name         string
	clientType   Type
	version      string
	beaconAPIURL string
	metricsURL   string
	p2pPort      int
	enr          string
	peerID       string
	serviceName  string
	containerID  string
}

func (c *ConsensusClientImpl) Name() string         { return c.name }
func (c *ConsensusClientImpl) Type() Type           { return c.clientType }
func (c *ConsensusClientImpl) Version() string      { return c.version }
func (c *ConsensusClientImpl) BeaconAPIURL() string { return c.beaconAPIURL }
func (c *ConsensusClientImpl) MetricsURL() string   { return c.metricsURL }
func (c *ConsensusClientImpl) P2PPort() int         { return c.p2pPort }
func (c *ConsensusClientImpl) ENR() string          { return c.enr }
func (c *ConsensusClientImpl) PeerID() string       { return c.peerID }
func (c *ConsensusClientImpl) ServiceName() string  { return c.serviceName }
func (c *ConsensusClientImpl) ContainerID() string  { return c.containerID }

// NodeIdentityResponse represents the response from /eth/v1/node/identity
type NodeIdentityResponse struct {
	Data struct {
		PeerID             string   `json:"peer_id"`
		ENR                string   `json:"enr"`
		P2PAddresses       []string `json:"p2p_addresses"`
		DiscoveryAddresses []string `json:"discovery_addresses"`
		Metadata           struct {
			SeqNumber       string `json:"seq_number"`
			Attnets         string `json:"attnets"`
			SyncCommitteeNets string `json:"syncnets,omitempty"`
		} `json:"metadata"`
	} `json:"data"`
}

// FetchPeerID fetches the live peer ID from the beacon API using /eth/v1/node/identity
func (c *ConsensusClientImpl) FetchPeerID(ctx context.Context) (string, error) {
	beaconURL := c.BeaconAPIURL()
	if beaconURL == "" {
		return "", fmt.Errorf("beacon API URL is empty")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Build the endpoint URL
	endpoint := fmt.Sprintf("%s/eth/v1/node/identity", beaconURL)

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request to %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("beacon API returned status %d for endpoint %s", resp.StatusCode, endpoint)
	}

	// Parse the response
	var nodeIdentity NodeIdentityResponse
	if err := json.NewDecoder(resp.Body).Decode(&nodeIdentity); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract peer ID
	peerID := nodeIdentity.Data.PeerID
	if peerID == "" {
		return "", fmt.Errorf("peer_id is empty in response")
	}

	return peerID, nil
}

// NewConsensusClient creates a new generic consensus client instance
func NewConsensusClient(clientType Type, name, version, beaconAPIURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) *ConsensusClientImpl {
	return &ConsensusClientImpl{
		name:         name,
		clientType:   clientType,
		version:      version,
		beaconAPIURL: beaconAPIURL,
		metricsURL:   metricsURL,
		p2pPort:      p2pPort,
		enr:          enr,
		peerID:       peerID,
		serviceName:  serviceName,
		containerID:  containerID,
	}
}

// ConsensusClients holds all consensus clients by type
type ConsensusClients struct {
	*Collection[ConsensusClient]
}

// NewConsensusClients creates a new ConsensusClients collection
func NewConsensusClients() *ConsensusClients {
	return &ConsensusClients{
		Collection: NewCollection[ConsensusClient](),
	}
}

// Add adds a consensus client to the collection
func (cc *ConsensusClients) Add(client ConsensusClient) {
	cc.Collection.Add(client.Type(), client)
}

// ByType returns all consensus clients of a specific type
func (cc *ConsensusClients) ByType(clientType Type) []ConsensusClient {
	return cc.Collection.ByType(clientType)
}

// PeerIDs fetches peer IDs for all consensus clients in the collection
func (cc *ConsensusClients) PeerIDs(ctx context.Context) (map[string]string, error) {
	clients := cc.All()
	peerIds := make(map[string]string)

	for _, client := range clients {
		peerID, err := client.FetchPeerID(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch peer ID for client %s: %w", client.Name(), err)
		}
		peerIds[client.Name()] = peerID
	}

	return peerIds, nil
}

// PeerIDsByType fetches peer IDs for all consensus clients of a specific type
func (cc *ConsensusClients) PeerIDsByType(ctx context.Context, clientType Type) (map[string]string, error) {
	clients := cc.ByType(clientType)
	peerIds := make(map[string]string)

	for _, client := range clients {
		peerID, err := client.FetchPeerID(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch peer ID for client %s: %w", client.Name(), err)
		}
		peerIds[client.Name()] = peerID
	}

	return peerIds, nil
}