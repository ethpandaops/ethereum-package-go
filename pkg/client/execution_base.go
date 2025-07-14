package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ClientConfig holds configuration for creating an execution client
type ClientConfig struct {
	Name       string
	RPCURL     string
	WSURL      string
	EngineURL  string
	P2PURL     string
	MetricsURL string
	Enode      string
	Image      string
	Entrypoint []string
	Cmd        []string
}

// BaseExecutionClient provides common functionality for all execution clients
type BaseExecutionClient struct {
	name       string
	clientType Type
	rpcURL     string
	wsURL      string
	engineURL  string
	p2pURL     string
	metricsURL string
	enode      string
	httpClient *http.Client
}

// NewBaseExecutionClient creates a new base execution client
func NewBaseExecutionClient(config ClientConfig) *BaseExecutionClient {
	return &BaseExecutionClient{
		name:       config.Name,
		rpcURL:     config.RPCURL,
		wsURL:      config.WSURL,
		engineURL:  config.EngineURL,
		p2pURL:     config.P2PURL,
		metricsURL: config.MetricsURL,
		enode:      config.Enode,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the client name
func (b *BaseExecutionClient) Name() string {
	return b.name
}

// Type returns the client type
func (b *BaseExecutionClient) Type() Type {
	return b.clientType
}

// RPCURL returns the RPC URL
func (b *BaseExecutionClient) RPCURL() string {
	return b.rpcURL
}

// WSURL returns the WebSocket URL
func (b *BaseExecutionClient) WSURL() string {
	return b.wsURL
}

// EngineURL returns the Engine API URL
func (b *BaseExecutionClient) EngineURL() string {
	return b.engineURL
}

// P2PURL returns the P2P URL
func (b *BaseExecutionClient) P2PURL() string {
	return b.p2pURL
}

// Enode returns the node's enode
func (b *BaseExecutionClient) Enode() string {
	return b.enode
}

// Metrics returns the metrics URL
func (b *BaseExecutionClient) Metrics() string {
	return b.metricsURL
}

// RPCResponse represents a JSON-RPC response
type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}

// makeRPCRequest makes a JSON-RPC request
func (b *BaseExecutionClient) makeRPCRequest(ctx context.Context, req interface{}) (*RPCResponse, error) {
	if b.rpcURL == "" {
		return nil, fmt.Errorf("RPC URL not configured")
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", b.rpcURL, bytes.NewReader(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := b.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var rpcResp RPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, rpcResp.Error
	}

	return &rpcResp, nil
}

// GetBlockNumber gets the current block number
func (b *BaseExecutionClient) GetBlockNumber(ctx context.Context) (uint64, error) {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_blockNumber",
		"params":  []interface{}{},
		"id":      1,
	}

	resp, err := b.makeRPCRequest(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to get block number: %w", err)
	}

	var blockNumberHex string
	if err := json.Unmarshal(resp.Result, &blockNumberHex); err != nil {
		return 0, fmt.Errorf("failed to parse block number: %w", err)
	}

	var blockNumber uint64
	if _, err := fmt.Sscanf(blockNumberHex, "0x%x", &blockNumber); err != nil {
		return 0, fmt.Errorf("failed to parse hex block number: %w", err)
	}

	return blockNumber, nil
}

// IsSyncing checks if the client is syncing
func (b *BaseExecutionClient) IsSyncing(ctx context.Context) (bool, error) {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_syncing",
		"params":  []interface{}{},
		"id":      1,
	}

	resp, err := b.makeRPCRequest(ctx, req)
	if err != nil {
		return false, fmt.Errorf("failed to check sync status: %w", err)
	}

	// eth_syncing returns false when not syncing, or a sync object when syncing
	var result interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return false, fmt.Errorf("failed to parse sync status: %w", err)
	}

	// If result is false, not syncing
	if syncing, ok := result.(bool); ok {
		return syncing, nil
	}

	// If result is an object, syncing
	return true, nil
}

