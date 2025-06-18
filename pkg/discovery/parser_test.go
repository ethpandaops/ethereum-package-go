package discovery

import (
	"testing"

	"github.com/ethpandaops/ethereum-package-go/pkg/client"
	"github.com/ethpandaops/ethereum-package-go/pkg/kurtosis"
	"github.com/ethpandaops/ethereum-package-go/pkg/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadataParser_ParseServiceMetadata(t *testing.T) {
	parser := NewMetadataParser()

	tests := []struct {
		name     string
		service  *kurtosis.ServiceInfo
		expected *network.ServiceMetadata
	}{
		{
			name: "execution client metadata",
			service: &kurtosis.ServiceInfo{
				Name:      "el-1-geth-lighthouse",
				UUID:      "uuid-1",
				Status:    "running",
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
				},
			},
			expected: &network.ServiceMetadata{
				Name:        "el-1-geth-lighthouse",
				ServiceType: network.ServiceTypeExecutionClient,
				ClientType:  client.Geth,
				Status:      "running",
				ContainerID: "uuid-1",
				IPAddress:   "10.0.0.1",
				NodeIndex:   1,
				NodeName:    "geth-lighthouse",
				Ports: map[string]network.PortMetadata{
					"rpc": {
						Name:          "rpc",
						Number:        8545,
						Protocol:      "TCP",
						URL:           "http://10.0.0.1:8545",
						ExposedToHost: true,
					},
					"ws": {
						Name:          "ws",
						Number:        8546,
						Protocol:      "TCP",
						URL:           "ws://10.0.0.1:8546",
						ExposedToHost: true,
					},
				},
			},
		},
		{
			name: "consensus client metadata",
			service: &kurtosis.ServiceInfo{
				Name:      "cl-2-lighthouse-geth",
				UUID:      "uuid-2",
				Status:    "running",
				IPAddress: "10.0.0.2",
				Ports: map[string]kurtosis.PortInfo{
					"http": {
						Number:   5052,
						Protocol: "TCP",
						MaybeURL: "http://10.0.0.2:5052",
					},
				},
			},
			expected: &network.ServiceMetadata{
				Name:        "cl-2-lighthouse-geth",
				ServiceType: network.ServiceTypeConsensusClient,
				ClientType:  client.Lighthouse,
				Status:      "running",
				ContainerID: "uuid-2",
				IPAddress:   "10.0.0.2",
				NodeIndex:   2,
				NodeName:    "lighthouse-geth",
				Ports: map[string]network.PortMetadata{
					"http": {
						Name:          "http",
						Number:        5052,
						Protocol:      "TCP",
						URL:           "http://10.0.0.2:5052",
						ExposedToHost: true,
					},
				},
			},
		},
		{
			name: "validator metadata",
			service: &kurtosis.ServiceInfo{
				Name:      "validator-1",
				UUID:      "uuid-3",
				Status:    "running",
				IPAddress: "10.0.0.3",
				Ports: map[string]kurtosis.PortInfo{
					"api": {
						Number:   7500,
						Protocol: "TCP",
						MaybeURL: "",
					},
				},
			},
			expected: &network.ServiceMetadata{
				Name:        "validator-1",
				ServiceType: network.ServiceTypeValidator,
				ClientType:  client.Unknown,
				Status:      "running",
				ContainerID: "uuid-3",
				IPAddress:   "10.0.0.3",
				NodeIndex:   0,
				NodeName:    "validator-1",
				Ports: map[string]network.PortMetadata{
					"api": {
						Name:          "api",
						Number:        7500,
						Protocol:      "TCP",
						URL:           "",
						ExposedToHost: false,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := parser.ParseServiceMetadata(tt.service)
			require.NoError(t, err)
			require.NotNil(t, metadata)

			assert.Equal(t, tt.expected.Name, metadata.Name)
			assert.Equal(t, tt.expected.ServiceType, metadata.ServiceType)
			assert.Equal(t, tt.expected.ClientType, metadata.ClientType)
			assert.Equal(t, tt.expected.Status, metadata.Status)
			assert.Equal(t, tt.expected.ContainerID, metadata.ContainerID)
			assert.Equal(t, tt.expected.IPAddress, metadata.IPAddress)
			assert.Equal(t, tt.expected.NodeIndex, metadata.NodeIndex)
			assert.Equal(t, tt.expected.NodeName, metadata.NodeName)
			assert.Equal(t, len(tt.expected.Ports), len(metadata.Ports))

			for portName, expectedPort := range tt.expected.Ports {
				actualPort, exists := metadata.Ports[portName]
				require.True(t, exists, "Port %s should exist", portName)
				assert.Equal(t, expectedPort, actualPort)
			}
		})
	}
}

func TestMetadataParser_ParseNodeInfo(t *testing.T) {
	tests := []struct {
		name          string
		serviceName   string
		expectedIndex int
		expectedName  string
	}{
		{
			name:          "el-1 pattern",
			serviceName:   "el-1-geth-lighthouse",
			expectedIndex: 1,
			expectedName:  "geth-lighthouse",
		},
		{
			name:          "cl-2 pattern",
			serviceName:   "cl-2-lighthouse-geth",
			expectedIndex: 2,
			expectedName:  "lighthouse-geth",
		},
		{
			name:          "el-10 pattern",
			serviceName:   "el-10-besu",
			expectedIndex: 10,
			expectedName:  "besu",
		},
		{
			name:          "no index pattern",
			serviceName:   "prometheus",
			expectedIndex: 0,
			expectedName:  "prometheus",
		},
		{
			name:          "invalid pattern",
			serviceName:   "el-abc-geth",
			expectedIndex: 0,
			expectedName:  "el-abc-geth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index, name := parseNodeInfo(tt.serviceName)
			assert.Equal(t, tt.expectedIndex, index)
			assert.Equal(t, tt.expectedName, name)
		})
	}
}

