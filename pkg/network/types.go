package network

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

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
	orphanOnExit     bool
	cleanupOnce      sync.Once
	signalHandler    func()
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
	OrphanOnExit     bool
}

// New creates a new Network instance
func New(config Config) Network {
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

func (n *network) Name() string                               { return n.name }
func (n *network) ChainID() uint64                            { return n.chainID }
func (n *network) EnclaveName() string                        { return n.enclaveName }
func (n *network) ExecutionClients() *client.ExecutionClients { return n.executionClients }
func (n *network) ConsensusClients() *client.ConsensusClients { return n.consensusClients }
func (n *network) Services() []Service                        { return n.services }
func (n *network) ApacheConfig() ApacheConfigServer           { return n.apacheConfig }

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
