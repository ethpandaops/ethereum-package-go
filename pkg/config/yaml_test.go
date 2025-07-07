package config

import (
	"strings"
	"testing"

	"github.com/ethpandaops/ethereum-package-go/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToYAML(t *testing.T) {
	config := &EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{
				ELType:         client.Geth,
				CLType:         client.Lighthouse,
				Count:          2,
				ValidatorCount: 32,
			},
		},
		NetworkParams: &NetworkParams{
			Network:        "kurtosis",
			NetworkID:      "12345",
			SecondsPerSlot: 12,
		},
		MEV: &MEVConfig{
			Type: "full",
		},
		AdditionalServices: []AdditionalService{
			{
				Name: "prometheus",
				Config: map[string]interface{}{
					"port": 9090,
				},
			},
		},
		GlobalLogLevel: "info",
	}

	yamlStr, err := ToYAML(config)
	require.NoError(t, err)
	assert.NotEmpty(t, yamlStr)

	// Check that key elements are present
	assert.Contains(t, yamlStr, "participants:")
	assert.Contains(t, yamlStr, "el_type: geth")
	assert.Contains(t, yamlStr, "cl_type: lighthouse")
	assert.Contains(t, yamlStr, "count: 2")
	assert.Contains(t, yamlStr, "validator_count: 32")
	assert.Contains(t, yamlStr, "network_params:")
	assert.Contains(t, yamlStr, "network_id: \"12345\"")
	assert.Contains(t, yamlStr, "mev_params:")
	assert.Contains(t, yamlStr, "type: full")
	assert.Contains(t, yamlStr, "additional_services:")
	assert.Contains(t, yamlStr, "name: prometheus")
	assert.Contains(t, yamlStr, "global_log_level: info")
}

func TestToYAMLMinimal(t *testing.T) {
	config := &EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{
				ELType: client.Geth,
				CLType: client.Lighthouse,
			},
		},
	}

	yamlStr, err := ToYAML(config)
	require.NoError(t, err)
	assert.NotEmpty(t, yamlStr)

	// Should only contain participants
	assert.Contains(t, yamlStr, "participants:")
	assert.Contains(t, yamlStr, "el_type: geth")
	assert.Contains(t, yamlStr, "cl_type: lighthouse")

	// Should not contain optional fields
	assert.NotContains(t, yamlStr, "network_params:")
	assert.NotContains(t, yamlStr, "mev_params:")
	assert.NotContains(t, yamlStr, "additional_services:")
	assert.NotContains(t, yamlStr, "global_log_level:")
}

func TestFromYAML(t *testing.T) {
	yamlContent := `
participants:
  - el_type: geth
    cl_type: lighthouse
    count: 2
    validator_count: 32
  - el_type: besu
    cl_type: teku
    count: 1

network_params:
  network: kurtosis
  network_id: "12345"
  seconds_per_slot: 12
  num_validator_keys_per_node: 64

mev_params:
  type: full
  relay_url: http://localhost:18550

additional_services:
  - name: prometheus
    config:
      port: 9090
  - name: grafana

global_log_level: debug
`

	config, err := FromYAML(yamlContent)
	require.NoError(t, err)

	// Check participants
	assert.Len(t, config.Participants, 2)
	assert.Equal(t, client.Geth, config.Participants[0].ELType)
	assert.Equal(t, client.Lighthouse, config.Participants[0].CLType)
	assert.Equal(t, 2, config.Participants[0].Count)
	assert.Equal(t, 32, config.Participants[0].ValidatorCount)
	assert.Equal(t, client.Besu, config.Participants[1].ELType)
	assert.Equal(t, client.Teku, config.Participants[1].CLType)

	// Check network params
	require.NotNil(t, config.NetworkParams)
	assert.Equal(t, "kurtosis", config.NetworkParams.Network)
	assert.Equal(t, "12345", config.NetworkParams.NetworkID)
	assert.Equal(t, 12, config.NetworkParams.SecondsPerSlot)
	assert.Equal(t, 64, config.NetworkParams.NumValidatorKeysPerNode)

	// Check MEV params
	require.NotNil(t, config.MEV)
	assert.Equal(t, "full", config.MEV.Type)
	assert.Equal(t, "http://localhost:18550", config.MEV.RelayURL)

	// Check additional services
	assert.Len(t, config.AdditionalServices, 2)
	assert.Equal(t, "prometheus", config.AdditionalServices[0].Name)
	assert.NotNil(t, config.AdditionalServices[0].Config)
	assert.Equal(t, "grafana", config.AdditionalServices[1].Name)

	// Check global log level
	assert.Equal(t, "debug", config.GlobalLogLevel)
}

