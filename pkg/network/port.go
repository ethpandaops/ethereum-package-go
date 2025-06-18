package network

// Port represents a network port mapping
type Port struct {
	Name          string
	InternalPort  int
	ExternalPort  int
	Protocol      string
	ExposedToHost bool
}

// PortMetadata represents detailed port information
type PortMetadata struct {
	Name          string
	Number        int
	Protocol      string
	URL           string
	ExposedToHost bool
}

// ExecutionEndpoints holds all endpoint URLs for execution clients
type ExecutionEndpoints struct {
	RPCURL     string
	WSURL      string
	EngineURL  string
	P2PURL     string
	MetricsURL string
}

// ConsensusEndpoints holds all endpoint URLs for consensus clients
type ConsensusEndpoints struct {
	BeaconURL  string
	P2PURL     string
	MetricsURL string
}

// ValidatorEndpoints holds all endpoint URLs for validator clients
type ValidatorEndpoints struct {
	APIURL     string
	MetricsURL string
}
