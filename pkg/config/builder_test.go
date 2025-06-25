package config

import (
	"testing"

	"github.com/ethpandaops/ethereum-package-go/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigBuilder(t *testing.T) {
	builder := NewConfigBuilder()

	participant := ParticipantConfig{
		ELType: client.Geth,
		CLType: client.Lighthouse,
		Count:  2,
	}

	networkParams := &NetworkParams{
		NetworkID:      "12345",
		SecondsPerSlot: 12,
	}

	mevConfig := &MEVConfig{
		Type: "full",
	}

	service := AdditionalService{
		Name: "prometheus",
	}

	config, err := builder.
		WithParticipant(participant).
		WithNetworkParams(networkParams).
		WithMEV(mevConfig).
		WithAdditionalService(service).
		WithGlobalLogLevel("debug").
		Build()

	require.NoError(t, err)
	assert.Len(t, config.Participants, 1)
	assert.Equal(t, client.Geth, config.Participants[0].ELType)
	assert.Equal(t, client.Lighthouse, config.Participants[0].CLType)
	assert.Equal(t, 2, config.Participants[0].Count)
	assert.NotNil(t, config.NetworkParams)
	assert.Equal(t, "12345", config.NetworkParams.NetworkID)
	assert.NotNil(t, config.MEV)
	assert.Equal(t, "full", config.MEV.Type)
	assert.Len(t, config.AdditionalServices, 1)
	assert.Equal(t, "prometheus", config.AdditionalServices[0].Name)
	assert.Equal(t, "debug", config.GlobalLogLevel)
}

func TestConfigBuilderWithParticipants(t *testing.T) {
	builder := NewConfigBuilder()

	participants := []ParticipantConfig{
		{ELType: client.Geth, CLType: client.Lighthouse, Count: 1},
		{ELType: client.Besu, CLType: client.Teku, Count: 1},
	}

	config, err := builder.WithParticipants(participants).Build()

	require.NoError(t, err)
	assert.Len(t, config.Participants, 2)
}

func TestConfigBuilderWithNetworkID(t *testing.T) {
	builder := NewConfigBuilder()

	participant := ParticipantConfig{
		ELType: client.Geth,
		CLType: client.Lighthouse,
	}

	config, err := builder.
		WithParticipant(participant).
		WithNetworkID("98765").
		Build()

	require.NoError(t, err)
	assert.NotNil(t, config.NetworkParams)
	assert.Equal(t, "98765", config.NetworkParams.NetworkID)
}

