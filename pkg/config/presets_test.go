package config

import (
	"testing"

	"github.com/ethpandaops/ethereum-package-go/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPresetConfig(t *testing.T) {
	tests := []struct {
		name          string
		preset        types.Preset
		expectErr     bool
		validateFunc  func(*testing.T, *types.EthereumPackageConfig)
	}{
		{
			name:      "all ELs preset",
			preset:    types.PresetAllELs,
			expectErr: false,
			validateFunc: func(t *testing.T, config *types.EthereumPackageConfig) {
				assert.Len(t, config.Participants, 5) // All 5 EL clients
				
				// Check that all use Lighthouse as CL
				for _, p := range config.Participants {
					assert.Equal(t, types.ClientLighthouse, p.CLType)
					assert.Equal(t, 1, p.Count)
				}

				// Check all different EL clients are present
				elTypes := make(map[types.ClientType]bool)
				for _, p := range config.Participants {
					elTypes[p.ELType] = true
				}
				assert.True(t, elTypes[types.ClientGeth])
				assert.True(t, elTypes[types.ClientBesu])
				assert.True(t, elTypes[types.ClientNethermind])
				assert.True(t, elTypes[types.ClientErigon])
				assert.True(t, elTypes[types.ClientReth])
			},
		},
		{
			name:      "all CLs preset",
			preset:    types.PresetAllCLs,
			expectErr: false,
			validateFunc: func(t *testing.T, config *types.EthereumPackageConfig) {
				assert.Len(t, config.Participants, 6) // All 6 CL clients
				
				// Check that all use Geth as EL
				for _, p := range config.Participants {
					assert.Equal(t, types.ClientGeth, p.ELType)
					assert.Equal(t, 1, p.Count)
				}

				// Check all different CL clients are present
				clTypes := make(map[types.ClientType]bool)
				for _, p := range config.Participants {
					clTypes[p.CLType] = true
				}
				assert.True(t, clTypes[types.ClientLighthouse])
				assert.True(t, clTypes[types.ClientTeku])
				assert.True(t, clTypes[types.ClientPrysm])
				assert.True(t, clTypes[types.ClientNimbus])
				assert.True(t, clTypes[types.ClientLodestar])
				assert.True(t, clTypes[types.ClientGrandine])
			},
		},
		{
			name:      "all clients matrix preset",
			preset:    types.PresetAllClientsMatrix,
			expectErr: false,
			validateFunc: func(t *testing.T, config *types.EthereumPackageConfig) {
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
			preset:    types.PresetMinimal,
			expectErr: false,
			validateFunc: func(t *testing.T, config *types.EthereumPackageConfig) {
				assert.Len(t, config.Participants, 1)
				assert.Equal(t, types.ClientGeth, config.Participants[0].ELType)
				assert.Equal(t, types.ClientLighthouse, config.Participants[0].CLType)
				assert.Equal(t, 1, config.Participants[0].Count)
				assert.Equal(t, 32, config.Participants[0].ValidatorCount)
			},
		},
		{
			name:      "invalid preset",
			preset:    types.Preset("invalid"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := GetPresetConfig(tt.preset)
			
			if tt.expectErr {
				assert.Error(t, err)
				assert.Equal(t, types.ErrInvalidPreset, err)
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
	builder, err := NewPresetBuilder(types.PresetMinimal)
	require.NoError(t, err)

	networkParams := &types.NetworkParams{
		ChainID:        99999,
		SecondsPerSlot: 6,
	}

	mevConfig := &types.MEVConfig{
		Type: "mock",
	}

	service := types.AdditionalService{
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
	assert.Equal(t, types.ClientGeth, config.Participants[0].ELType)
	assert.Equal(t, types.ClientLighthouse, config.Participants[0].CLType)
	
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
	_, err := NewPresetBuilder(types.Preset("invalid"))
	assert.Error(t, err)
	assert.Equal(t, types.ErrInvalidPreset, err)
}

func TestAllPresetsValidYAML(t *testing.T) {
	presets := []types.Preset{
		types.PresetAllELs,
		types.PresetAllCLs,
		types.PresetAllClientsMatrix,
		types.PresetMinimal,
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
	presets := []types.Preset{
		types.PresetAllELs,
		types.PresetAllCLs,
		types.PresetAllClientsMatrix,
		types.PresetMinimal,
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