// ValidatorInfo is used for testing validator parsing
type ValidatorInfo struct {
	Count      int
	StartIndex int
}

func TestMetadataParser_ParseValidatorInfo(t *testing.T) {
	t.Skip("parseValidatorInfo is no longer a method on MetadataParser")
}

func TestMetadataParser_ParseConnectionString(t *testing.T) {
	t.Skip("ParseConnectionString is no longer a method on MetadataParser")
}

func TestMetadataParser_SerializeDeserialize(t *testing.T) {
	t.Skip("SerializeMetadata/DeserializeMetadata are no longer methods on MetadataParser")
}
func TestDetectServiceType(t *testing.T) {
	tests := []struct {
		name         string
		serviceName  string
		expectedType network.ServiceType
	}{
		// Execution clients
		{"geth", "el-1-geth-lighthouse", network.ServiceTypeExecutionClient},
		{"besu", "besu-node", network.ServiceTypeExecutionClient},
		{"nethermind", "nethermind-1", network.ServiceTypeExecutionClient},
		{"erigon", "erigon", network.ServiceTypeExecutionClient},
		{"reth", "reth-node", network.ServiceTypeExecutionClient},

		// Consensus clients
		{"lighthouse", "cl-1-lighthouse", network.ServiceTypeConsensusClient},
		{"teku", "teku-beacon", network.ServiceTypeConsensusClient},
		{"prysm", "prysm-beacon", network.ServiceTypeConsensusClient},
		{"nimbus", "nimbus-bn", network.ServiceTypeConsensusClient},
		{"lodestar", "lodestar", network.ServiceTypeConsensusClient},
		{"grandine", "grandine-beacon", network.ServiceTypeConsensusClient},

		// Other services
		{"validator", "validator-1", network.ServiceTypeValidator},
		{"prometheus", "prometheus", network.ServiceTypePrometheus},
		{"grafana", "grafana", network.ServiceTypeGrafana},
		{"blockscout", "blockscout", network.ServiceTypeBlockscout},
		{"apache", "apache", network.ServiceTypeApache},
		{"apache config", "apache-config-server", network.ServiceTypeApache},

		// Unknown
		{"unknown", "random-service", network.ServiceTypeOther},
		{"empty", "", network.ServiceTypeOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectServiceType(tt.serviceName)
			assert.Equal(t, tt.expectedType, result)
		})
	}
}
