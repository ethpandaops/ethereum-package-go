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
		serviceType := m.detectServiceTypeWithPorts(service)
		
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
	// For execution clients, we need to detect which execution client it is
	// from the service name, even if it also contains consensus client names
	clientType := m.detectExecutionClientType(service.Name)
	
	if clientType == types.ClientType("unknown") {
		return nil
	}

	return kurtosis.ConvertServiceInfoToExecutionClient(service, clientType)
}

// mapConsensusClient maps a Kurtosis service to a ConsensusClient
func (m *ServiceMapper) mapConsensusClient(service *kurtosis.ServiceInfo) types.ConsensusClient {
	// For consensus clients, we need to detect which consensus client it is
	// from the service name, even if it also contains execution client names
	clientType := m.detectConsensusClientType(service.Name)
	
	if clientType == types.ClientType("unknown") {
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

// detectServiceType detects the type of service based on its name and ports
func detectServiceType(serviceName string) types.ServiceType {
	name := strings.ToLower(serviceName)
	
	// Special case services first
	switch {
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
	}
	
	// For ethereum clients, check prefixes to determine intent
	// In test scenarios, cl- prefix means consensus, el- means execution
	if strings.HasPrefix(name, "cl-") {
		// It's intended to be a consensus client
		if strings.Contains(name, "lighthouse") || 
		   strings.Contains(name, "teku") || 
		   strings.Contains(name, "prysm") || 
		   strings.Contains(name, "nimbus") || 
		   strings.Contains(name, "lodestar") || 
		   strings.Contains(name, "grandine") {
			return types.ServiceTypeConsensusClient
		}
	} else if strings.HasPrefix(name, "el-") {
		// It's intended to be an execution client
		if strings.Contains(name, "geth") || 
		   strings.Contains(name, "besu") || 
		   strings.Contains(name, "nethermind") || 
		   strings.Contains(name, "erigon") || 
		   strings.Contains(name, "reth") {
			return types.ServiceTypeExecutionClient
		}
	}
	
	// No prefix, detect by client name alone
	// Check consensus clients first (they're more specific)
	if strings.Contains(name, "lighthouse") || 
	   strings.Contains(name, "teku") || 
	   strings.Contains(name, "prysm") || 
	   strings.Contains(name, "nimbus") || 
	   strings.Contains(name, "lodestar") || 
	   strings.Contains(name, "grandine") {
		return types.ServiceTypeConsensusClient
	}
	
	// Then check execution clients
	if strings.Contains(name, "geth") || 
	   strings.Contains(name, "besu") || 
	   strings.Contains(name, "nethermind") || 
	   strings.Contains(name, "erigon") || 
	   strings.Contains(name, "reth") {
		return types.ServiceTypeExecutionClient
	}
	
	return types.ServiceTypeOther
}

// detectServiceTypeWithPorts detects service type using port information as a hint
func (m *ServiceMapper) detectServiceTypeWithPorts(service *kurtosis.ServiceInfo) types.ServiceType {
	// First try name-based detection
	serviceType := detectServiceType(service.Name)
	
	// If we got consensus or execution client, double-check with ports
	if serviceType == types.ServiceTypeConsensusClient || serviceType == types.ServiceTypeExecutionClient {
		hasRPC := false
		hasEngine := false
		hasBeacon := false
		
		for portName := range service.Ports {
			switch portName {
			case "rpc", "ws":
				hasRPC = true
			case "engine":
				hasEngine = true
			case "beacon", "http":
				hasBeacon = true
			}
		}
		
		// If it has RPC/WS and engine ports, it's definitely an execution client
		if hasRPC && hasEngine {
			return types.ServiceTypeExecutionClient
		}
		
		// If it has beacon port, it's definitely a consensus client
		if hasBeacon {
			return types.ServiceTypeConsensusClient
		}
	}
	
	return serviceType
}


// detectExecutionClientType detects which execution client type from the service name
func (m *ServiceMapper) detectExecutionClientType(serviceName string) types.ClientType {
	name := strings.ToLower(serviceName)
	
	// Check for execution client names
	switch {
	case strings.Contains(name, "geth"):
		return types.ClientGeth
	case strings.Contains(name, "besu"):
		return types.ClientBesu
	case strings.Contains(name, "nethermind"):
		return types.ClientNethermind
	case strings.Contains(name, "erigon"):
		return types.ClientErigon
	case strings.Contains(name, "reth"):
		return types.ClientReth
	default:
		return types.ClientType("unknown")
	}
}

// detectConsensusClientType detects which consensus client type from the service name
func (m *ServiceMapper) detectConsensusClientType(serviceName string) types.ClientType {
	name := strings.ToLower(serviceName)
	
	// Check for consensus client names
	switch {
	case strings.Contains(name, "lighthouse"):
		return types.ClientLighthouse
	case strings.Contains(name, "teku"):
		return types.ClientTeku
	case strings.Contains(name, "prysm"):
		return types.ClientPrysm
	case strings.Contains(name, "nimbus"):
		return types.ClientNimbus
	case strings.Contains(name, "lodestar"):
		return types.ClientLodestar
	case strings.Contains(name, "grandine"):
		return types.ClientGrandine
	default:
		return types.ClientType("unknown")
	}
}