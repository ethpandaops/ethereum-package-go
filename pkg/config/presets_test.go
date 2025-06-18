package config

import (
	"testing"

	"github.com/ethpandaops/ethereum-package-go/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPresetConfig(t *testing.T) {
	tests := []struct {
		name         string
		preset       Preset
		expectErr    bool
		validateFunc func(*testing.T, *EthereumPackageConfig)
	}{
		{
			name:      "all ELs preset",
			preset:    PresetAllELs,
			expectErr: false,
			validateFunc: func(t *testing.T, config *EthereumPackageConfig) {
				assert.Len(t, config.Participants, 5) // All 5 EL clients

				// Check that all use Lighthouse as CL
				for _, p := range config.Participants {
					assert.Equal(t, client.Lighthouse, p.CLType)
					assert.Equal(t, 1, p.Count)
				}

				// Check all different EL clients are present
				elTypes := make(map[client.Type]bool)
				for _, p := range config.Participants {
					elTypes[p.ELType] = true
				}
				assert.True(t, elTypes[client.Geth])
				assert.True(t, elTypes[client.Besu])
				assert.True(t, elTypes[client.Nethermind])
				assert.True(t, elTypes[client.Erigon])
				assert.True(t, elTypes[client.Reth])
			},
		},
		{
			name:      "all CLs preset",
			preset:    PresetAllCLs,
			expectErr: false,
			validateFunc: func(t *testing.T, config *EthereumPackageConfig) {
				assert.Len(t, config.Participants, 6) // All 6 CL clients

				// Check that all use Geth as EL
				for _, p := range config.Participants {
					assert.Equal(t, client.Geth, p.ELType)
					assert.Equal(t, 1, p.Count)
				}

				// Check all different CL clients are present
				clTypes := make(map[client.Type]bool)
				for _, p := range config.Participants {
					clTypes[p.CLType] = true
				}
				assert.True(t, clTypes[client.Lighthouse])
				assert.True(t, clTypes[client.Teku])
				assert.True(t, clTypes[client.Prysm])
				assert.True(t, clTypes[client.Nimbus])
				assert.True(t, clTypes[client.Lodestar])
				assert.True(t, clTypes[client.Grandine])
			},
		},
		{
			name:      "all clients matrix preset",
			preset:    PresetAllClientsMatrix,
			expectErr: false,
			validateFunc: func(t *testing.T, config *EthereumPackageConfig) {
				// 5 EL clients * 6 CL clients = 30 combinations
				assert.Len(t, config.Participants, 30)

				// Check that all combinations are present
				combinations := make(map[string]bool)
				for _, p := range config.Participants {
					key := string(p.ELType) + "-" + string(p.CLType)
					combinations[key] = true
					assert.Equal(t, 1, p.Count)
				}
				assert.Len(t, combinations, 30)
			},
		},
		{
			name:      "minimal preset",
			preset:    PresetMinimal,
			expectErr: false,
			validateFunc: func(t *testing.T, config *EthereumPackageConfig) {
				assert.Len(t, config.Participants, 1)
				assert.Equal(t, client.Geth, config.Participants[0].ELType)
				assert.Equal(t, client.Lighthouse, config.Participants[0].CLType)
				assert.Equal(t, 1, config.Participants[0].Count)
				assert.Equal(t, 32, config.Participants[0].ValidatorCount)
			},
		},
		{
			name:      "invalid preset",
			preset:    Preset("invalid"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := GetPresetConfig(tt.preset)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Equal(t, ErrInvalidPreset, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				if tt.validateFunc != nil {
					tt.validateFunc(t, config)
				}
			}
		})
	}
}

func TestPresetBuilder(t *testing.T) {
	builder, err := NewPresetBuilder(PresetMinimal)
	require.NoError(t, err)

	networkParams := &NetworkParams{
		ChainID:        99999,
		SecondsPerSlot: 6,
	}

	mevConfig := &MEVConfig{
		Type: "mock",
	}

	service := AdditionalService{
		Name: "blockscout",
	}

	config, err := builder.
		WithChainID(12345). // This will be overridden by network params
		WithNetworkParams(networkParams).
		WithMEV(mevConfig).
		WithAdditionalService(service).
		WithGlobalLogLevel("debug").
		Build()

	require.NoError(t, err)

	// Should have minimal preset participants
	assert.Len(t, config.Participants, 1)
	assert.Equal(t, client.Geth, config.Participants[0].ELType)
	assert.Equal(t, client.Lighthouse, config.Participants[0].CLType)

	// Should have custom network params
	require.NotNil(t, config.NetworkParams)
	assert.Equal(t, uint64(99999), config.NetworkParams.ChainID)
	assert.Equal(t, 6, config.NetworkParams.SecondsPerSlot)

	// Should have MEV config
	require.NotNil(t, config.MEV)
	assert.Equal(t, "mock", config.MEV.Type)

	// Should have additional service
	assert.Len(t, config.AdditionalServices, 1)
	assert.Equal(t, "blockscout", config.AdditionalServices[0].Name)

	// Should have global log level
	assert.Equal(t, "debug", config.GlobalClientLogLevel)
}

func TestPresetBuilderInvalidPreset(t *testing.T) {
	_, err := NewPresetBuilder(Preset("invalid"))
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPreset, err)
}

func TestAllPresetsValidYAML(t *testing.T) {
	presets := []Preset{
		PresetAllELs,
		PresetAllCLs,
		PresetAllClientsMatrix,
		PresetMinimal,
	}

	for _, preset := range presets {
		t.Run(string(preset), func(t *testing.T) {
			config, err := GetPresetConfig(preset)
			require.NoError(t, err)

			// Convert to YAML
			yamlStr, err := ToYAML(config)
			require.NoError(t, err)
			assert.NotEmpty(t, yamlStr)

			// Parse back from YAML
			parsed, err := FromYAML(yamlStr)
			require.NoError(t, err)

			// Verify participants match
			assert.Len(t, parsed.Participants, len(config.Participants))
		})
	}
}

func TestPresetConsistency(t *testing.T) {
	// Verify that all presets produce valid configurations
	presets := []Preset{
		PresetAllELs,
		PresetAllCLs,
		PresetAllClientsMatrix,
		PresetMinimal,
	}

	for _, preset := range presets {
		t.Run(string(preset), func(t *testing.T) {
			config, err := GetPresetConfig(preset)
			require.NoError(t, err)

			// Build using ConfigBuilder to ensure validation passes
			builder := NewConfigBuilder()
			builder.WithParticipants(config.Participants)

			_, err = builder.Build()
			assert.NoError(t, err, "preset %s should produce valid configuration", preset)
		})
	}
}
