package network

import (
	"context"

	"github.com/ethpandaops/ethereum-package-go/pkg/client"
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
	ServiceTypeDora            ServiceType = "dora"
	ServiceTypeApache          ServiceType = "apache"
	ServiceTypeSpamoor         ServiceType = "spamoor"
	ServiceTypeOther           ServiceType = "other"
)

// Network represents an Ethereum network with all its services
type Network interface {
	// Network information
	Name() string
	ChainID() uint64
	EnclaveName() string

	// Client accessors
	ExecutionClients() *client.ExecutionClients
	ConsensusClients() *client.ConsensusClients

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
	executionClients *client.ExecutionClients
	consensusClients *client.ConsensusClients
	services         []Service
	apacheConfig     ApacheConfigServer
	cleanupFunc      func(context.Context) error
}

// Config holds configuration for creating a new network
type Config struct {
	Name             string
	ChainID          uint64
	EnclaveName      string
	ExecutionClients *client.ExecutionClients
	ConsensusClients *client.ConsensusClients
	Services         []Service
	ApacheConfig     ApacheConfigServer
	CleanupFunc      func(context.Context) error
}

// New creates a new Network instance
func New(config Config) Network {
	return &network{
		name:             config.Name,
		chainID:          config.ChainID,
		enclaveName:      config.EnclaveName,
		executionClients: config.ExecutionClients,
		consensusClients: config.ConsensusClients,
		services:         config.Services,
		apacheConfig:     config.ApacheConfig,
		cleanupFunc:      config.CleanupFunc,
	}
}

func (n *network) Name() string                           { return n.name }
func (n *network) ChainID() uint64                        { return n.chainID }
func (n *network) EnclaveName() string                    { return n.enclaveName }
func (n *network) ExecutionClients() *client.ExecutionClients { return n.executionClients }
func (n *network) ConsensusClients() *client.ConsensusClients { return n.consensusClients }
func (n *network) Services() []Service                    { return n.services }
func (n *network) ApacheConfig() ApacheConfigServer       { return n.apacheConfig }

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