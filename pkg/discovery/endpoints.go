package discovery

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/ethpandaops/ethereum-package-go/pkg/kurtosis"
	"github.com/ethpandaops/ethereum-package-go/pkg/network"
)

// EndpointExtractor extracts and formats endpoints from Kurtosis service information
type EndpointExtractor struct{}

// NewEndpointExtractor creates a new endpoint extractor
func NewEndpointExtractor() *EndpointExtractor {
	return &EndpointExtractor{}
}

// ExtractExecutionEndpoints extracts all endpoints for an execution client
func (e *EndpointExtractor) ExtractExecutionEndpoints(service *kurtosis.ServiceInfo) (*network.ExecutionEndpoints, error) {
	endpoints := &network.ExecutionEndpoints{}

	for portName, portInfo := range service.Ports {
		portNameLower := strings.ToLower(portName)

		switch {
		case strings.Contains(portNameLower, "rpc") && !strings.Contains(portNameLower, "ws") && !strings.Contains(portNameLower, "engine"):
			endpoints.RPCURL = e.buildURL(service, portInfo, "http")
		case strings.Contains(portNameLower, "ws") || strings.Contains(portNameLower, "websocket"):
			endpoints.WSURL = e.buildURL(service, portInfo, "ws")
		case strings.Contains(portNameLower, "engine") || strings.Contains(portNameLower, "auth"):
			endpoints.EngineURL = e.buildURL(service, portInfo, "http")
		case strings.Contains(portNameLower, "p2p") || strings.Contains(portNameLower, "tcp"):
			endpoints.P2PURL = e.buildURL(service, portInfo, "tcp")
		case strings.Contains(portNameLower, "metrics"):
			endpoints.MetricsURL = e.buildURL(service, portInfo, "http")
		}
	}

	// Fallback attempts if certain endpoints are missing
	if endpoints.RPCURL == "" {
		endpoints.RPCURL = e.findFallbackEndpoint(service, []string{"http-rpc", "json-rpc", "ws-rpc", "http", "rpc"}, "http")
	}
	if endpoints.EngineURL == "" {
		endpoints.EngineURL = e.findFallbackEndpoint(service, []string{"engine", "auth", "auth-rpc"}, "http")
	}

	return endpoints, nil
}

// ExtractConsensusEndpoints extracts all endpoints for a consensus client
func (e *EndpointExtractor) ExtractConsensusEndpoints(service *kurtosis.ServiceInfo) (*network.ConsensusEndpoints, error) {
	endpoints := &network.ConsensusEndpoints{}

	for portName, portInfo := range service.Ports {
		portNameLower := strings.ToLower(portName)

		switch {
		case strings.Contains(portNameLower, "beacon") || strings.Contains(portNameLower, "http"):
			endpoints.BeaconURL = e.buildURL(service, portInfo, "http")
		case strings.Contains(portNameLower, "p2p") || strings.Contains(portNameLower, "tcp"):
			endpoints.P2PURL = e.buildURL(service, portInfo, "tcp")
		case strings.Contains(portNameLower, "metrics"):
			endpoints.MetricsURL = e.buildURL(service, portInfo, "http")
		}
	}

	// Fallback attempts if beacon endpoint is missing
	if endpoints.BeaconURL == "" {
		endpoints.BeaconURL = e.findFallbackEndpoint(service, []string{"api", "rest", "http"}, "http")
	}

	return endpoints, nil
}

// ExtractValidatorEndpoints extracts all endpoints for a validator client
func (e *EndpointExtractor) ExtractValidatorEndpoints(service *kurtosis.ServiceInfo) (*network.ValidatorEndpoints, error) {
	endpoints := &network.ValidatorEndpoints{}

	for portName, portInfo := range service.Ports {
		portNameLower := strings.ToLower(portName)

		switch {
		case strings.Contains(portNameLower, "api") || strings.Contains(portNameLower, "http"):
			endpoints.APIURL = e.buildURL(service, portInfo, "http")
		case strings.Contains(portNameLower, "metrics"):
			endpoints.MetricsURL = e.buildURL(service, portInfo, "http")
		}
	}

	if endpoints.APIURL == "" {
		return nil, fmt.Errorf("no API endpoint found for validator service %s", service.Name)
	}

	return endpoints, nil
}

// buildURL constructs a URL from service info and port information
func (e *EndpointExtractor) buildURL(service *kurtosis.ServiceInfo, port kurtosis.PortInfo, scheme string) string {
	// Use MaybeURL if available
	if port.MaybeURL != "" {
		return port.MaybeURL
	}

	// Construct URL from parts
	if service.IPAddress != "" {
		return fmt.Sprintf("%s://%s:%d", scheme, service.IPAddress, port.Number)
	}

	// Fallback to service name or localhost
	host := service.Name
	if host == "" {
		host = "localhost"
	}
	return fmt.Sprintf("%s://%s:%d", scheme, host, port.Number)
}

// findFallbackEndpoint attempts to find an endpoint based on port name patterns
func (e *EndpointExtractor) findFallbackEndpoint(service *kurtosis.ServiceInfo, patterns []string, scheme string) string {
	for _, pattern := range patterns {
		for portName, portInfo := range service.Ports {
			if strings.Contains(strings.ToLower(portName), pattern) {
				return e.buildURL(service, portInfo, scheme)
			}
		}
	}
	return ""
}

// ParseEndpointURL validates and parses an endpoint URL
func (e *EndpointExtractor) ParseEndpointURL(endpoint string) (*url.URL, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("empty endpoint")
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Validate that we have a scheme
	if u.Scheme == "" {
		return nil, fmt.Errorf("missing URL scheme")
	}

	// Validate supported schemes
	switch u.Scheme {
	case "http", "https", "ws", "wss":
		// Supported schemes
	default:
		return nil, fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}

	// Validate that we have a host
	if u.Host == "" {
		return nil, fmt.Errorf("missing host")
	}

	// Set default ports if not specified
	if u.Port() == "" {
		switch u.Scheme {
		case "http":
			u.Host = u.Host + ":80"
		case "https":
			u.Host = u.Host + ":443"
		case "ws":
			u.Host = u.Host + ":80"
		case "wss":
			u.Host = u.Host + ":443"
		}
	}

	return u, nil
}

// ValidateEndpoint checks if an endpoint is valid
func (e *EndpointExtractor) ValidateEndpoint(endpoint string) error {
	_, err := e.ParseEndpointURL(endpoint)
	return err
}

// GetPortNumber extracts the port number from an endpoint URL
func (e *EndpointExtractor) GetPortNumber(endpoint string) (int, error) {
	u, err := e.ParseEndpointURL(endpoint)
	if err != nil {
		return 0, err
	}

	portStr := u.Port()
	if portStr == "" {
		// Return default ports
		switch u.Scheme {
		case "http", "ws":
			return 80, nil
		case "https", "wss":
			return 443, nil
		default:
			return 0, fmt.Errorf("unknown scheme: %s", u.Scheme)
		}
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("invalid port: %s", portStr)
	}

	return port, nil
}
