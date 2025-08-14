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
			"prometheus",
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
	assert.Contains(t, yamlStr, "- prometheus")
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
  - prometheus
  - grafana

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
	assert.Equal(t, AdditionalService("prometheus"), config.AdditionalServices[0])
	assert.Equal(t, AdditionalService("grafana"), config.AdditionalServices[1])

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
				ELImage:        &[]string{"v1.13.0"}[0],
				CLImage:        &[]string{"v4.5.0"}[0],
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
			"prometheus",
			"grafana",
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

func TestToYAMLWithExtraConfig(t *testing.T) {
	config := &EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{
				ELType:        client.Geth,
				CLType:        client.Prysm,
				Count:         1,
				ELExtraParams: []string{"--metrics", "--metrics.addr=0.0.0.0"},
				ELExtraMounts: map[string]string{
					"/data/scripts": "scripts/monitoring",
				},
				ELExtraEnvVars: map[string]string{
					"GETH_DEBUG": "true",
				},
				ELExtraLabels: map[string]string{
					"team": "devops",
				},
				CLExtraParams: []string{"--graffiti=test"},
				CLExtraMounts: map[string]string{
					"/etc/tysm": "static_files/tysm",
				},
				CLExtraEnvVars: map[string]string{
					"LOG_LEVEL": "debug",
				},
				CLExtraLabels: map[string]string{
					"environment": "testing",
				},
				VCExtraParams: []string{"--suggested-fee-recipient=0x0000000000000000000000000000000000000000"},
				VCExtraMounts: map[string]string{
					"/validator/keys": "validator_keys/production",
				},
				VCExtraEnvVars: map[string]string{
					"JAVA_OPTS": "-Xmx2g",
				},
				VCExtraLabels: map[string]string{
					"security": "high",
				},
			},
		},
	}

	yamlStr, err := ToYAML(config)
	require.NoError(t, err)
	assert.NotEmpty(t, yamlStr)

	// Check that extra config elements are present
	assert.Contains(t, yamlStr, "el_extra_params:")
	assert.Contains(t, yamlStr, "- --metrics")
	assert.Contains(t, yamlStr, "- --metrics.addr=0.0.0.0")
	assert.Contains(t, yamlStr, "el_extra_mounts:")
	assert.Contains(t, yamlStr, "/data/scripts: scripts/monitoring")
	assert.Contains(t, yamlStr, "el_extra_env_vars:")
	assert.Contains(t, yamlStr, "GETH_DEBUG: \"true\"")
	assert.Contains(t, yamlStr, "el_extra_labels:")
	assert.Contains(t, yamlStr, "team: devops")

	assert.Contains(t, yamlStr, "cl_extra_params:")
	assert.Contains(t, yamlStr, "- --graffiti=test")
	assert.Contains(t, yamlStr, "cl_extra_mounts:")
	assert.Contains(t, yamlStr, "/etc/tysm: static_files/tysm")
	assert.Contains(t, yamlStr, "cl_extra_env_vars:")
	assert.Contains(t, yamlStr, "LOG_LEVEL: debug")
	assert.Contains(t, yamlStr, "cl_extra_labels:")
	assert.Contains(t, yamlStr, "environment: testing")

	assert.Contains(t, yamlStr, "vc_extra_params:")
	assert.Contains(t, yamlStr, "- --suggested-fee-recipient=0x0000000000000000000000000000000000000000")
	assert.Contains(t, yamlStr, "vc_extra_mounts:")
	assert.Contains(t, yamlStr, "/validator/keys: validator_keys/production")
	assert.Contains(t, yamlStr, "vc_extra_env_vars:")
	assert.Contains(t, yamlStr, "JAVA_OPTS: -Xmx2g")
	assert.Contains(t, yamlStr, "vc_extra_labels:")
	assert.Contains(t, yamlStr, "security: high")
}

