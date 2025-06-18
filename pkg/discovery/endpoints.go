package discovery

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/ethpandaops/ethereum-package-go/pkg/kurtosis"
	"github.com/ethpandaops/ethereum-package-go/pkg/types"
)

// EndpointExtractor extracts and formats endpoints from Kurtosis service information
type EndpointExtractor struct{}

// NewEndpointExtractor creates a new endpoint extractor
func NewEndpointExtractor() *EndpointExtractor {
	return &EndpointExtractor{}
}

// ExtractExecutionEndpoints extracts all endpoints for an execution client
func (e *EndpointExtractor) ExtractExecutionEndpoints(service *kurtosis.ServiceInfo) (*types.ExecutionEndpoints, error) {
	endpoints := &types.ExecutionEndpoints{}
	
	for portName, portInfo := range service.Ports {
		portNameLower := strings.ToLower(portName)
		
		switch {
		case strings.Contains(portNameLower, "rpc") && !strings.Contains(portNameLower, "ws"):
			endpoints.RPCURL = e.buildURL(service, portInfo, "http")
		case strings.Contains(portNameLower, "ws") || strings.Contains(portNameLower, "websocket"):
			endpoints.WSURL = e.buildURL(service, portInfo, "ws")
		case strings.Contains(portNameLower, "engine"):
			endpoints.EngineURL = e.buildURL(service, portInfo, "http")
		case strings.Contains(portNameLower, "discovery") || strings.Contains(portNameLower, "p2p"):
			endpoints.P2PURL = e.buildURL(service, portInfo, "tcp")
		case strings.Contains(portNameLower, "metrics"):
			endpoints.MetricsURL = e.buildURL(service, portInfo, "http")
		}
	}
	
	// Set defaults if specific ports not found
	if endpoints.RPCURL == "" {
		endpoints.RPCURL = e.findFirstHTTPPort(service)
	}
	
	return endpoints, nil
}

// ExtractConsensusEndpoints extracts all endpoints for a consensus client
func (e *EndpointExtractor) ExtractConsensusEndpoints(service *kurtosis.ServiceInfo) (*types.ConsensusEndpoints, error) {
	endpoints := &types.ConsensusEndpoints{}
	
	for portName, portInfo := range service.Ports {
		portNameLower := strings.ToLower(portName)
		
		switch {
		case strings.Contains(portNameLower, "http") && !strings.Contains(portNameLower, "metrics"):
			endpoints.BeaconURL = e.buildURL(service, portInfo, "http")
		case strings.Contains(portNameLower, "p2p") || strings.Contains(portNameLower, "tcp"):
			endpoints.P2PURL = e.buildURL(service, portInfo, "tcp")
		case strings.Contains(portNameLower, "metrics"):
			endpoints.MetricsURL = e.buildURL(service, portInfo, "http")
		}
	}
	
	// Set defaults if specific ports not found
	if endpoints.BeaconURL == "" {
		endpoints.BeaconURL = e.findFirstHTTPPort(service)
	}
	
	return endpoints, nil
}

// ExtractValidatorEndpoints extracts all endpoints for a validator client
func (e *EndpointExtractor) ExtractValidatorEndpoints(service *kurtosis.ServiceInfo) (*types.ValidatorEndpoints, error) {
	endpoints := &types.ValidatorEndpoints{}
	
	for portName, portInfo := range service.Ports {
		portNameLower := strings.ToLower(portName)
		
		switch {
		case strings.Contains(portNameLower, "http") || strings.Contains(portNameLower, "api"):
			endpoints.APIURL = e.buildURL(service, portInfo, "http")
		case strings.Contains(portNameLower, "metrics"):
			endpoints.MetricsURL = e.buildURL(service, portInfo, "http")
		}
	}
	
	return endpoints, nil
}

// buildURL constructs a URL from service and port information
func (e *EndpointExtractor) buildURL(service *kurtosis.ServiceInfo, portInfo kurtosis.PortInfo, protocol string) string {
	// Use the pre-built URL if available
	if portInfo.MaybeURL != "" {
		return portInfo.MaybeURL
	}
	
	// Construct URL manually
	host := service.IPAddress
	if host == "" {
		host = "localhost"
	}
	
	return fmt.Sprintf("%s://%s:%d", protocol, host, portInfo.Number)
}

// findFirstHTTPPort finds the first HTTP port available
func (e *EndpointExtractor) findFirstHTTPPort(service *kurtosis.ServiceInfo) string {
	for portName, portInfo := range service.Ports {
		portNameLower := strings.ToLower(portName)
		if strings.Contains(portNameLower, "http") || strings.Contains(portNameLower, "api") {
			return e.buildURL(service, portInfo, "http")
		}
	}
	
	// Fallback to first port with HTTP protocol
	for _, portInfo := range service.Ports {
		if strings.ToLower(portInfo.Protocol) == "tcp" {
			return e.buildURL(service, portInfo, "http")
		}
	}
	
	return ""
}

// ParseEndpointURL parses an endpoint URL and returns host and port
func (e *EndpointExtractor) ParseEndpointURL(endpoint string) (host string, port int, err error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse URL %s: %w", endpoint, err)
	}
	
	host = u.Hostname()
	portStr := u.Port()
	
	if portStr == "" {
		// Use default ports
		switch u.Scheme {
		case "http":
			port = 80
		case "https":
			port = 443
		case "ws":
			port = 80
		case "wss":
			port = 443
		default:
			return "", 0, fmt.Errorf("unknown scheme %s and no port specified", u.Scheme)
		}
	} else {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return "", 0, fmt.Errorf("invalid port %s: %w", portStr, err)
		}
	}
	
	return host, port, nil
}

// ValidateEndpoint validates that an endpoint is reachable and properly formatted
func (e *EndpointExtractor) ValidateEndpoint(endpoint string) error {
	if endpoint == "" {
		return fmt.Errorf("endpoint cannot be empty")
	}
	
	_, _, err := e.ParseEndpointURL(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint format: %w", err)
	}
	
	return nil
}