package discovery

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ethpandaops/ethereum-package-go/pkg/client"
	"github.com/ethpandaops/ethereum-package-go/pkg/kurtosis"
	"github.com/ethpandaops/ethereum-package-go/pkg/network"
)

// MetadataParser parses service metadata and extracts useful information
type MetadataParser struct {
	endpointExtractor *EndpointExtractor
}

// NewMetadataParser creates a new metadata parser
func NewMetadataParser() *MetadataParser {
	return &MetadataParser{
		endpointExtractor: NewEndpointExtractor(),
	}
}

// ParseServiceMetadata parses metadata from a Kurtosis service
func (p *MetadataParser) ParseServiceMetadata(service *kurtosis.ServiceInfo) (*network.ServiceMetadata, error) {
	// First detect the service type
	serviceType := detectServiceType(service.Name)
	
	// Then detect client type based on service type
	var clientType client.Type
	if serviceType == network.ServiceTypeExecutionClient {
		clientType = detectExecutionClientType(service.Name)
	} else if serviceType == network.ServiceTypeConsensusClient {
		clientType = detectConsensusClientType(service.Name)
	} else {
		clientType = client.Unknown
	}
	
	metadata := &network.ServiceMetadata{
		Name:        service.Name,
		ServiceType: serviceType,
		ClientType:  clientType,
		Status:      service.Status,
		ContainerID: service.UUID,
		IPAddress:   service.IPAddress,
		Ports:       make(map[string]network.PortMetadata),
	}

	// Parse port metadata
	for portName, portInfo := range service.Ports {
		metadata.Ports[portName] = network.PortMetadata{
			Name:          portName,
			Number:        int(portInfo.Number),
			Protocol:      portInfo.Protocol,
			URL:           portInfo.MaybeURL,
			ExposedToHost: portInfo.MaybeURL != "",
		}
	}

	// Extract client-specific metadata
	extractClientSpecificMetadata(metadata, service)

	return metadata, nil
}

// extractClientSpecificMetadata extracts metadata specific to client types
func extractClientSpecificMetadata(metadata *network.ServiceMetadata, service *kurtosis.ServiceInfo) {
	// Parse node index and name
	metadata.NodeIndex, metadata.NodeName = parseNodeInfo(service.Name)

	// Extract version from container or other metadata
	metadata.Version = extractVersion(service)

	// Extract network-specific data
	// Note: Config field is not available in current Kurtosis ServiceInfo
	// This would need to be extracted from service metadata or environment

	// Extract validator information
	if metadata.ServiceType == network.ServiceTypeValidator {
		metadata.ValidatorCount, metadata.ValidatorStartIndex = parseValidatorInfo(service)
	}

	// Extract P2P port
	for portName, portInfo := range service.Ports {
		if strings.Contains(strings.ToLower(portName), "p2p") {
			metadata.P2PPort = int(portInfo.Number)
			break
		}
	}

	// Extract enode/ENR/peer ID based on client type
	if metadata.ServiceType == network.ServiceTypeExecutionClient {
		metadata.Enode = extractEnode(service)
	} else if metadata.ServiceType == network.ServiceTypeConsensusClient {
		metadata.ENR = extractENR(service)
		metadata.PeerID = extractPeerID(service)
	}
}

// parseNodeInfo extracts node index and name from service name
func parseNodeInfo(serviceName string) (int, string) {
	// Pattern: el-1-geth-lighthouse, cl-2-teku-geth, etc.
	re := regexp.MustCompile(`^(el|cl)-(\d+)-(.+)$`)
	matches := re.FindStringSubmatch(serviceName)
	
	if len(matches) >= 4 {
		index, _ := strconv.Atoi(matches[2])
		return index, matches[3]
	}
	
	return 0, serviceName
}

// parseValidatorInfo extracts validator count and start index
func parseValidatorInfo(service *kurtosis.ServiceInfo) (int, int) {
	// Check config for validator information
	// Note: Config field is not available in current Kurtosis ServiceInfo
	// Will try to parse from service name instead
	
	// Try to parse from service name (e.g., "validator-5-10" means 5 validators starting at index 10)
	re := regexp.MustCompile(`validator-(\d+)-(\d+)`)
	if matches := re.FindStringSubmatch(service.Name); len(matches) >= 3 {
		count, _ := strconv.Atoi(matches[1])
		startIndex, _ := strconv.Atoi(matches[2])
		return count, startIndex
	}
	
	return 0, 0
}

// extractVersion attempts to extract version information
func extractVersion(service *kurtosis.ServiceInfo) string {
	// Check config
	// Note: Config and Container fields are not available in current Kurtosis ServiceInfo
	// Will return a default version based on client type
	// In a real implementation, this would be fetched from the service's API
	
	return "unknown"
}

// extractEnode extracts enode URL for execution clients
func extractEnode(service *kurtosis.ServiceInfo) string {
	// Note: Config field is not available in current Kurtosis ServiceInfo
	// In a real implementation, this would be fetched from the service's admin API
	return ""
}

// extractENR extracts ENR for consensus clients
func extractENR(service *kurtosis.ServiceInfo) string {
	// Note: Config field is not available in current Kurtosis ServiceInfo
	// In a real implementation, this would be fetched from the beacon node's API
	return ""
}

// extractPeerID extracts peer ID for consensus clients
func extractPeerID(service *kurtosis.ServiceInfo) string {
	// Note: Config field is not available in current Kurtosis ServiceInfo
	// In a real implementation, this would be fetched from the beacon node's API
	return ""
}


// ParseConnectionString parses a connection string into components
func ParseConnectionString(connectionStr string) (protocol, address string, port int, err error) {
	// Format: protocol://address:port
	parts := strings.Split(connectionStr, "://")
	if len(parts) != 2 {
		return "", "", 0, fmt.Errorf("invalid connection string format")
	}
	
	protocol = parts[0]
	
	// Split address and port
	addressParts := strings.Split(parts[1], ":")
	if len(addressParts) != 2 {
		return "", "", 0, fmt.Errorf("invalid address format")
	}
	
	address = addressParts[0]
	port, err = strconv.Atoi(addressParts[1])
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid port: %w", err)
	}
	
	return protocol, address, port, nil
}

// SerializeMetadata converts metadata to JSON
func SerializeMetadata(metadata *network.ServiceMetadata) (string, error) {
	data, err := json.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("failed to serialize metadata: %w", err)
	}
	return string(data), nil
}

// DeserializeMetadata converts JSON to metadata
func DeserializeMetadata(data string) (*network.ServiceMetadata, error) {
	var metadata network.ServiceMetadata
	if err := json.Unmarshal([]byte(data), &metadata); err != nil {
		return nil, fmt.Errorf("failed to deserialize metadata: %w", err)
	}
	return &metadata, nil
}