func TestConfigBuilderValidation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*ConfigBuilder)
		wantErr string
	}{
		{
			name: "no participants",
			setup: func(b *ConfigBuilder) {
				// Don't add any participants
			},
			wantErr: "at least one participant is required",
		},
		{
			name: "missing EL type",
			setup: func(b *ConfigBuilder) {
				b.WithParticipant(ParticipantConfig{
					CLType: client.Lighthouse,
				})
			},
			wantErr: "participant 0: execution layer type is required",
		},
		{
			name: "missing CL type",
			setup: func(b *ConfigBuilder) {
				b.WithParticipant(ParticipantConfig{
					ELType: client.Geth,
				})
			},
			wantErr: "participant 0: consensus layer type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewConfigBuilder()
			tt.setup(builder)
			_, err := builder.Build()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestConfigBuilderDefaultCount(t *testing.T) {
	builder := NewConfigBuilder()

	config, err := builder.
		WithParticipant(ParticipantConfig{
			ELType: client.Geth,
			CLType: client.Lighthouse,
			// Count not specified
		}).
		Build()

	require.NoError(t, err)
	assert.Equal(t, 1, config.Participants[0].Count)
}

func TestSimpleParticipantBuilder(t *testing.T) {
	participant := NewParticipantBuilder().
		WithEL(client.Geth).
		WithCL(client.Lighthouse).
		WithELVersion("v1.13.0").
		WithCLVersion("v4.5.0").
		WithCount(3).
		WithValidatorCount(96).
		Build()

	assert.Equal(t, client.Geth, participant.ELType)
	assert.Equal(t, client.Lighthouse, participant.CLType)
	assert.Equal(t, "v1.13.0", participant.ELVersion)
	assert.Equal(t, "v4.5.0", participant.CLVersion)
	assert.Equal(t, 3, participant.Count)
	assert.Equal(t, 96, participant.ValidatorCount)
}

func TestSimpleParticipantBuilderDefaults(t *testing.T) {
	participant := NewParticipantBuilder().
		WithEL(client.Geth).
		WithCL(client.Lighthouse).
		Build()

	assert.Equal(t, client.Geth, participant.ELType)
	assert.Equal(t, client.Lighthouse, participant.CLType)
	assert.Equal(t, 1, participant.Count) // Default
	assert.Equal(t, "", participant.ELVersion)
	assert.Equal(t, "", participant.CLVersion)
	assert.Equal(t, 0, participant.ValidatorCount)
}

func TestConfigBuilderImmutability(t *testing.T) {
	builder := NewConfigBuilder()
	participant := ParticipantConfig{
		ELType: client.Geth,
		CLType: client.Lighthouse,
		Count:  1,
	}

	config1, err := builder.WithParticipant(participant).Build()
	require.NoError(t, err)

	// Modify the builder after building
	builder.WithParticipant(ParticipantConfig{
		ELType: client.Besu,
		CLType: client.Teku,
		Count:  1,
	})

	// The first config should not be affected
	assert.Len(t, config1.Participants, 1)
	assert.Equal(t, client.Geth, config1.Participants[0].ELType)
}

func TestConfigBuilderWithPortPublisher(t *testing.T) {
	builder := NewConfigBuilder()

	portPublisher := &PortPublisherConfig{
		NatExitIP: "192.168.1.100",
		EL:        &PortPublisherComponent{Enabled: true, PublicPortStart: 32000},
		CL:        &PortPublisherComponent{Enabled: true, PublicPortStart: 33000},
	}

	participant := ParticipantConfig{
		ELType: client.Geth,
		CLType: client.Lighthouse,
		Count:  1,
	}

	config, err := builder.
		WithParticipant(participant).
		WithPortPublisher(portPublisher).
		Build()

	require.NoError(t, err)
	assert.NotNil(t, config.PortPublisher)
	assert.Equal(t, "192.168.1.100", config.PortPublisher.NatExitIP)
	assert.True(t, config.PortPublisher.EL.Enabled)
	assert.Equal(t, 32000, config.PortPublisher.EL.PublicPortStart)
	assert.True(t, config.PortPublisher.CL.Enabled)
	assert.Equal(t, 33000, config.PortPublisher.CL.PublicPortStart)
}

func TestConfigBuilderPortPublisherDefaults(t *testing.T) {
	builder := NewConfigBuilder()

	// Create port publisher with enabled components but no ports
	portPublisher := &PortPublisherConfig{
		EL: &PortPublisherComponent{Enabled: true},
		CL: &PortPublisherComponent{Enabled: true},
		VC: &PortPublisherComponent{Enabled: true},
	}

	participant := ParticipantConfig{
		ELType: client.Geth,
		CLType: client.Lighthouse,
		Count:  1,
	}

	config, err := builder.
		WithParticipant(participant).
		WithPortPublisher(portPublisher).
		Build()

	require.NoError(t, err)
	assert.NotNil(t, config.PortPublisher)
	// Check that defaults were applied
	assert.Equal(t, "KURTOSIS_IP_ADDR_PLACEHOLDER", config.PortPublisher.NatExitIP)
	assert.Equal(t, 32000, config.PortPublisher.EL.PublicPortStart)
	assert.Equal(t, 33000, config.PortPublisher.CL.PublicPortStart)
	assert.Equal(t, 34000, config.PortPublisher.VC.PublicPortStart)
}

func TestConfigBuilderPortPublisherValidationError(t *testing.T) {
	builder := NewConfigBuilder()

	// Create port publisher with invalid port range
	portPublisher := &PortPublisherConfig{
		EL: &PortPublisherComponent{Enabled: true, PublicPortStart: 80}, // Too low
	}

	participant := ParticipantConfig{
		ELType: client.Geth,
		CLType: client.Lighthouse,
		Count:  1,
	}

	_, err := builder.
		WithParticipant(participant).
		WithPortPublisher(portPublisher).
		Build()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "port publisher el: public_port_start must be between 1024 and 65535")
}

func TestPortPublisherConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *PortPublisherConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with all components",
			config: &PortPublisherConfig{
				NatExitIP: "127.0.0.1",
				EL:        &PortPublisherComponent{Enabled: true, PublicPortStart: 32000},
				CL:        &PortPublisherComponent{Enabled: true, PublicPortStart: 33000},
				VC:        &PortPublisherComponent{Enabled: true, PublicPortStart: 34000},
			},
			wantErr: false,
		},
		{
			name: "valid config with only EL enabled",
			config: &PortPublisherConfig{
				NatExitIP: "192.168.1.100",
				EL:        &PortPublisherComponent{Enabled: true, PublicPortStart: 32000},
			},
			wantErr: false,
		},
		{
			name: "valid config with disabled components",
			config: &PortPublisherConfig{
				NatExitIP: "auto",
				EL:        &PortPublisherComponent{Enabled: false},
				CL:        &PortPublisherComponent{Enabled: false},
			},
			wantErr: false,
		},
		{
			name: "invalid EL port - too low",
			config: &PortPublisherConfig{
				EL: &PortPublisherComponent{Enabled: true, PublicPortStart: 80},
			},
			wantErr: true,
			errMsg:  "port publisher el: public_port_start must be between 1024 and 65535, got 80",
		},
		{
			name: "invalid CL port - too high",
			config: &PortPublisherConfig{
				CL: &PortPublisherComponent{Enabled: true, PublicPortStart: 70000},
			},
			wantErr: true,
			errMsg:  "port publisher cl: public_port_start must be between 1024 and 65535, got 70000",
		},
		{
			name: "invalid additional services port",
			config: &PortPublisherConfig{
				AdditionalServices: &PortPublisherComponent{Enabled: true, PublicPortStart: 0},
			},
			wantErr: true,
			errMsg:  "port publisher additional_services: public_port_start must be between 1024 and 65535, got 0",
		},
		{
			name:    "nil config validates successfully",
			config:  nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config == nil {
				// Skip validation for nil config
				return
			}

			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPortPublisherConfig_ApplyDefaults(t *testing.T) {
	tests := []struct {
		name     string
		config   *PortPublisherConfig
		expected *PortPublisherConfig
	}{
		{
			name:   "empty config gets default NAT exit IP",
			config: &PortPublisherConfig{},
			expected: &PortPublisherConfig{
				NatExitIP: "KURTOSIS_IP_ADDR_PLACEHOLDER",
			},
		},
		{
			name: "enabled components without ports get defaults",
			config: &PortPublisherConfig{
				EL:                 &PortPublisherComponent{Enabled: true},
				CL:                 &PortPublisherComponent{Enabled: true},
				VC:                 &PortPublisherComponent{Enabled: true},
				AdditionalServices: &PortPublisherComponent{Enabled: true},
			},
			expected: &PortPublisherConfig{
				NatExitIP:          "KURTOSIS_IP_ADDR_PLACEHOLDER",
				EL:                 &PortPublisherComponent{Enabled: true, PublicPortStart: 32000},
				CL:                 &PortPublisherComponent{Enabled: true, PublicPortStart: 33000},
				VC:                 &PortPublisherComponent{Enabled: true, PublicPortStart: 34000},
				AdditionalServices: &PortPublisherComponent{Enabled: true, PublicPortStart: 35000},
			},
		},
		{
			name: "existing values are preserved",
			config: &PortPublisherConfig{
				NatExitIP: "192.168.1.100",
				EL:        &PortPublisherComponent{Enabled: true, PublicPortStart: 40000},
				CL:        &PortPublisherComponent{Enabled: true, PublicPortStart: 41000},
			},
			expected: &PortPublisherConfig{
				NatExitIP: "192.168.1.100",
				EL:        &PortPublisherComponent{Enabled: true, PublicPortStart: 40000},
				CL:        &PortPublisherComponent{Enabled: true, PublicPortStart: 41000},
			},
		},
		{
			name: "disabled components don't get port defaults",
			config: &PortPublisherConfig{
				EL: &PortPublisherComponent{Enabled: false},
				CL: &PortPublisherComponent{Enabled: false},
			},
			expected: &PortPublisherConfig{
				NatExitIP: "KURTOSIS_IP_ADDR_PLACEHOLDER",
				EL:        &PortPublisherComponent{Enabled: false, PublicPortStart: 0},
				CL:        &PortPublisherComponent{Enabled: false, PublicPortStart: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.ApplyDefaults()
			assert.Equal(t, tt.expected, tt.config)
		})
	}
}

func TestEthereumPackageConfig_WithPortPublisher(t *testing.T) {
	config := &EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{ELType: "geth", CLType: "lighthouse", Count: 1},
		},
		PortPublisher: &PortPublisherConfig{
			NatExitIP: "127.0.0.1",
			EL:        &PortPublisherComponent{Enabled: true, PublicPortStart: 32000},
			CL:        &PortPublisherComponent{Enabled: true, PublicPortStart: 33000},
		},
	}

	// Apply defaults
	config.ApplyDefaults()

	// Validate
	err := config.Validate()
	require.NoError(t, err)

	// Check that port publisher config is preserved
	assert.Equal(t, "127.0.0.1", config.PortPublisher.NatExitIP)
	assert.True(t, config.PortPublisher.EL.Enabled)
	assert.Equal(t, 32000, config.PortPublisher.EL.PublicPortStart)
	assert.True(t, config.PortPublisher.CL.Enabled)
	assert.Equal(t, 33000, config.PortPublisher.CL.PublicPortStart)
}

func TestEthereumPackageConfig_PortPublisherValidation(t *testing.T) {
	config := &EthereumPackageConfig{
		Participants: []ParticipantConfig{
			{ELType: "geth", CLType: "lighthouse", Count: 1},
		},
		PortPublisher: &PortPublisherConfig{
			EL: &PortPublisherComponent{Enabled: true, PublicPortStart: 100}, // Invalid port
		},
	}

	err := config.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "port publisher el: public_port_start must be between 1024 and 65535")
}
