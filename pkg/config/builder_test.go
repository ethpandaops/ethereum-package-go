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
