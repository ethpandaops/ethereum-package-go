package discovery

import (
	"testing"

	"github.com/ethpandaops/ethereum-package-go/pkg/kurtosis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndpointExtractor_ExtractExecutionEndpoints(t *testing.T) {
	extractor := NewEndpointExtractor()

	tests := []struct {
		name     string
		service  *kurtosis.ServiceInfo
		expected struct {
			rpcURL     string
			wsURL      string
			engineURL  string
			metricsURL string
			p2pURL     string
		}
	}{
		{
			name: "full execution client ports",
			service: &kurtosis.ServiceInfo{
				Name:      "el-1-geth",
				IPAddress: "10.0.0.1",
				Ports: map[string]kurtosis.PortInfo{
					"rpc": {
						Number:   8545,
						Protocol: "TCP",
						MaybeURL: "http://10.0.0.1:8545",
					},
					"ws": {
						Number:   8546,
						Protocol: "TCP",
						MaybeURL: "ws://10.0.0.1:8546",
					},
					"engine": {
						Number:   8551,
						Protocol: "TCP",
						MaybeURL: "http://10.0.0.1:8551",
					},
					"metrics": {
						Number:   6060,
						Protocol: "TCP",
						MaybeURL: "http://10.0.0.1:6060",
					},
					"p2p": {
						Number:   30303,
						Protocol: "TCP",
						MaybeURL: "",
					},
				},
			},
			expected: struct {
				rpcURL     string
				wsURL      string
				engineURL  string
				metricsURL string
				p2pURL     string
			}{
				rpcURL:     "http://10.0.0.1:8545",
				wsURL:      "ws://10.0.0.1:8546",
				engineURL:  "http://10.0.0.1:8551",
				metricsURL: "http://10.0.0.1:6060",
				p2pURL:     "tcp://10.0.0.1:30303",
			},
		},
		{
			name: "ports without MaybeURL",
			service: &kurtosis.ServiceInfo{
				Name:      "el-2-besu",
				IPAddress: "10.0.0.2",
				Ports: map[string]kurtosis.PortInfo{
					"rpc": {
						Number:   8545,
						Protocol: "TCP",
						MaybeURL: "",
					},
					"websocket": {
						Number:   8546,
						Protocol: "TCP",
						MaybeURL: "",
					},
				},
			},
			expected: struct {
				rpcURL     string
				wsURL      string
				engineURL  string
				metricsURL string
				p2pURL     string
			}{
				rpcURL: "http://10.0.0.2:8545",
				wsURL:  "ws://10.0.0.2:8546",
			},
		},
		{
			name: "minimal ports with fallback",
			service: &kurtosis.ServiceInfo{
				Name:      "el-3-nethermind",
				IPAddress: "10.0.0.3",
				Ports: map[string]kurtosis.PortInfo{
					"http": {
						Number:   8080,
						Protocol: "TCP",
						MaybeURL: "http://10.0.0.3:8080",
					},
				},
			},
			expected: struct {
				rpcURL     string
				wsURL      string
				engineURL  string
				metricsURL string
				p2pURL     string
			}{
				rpcURL: "http://10.0.0.3:8080", // Should fallback to first HTTP port
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoints, err := extractor.ExtractExecutionEndpoints(tt.service)
			require.NoError(t, err)
			require.NotNil(t, endpoints)

			assert.Equal(t, tt.expected.rpcURL, endpoints.RPCURL)
			assert.Equal(t, tt.expected.wsURL, endpoints.WSURL)
			assert.Equal(t, tt.expected.engineURL, endpoints.EngineURL)
			assert.Equal(t, tt.expected.metricsURL, endpoints.MetricsURL)
			assert.Equal(t, tt.expected.p2pURL, endpoints.P2PURL)
		})
	}
}

func TestEndpointExtractor_ExtractConsensusEndpoints(t *testing.T) {
	extractor := NewEndpointExtractor()

	tests := []struct {
		name     string
		service  *kurtosis.ServiceInfo
		expected struct {
			beaconURL  string
			p2pURL     string
			metricsURL string
		}
	}{
		{
			name: "full consensus client ports",
			service: &kurtosis.ServiceInfo{
				Name:      "cl-1-lighthouse",
				IPAddress: "10.0.0.1",
				Ports: map[string]kurtosis.PortInfo{
					"http": {
						Number:   5052,
						Protocol: "TCP",
						MaybeURL: "http://10.0.0.1:5052",
					},
					"p2p": {
						Number:   9000,
						Protocol: "TCP",
						MaybeURL: "",
					},
					"metrics": {
						Number:   5054,
						Protocol: "TCP",
						MaybeURL: "http://10.0.0.1:5054",
					},
				},
			},
			expected: struct {
				beaconURL  string
				p2pURL     string
				metricsURL string
			}{
				beaconURL:  "http://10.0.0.1:5052",
				p2pURL:     "tcp://10.0.0.1:9000",
				metricsURL: "http://10.0.0.1:5054",
			},
		},
		{
			name: "alternative port names",
			service: &kurtosis.ServiceInfo{
				Name:      "cl-2-teku",
				IPAddress: "10.0.0.2",
				Ports: map[string]kurtosis.PortInfo{
					"beacon": {
						Number:   5052,
						Protocol: "TCP",
						MaybeURL: "http://10.0.0.2:5052",
					},
					"tcp": {
						Number:   9000,
						Protocol: "TCP",
						MaybeURL: "",
					},
				},
			},
			expected: struct {
				beaconURL  string
				p2pURL     string
				metricsURL string
			}{
				beaconURL: "http://10.0.0.2:5052", // Should still find first HTTP port
				p2pURL:    "tcp://10.0.0.2:9000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoints, err := extractor.ExtractConsensusEndpoints(tt.service)
			require.NoError(t, err)
			require.NotNil(t, endpoints)

			assert.Equal(t, tt.expected.beaconURL, endpoints.BeaconURL)
			assert.Equal(t, tt.expected.p2pURL, endpoints.P2PURL)
			assert.Equal(t, tt.expected.metricsURL, endpoints.MetricsURL)
		})
	}
}

