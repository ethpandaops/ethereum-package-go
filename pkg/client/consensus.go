package client

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