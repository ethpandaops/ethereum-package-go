package discovery

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ethpandaops/ethereum-package-go/pkg/kurtosis"
	"github.com/ethpandaops/ethereum-package-go/pkg/types"
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
func (p *MetadataParser) ParseServiceMetadata(service *kurtosis.ServiceInfo) (*types.ServiceMetadata, error) {
	metadata := &types.ServiceMetadata{
		Name:        service.Name,
		ServiceType: detectServiceType(service.Name),
		ClientType:  kurtosis.DetectClientType(service.Name),
		Status:      service.Status,
		ContainerID: service.UUID,
		IPAddress:   service.IPAddress,
		Ports:       make(map[string]types.PortMetadata),
	}

	// Parse port metadata
	for portName, portInfo := range service.Ports {
		metadata.Ports[portName] = types.PortMetadata{
			Name:         portName,
			Number:       int(portInfo.Number),
			Protocol:     portInfo.Protocol,
			URL:          portInfo.MaybeURL,
			ExposedToHost: portInfo.MaybeURL != "",
		}
	}

	// Extract client-specific metadata
	if err := p.extractClientSpecificMetadata(metadata, service); err != nil {
		return nil, fmt.Errorf("failed to extract client-specific metadata: %w", err)
	}

	return metadata, nil
}

// extractClientSpecificMetadata extracts metadata specific to client types
func (p *MetadataParser) extractClientSpecificMetadata(metadata *types.ServiceMetadata, service *kurtosis.ServiceInfo) error {
	switch metadata.ServiceType {
	case types.ServiceTypeExecutionClient:
		return p.extractExecutionClientMetadata(metadata, service)
	case types.ServiceTypeConsensusClient:
		return p.extractConsensusClientMetadata(metadata, service)
	case types.ServiceTypeValidator:
		return p.extractValidatorMetadata(metadata, service)
	default:
		// No specific metadata needed for other service types
		return nil
	}
}

// extractExecutionClientMetadata extracts execution client specific metadata
func (p *MetadataParser) extractExecutionClientMetadata(metadata *types.ServiceMetadata, service *kurtosis.ServiceInfo) error {
	// Extract node info from service name or environment
	nodeInfo := p.parseNodeInfo(service.Name)
	metadata.NodeIndex = nodeInfo.Index
	metadata.NodeName = nodeInfo.Name

	// Extract network ID and chain ID if available
	if chainID := p.extractChainID(service); chainID != 0 {
		metadata.ChainID = chainID
	}

	return nil
}

// extractConsensusClientMetadata extracts consensus client specific metadata
func (p *MetadataParser) extractConsensusClientMetadata(metadata *types.ServiceMetadata, service *kurtosis.ServiceInfo) error {
	// Extract validator info
	validatorInfo := p.parseValidatorInfo(service.Name)
	metadata.ValidatorCount = validatorInfo.Count
	metadata.ValidatorStartIndex = validatorInfo.StartIndex

	// Extract node info
	nodeInfo := p.parseNodeInfo(service.Name)
	metadata.NodeIndex = nodeInfo.Index
	metadata.NodeName = nodeInfo.Name

	return nil
}

// extractValidatorMetadata extracts validator specific metadata
func (p *MetadataParser) extractValidatorMetadata(metadata *types.ServiceMetadata, service *kurtosis.ServiceInfo) error {
	// Extract validator-specific information
	validatorInfo := p.parseValidatorInfo(service.Name)
	metadata.ValidatorCount = validatorInfo.Count
	metadata.ValidatorStartIndex = validatorInfo.StartIndex

	return nil
}

// NodeInfo represents parsed node information
type NodeInfo struct {
	Index int
	Name  string
}

// parseNodeInfo extracts node information from service name
func (p *MetadataParser) parseNodeInfo(serviceName string) NodeInfo {
	// Common patterns: "el-1-geth-lighthouse", "cl-2-lighthouse", etc.
	indexPattern := regexp.MustCompile(`(?:el|cl)-(\d+)`)
	matches := indexPattern.FindStringSubmatch(serviceName)
	
	nodeInfo := NodeInfo{
		Name: serviceName,
	}
	
	if len(matches) > 1 {
		if index, err := strconv.Atoi(matches[1]); err == nil {
			nodeInfo.Index = index
		}
	}
	
	return nodeInfo
}

// ValidatorInfo represents parsed validator information
type ValidatorInfo struct {
	Count      int
	StartIndex int
}

// parseValidatorInfo extracts validator information from service name
func (p *MetadataParser) parseValidatorInfo(serviceName string) ValidatorInfo {
	// Look for validator count patterns
	countPattern := regexp.MustCompile(`validator.*?(\d+)`)
	matches := countPattern.FindStringSubmatch(strings.ToLower(serviceName))
	
	validatorInfo := ValidatorInfo{}
	
	if len(matches) > 1 {
		if count, err := strconv.Atoi(matches[1]); err == nil {
			validatorInfo.Count = count
		}
	}
	
	return validatorInfo
}

// extractChainID extracts chain ID from service metadata
func (p *MetadataParser) extractChainID(service *kurtosis.ServiceInfo) uint64 {
	// This would typically come from environment variables or config
	// For now, return a default value
	return 0
}

// ParseConnectionString parses connection strings for various protocols
func (p *MetadataParser) ParseConnectionString(protocol, endpoint string) (map[string]string, error) {
	switch strings.ToLower(protocol) {
	case "http", "https":
		return p.parseHTTPConnection(endpoint)
	case "ws", "wss":
		return p.parseWebSocketConnection(endpoint)
	case "tcp":
		return p.parseTCPConnection(endpoint)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// parseHTTPConnection parses HTTP connection details
func (p *MetadataParser) parseHTTPConnection(endpoint string) (map[string]string, error) {
	host, port, err := p.endpointExtractor.ParseEndpointURL(endpoint)
	if err != nil {
		return nil, err
	}
	
	return map[string]string{
		"type":     "http",
		"host":     host,
		"port":     strconv.Itoa(port),
		"endpoint": endpoint,
	}, nil
}

// parseWebSocketConnection parses WebSocket connection details
func (p *MetadataParser) parseWebSocketConnection(endpoint string) (map[string]string, error) {
	host, port, err := p.endpointExtractor.ParseEndpointURL(endpoint)
	if err != nil {
		return nil, err
	}
	
	return map[string]string{
		"type":     "websocket",
		"host":     host,
		"port":     strconv.Itoa(port),
		"endpoint": endpoint,
	}, nil
}

// parseTCPConnection parses TCP connection details
func (p *MetadataParser) parseTCPConnection(endpoint string) (map[string]string, error) {
	host, port, err := p.endpointExtractor.ParseEndpointURL(endpoint)
	if err != nil {
		return nil, err
	}
	
	return map[string]string{
		"type":     "tcp",
		"host":     host,
		"port":     strconv.Itoa(port),
		"endpoint": endpoint,
	}, nil
}

// SerializeMetadata serializes service metadata to JSON
func (p *MetadataParser) SerializeMetadata(metadata *types.ServiceMetadata) ([]byte, error) {
	return json.MarshalIndent(metadata, "", "  ")
}

// DeserializeMetadata deserializes service metadata from JSON
func (p *MetadataParser) DeserializeMetadata(data []byte) (*types.ServiceMetadata, error) {
	var metadata types.ServiceMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to deserialize metadata: %w", err)
	}
	return &metadata, nil
}