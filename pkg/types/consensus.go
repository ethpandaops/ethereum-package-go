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

// baseConsensusClient provides common implementation for ConsensusClient
type baseConsensusClient struct {
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

func (b *baseConsensusClient) Name() string         { return b.name }
func (b *baseConsensusClient) Type() ClientType     { return b.clientType }
func (b *baseConsensusClient) Version() string      { return b.version }
func (b *baseConsensusClient) BeaconAPIURL() string { return b.beaconAPIURL }
func (b *baseConsensusClient) MetricsURL() string   { return b.metricsURL }
func (b *baseConsensusClient) P2PPort() int         { return b.p2pPort }
func (b *baseConsensusClient) ENR() string          { return b.enr }
func (b *baseConsensusClient) PeerID() string       { return b.peerID }
func (b *baseConsensusClient) ServiceName() string  { return b.serviceName }
func (b *baseConsensusClient) ContainerID() string  { return b.containerID }

// LighthouseClient represents a Lighthouse consensus client
type LighthouseClient struct {
	baseConsensusClient
}

// NewLighthouseClient creates a new Lighthouse client instance
func NewLighthouseClient(name, version, beaconAPIURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) *LighthouseClient {
	return &LighthouseClient{
		baseConsensusClient: baseConsensusClient{
			name:         name,
			clientType:   ClientLighthouse,
			version:      version,
			beaconAPIURL: beaconAPIURL,
			metricsURL:   metricsURL,
			p2pPort:      p2pPort,
			enr:          enr,
			peerID:       peerID,
			serviceName:  serviceName,
			containerID:  containerID,
		},
	}
}

// TekuClient represents a Teku consensus client
type TekuClient struct {
	baseConsensusClient
}

// NewTekuClient creates a new Teku client instance
func NewTekuClient(name, version, beaconAPIURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) *TekuClient {
	return &TekuClient{
		baseConsensusClient: baseConsensusClient{
			name:         name,
			clientType:   ClientTeku,
			version:      version,
			beaconAPIURL: beaconAPIURL,
			metricsURL:   metricsURL,
			p2pPort:      p2pPort,
			enr:          enr,
			peerID:       peerID,
			serviceName:  serviceName,
			containerID:  containerID,
		},
	}
}

// PrysmClient represents a Prysm consensus client
type PrysmClient struct {
	baseConsensusClient
}

// NewPrysmClient creates a new Prysm client instance
func NewPrysmClient(name, version, beaconAPIURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) *PrysmClient {
	return &PrysmClient{
		baseConsensusClient: baseConsensusClient{
			name:         name,
			clientType:   ClientPrysm,
			version:      version,
			beaconAPIURL: beaconAPIURL,
			metricsURL:   metricsURL,
			p2pPort:      p2pPort,
			enr:          enr,
			peerID:       peerID,
			serviceName:  serviceName,
			containerID:  containerID,
		},
	}
}

// NimbusClient represents a Nimbus consensus client
type NimbusClient struct {
	baseConsensusClient
}

// NewNimbusClient creates a new Nimbus client instance
func NewNimbusClient(name, version, beaconAPIURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) *NimbusClient {
	return &NimbusClient{
		baseConsensusClient: baseConsensusClient{
			name:         name,
			clientType:   ClientNimbus,
			version:      version,
			beaconAPIURL: beaconAPIURL,
			metricsURL:   metricsURL,
			p2pPort:      p2pPort,
			enr:          enr,
			peerID:       peerID,
			serviceName:  serviceName,
			containerID:  containerID,
		},
	}
}

// LodestarClient represents a Lodestar consensus client
type LodestarClient struct {
	baseConsensusClient
}

// NewLodestarClient creates a new Lodestar client instance
func NewLodestarClient(name, version, beaconAPIURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) *LodestarClient {
	return &LodestarClient{
		baseConsensusClient: baseConsensusClient{
			name:         name,
			clientType:   ClientLodestar,
			version:      version,
			beaconAPIURL: beaconAPIURL,
			metricsURL:   metricsURL,
			p2pPort:      p2pPort,
			enr:          enr,
			peerID:       peerID,
			serviceName:  serviceName,
			containerID:  containerID,
		},
	}
}

// GrandineClient represents a Grandine consensus client
type GrandineClient struct {
	baseConsensusClient
}

// NewGrandineClient creates a new Grandine client instance
func NewGrandineClient(name, version, beaconAPIURL, metricsURL, enr, peerID, serviceName, containerID string, p2pPort int) *GrandineClient {
	return &GrandineClient{
		baseConsensusClient: baseConsensusClient{
			name:         name,
			clientType:   ClientGrandine,
			version:      version,
			beaconAPIURL: beaconAPIURL,
			metricsURL:   metricsURL,
			p2pPort:      p2pPort,
			enr:          enr,
			peerID:       peerID,
			serviceName:  serviceName,
			containerID:  containerID,
		},
	}
}

// ConsensusClients holds all consensus clients by type
type ConsensusClients struct {
	clients map[ClientType][]ConsensusClient
}

// NewConsensusClients creates a new ConsensusClients collection
func NewConsensusClients() *ConsensusClients {
	return &ConsensusClients{
		clients: make(map[ClientType][]ConsensusClient),
	}
}

// Add adds a consensus client to the collection
func (cc *ConsensusClients) Add(client ConsensusClient) {
	cc.clients[client.Type()] = append(cc.clients[client.Type()], client)
}

// All returns all consensus clients
func (cc *ConsensusClients) All() []ConsensusClient {
	var all []ConsensusClient
	for _, clients := range cc.clients {
		all = append(all, clients...)
	}
	return all
}

// Lighthouse returns all Lighthouse clients
func (cc *ConsensusClients) Lighthouse() []*LighthouseClient {
	var lighthouseClients []*LighthouseClient
	for _, client := range cc.clients[ClientLighthouse] {
		if lc, ok := client.(*LighthouseClient); ok {
			lighthouseClients = append(lighthouseClients, lc)
		}
	}
	return lighthouseClients
}

// Teku returns all Teku clients
func (cc *ConsensusClients) Teku() []*TekuClient {
	var tekuClients []*TekuClient
	for _, client := range cc.clients[ClientTeku] {
		if tc, ok := client.(*TekuClient); ok {
			tekuClients = append(tekuClients, tc)
		}
	}
	return tekuClients
}

// Prysm returns all Prysm clients
func (cc *ConsensusClients) Prysm() []*PrysmClient {
	var prysmClients []*PrysmClient
	for _, client := range cc.clients[ClientPrysm] {
		if pc, ok := client.(*PrysmClient); ok {
			prysmClients = append(prysmClients, pc)
		}
	}
	return prysmClients
}

// Nimbus returns all Nimbus clients
func (cc *ConsensusClients) Nimbus() []*NimbusClient {
	var nimbusClients []*NimbusClient
	for _, client := range cc.clients[ClientNimbus] {
		if nc, ok := client.(*NimbusClient); ok {
			nimbusClients = append(nimbusClients, nc)
		}
	}
	return nimbusClients
}

// Lodestar returns all Lodestar clients
func (cc *ConsensusClients) Lodestar() []*LodestarClient {
	var lodestarClients []*LodestarClient
	for _, client := range cc.clients[ClientLodestar] {
		if lc, ok := client.(*LodestarClient); ok {
			lodestarClients = append(lodestarClients, lc)
		}
	}
	return lodestarClients
}

// Grandine returns all Grandine clients
func (cc *ConsensusClients) Grandine() []*GrandineClient {
	var grandineClients []*GrandineClient
	for _, client := range cc.clients[ClientGrandine] {
		if gc, ok := client.(*GrandineClient); ok {
			grandineClients = append(grandineClients, gc)
		}
	}
	return grandineClients
}