// WaitForSync waits for the client to finish syncing
func (b *BaseExecutionClient) WaitForSync(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			syncing, err := b.IsSyncing(ctx)
			if err != nil {
				return fmt.Errorf("failed to check sync status: %w", err)
			}
			if !syncing {
				return nil
			}
		}
	}
}

// SyncProgress gets the sync progress object if syncing, returns nil if not syncing
func (b *BaseExecutionClient) SyncProgress(ctx context.Context) (*SyncProgress, error) {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_syncing",
		"params":  []interface{}{},
		"id":      1,
	}

	resp, err := b.makeRPCRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync progress: %w", err)
	}

	// eth_syncing returns false when not syncing, or a sync object when syncing
	var result interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse sync progress: %w", err)
	}

	// If result is false, not syncing
	if syncing, ok := result.(bool); ok && !syncing {
		return nil, nil
	}

	// If result is an object, parse it as sync progress with hex values
	var rawProgress struct {
		CurrentBlock  string `json:"currentBlock"`
		HighestBlock  string `json:"highestBlock"`
		StartingBlock string `json:"startingBlock"`
	}
	if err := json.Unmarshal(resp.Result, &rawProgress); err != nil {
		return nil, fmt.Errorf("failed to parse sync progress object: %w", err)
	}

	// Parse hex strings to uint64
	progress := &SyncProgress{}
	if _, err := fmt.Sscanf(rawProgress.CurrentBlock, "0x%x", &progress.CurrentBlock); err != nil {
		return nil, fmt.Errorf("failed to parse currentBlock hex: %w", err)
	}
	if _, err := fmt.Sscanf(rawProgress.HighestBlock, "0x%x", &progress.HighestBlock); err != nil {
		return nil, fmt.Errorf("failed to parse highestBlock hex: %w", err)
	}
	if _, err := fmt.Sscanf(rawProgress.StartingBlock, "0x%x", &progress.StartingBlock); err != nil {
		return nil, fmt.Errorf("failed to parse startingBlock hex: %w", err)
	}

	return progress, nil
}

// NodeInfo represents node information
type NodeInfo struct {
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Enode string                 `json:"enode"`
	ENR   string                 `json:"enr"`
	IP    string                 `json:"ip"`
	Ports map[string]interface{} `json:"ports"`
}

// PeerInfo represents peer information
type PeerInfo struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Enode     string                 `json:"enode"`
	ENR       string                 `json:"enr"`
	Caps      []string               `json:"caps"`
	Network   map[string]interface{} `json:"network"`
	Protocols map[string]interface{} `json:"protocols"`
}

// TraceResult represents the result of a transaction trace
type TraceResult struct {
	Type    string        `json:"type"`
	From    string        `json:"from"`
	To      string        `json:"to"`
	Value   string        `json:"value"`
	Gas     string        `json:"gas"`
	GasUsed string        `json:"gasUsed"`
	Input   string        `json:"input"`
	Output  string        `json:"output"`
	Error   string        `json:"error,omitempty"`
	Calls   []TraceResult `json:"calls,omitempty"`
}

// TxPoolStatus represents transaction pool status
type TxPoolStatus struct {
	Pending string `json:"pending"`
	Queued  string `json:"queued"`
}

// SyncProgress represents the sync progress when syncing
type SyncProgress struct {
	CurrentBlock  uint64 `json:"currentBlock"`
	HighestBlock  uint64 `json:"highestBlock"`
	StartingBlock uint64 `json:"startingBlock"`
}

// GetPeerCount gets the number of connected peers
func (b *BaseExecutionClient) GetPeerCount(ctx context.Context) (int, error) {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "net_peerCount",
		"params":  []interface{}{},
		"id":      1,
	}

	resp, err := b.makeRPCRequest(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to get peer count: %w", err)
	}

	var peerCountHex string
	if err := json.Unmarshal(resp.Result, &peerCountHex); err != nil {
		return 0, fmt.Errorf("failed to parse peer count: %w", err)
	}

	var peerCount int
	if _, err := fmt.Sscanf(peerCountHex, "0x%x", &peerCount); err != nil {
		return 0, fmt.Errorf("failed to parse hex peer count: %w", err)
	}

	return peerCount, nil
}
