package types

import "context"

// ServiceType represents the type of service in the network
type ServiceType string

const (
	ServiceTypeExecutionClient ServiceType = "execution"
	ServiceTypeConsensusClient ServiceType = "consensus"
	ServiceTypeValidator       ServiceType = "validator"
	ServiceTypePrometheus      ServiceType = "prometheus"
	ServiceTypeGrafana         ServiceType = "grafana"
	ServiceTypeBlockscout      ServiceType = "blockscout"
	ServiceTypeApache          ServiceType = "apache"
	ServiceTypeOther           ServiceType = "other"
)

// Port represents a network port mapping
type Port struct {
	Name          string
	InternalPort  int
	ExternalPort  int
	Protocol      string
	ExposedToHost bool
}

// Service represents a generic service in the network
type Service struct {
	Name        string
	Type        ServiceType
	ContainerID string
	Ports       []Port
	Status      string
}

// ApacheConfigServer represents the Apache server that hosts network configuration files
type ApacheConfigServer interface {
	URL() string
	GenesisSSZURL() string
	ConfigYAMLURL() string
	BootnodesYAMLURL() string
	DepositContractBlockURL() string
}

// apacheConfigServer is the concrete implementation
type apacheConfigServer struct {
	url string
}

// NewApacheConfigServer creates a new Apache config server instance
func NewApacheConfigServer(url string) ApacheConfigServer {
	return &apacheConfigServer{url: url}
}

func (a *apacheConfigServer) URL() string {
	return a.url
}

func (a *apacheConfigServer) GenesisSSZURL() string {
	return a.url + "/network-configs/genesis.ssz"
}

func (a *apacheConfigServer) ConfigYAMLURL() string {
	return a.url + "/network-configs/config.yaml"
}

func (a *apacheConfigServer) BootnodesYAMLURL() string {
	return a.url + "/network-configs/boot_enr.yaml"
}

func (a *apacheConfigServer) DepositContractBlockURL() string {
	return a.url + "/network-configs/deposit_contract_block.txt"
}

// Network represents an Ethereum network with all its services
type Network interface {
	// Network information
	Name() string
	ChainID() uint64
	EnclaveName() string

	// Client accessors
	ExecutionClients() *ExecutionClients
	ConsensusClients() *ConsensusClients

	// Service accessors
	Services() []Service
	ApacheConfig() ApacheConfigServer
	PrometheusURL() string
	GrafanaURL() string
	BlockscoutURL() string

	// Lifecycle management
	Stop(ctx context.Context) error
	Cleanup(ctx context.Context) error
}

// network is the concrete implementation of Network
type network struct {
	name             string
	chainID          uint64
	enclaveName      string
	executionClients *ExecutionClients
	consensusClients *ConsensusClients
	services         []Service
	apacheConfig     ApacheConfigServer
	prometheusURL    string
	grafanaURL       string
	blockscoutURL    string
	cleanupFunc      func(context.Context) error
}

// NetworkConfig holds configuration for creating a new network
type NetworkConfig struct {
	Name             string
	ChainID          uint64
	EnclaveName      string
	ExecutionClients *ExecutionClients
	ConsensusClients *ConsensusClients
	Services         []Service
	ApacheConfig     ApacheConfigServer
	PrometheusURL    string
	GrafanaURL       string
	BlockscoutURL    string
	CleanupFunc      func(context.Context) error
}

// NewNetwork creates a new Network instance
func NewNetwork(config NetworkConfig) Network {
	return &network{
		name:             config.Name,
		chainID:          config.ChainID,
		enclaveName:      config.EnclaveName,
		executionClients: config.ExecutionClients,
		consensusClients: config.ConsensusClients,
		services:         config.Services,
		apacheConfig:     config.ApacheConfig,
		prometheusURL:    config.PrometheusURL,
		grafanaURL:       config.GrafanaURL,
		blockscoutURL:    config.BlockscoutURL,
		cleanupFunc:      config.CleanupFunc,
	}
}

func (n *network) Name() string                       { return n.name }
func (n *network) ChainID() uint64                    { return n.chainID }
func (n *network) EnclaveName() string                { return n.enclaveName }
func (n *network) ExecutionClients() *ExecutionClients { return n.executionClients }
func (n *network) ConsensusClients() *ConsensusClients { return n.consensusClients }
func (n *network) Services() []Service                { return n.services }
func (n *network) ApacheConfig() ApacheConfigServer   { return n.apacheConfig }
func (n *network) PrometheusURL() string              { return n.prometheusURL }
func (n *network) GrafanaURL() string                 { return n.grafanaURL }
func (n *network) BlockscoutURL() string              { return n.blockscoutURL }

func (n *network) Stop(ctx context.Context) error {
	// In a real implementation, this would stop the Kurtosis enclave
	// For now, we'll just return nil
	return nil
}

func (n *network) Cleanup(ctx context.Context) error {
	if n.cleanupFunc != nil {
		return n.cleanupFunc(ctx)
	}
	return nil
}