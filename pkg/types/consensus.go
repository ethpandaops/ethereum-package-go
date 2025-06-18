package types

// Consensus client types
const (
	ClientLighthouse ClientType = "lighthouse"
	ClientTeku       ClientType = "teku"
	ClientPrysm      ClientType = "prysm"
	ClientNimbus     ClientType = "nimbus"
	ClientLodestar   ClientType = "lodestar"
	ClientGrandine   ClientType = "grandine"
)

// ConsensusClient represents a common interface for all consensus layer clients
type ConsensusClient interface {
	// Basic information
	Name() string
	Type() ClientType
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
	clientType   ClientType
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
func (c *ConsensusClientImpl) Type() ClientType     { return c.clientType }
func (c *ConsensusClientImpl) Version() string      { return c.version }
func (c *ConsensusClientImpl) BeaconAPIURL() string { return c.beaconAPIURL }
func (c *ConsensusClientImpl) MetricsURL() string   { return c.metricsURL }
func (c *ConsensusClientImpl) P2PPort() int         { return c.p2pPort }
func (c *ConsensusClientImpl) ENR() string          { return c.enr }
func (c *ConsensusClientImpl) PeerID() string       { return c.peerID }
func (c *ConsensusClientImpl) ServiceName() string  { return c.serviceName }
func (c *ConsensusClientImpl) ContainerID() string  { return c.containerID }

// NewConsensusClient creates a new generic consensus client instance
func NewConsensusClient(clientType ClientType, name, version, beaconAPIURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) *ConsensusClientImpl {
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

// Deprecated: Use NewConsensusClient instead
type LighthouseClient = ConsensusClientImpl

// Deprecated: Use NewConsensusClient with ClientLighthouse instead
func NewLighthouseClient(name, version, beaconAPIURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) *ConsensusClientImpl {
	return NewConsensusClient(ClientLighthouse, name, version, beaconAPIURL, metricsURL, enr, peerID, serviceName, containerID, p2pPort)
}

// Deprecated: Use NewConsensusClient instead
type TekuClient = ConsensusClientImpl

// Deprecated: Use NewConsensusClient with ClientTeku instead
func NewTekuClient(name, version, beaconAPIURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) *ConsensusClientImpl {
	return NewConsensusClient(ClientTeku, name, version, beaconAPIURL, metricsURL, enr, peerID, serviceName, containerID, p2pPort)
}

// Deprecated: Use NewConsensusClient instead
type PrysmClient = ConsensusClientImpl

// Deprecated: Use NewConsensusClient with ClientPrysm instead
func NewPrysmClient(name, version, beaconAPIURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) *ConsensusClientImpl {
	return NewConsensusClient(ClientPrysm, name, version, beaconAPIURL, metricsURL, enr, peerID, serviceName, containerID, p2pPort)
}

// Deprecated: Use NewConsensusClient instead
type NimbusClient = ConsensusClientImpl

// Deprecated: Use NewConsensusClient with ClientNimbus instead
func NewNimbusClient(name, version, beaconAPIURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) *ConsensusClientImpl {
	return NewConsensusClient(ClientNimbus, name, version, beaconAPIURL, metricsURL, enr, peerID, serviceName, containerID, p2pPort)
}

// Deprecated: Use NewConsensusClient instead
type LodestarClient = ConsensusClientImpl

// Deprecated: Use NewConsensusClient with ClientLodestar instead
func NewLodestarClient(name, version, beaconAPIURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) *ConsensusClientImpl {
	return NewConsensusClient(ClientLodestar, name, version, beaconAPIURL, metricsURL, enr, peerID, serviceName, containerID, p2pPort)
}

// Deprecated: Use NewConsensusClient instead
type GrandineClient = ConsensusClientImpl

// Deprecated: Use NewConsensusClient with ClientGrandine instead
func NewGrandineClient(name, version, beaconAPIURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) *ConsensusClientImpl {
	return NewConsensusClient(ClientGrandine, name, version, beaconAPIURL, metricsURL, enr, peerID, serviceName, containerID, p2pPort)
}

// ConsensusClients holds all consensus clients by type
type ConsensusClients struct {
	*ClientCollection[ConsensusClient]
}

// NewConsensusClients creates a new ConsensusClients collection
func NewConsensusClients() *ConsensusClients {
	return &ConsensusClients{
		ClientCollection: NewClientCollection[ConsensusClient](),
	}
}

// Add adds a consensus client to the collection
func (cc *ConsensusClients) Add(client ConsensusClient) {
	cc.ClientCollection.Add(client.Type(), client)
}

// ByType returns all consensus clients of a specific type
func (cc *ConsensusClients) ByType(clientType ClientType) []ConsensusClient {
	return cc.ClientCollection.ByType(clientType)
}

// Lighthouse returns all Lighthouse clients
// Deprecated: Use ByType(ClientLighthouse) instead
func (cc *ConsensusClients) Lighthouse() []*ConsensusClientImpl {
	var lighthouseClients []*ConsensusClientImpl
	for _, client := range cc.ByType(ClientLighthouse) {
		if lc, ok := client.(*ConsensusClientImpl); ok {
			lighthouseClients = append(lighthouseClients, lc)
		}
	}
	return lighthouseClients
}

// Teku returns all Teku clients
// Deprecated: Use ByType(ClientTeku) instead
func (cc *ConsensusClients) Teku() []*ConsensusClientImpl {
	var tekuClients []*ConsensusClientImpl
	for _, client := range cc.ByType(ClientTeku) {
		if tc, ok := client.(*ConsensusClientImpl); ok {
			tekuClients = append(tekuClients, tc)
		}
	}
	return tekuClients
}

// Prysm returns all Prysm clients
// Deprecated: Use ByType(ClientPrysm) instead
func (cc *ConsensusClients) Prysm() []*ConsensusClientImpl {
	var prysmClients []*ConsensusClientImpl
	for _, client := range cc.ByType(ClientPrysm) {
		if pc, ok := client.(*ConsensusClientImpl); ok {
			prysmClients = append(prysmClients, pc)
		}
	}
	return prysmClients
}

// Nimbus returns all Nimbus clients
// Deprecated: Use ByType(ClientNimbus) instead
func (cc *ConsensusClients) Nimbus() []*ConsensusClientImpl {
	var nimbusClients []*ConsensusClientImpl
	for _, client := range cc.ByType(ClientNimbus) {
		if nc, ok := client.(*ConsensusClientImpl); ok {
			nimbusClients = append(nimbusClients, nc)
		}
	}
	return nimbusClients
}

// Lodestar returns all Lodestar clients
// Deprecated: Use ByType(ClientLodestar) instead
func (cc *ConsensusClients) Lodestar() []*ConsensusClientImpl {
	var lodestarClients []*ConsensusClientImpl
	for _, client := range cc.ByType(ClientLodestar) {
		if lc, ok := client.(*ConsensusClientImpl); ok {
			lodestarClients = append(lodestarClients, lc)
		}
	}
	return lodestarClients
}

// Grandine returns all Grandine clients
// Deprecated: Use ByType(ClientGrandine) instead
func (cc *ConsensusClients) Grandine() []*ConsensusClientImpl {
	var grandineClients []*ConsensusClientImpl
	for _, client := range cc.ByType(ClientGrandine) {
		if gc, ok := client.(*ConsensusClientImpl); ok {
			grandineClients = append(grandineClients, gc)
		}
	}
	return grandineClients
}
