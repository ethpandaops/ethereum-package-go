package types

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
)

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

// ServiceMetadata contains detailed information about a service
type ServiceMetadata struct {
	Name                string
	ServiceType         ServiceType
	ClientType          ClientType
	Status              string
	ContainerID         string
	IPAddress           string
	Ports               map[string]PortMetadata
	NodeIndex           int
	NodeName            string
	ChainID             uint64
	ValidatorCount      int
	ValidatorStartIndex int
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
	cleanupFunc      func(context.Context) error
	orphanOnExit     bool
	cleanupOnce      sync.Once
	signalHandler    func()
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
	CleanupFunc      func(context.Context) error
	OrphanOnExit     bool
}

// NewNetwork creates a new Network instance
func NewNetwork(config NetworkConfig) Network {
	n := &network{
		name:             config.Name,
		chainID:          config.ChainID,
		enclaveName:      config.EnclaveName,
		executionClients: config.ExecutionClients,
		consensusClients: config.ConsensusClients,
		services:         config.Services,
		apacheConfig:     config.ApacheConfig,
		cleanupFunc:      config.CleanupFunc,
		orphanOnExit:     config.OrphanOnExit,
	}

	// Set up automatic cleanup on process exit unless orphaned
	if !config.OrphanOnExit {
		n.setupAutoCleanup()
		// Set up a finalizer as last resort cleanup
		runtime.SetFinalizer(n, (*network).finalize)
	}

	return n
}

func (n *network) Name() string                        { return n.name }
func (n *network) ChainID() uint64                     { return n.chainID }
func (n *network) EnclaveName() string                 { return n.enclaveName }
func (n *network) ExecutionClients() *ExecutionClients { return n.executionClients }
func (n *network) ConsensusClients() *ConsensusClients { return n.consensusClients }
func (n *network) Services() []Service                 { return n.services }
func (n *network) ApacheConfig() ApacheConfigServer    { return n.apacheConfig }

func (n *network) Stop(ctx context.Context) error {
	// In a real implementation, this would stop the Kurtosis enclave
	// For now, we'll just return nil
	return nil
}

func (n *network) Cleanup(ctx context.Context) error {
	var err error
	n.cleanupOnce.Do(func() {
		if n.cleanupFunc != nil {
			err = n.cleanupFunc(ctx)
		}
		// Remove signal handler if it exists
		if n.signalHandler != nil {
			n.signalHandler()
		}
		// Clear finalizer since we're explicitly cleaning up
		runtime.SetFinalizer(n, nil)
	})
	return err
}

// setupAutoCleanup sets up signal handlers for automatic cleanup
func (n *network) setupAutoCleanup() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start a goroutine to handle signals
	go func() {
		<-sigChan
		ctx := context.Background()
		_ = n.Cleanup(ctx) // Best effort cleanup on signal
		os.Exit(0)
	}()

	// Store cleanup function to remove signal handler
	n.signalHandler = func() {
		signal.Reset(syscall.SIGINT, syscall.SIGTERM)
	}
}

// finalize is called by the garbage collector as a last resort
func (n *network) finalize() {
	if !n.orphanOnExit {
		ctx := context.Background()
		_ = n.Cleanup(ctx) // Best effort cleanup
	}
}