func TestFromYAMLMinimal(t *testing.T) {
	yamlContent := `
participants:
  - el_type: geth
    cl_type: lighthouse
`

	config, err := FromYAML(yamlContent)
	require.NoError(t, err)

	assert.Len(t, config.Participants, 1)
	assert.Equal(t, client.Geth, config.Participants[0].ELType)
	assert.Equal(t, client.Lighthouse, config.Participants[0].CLType)
	assert.Nil(t, config.NetworkParams)
	assert.Nil(t, config.MEV)
	assert.Len(t, config.AdditionalServices, 0)
	assert.Empty(t, config.GlobalLogLevel)
}

func TestFromYAMLInvalid(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name:    "invalid yaml",
			yaml:    "invalid: yaml: content:",
			wantErr: true,
		},
		{
			name:    "empty yaml",
			yaml:    "",
			wantErr: false, // Empty YAML is valid, just produces empty config
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FromYAML(tt.yaml)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	// Create a comprehensive config
	original := &EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{
				ELType:         client.Geth,
				CLType:         client.Lighthouse,
				ELVersion:      "v1.13.0",
				CLVersion:      "v4.5.0",
				Count:          3,
				ValidatorCount: 96,
			},
			{
				ELType: client.Besu,
				CLType: client.Teku,
				Count:  1,
			},
		},
		NetworkParams: &NetworkParams{
			Network:                 "kurtosis",
			NetworkID:               "98765",
			SecondsPerSlot:          12,
			NumValidatorKeysPerNode: 64,
			AltairForkEpoch:         0,
			BellatrixForkEpoch:      0,
			CapellaForkEpoch:        10,
			DenebForkEpoch:          20,
			ElectraForkEpoch:        30,
		},
		MEV: &MEVConfig{
			Type:            "full",
			RelayURL:        "http://relay:18550",
			MinBidEth:       "0.01",
			MaxBundleLength: 3,
		},
		AdditionalServices: []AdditionalService{
			{
				Name: "prometheus",
				Config: map[string]interface{}{
					"port":      9090,
					"retention": "15d",
				},
			},
			{
				Name: "grafana",
			},
		},
		GlobalLogLevel: "info",
	}

	// Convert to YAML
	yamlStr, err := ToYAML(original)
	require.NoError(t, err)

	// Parse back from YAML
	parsed, err := FromYAML(yamlStr)
	require.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, len(original.Participants), len(parsed.Participants))
	for i := range original.Participants {
		assert.Equal(t, original.Participants[i].ELType, parsed.Participants[i].ELType)
		assert.Equal(t, original.Participants[i].CLType, parsed.Participants[i].CLType)
		assert.Equal(t, original.Participants[i].Count, parsed.Participants[i].Count)
		assert.Equal(t, original.Participants[i].ValidatorCount, parsed.Participants[i].ValidatorCount)
	}

	assert.Equal(t, original.NetworkParams.NetworkID, parsed.NetworkParams.NetworkID)
	assert.Equal(t, original.MEV.Type, parsed.MEV.Type)
	assert.Equal(t, len(original.AdditionalServices), len(parsed.AdditionalServices))
	assert.Equal(t, original.GlobalLogLevel, parsed.GlobalLogLevel)
}

