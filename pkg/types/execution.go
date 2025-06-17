package types

// ClientType represents the type of Ethereum client
type ClientType string

const (
	// Execution clients
	ClientGeth       ClientType = "geth"
	ClientBesu       ClientType = "besu"
	ClientNethermind ClientType = "nethermind"
	ClientErigon     ClientType = "erigon"
	ClientReth       ClientType = "reth"
)

// ExecutionClient represents a common interface for all execution layer clients
type ExecutionClient interface {
	// Basic information
	Name() string
	Type() ClientType
	Version() string

	// Network endpoints
	RPCURL() string
	WSURL() string
	EngineURL() string
	MetricsURL() string

	// P2P information
	Enode() string
	P2PPort() int

	// Service information
	ServiceName() string
	ContainerID() string
}

// baseExecutionClient provides common implementation for ExecutionClient
type baseExecutionClient struct {
	name        string
	clientType  ClientType
	version     string
	rpcURL      string
	wsURL       string
	engineURL   string
	metricsURL  string
	enode       string
	p2pPort     int
	serviceName string
	containerID string
}

func (b *baseExecutionClient) Name() string         { return b.name }
func (b *baseExecutionClient) Type() ClientType     { return b.clientType }
func (b *baseExecutionClient) Version() string      { return b.version }
func (b *baseExecutionClient) RPCURL() string       { return b.rpcURL }
func (b *baseExecutionClient) WSURL() string        { return b.wsURL }
func (b *baseExecutionClient) EngineURL() string    { return b.engineURL }
func (b *baseExecutionClient) MetricsURL() string   { return b.metricsURL }
func (b *baseExecutionClient) Enode() string        { return b.enode }
func (b *baseExecutionClient) P2PPort() int         { return b.p2pPort }
func (b *baseExecutionClient) ServiceName() string  { return b.serviceName }
func (b *baseExecutionClient) ContainerID() string  { return b.containerID }

// GethClient represents a Geth execution client
type GethClient struct {
	baseExecutionClient
}

// NewGethClient creates a new Geth client instance
func NewGethClient(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID string, p2pPort int) *GethClient {
	return &GethClient{
		baseExecutionClient: baseExecutionClient{
			name:        name,
			clientType:  ClientGeth,
			version:     version,
			rpcURL:      rpcURL,
			wsURL:       wsURL,
			engineURL:   engineURL,
			metricsURL:  metricsURL,
			enode:       enode,
			p2pPort:     p2pPort,
			serviceName: serviceName,
			containerID: containerID,
		},
	}
}

// BesuClient represents a Besu execution client
type BesuClient struct {
	baseExecutionClient
}

// NewBesuClient creates a new Besu client instance
func NewBesuClient(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID string, p2pPort int) *BesuClient {
	return &BesuClient{
		baseExecutionClient: baseExecutionClient{
			name:        name,
			clientType:  ClientBesu,
			version:     version,
			rpcURL:      rpcURL,
			wsURL:       wsURL,
			engineURL:   engineURL,
			metricsURL:  metricsURL,
			enode:       enode,
			p2pPort:     p2pPort,
			serviceName: serviceName,
			containerID: containerID,
		},
	}
}

// NethermindClient represents a Nethermind execution client
type NethermindClient struct {
	baseExecutionClient
}

// NewNethermindClient creates a new Nethermind client instance
func NewNethermindClient(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID string, p2pPort int) *NethermindClient {
	return &NethermindClient{
		baseExecutionClient: baseExecutionClient{
			name:        name,
			clientType:  ClientNethermind,
			version:     version,
			rpcURL:      rpcURL,
			wsURL:       wsURL,
			engineURL:   engineURL,
			metricsURL:  metricsURL,
			enode:       enode,
			p2pPort:     p2pPort,
			serviceName: serviceName,
			containerID: containerID,
		},
	}
}

// ErigonClient represents an Erigon execution client
type ErigonClient struct {
	baseExecutionClient
}

// NewErigonClient creates a new Erigon client instance
func NewErigonClient(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID string, p2pPort int) *ErigonClient {
	return &ErigonClient{
		baseExecutionClient: baseExecutionClient{
			name:        name,
			clientType:  ClientErigon,
			version:     version,
			rpcURL:      rpcURL,
			wsURL:       wsURL,
			engineURL:   engineURL,
			metricsURL:  metricsURL,
			enode:       enode,
			p2pPort:     p2pPort,
			serviceName: serviceName,
			containerID: containerID,
		},
	}
}

// RethClient represents a Reth execution client
type RethClient struct {
	baseExecutionClient
}

// NewRethClient creates a new Reth client instance
func NewRethClient(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID string, p2pPort int) *RethClient {
	return &RethClient{
		baseExecutionClient: baseExecutionClient{
			name:        name,
			clientType:  ClientReth,
			version:     version,
			rpcURL:      rpcURL,
			wsURL:       wsURL,
			engineURL:   engineURL,
			metricsURL:  metricsURL,
			enode:       enode,
			p2pPort:     p2pPort,
			serviceName: serviceName,
			containerID: containerID,
		},
	}
}

// ExecutionClients holds all execution clients by type
type ExecutionClients struct {
	clients map[ClientType][]ExecutionClient
}

// NewExecutionClients creates a new ExecutionClients collection
func NewExecutionClients() *ExecutionClients {
	return &ExecutionClients{
		clients: make(map[ClientType][]ExecutionClient),
	}
}

// Add adds an execution client to the collection
func (ec *ExecutionClients) Add(client ExecutionClient) {
	ec.clients[client.Type()] = append(ec.clients[client.Type()], client)
}

// All returns all execution clients
func (ec *ExecutionClients) All() []ExecutionClient {
	var all []ExecutionClient
	for _, clients := range ec.clients {
		all = append(all, clients...)
	}
	return all
}

// Geth returns all Geth clients
func (ec *ExecutionClients) Geth() []*GethClient {
	var gethClients []*GethClient
	for _, client := range ec.clients[ClientGeth] {
		if gc, ok := client.(*GethClient); ok {
			gethClients = append(gethClients, gc)
		}
	}
	return gethClients
}

// Besu returns all Besu clients
func (ec *ExecutionClients) Besu() []*BesuClient {
	var besuClients []*BesuClient
	for _, client := range ec.clients[ClientBesu] {
		if bc, ok := client.(*BesuClient); ok {
			besuClients = append(besuClients, bc)
		}
	}
	return besuClients
}

// Nethermind returns all Nethermind clients
func (ec *ExecutionClients) Nethermind() []*NethermindClient {
	var nethermindClients []*NethermindClient
	for _, client := range ec.clients[ClientNethermind] {
		if nc, ok := client.(*NethermindClient); ok {
			nethermindClients = append(nethermindClients, nc)
		}
	}
	return nethermindClients
}

// Erigon returns all Erigon clients
func (ec *ExecutionClients) Erigon() []*ErigonClient {
	var erigonClients []*ErigonClient
	for _, client := range ec.clients[ClientErigon] {
		if ec, ok := client.(*ErigonClient); ok {
			erigonClients = append(erigonClients, ec)
		}
	}
	return erigonClients
}

// Reth returns all Reth clients
func (ec *ExecutionClients) Reth() []*RethClient {
	var rethClients []*RethClient
	for _, client := range ec.clients[ClientReth] {
		if rc, ok := client.(*RethClient); ok {
			rethClients = append(rethClients, rc)
		}
	}
	return rethClients
}