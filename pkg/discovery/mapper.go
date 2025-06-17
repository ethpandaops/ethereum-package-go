package discovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethpandaops/ethereum-package-go/pkg/kurtosis"
	"github.com/ethpandaops/ethereum-package-go/pkg/types"
)

// ServiceMapper maps Kurtosis services to typed Ethereum clients and services
type ServiceMapper struct {
	kurtosisClient kurtosis.Client
}

// NewServiceMapper creates a new service mapper
func NewServiceMapper(kurtosisClient kurtosis.Client) *ServiceMapper {
	return &ServiceMapper{
		kurtosisClient: kurtosisClient,
	}
}

// MapToNetwork discovers services and creates a Network instance
func (m *ServiceMapper) MapToNetwork(ctx context.Context, enclaveName string, config *types.EthereumPackageConfig) (types.Network, error) {
	// Get all services from Kurtosis
	services, err := m.kurtosisClient.GetServices(ctx, enclaveName)
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	// Initialize client collections
	executionClients := types.NewExecutionClients()
	consensusClients := types.NewConsensusClients()
	var networkServices []types.Service
	var apacheConfigServer types.ApacheConfigServer
	var prometheusURL, grafanaURL, blockscoutURL string

	// Process each service
	for _, service := range services {
		serviceType := detectServiceType(service.Name)
		
		switch serviceType {
		case types.ServiceTypeExecutionClient:
			client := m.mapExecutionClient(service)
			if client != nil {
				executionClients.Add(client)
			}
			
		case types.ServiceTypeConsensusClient:
			client := m.mapConsensusClient(service)
			if client != nil {
				consensusClients.Add(client)
			}
			
		case types.ServiceTypeApache:
			apacheConfigServer = m.mapApacheConfigServer(service)
			
		case types.ServiceTypePrometheus:
			prometheusURL = m.buildServiceURL(service, "http")
			
		case types.ServiceTypeGrafana:
			grafanaURL = m.buildServiceURL(service, "http")
			
		case types.ServiceTypeBlockscout:
			blockscoutURL = m.buildServiceURL(service, "http")
		}

		// Add to network services
		networkServices = append(networkServices, types.Service{
			Name:        service.Name,
			Type:        serviceType,
			ContainerID: service.UUID,
			Ports:       m.convertPorts(service.Ports),
			Status:      service.Status,
		})
	}

	// Determine chain ID
	chainID := uint64(12345) // Default
	if config.NetworkParams != nil && config.NetworkParams.ChainID != 0 {
		chainID = config.NetworkParams.ChainID
	}

	// Create network configuration
	networkConfig := types.NetworkConfig{
		Name:             fmt.Sprintf("ethereum-network-%s", enclaveName),
		ChainID:          chainID,
		EnclaveName:      enclaveName,
		ExecutionClients: executionClients,
		ConsensusClients: consensusClients,
		Services:         networkServices,
		ApacheConfig:     apacheConfigServer,
		PrometheusURL:    prometheusURL,
		GrafanaURL:       grafanaURL,
		BlockscoutURL:    blockscoutURL,
		CleanupFunc: func(ctx context.Context) error {
			return m.kurtosisClient.DestroyEnclave(ctx, enclaveName)
		},
	}

	return types.NewNetwork(networkConfig), nil
}

// mapExecutionClient maps a Kurtosis service to an ExecutionClient
func (m *ServiceMapper) mapExecutionClient(service *kurtosis.ServiceInfo) types.ExecutionClient {
	clientType := kurtosis.DetectClientType(service.Name)
	
	// Only map if it's an execution client type
	if !isExecutionClient(clientType) {
		return nil
	}

	return kurtosis.ConvertServiceInfoToExecutionClient(service, clientType)
}

// mapConsensusClient maps a Kurtosis service to a ConsensusClient
func (m *ServiceMapper) mapConsensusClient(service *kurtosis.ServiceInfo) types.ConsensusClient {
	clientType := kurtosis.DetectClientType(service.Name)
	
	// Only map if it's a consensus client type
	if !isConsensusClient(clientType) {
		return nil
	}

	return kurtosis.ConvertServiceInfoToConsensusClient(service, clientType)
}