func TestYAMLFormatting(t *testing.T) {
	config := &EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{
				ELType: client.Geth,
				CLType: client.Lighthouse,
			},
		},
	}

	yamlStr, err := ToYAML(config)
	require.NoError(t, err)

	// Check proper indentation (2 spaces)
	lines := strings.Split(yamlStr, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") {
			// Second level should have 2 spaces
			assert.True(t, strings.HasPrefix(line, "  "))
		}
	}
}

func TestToYAMLWithPortPublisher(t *testing.T) {
	config := &EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{
				ELType: client.Geth,
				CLType: client.Lighthouse,
				Count:  1,
			},
		},
		PortPublisher: &PortPublisherConfig{
			NatExitIP: "192.168.1.100",
			EL: &PortPublisherComponent{
				Enabled:         true,
				PublicPortStart: 32000,
			},
			CL: &PortPublisherComponent{
				Enabled:         true,
				PublicPortStart: 33000,
			},
			VC: &PortPublisherComponent{
				Enabled:         true,
				PublicPortStart: 34000,
			},
			AdditionalServices: &PortPublisherComponent{
				Enabled: false,
			},
		},
	}

	yamlStr, err := ToYAML(config)
	require.NoError(t, err)
	assert.NotEmpty(t, yamlStr)

	// Check that port publisher elements are present
	assert.Contains(t, yamlStr, "port_publisher:")
	assert.Contains(t, yamlStr, "nat_exit_ip: 192.168.1.100")
	assert.Contains(t, yamlStr, "el:")
	assert.Contains(t, yamlStr, "enabled: true")
	assert.Contains(t, yamlStr, "public_port_start: 32000")
	assert.Contains(t, yamlStr, "cl:")
	assert.Contains(t, yamlStr, "public_port_start: 33000")
	assert.Contains(t, yamlStr, "vc:")
	assert.Contains(t, yamlStr, "public_port_start: 34000")
	assert.Contains(t, yamlStr, "additional_services:")
	assert.Contains(t, yamlStr, "enabled: false")
}

func TestFromYAMLWithPortPublisher(t *testing.T) {
	yamlContent := `
participants:
  - el_type: geth
    cl_type: lighthouse
    count: 2

port_publisher:
  nat_exit_ip: "127.0.0.1"
  el:
    enabled: true
    public_port_start: 32000
  cl:
    enabled: true
    public_port_start: 33000
  vc:
    enabled: false
  additional_services:
    enabled: true
    public_port_start: 35000
`

	config, err := FromYAML(yamlContent)
	require.NoError(t, err)

	// Check participants
	assert.Len(t, config.Participants, 1)
	assert.Equal(t, client.Geth, config.Participants[0].ELType)
	assert.Equal(t, client.Lighthouse, config.Participants[0].CLType)
	assert.Equal(t, 2, config.Participants[0].Count)

	// Check port publisher
	require.NotNil(t, config.PortPublisher)
	assert.Equal(t, "127.0.0.1", config.PortPublisher.NatExitIP)

	require.NotNil(t, config.PortPublisher.EL)
	assert.True(t, config.PortPublisher.EL.Enabled)
	assert.Equal(t, 32000, config.PortPublisher.EL.PublicPortStart)

	require.NotNil(t, config.PortPublisher.CL)
	assert.True(t, config.PortPublisher.CL.Enabled)
	assert.Equal(t, 33000, config.PortPublisher.CL.PublicPortStart)

	require.NotNil(t, config.PortPublisher.VC)
	assert.False(t, config.PortPublisher.VC.Enabled)

	require.NotNil(t, config.PortPublisher.AdditionalServices)
	assert.True(t, config.PortPublisher.AdditionalServices.Enabled)
	assert.Equal(t, 35000, config.PortPublisher.AdditionalServices.PublicPortStart)
}

