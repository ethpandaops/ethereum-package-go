package client

// ExecutionClient represents a common interface for all execution layer clients
type ExecutionClient interface {
	// Basic information
	Name() string
	Type() Type
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
	clientType  Type
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
func (e *ExecutionClientImpl) Type() Type           { return e.clientType }
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
func NewExecutionClient(clientType Type, name, version, rpcURL, wsURL, engineURL, metricsURL, enode, serviceName, containerID string, p2pPort int) *ExecutionClientImpl {
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

// ExecutionClients holds all execution clients by type
type ExecutionClients struct {
	*Collection[ExecutionClient]
}

// NewExecutionClients creates a new ExecutionClients collection
func NewExecutionClients() *ExecutionClients {
	return &ExecutionClients{
		Collection: NewCollection[ExecutionClient](),
	}
}

// Add adds an execution client to the collection
func (ec *ExecutionClients) Add(client ExecutionClient) {
	ec.Collection.Add(client.Type(), client)
}

// ByType returns all execution clients of a specific type
func (ec *ExecutionClients) ByType(clientType Type) []ExecutionClient {
	return ec.Collection.ByType(clientType)
}