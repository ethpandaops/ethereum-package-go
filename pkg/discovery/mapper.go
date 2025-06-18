package discovery

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethpandaops/ethereum-package-go/pkg/client"
	"github.com/ethpandaops/ethereum-package-go/pkg/config"
	"github.com/ethpandaops/ethereum-package-go/pkg/kurtosis"
	"github.com/ethpandaops/ethereum-package-go/pkg/network"
)

// ServiceMapper maps Kurtosis services to typed Ethereum clients and services
type ServiceMapper struct {
	kurtosisClient kurtosis.Client
	metadataParser *MetadataParser
}

// NewServiceMapper creates a new service mapper
func NewServiceMapper(kurtosisClient kurtosis.Client) *ServiceMapper {
	return &ServiceMapper{
		kurtosisClient: kurtosisClient,
		metadataParser: NewMetadataParser(),
	}
}

// MapToNetwork discovers services and creates a Network instance
func (m *ServiceMapper) MapToNetwork(ctx context.Context, enclaveName string, cfg *config.EthereumPackageConfig) (network.Network, error) {
	// Get all services from Kurtosis
	services, err := m.kurtosisClient.GetServices(ctx, enclaveName)
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	// Initialize client collections
	executionClients := client.NewExecutionClients()
	consensusClients := client.NewConsensusClients()
	var networkServices []network.Service
	var apacheConfigServer network.ApacheConfigServer

	// Process each service
	for _, service := range services {
		serviceType := m.detectServiceTypeWithPorts(service)

		switch serviceType {
		case network.ServiceTypeExecutionClient:
			client := m.mapExecutionClient(service)
			if client != nil {
				executionClients.Add(client)
			}

		case network.ServiceTypeConsensusClient:
			client := m.mapConsensusClient(service)
			if client != nil {
				consensusClients.Add(client)
			}

		case network.ServiceTypeApache:
			apacheConfigServer = m.mapApacheConfigServer(service)
		}

		// Add to network services
		networkServices = append(networkServices, network.Service{
			Name:        service.Name,
			Type:        serviceType,
			ContainerID: service.UUID,
			Ports:       m.convertPorts(service.Ports),
			Status:      service.Status,
		})
	}

	// Determine chain ID from network ID
	chainID := uint64(12345) // Default
	if cfg.NetworkParams != nil && cfg.NetworkParams.NetworkID != "" {
		if parsedID, err := strconv.ParseUint(cfg.NetworkParams.NetworkID, 10, 64); err == nil {
			chainID = parsedID
		}
	}

	// Create network configuration
	networkConfig := network.Config{
		Name:             fmt.Sprintf("ethereum-network-%s", enclaveName),
		ChainID:          chainID,
		EnclaveName:      enclaveName,
		ExecutionClients: executionClients,
		ConsensusClients: consensusClients,
		Services:         networkServices,
		ApacheConfig:     apacheConfigServer,
		CleanupFunc:      m.createCleanupFunc(enclaveName),
	}

	return network.New(networkConfig), nil
}

// detectServiceTypeWithPorts detects the service type based on name and ports
func (m *ServiceMapper) detectServiceTypeWithPorts(service *kurtosis.ServiceInfo) network.ServiceType {
	// Check by name patterns
	serviceType := detectServiceType(service.Name)
	if serviceType != network.ServiceTypeOther {
		return serviceType
	}

	// If name doesn't match, check ports for execution/consensus patterns
	for portName := range service.Ports {
		if strings.Contains(portName, "rpc") || strings.Contains(portName, "engine") {
			return network.ServiceTypeExecutionClient
		}
		if strings.Contains(portName, "beacon") || strings.Contains(portName, "http") {
			return network.ServiceTypeConsensusClient
		}
	}

	return network.ServiceTypeOther
}

// mapExecutionClient maps a Kurtosis service to an ExecutionClient
func (m *ServiceMapper) mapExecutionClient(service *kurtosis.ServiceInfo) client.ExecutionClient {
	// Extract endpoints
	extractor := NewEndpointExtractor()
	endpoints, _ := extractor.ExtractExecutionEndpoints(service)

	// Detect client type
	clientType := detectExecutionClientType(service.Name)

	// Extract metadata
	metadata, _ := m.metadataParser.ParseServiceMetadata(service)

	return client.NewExecutionClient(
		clientType,
		service.Name,
		metadata.Version,
		endpoints.RPCURL,
		endpoints.WSURL,
		endpoints.EngineURL,
		endpoints.MetricsURL,
		metadata.Enode,
		service.Name,
		service.UUID,
		metadata.P2PPort,
	)
}

// mapConsensusClient maps a Kurtosis service to a ConsensusClient
func (m *ServiceMapper) mapConsensusClient(service *kurtosis.ServiceInfo) client.ConsensusClient {
	// Extract endpoints
	extractor := NewEndpointExtractor()
	endpoints, _ := extractor.ExtractConsensusEndpoints(service)

	// Detect client type
	clientType := detectConsensusClientType(service.Name)

	// Extract metadata
	metadata, _ := m.metadataParser.ParseServiceMetadata(service)

	return client.NewConsensusClient(
		clientType,
		service.Name,
		metadata.Version,
		endpoints.BeaconURL,
		endpoints.MetricsURL,
		metadata.ENR,
		metadata.PeerID,
		service.Name,
		service.UUID,
		metadata.P2PPort,
	)
}