func TestEndpointExtractor_ExtractValidatorEndpoints(t *testing.T) {
	extractor := NewEndpointExtractor()

	service := &kurtosis.ServiceInfo{
		Name:      "validator-1",
		IPAddress: "10.0.0.1",
		Ports: map[string]kurtosis.PortInfo{
			"api": {
				Number:   7500,
				Protocol: "TCP",
				MaybeURL: "http://10.0.0.1:7500",
			},
			"metrics": {
				Number:   8080,
				Protocol: "TCP",
				MaybeURL: "http://10.0.0.1:8080",
			},
		},
	}

	endpoints, err := extractor.ExtractValidatorEndpoints(service)
	require.NoError(t, err)
	require.NotNil(t, endpoints)

	assert.Equal(t, "http://10.0.0.1:7500", endpoints.APIURL)
	assert.Equal(t, "http://10.0.0.1:8080", endpoints.MetricsURL)
}

func TestEndpointExtractor_ParseEndpointURL(t *testing.T) {
	extractor := NewEndpointExtractor()

	tests := []struct {
		name         string
		endpoint     string
		expectedHost string
		expectedPort int
		expectError  bool
	}{
		{
			name:         "http with port",
			endpoint:     "http://localhost:8545",
			expectedHost: "localhost",
			expectedPort: 8545,
		},
		{
			name:         "https with port",
			endpoint:     "https://example.com:443",
			expectedHost: "example.com",
			expectedPort: 443,
		},
		{
			name:         "ws with port",
			endpoint:     "ws://10.0.0.1:8546",
			expectedHost: "10.0.0.1",
			expectedPort: 8546,
		},
		{
			name:         "http without port",
			endpoint:     "http://example.com",
			expectedHost: "example.com",
			expectedPort: 80,
		},
		{
			name:         "https without port",
			endpoint:     "https://example.com",
			expectedHost: "example.com",
			expectedPort: 443,
		},
		{
			name:         "ws without port",
			endpoint:     "ws://example.com",
			expectedHost: "example.com",
			expectedPort: 80,
		},
		{
			name:         "wss without port",
			endpoint:     "wss://example.com",
			expectedHost: "example.com",
			expectedPort: 443,
		},
		{
			name:        "invalid URL",
			endpoint:    "not-a-url",
			expectError: true,
		},
		{
			name:        "unknown scheme",
			endpoint:    "ftp://example.com",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := extractor.ParseEndpointURL(tt.endpoint)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, u)
				assert.Equal(t, tt.expectedHost, u.Hostname())
				port, err := extractor.GetPortNumber(tt.endpoint)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedPort, port)
			}
		})
	}
}

func TestEndpointExtractor_ValidateEndpoint(t *testing.T) {
	extractor := NewEndpointExtractor()

	tests := []struct {
		name        string
		endpoint    string
		expectError bool
	}{
		{
			name:     "valid http endpoint",
			endpoint: "http://localhost:8545",
		},
		{
			name:     "valid https endpoint",
			endpoint: "https://example.com:443",
		},
		{
			name:     "valid ws endpoint",
			endpoint: "ws://10.0.0.1:8546",
		},
		{
			name:        "empty endpoint",
			endpoint:    "",
			expectError: true,
		},
		{
			name:        "invalid URL",
			endpoint:    "not-a-url",
			expectError: true,
		},
		{
			name:        "missing scheme",
			endpoint:    "example.com:8080",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := extractor.ValidateEndpoint(tt.endpoint)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEndpointExtractor_BuildURL(t *testing.T) {
	extractor := NewEndpointExtractor()

	tests := []struct {
		name     string
		service  *kurtosis.ServiceInfo
		portInfo kurtosis.PortInfo
		protocol string
		expected string
	}{
		{
			name: "with MaybeURL",
			service: &kurtosis.ServiceInfo{
				IPAddress: "10.0.0.1",
			},
			portInfo: kurtosis.PortInfo{
				Number:   8545,
				MaybeURL: "http://external.host:8545",
			},
			protocol: "http",
			expected: "http://external.host:8545",
		},
		{
			name: "without MaybeURL",
			service: &kurtosis.ServiceInfo{
				IPAddress: "10.0.0.1",
			},
			portInfo: kurtosis.PortInfo{
				Number:   8545,
				MaybeURL: "",
			},
			protocol: "http",
			expected: "http://10.0.0.1:8545",
		},
		{
			name: "no IP address fallback",
			service: &kurtosis.ServiceInfo{
				IPAddress: "",
			},
			portInfo: kurtosis.PortInfo{
				Number:   8545,
				MaybeURL: "",
			},
			protocol: "http",
			expected: "http://localhost:8545",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := extractor.buildURL(tt.service, tt.portInfo, tt.protocol)
			assert.Equal(t, tt.expected, url)
		})
	}
}