func TestFromYAMLWithExtraConfig(t *testing.T) {
	yamlContent := `
participants:
  - el_type: geth
    cl_type: prysm
    count: 1
    el_extra_params:
      - --metrics
      - --metrics.addr=0.0.0.0
    el_extra_mounts:
      /data/scripts: scripts/monitoring
    el_extra_env_vars:
      GETH_DEBUG: "true"
    el_extra_labels:
      team: devops
    cl_extra_params:
      - --graffiti=test
    cl_extra_mounts:
      /etc/tysm: static_files/tysm
    cl_extra_env_vars:
      LOG_LEVEL: debug
    cl_extra_labels:
      environment: testing
    vc_extra_params:
      - --suggested-fee-recipient=0x0000000000000000000000000000000000000000
    vc_extra_mounts:
      /validator/keys: validator_keys/production
    vc_extra_env_vars:
      JAVA_OPTS: "-Xmx2g"
    vc_extra_labels:
      security: high
`

	config, err := FromYAML(yamlContent)
	require.NoError(t, err)

	// Check participant
	assert.Len(t, config.Participants, 1)
	participant := config.Participants[0]

	// Check EL extras
	assert.Equal(t, []string{"--metrics", "--metrics.addr=0.0.0.0"}, participant.ELExtraParams)
	assert.Equal(t, map[string]string{"/data/scripts": "scripts/monitoring"}, participant.ELExtraMounts)
	assert.Equal(t, map[string]string{"GETH_DEBUG": "true"}, participant.ELExtraEnvVars)
	assert.Equal(t, map[string]string{"team": "devops"}, participant.ELExtraLabels)

	// Check CL extras
	assert.Equal(t, []string{"--graffiti=test"}, participant.CLExtraParams)
	assert.Equal(t, map[string]string{"/etc/tysm": "static_files/tysm"}, participant.CLExtraMounts)
	assert.Equal(t, map[string]string{"LOG_LEVEL": "debug"}, participant.CLExtraEnvVars)
	assert.Equal(t, map[string]string{"environment": "testing"}, participant.CLExtraLabels)

	// Check VC extras
	assert.Equal(t, []string{"--suggested-fee-recipient=0x0000000000000000000000000000000000000000"}, participant.VCExtraParams)
	assert.Equal(t, map[string]string{"/validator/keys": "validator_keys/production"}, participant.VCExtraMounts)
	assert.Equal(t, map[string]string{"JAVA_OPTS": "-Xmx2g"}, participant.VCExtraEnvVars)
	assert.Equal(t, map[string]string{"security": "high"}, participant.VCExtraLabels)
}

func TestExtraConfigRoundTrip(t *testing.T) {
	// Create a config with extra configuration
	original := &EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{
				ELType:        client.Geth,
				CLType:        client.Prysm,
				Count:         2,
				ELExtraParams: []string{"--param1", "--param2=value"},
				ELExtraMounts: map[string]string{
					"/mount1": "source1",
					"/mount2": "source2",
				},
				ELExtraEnvVars: map[string]string{
					"ENV1": "value1",
					"ENV2": "value2",
				},
				ELExtraLabels: map[string]string{
					"label1": "value1",
					"label2": "value2",
				},
				CLExtraParams: []string{"--cl-param1"},
				CLExtraMounts: map[string]string{
					"/cl/mount": "cl/source",
				},
				CLExtraEnvVars: map[string]string{
					"CL_ENV": "cl_value",
				},
				CLExtraLabels: map[string]string{
					"cl_label": "cl_value",
				},
				VCExtraParams: []string{"--vc-param1"},
				VCExtraMounts: map[string]string{
					"/vc/mount": "vc/source",
				},
				VCExtraEnvVars: map[string]string{
					"VC_ENV": "vc_value",
				},
				VCExtraLabels: map[string]string{
					"vc_label": "vc_value",
				},
			},
		},
	}

	// Convert to YAML
	yamlStr, err := ToYAML(original)
	require.NoError(t, err)

	// Parse back from YAML
	parsed, err := FromYAML(yamlStr)
	require.NoError(t, err)

	// Verify all extra config fields match
	require.Len(t, parsed.Participants, 1)
	participant := parsed.Participants[0]

	assert.Equal(t, original.Participants[0].ELExtraParams, participant.ELExtraParams)
	assert.Equal(t, original.Participants[0].ELExtraMounts, participant.ELExtraMounts)
	assert.Equal(t, original.Participants[0].ELExtraEnvVars, participant.ELExtraEnvVars)
	assert.Equal(t, original.Participants[0].ELExtraLabels, participant.ELExtraLabels)

	assert.Equal(t, original.Participants[0].CLExtraParams, participant.CLExtraParams)
	assert.Equal(t, original.Participants[0].CLExtraMounts, participant.CLExtraMounts)
	assert.Equal(t, original.Participants[0].CLExtraEnvVars, participant.CLExtraEnvVars)
	assert.Equal(t, original.Participants[0].CLExtraLabels, participant.CLExtraLabels)

	assert.Equal(t, original.Participants[0].VCExtraParams, participant.VCExtraParams)
	assert.Equal(t, original.Participants[0].VCExtraMounts, participant.VCExtraMounts)
	assert.Equal(t, original.Participants[0].VCExtraEnvVars, participant.VCExtraEnvVars)
	assert.Equal(t, original.Participants[0].VCExtraLabels, participant.VCExtraLabels)
}