func TestPortPublisherRoundTrip(t *testing.T) {
	// Create a config with port publisher
	original := &EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{
				ELType: client.Geth,
				CLType: client.Prysm,
				Count:  1,
			},
		},
		NetworkParams: &NetworkParams{
			NetworkID: "12345",
		},
		PortPublisher: &PortPublisherConfig{
			NatExitIP: "auto",
			EL: &PortPublisherComponent{
				Enabled:         true,
				PublicPortStart: 40000,
			},
			CL: &PortPublisherComponent{
				Enabled:         true,
				PublicPortStart: 41000,
			},
		},
	}

	// Convert to YAML
	yamlStr, err := ToYAML(original)
	require.NoError(t, err)

	// Parse back from YAML
	parsed, err := FromYAML(yamlStr)
	require.NoError(t, err)

	// Verify port publisher fields match
	require.NotNil(t, parsed.PortPublisher)
	assert.Equal(t, original.PortPublisher.NatExitIP, parsed.PortPublisher.NatExitIP)
	assert.Equal(t, original.PortPublisher.EL.Enabled, parsed.PortPublisher.EL.Enabled)
	assert.Equal(t, original.PortPublisher.EL.PublicPortStart, parsed.PortPublisher.EL.PublicPortStart)
	assert.Equal(t, original.PortPublisher.CL.Enabled, parsed.PortPublisher.CL.Enabled)
	assert.Equal(t, original.PortPublisher.CL.PublicPortStart, parsed.PortPublisher.CL.PublicPortStart)
}

func TestToYAMLWithDockerCacheParams(t *testing.T) {
	config := &EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{
				ELType: client.Geth,
				CLType: client.Lighthouse,
				Count:  1,
			},
		},
		DockerCacheParams: &DockerCacheParams{
			Enabled: true,
			URL:     "docker.ethquokkaops.io",
		},
	}

	yamlStr, err := ToYAML(config)
	require.NoError(t, err)
	assert.NotEmpty(t, yamlStr)

	// Check that docker cache params elements are present
	assert.Contains(t, yamlStr, "docker_cache_params:")
	assert.Contains(t, yamlStr, "enabled: true")
	assert.Contains(t, yamlStr, "url: docker.ethquokkaops.io")
}

func TestFromYAMLWithDockerCacheParams(t *testing.T) {
	yamlContent := `
participants:
  - el_type: geth
    cl_type: lighthouse
    count: 1

docker_cache_params:
  enabled: true
  url: "docker.ethquokkaops.io"
`

	config, err := FromYAML(yamlContent)
	require.NoError(t, err)

	// Check participants
	assert.Len(t, config.Participants, 1)
	assert.Equal(t, client.Geth, config.Participants[0].ELType)
	assert.Equal(t, client.Lighthouse, config.Participants[0].CLType)

	// Check docker cache params
	require.NotNil(t, config.DockerCacheParams)
	assert.True(t, config.DockerCacheParams.Enabled)
	assert.Equal(t, "docker.ethquokkaops.io", config.DockerCacheParams.URL)
}

func TestDockerCacheParamsRoundTrip(t *testing.T) {
	// Create a config with docker cache params
	original := &EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{
				ELType: client.Geth,
				CLType: client.Prysm,
				Count:  1,
			},
		},
		NetworkParams: &NetworkParams{
			NetworkID: "12345",
		},
		DockerCacheParams: &DockerCacheParams{
			Enabled: true,
			URL:     "docker.ethquokkaops.io",
		},
	}

	// Convert to YAML
	yamlStr, err := ToYAML(original)
	require.NoError(t, err)

	// Parse back from YAML
	parsed, err := FromYAML(yamlStr)
	require.NoError(t, err)

	// Verify docker cache params fields match
	require.NotNil(t, parsed.DockerCacheParams)
	assert.Equal(t, original.DockerCacheParams.Enabled, parsed.DockerCacheParams.Enabled)
	assert.Equal(t, original.DockerCacheParams.URL, parsed.DockerCacheParams.URL)
}