// mapApacheConfigServer maps a Kurtosis service to an ApacheConfigServer
func (m *ServiceMapper) mapApacheConfigServer(service *kurtosis.ServiceInfo) types.ApacheConfigServer {
	url := m.buildServiceURL(service, "http")
	if url == "" {
		return nil
	}
	return types.NewApacheConfigServer(url)
}

// buildServiceURL builds a service URL from service info
func (m *ServiceMapper) buildServiceURL(service *kurtosis.ServiceInfo, protocol string) string {
	// Look for HTTP port
	for portName, portInfo := range service.Ports {
		if strings.Contains(strings.ToLower(portName), "http") || 
		   strings.Contains(strings.ToLower(portName), protocol) {
			if portInfo.MaybeURL != "" {
				return portInfo.MaybeURL
			}
			return fmt.Sprintf("%s://%s:%d", protocol, service.IPAddress, portInfo.Number)
		}
	}
	
	// Fallback to first port if no specific port found
	for _, portInfo := range service.Ports {
		if portInfo.MaybeURL != "" {
			return portInfo.MaybeURL
		}
		return fmt.Sprintf("%s://%s:%d", protocol, service.IPAddress, portInfo.Number)
	}
	
	return ""
}

// convertPorts converts Kurtosis port info to our port format
func (m *ServiceMapper) convertPorts(kurtosisePorts map[string]kurtosis.PortInfo) []types.Port {
	var ports []types.Port
	
	for name, portInfo := range kurtosisePorts {
		ports = append(ports, types.Port{
			Name:          name,
			InternalPort:  int(portInfo.Number),
			ExternalPort:  int(portInfo.Number),
			Protocol:      strings.ToLower(portInfo.Protocol),
			ExposedToHost: portInfo.MaybeURL != "",
		})
	}
	
	return ports
}

// detectServiceType detects the type of service based on its name
func detectServiceType(serviceName string) types.ServiceType {
	name := strings.ToLower(serviceName)
	
	switch {
	case strings.Contains(name, "geth") || 
		 strings.Contains(name, "besu") || 
		 strings.Contains(name, "nethermind") || 
		 strings.Contains(name, "erigon") || 
		 strings.Contains(name, "reth"):
		return types.ServiceTypeExecutionClient
		
	case strings.Contains(name, "lighthouse") || 
		 strings.Contains(name, "teku") || 
		 strings.Contains(name, "prysm") || 
		 strings.Contains(name, "nimbus") || 
		 strings.Contains(name, "lodestar") || 
		 strings.Contains(name, "grandine"):
		return types.ServiceTypeConsensusClient
		
	case strings.Contains(name, "validator"):
		return types.ServiceTypeValidator
		
	case strings.Contains(name, "prometheus"):
		return types.ServiceTypePrometheus
		
	case strings.Contains(name, "grafana"):
		return types.ServiceTypeGrafana
		
	case strings.Contains(name, "blockscout"):
		return types.ServiceTypeBlockscout
		
	case strings.Contains(name, "apache") || strings.Contains(name, "config"):
		return types.ServiceTypeApache
		
	default:
		return types.ServiceTypeOther
	}
}

// isExecutionClient checks if the client type is an execution client
func isExecutionClient(clientType types.ClientType) bool {
	switch clientType {
	case types.ClientGeth, types.ClientBesu, types.ClientNethermind, types.ClientErigon, types.ClientReth:
		return true
	default:
		return false
	}
}

// isConsensusClient checks if the client type is a consensus client
func isConsensusClient(clientType types.ClientType) bool {
	switch clientType {
	case types.ClientLighthouse, types.ClientTeku, types.ClientPrysm, types.ClientNimbus, types.ClientLodestar, types.ClientGrandine:
		return true
	default:
		return false
	}
}