// mapApacheConfigServer maps a Kurtosis service to an ApacheConfigServer
func (m *ServiceMapper) mapApacheConfigServer(service *kurtosis.ServiceInfo) network.ApacheConfigServer {
	// Find the HTTP port
	for portName, port := range service.Ports {
		if strings.Contains(portName, "http") || portName == "http" {
			url := fmt.Sprintf("http://%s:%d", service.IPAddress, port.Number)
			return network.NewApacheConfigServer(url)
		}
	}

	// Fallback to default port
	url := fmt.Sprintf("http://%s:80", service.IPAddress)
	return network.NewApacheConfigServer(url)
}

// convertPorts converts Kurtosis ports to network Port types
func (m *ServiceMapper) convertPorts(ports map[string]kurtosis.PortInfo) []network.Port {
	var result []network.Port
	for name, port := range ports {
		result = append(result, network.Port{
			Name:          name,
			InternalPort:  int(port.Number),
			ExternalPort:  int(port.Number),
			Protocol:      port.Protocol,
			ExposedToHost: true, // Assume exposed for now
		})
	}
	return result
}

// createCleanupFunc creates a cleanup function for the network
func (m *ServiceMapper) createCleanupFunc(enclaveName string) func(context.Context) error {
	return func(ctx context.Context) error {
		return m.kurtosisClient.DestroyEnclave(ctx, enclaveName)
	}
}

// detectExecutionClientType detects the execution client type from the service name
func detectExecutionClientType(name string) client.Type {
	nameLower := strings.ToLower(name)

	switch {
	case strings.Contains(nameLower, "geth"):
		return client.Geth
	case strings.Contains(nameLower, "besu"):
		return client.Besu
	case strings.Contains(nameLower, "nethermind"):
		return client.Nethermind
	case strings.Contains(nameLower, "erigon"):
		return client.Erigon
	case strings.Contains(nameLower, "reth"):
		return client.Reth
	default:
		return client.Unknown
	}
}

// detectConsensusClientType detects the consensus client type from the service name
func detectConsensusClientType(name string) client.Type {
	nameLower := strings.ToLower(name)

	switch {
	case strings.Contains(nameLower, "lighthouse"):
		return client.Lighthouse
	case strings.Contains(nameLower, "teku"):
		return client.Teku
	case strings.Contains(nameLower, "prysm"):
		return client.Prysm
	case strings.Contains(nameLower, "nimbus"):
		return client.Nimbus
	case strings.Contains(nameLower, "lodestar"):
		return client.Lodestar
	case strings.Contains(nameLower, "grandine"):
		return client.Grandine
	default:
		return client.Unknown
	}
}

// detectServiceType detects the service type from the service name
func detectServiceType(name string) network.ServiceType {
	nameLower := strings.ToLower(name)

	// Check for validator services first (most specific)
	if strings.Contains(nameLower, "validator-key-generation") || 
		strings.HasPrefix(nameLower, "vc-") || 
		(strings.Contains(nameLower, "validator") && !strings.HasPrefix(nameLower, "cl-") && !strings.HasPrefix(nameLower, "el-")) {
		return network.ServiceTypeValidator
	}

	// Check for consensus clients (cl- prefix takes precedence)
	if strings.HasPrefix(nameLower, "cl-") || strings.Contains(nameLower, "beacon") {
		return network.ServiceTypeConsensusClient
	}

	// Check for execution clients
	if strings.HasPrefix(nameLower, "el-") || strings.Contains(nameLower, "execution") {
		return network.ServiceTypeExecutionClient
	}

	// Check by client name patterns (only if no prefix found)
	if !strings.Contains(nameLower, "-") ||
		(!strings.HasPrefix(nameLower, "cl-") && !strings.HasPrefix(nameLower, "el-")) {
		// Execution clients
		if strings.Contains(nameLower, "geth") ||
			strings.Contains(nameLower, "besu") ||
			strings.Contains(nameLower, "nethermind") ||
			strings.Contains(nameLower, "erigon") ||
			strings.Contains(nameLower, "reth") {
			return network.ServiceTypeExecutionClient
		}

		// Consensus clients
		if strings.Contains(nameLower, "lighthouse") ||
			strings.Contains(nameLower, "teku") ||
			strings.Contains(nameLower, "prysm") ||
			strings.Contains(nameLower, "nimbus") ||
			strings.Contains(nameLower, "lodestar") ||
			strings.Contains(nameLower, "grandine") {
			return network.ServiceTypeConsensusClient
		}
	}

	// Validator check already done above, skip duplicate

	// Check for other services
	if strings.Contains(nameLower, "prometheus") {
		return network.ServiceTypePrometheus
	}
	if strings.Contains(nameLower, "grafana") {
		return network.ServiceTypeGrafana
	}
	if strings.Contains(nameLower, "blockscout") {
		return network.ServiceTypeBlockscout
	}
	if strings.Contains(nameLower, "dora") {
		return network.ServiceTypeDora
	}
	if strings.Contains(nameLower, "apache") {
		return network.ServiceTypeApache
	}
	if strings.Contains(nameLower, "spamoor") {
		return network.ServiceTypeSpamoor
	}

	return network.ServiceTypeOther
}