func TestExtraConfigOmitEmpty(t *testing.T) {
	// Create a config with only some extra config fields set
	config := &EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{
				ELType: client.Geth,
				CLType: client.Lighthouse,
				Count:  1,
				// Only set CL extra mounts, leave everything else empty
				CLExtraMounts: map[string]string{
					"/etc/config": "config/cl",
				},
			},
		},
	}

	yamlStr, err := ToYAML(config)
	require.NoError(t, err)
	assert.NotEmpty(t, yamlStr)

	// Check that only the set field is present
	assert.Contains(t, yamlStr, "cl_extra_mounts:")
	assert.Contains(t, yamlStr, "/etc/config: config/cl")

	// Check that empty fields are omitted
	assert.NotContains(t, yamlStr, "el_extra_params:")
	assert.NotContains(t, yamlStr, "el_extra_mounts:")
	assert.NotContains(t, yamlStr, "el_extra_env_vars:")
	assert.NotContains(t, yamlStr, "el_extra_labels:")
	assert.NotContains(t, yamlStr, "cl_extra_params:")
	assert.NotContains(t, yamlStr, "cl_extra_env_vars:")
	assert.NotContains(t, yamlStr, "cl_extra_labels:")
	assert.NotContains(t, yamlStr, "vc_extra_params:")
	assert.NotContains(t, yamlStr, "vc_extra_mounts:")
	assert.NotContains(t, yamlStr, "vc_extra_env_vars:")
	assert.NotContains(t, yamlStr, "vc_extra_labels:")
}

func TestToYAMLWithExtraFiles(t *testing.T) {
	config := &EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{
				ELType: client.Geth,
				CLType: client.Lighthouse,
				Count:  1,
				CLExtraMounts: map[string]string{
					"/configs/beacon.yaml": "beacon-config.yaml",
				},
			},
		},
		ExtraFiles: map[string]string{
			"beacon-config.yaml": "metrics:\n  enabled: true\n  port: 8080",
			"validator.yaml":     "graffiti: test",
		},
	}

	yamlStr, err := ToYAML(config)
	require.NoError(t, err)
	assert.NotEmpty(t, yamlStr)

	// Check that extra_files is present at root level
	assert.Contains(t, yamlStr, "extra_files:")
	assert.Contains(t, yamlStr, "beacon-config.yaml:")
	assert.Contains(t, yamlStr, "validator.yaml:")
	assert.Contains(t, yamlStr, "graffiti: test")
}

func TestFromYAMLWithExtraFiles(t *testing.T) {
	yamlContent := `
participants:
  - el_type: geth
    cl_type: lighthouse
    cl_extra_mounts:
      /configs/beacon.yaml: beacon-config.yaml

extra_files:
  beacon-config.yaml: |
    metrics:
      enabled: true
      port: 8080
  validator.yaml: "graffiti: test"
`

	config, err := FromYAML(yamlContent)
	require.NoError(t, err)

	// Check participants
	assert.Len(t, config.Participants, 1)
	assert.Equal(t, client.Geth, config.Participants[0].ELType)
	assert.Equal(t, client.Lighthouse, config.Participants[0].CLType)

	// Check extra files
	require.NotNil(t, config.ExtraFiles)
	assert.Len(t, config.ExtraFiles, 2)
	assert.Contains(t, config.ExtraFiles["beacon-config.yaml"], "metrics:")
	assert.Equal(t, "graffiti: test", config.ExtraFiles["validator.yaml"])

	// Check extra mounts reference extra files
	assert.Equal(t, "beacon-config.yaml", config.Participants[0].CLExtraMounts["/configs/beacon.yaml"])
}

func TestExtraFilesRoundTrip(t *testing.T) {
	// Create a config with extra files
	original := &EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{
				ELType: client.Geth,
				CLType: client.Prysm,
				Count:  1,
				CLExtraMounts: map[string]string{
					"/etc/config.yaml":   "config.yaml",
					"/etc/features.json": "features.json",
				},
			},
		},
		ExtraFiles: map[string]string{
			"config.yaml":   "key: value\nother: data",
			"features.json": `{"feature1": true, "feature2": false}`,
		},
	}

	// Convert to YAML
	yamlStr, err := ToYAML(original)
	require.NoError(t, err)

	// Parse back from YAML
	parsed, err := FromYAML(yamlStr)
	require.NoError(t, err)

	// Verify extra files match
	require.NotNil(t, parsed.ExtraFiles)
	assert.Equal(t, len(original.ExtraFiles), len(parsed.ExtraFiles))
	for key, value := range original.ExtraFiles {
		assert.Equal(t, value, parsed.ExtraFiles[key])
	}
}

func TestExtraFilesOmitEmpty(t *testing.T) {
	// Create a config without extra files
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

	// Check that extra_files is omitted when empty
	assert.NotContains(t, yamlStr, "extra_files:")
}
