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
	
	// Unknown client
	ClientUnknown    ClientType = "unknown"
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

// ExecutionClientImpl is a generic implementation of the ExecutionClient interface
type ExecutionClientImpl struct {
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

func (e *ExecutionClientImpl) Name() string         { return e.name }
func (e *ExecutionClientImpl) Type() ClientType     { return e.clientType }
func (e *ExecutionClientImpl) Version() string      { return e.version }
func (e *ExecutionClientImpl) RPCURL() string       { return e.rpcURL }
func (e *ExecutionClientImpl) WSURL() string        { return e.wsURL }
func (e *ExecutionClientImpl) EngineURL() string    { return e.engineURL }
func (e *ExecutionClientImpl) MetricsURL() string   { return e.metricsURL }
func (e *ExecutionClientImpl) Enode() string        { return e.enode }
func (e *ExecutionClientImpl) P2PPort() int         { return e.p2pPort }
func (e *ExecutionClientImpl) ServiceName() string  { return e.serviceName }
func (e *ExecutionClientImpl) ContainerID() string  { return e.containerID }

// NewExecutionClient creates a new generic execution client instance
func NewExecutionClient(clientType ClientType, name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID string, p2pPort int) *ExecutionClientImpl {
	return &ExecutionClientImpl{
		name:        name,
		clientType:  clientType,
		version:     version,
		rpcURL:      rpcURL,
		wsURL:       wsURL,
		engineURL:   engineURL,
		metricsURL:  metricsURL,
		enode:       enode,
		p2pPort:     p2pPort,
		serviceName: serviceName,
		containerID: containerID,
	}
}

// Deprecated: Use NewExecutionClient instead
type GethClient = ExecutionClientImpl

// Deprecated: Use NewExecutionClient with ClientGeth instead
func NewGethClient(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID string, p2pPort int) *ExecutionClientImpl {
	return NewExecutionClient(ClientGeth, name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID, p2pPort)
}

// Deprecated: Use NewExecutionClient instead
type BesuClient = ExecutionClientImpl

// Deprecated: Use NewExecutionClient with ClientBesu instead
func NewBesuClient(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID string, p2pPort int) *ExecutionClientImpl {
	return NewExecutionClient(ClientBesu, name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID, p2pPort)
}

// Deprecated: Use NewExecutionClient instead
type NethermindClient = ExecutionClientImpl

// Deprecated: Use NewExecutionClient with ClientNethermind instead
func NewNethermindClient(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID string, p2pPort int) *ExecutionClientImpl {
	return NewExecutionClient(ClientNethermind, name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID, p2pPort)
}

// Deprecated: Use NewExecutionClient instead
type ErigonClient = ExecutionClientImpl

// Deprecated: Use NewExecutionClient with ClientErigon instead
func NewErigonClient(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID string, p2pPort int) *ExecutionClientImpl {
	return NewExecutionClient(ClientErigon, name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID, p2pPort)
}

// Deprecated: Use NewExecutionClient instead
type RethClient = ExecutionClientImpl

// Deprecated: Use NewExecutionClient with ClientReth instead
func NewRethClient(name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID string, p2pPort int) *ExecutionClientImpl {
	return NewExecutionClient(ClientReth, name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID, p2pPort)
}

// ExecutionClients holds all execution clients by type
type ExecutionClients struct {
	*ClientCollection[ExecutionClient]
}

// NewExecutionClients creates a new ExecutionClients collection
func NewExecutionClients() *ExecutionClients {
	return &ExecutionClients{
		ClientCollection: NewClientCollection[ExecutionClient](),
	}
}

// Add adds an execution client to the collection
func (ec *ExecutionClients) Add(client ExecutionClient) {
	ec.ClientCollection.Add(client.Type(), client)
}

// ByType returns all execution clients of a specific type
func (ec *ExecutionClients) ByType(clientType ClientType) []ExecutionClient {
	return ec.ClientCollection.ByType(clientType)
}

// Geth returns all Geth clients
// Deprecated: Use ByType(ClientGeth) instead
func (ec *ExecutionClients) Geth() []*ExecutionClientImpl {
	var gethClients []*ExecutionClientImpl
	for _, client := range ec.ByType(ClientGeth) {
		if gc, ok := client.(*ExecutionClientImpl); ok {
			gethClients = append(gethClients, gc)
		}
	}
	return gethClients
}

// Besu returns all Besu clients
// Deprecated: Use ByType(ClientBesu) instead
func (ec *ExecutionClients) Besu() []*ExecutionClientImpl {
	var besuClients []*ExecutionClientImpl
	for _, client := range ec.ByType(ClientBesu) {
		if bc, ok := client.(*ExecutionClientImpl); ok {
			besuClients = append(besuClients, bc)
		}
	}
	return besuClients
}

// Nethermind returns all Nethermind clients
// Deprecated: Use ByType(ClientNethermind) instead
func (ec *ExecutionClients) Nethermind() []*ExecutionClientImpl {
	var nethermindClients []*ExecutionClientImpl
	for _, client := range ec.ByType(ClientNethermind) {
		if nc, ok := client.(*ExecutionClientImpl); ok {
			nethermindClients = append(nethermindClients, nc)
		}
	}
	return nethermindClients
}

// Erigon returns all Erigon clients
// Deprecated: Use ByType(ClientErigon) instead
func (ec *ExecutionClients) Erigon() []*ExecutionClientImpl {
	var erigonClients []*ExecutionClientImpl
	for _, client := range ec.ByType(ClientErigon) {
		if ec, ok := client.(*ExecutionClientImpl); ok {
			erigonClients = append(erigonClients, ec)
		}
	}
	return erigonClients
}

// Reth returns all Reth clients
// Deprecated: Use ByType(ClientReth) instead
func (ec *ExecutionClients) Reth() []*ExecutionClientImpl {
	var rethClients []*ExecutionClientImpl
	for _, client := range ec.ByType(ClientReth) {
		if rc, ok := client.(*ExecutionClientImpl); ok {
			rethClients = append(rethClients, rc)
		}
	}
	return rethClients
}