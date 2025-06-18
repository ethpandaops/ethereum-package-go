package ethereum

import (
	"testing"
	"time"

	"github.com/ethpandaops/ethereum-package-go/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithParticipants(t *testing.T) {
	participants := []types.ParticipantConfig{
		{
			ELType: "geth",
			CLType: "lighthouse",
			Count:  2,
		},
		{
			ELType: "besu",
			CLType: "teku",
			Count:  1,
		},
		{
			ELType: "nethermind",
			CLType: "prysm",
			Count:  1,
		},
	}

	cfg := defaultRunConfig()
	opt := WithParticipants(participants)
	opt(cfg)

	require.NotNil(t, cfg.ConfigSource)
	assert.Equal(t, "inline", cfg.ConfigSource.Type())

	inlineSource, ok := cfg.ConfigSource.(*types.InlineConfigSource)
	require.True(t, ok)

	ethConfig := inlineSource.GetConfig()
	require.NotNil(t, ethConfig)
	assert.Len(t, ethConfig.Participants, 3)
	assert.Equal(t, participants, ethConfig.Participants)
}

func TestWithCustomChain(t *testing.T) {
	cfg := defaultRunConfig()
	
	chainID := uint64(99999)
	secondsPerSlot := 6
	slotsPerEpoch := 16
	
	opt := WithCustomChain(chainID, secondsPerSlot, slotsPerEpoch)
	opt(cfg)

	require.NotNil(t, cfg.NetworkParams)
	assert.Equal(t, chainID, cfg.NetworkParams.ChainID)
	assert.Equal(t, chainID, cfg.NetworkParams.NetworkID)
	assert.Equal(t, secondsPerSlot, cfg.NetworkParams.SecondsPerSlot)
	assert.Equal(t, slotsPerEpoch, cfg.NetworkParams.SlotsPerEpoch)
}

func TestAdvancedConfigurationCombinations(t *testing.T) {
	tests := []struct {
		name     string
		options  []RunOption
		validate func(t *testing.T, cfg *RunConfig)
	}{
		{
			name: "custom participants with MEV",
			options: []RunOption{
				WithParticipants([]types.ParticipantConfig{
					{ELType: "geth", CLType: "lighthouse", Count: 2},
				}),
				WithMEVBoost(),
				WithChainID(12345),
			},
			validate: func(t *testing.T, cfg *RunConfig) {
				// Check participants
				inlineSource, ok := cfg.ConfigSource.(*types.InlineConfigSource)
				require.True(t, ok)
				assert.Len(t, inlineSource.GetConfig().Participants, 1)
				
				// Check MEV
				require.NotNil(t, cfg.MEV)
				assert.Equal(t, "full", cfg.MEV.Type)
				
				// Check chain ID
				assert.Equal(t, uint64(12345), cfg.ChainID)
			},
		},
		{
			name: "custom network params with monitoring",
			options: []RunOption{
				WithCustomChain(88888, 12, 32),
				WithMonitoring(),
				WithGlobalLogLevel("debug"),
			},
			validate: func(t *testing.T, cfg *RunConfig) {
				// Check network params
				require.NotNil(t, cfg.NetworkParams)
				assert.Equal(t, uint64(88888), cfg.NetworkParams.ChainID)
				assert.Equal(t, 12, cfg.NetworkParams.SecondsPerSlot)
				
				// Check monitoring services
				assert.Len(t, cfg.AdditionalServices, 2)
				assert.Equal(t, "prometheus", cfg.AdditionalServices[0].Name)
				assert.Equal(t, "grafana", cfg.AdditionalServices[1].Name)
				
				// Check log level
				assert.Equal(t, "debug", cfg.GlobalLogLevel)
			},
		},
		{
			name: "preset override with participants",
			options: []RunOption{
				AllELs(), // First set a preset
				WithParticipants([]types.ParticipantConfig{ // Then override with participants
					{ELType: "geth", CLType: "lighthouse", Count: 1},
				}),
			},
			validate: func(t *testing.T, cfg *RunConfig) {
				// Participants should override preset
				inlineSource, ok := cfg.ConfigSource.(*types.InlineConfigSource)
				require.True(t, ok)
				assert.Len(t, inlineSource.GetConfig().Participants, 1)
				assert.Equal(t, types.ClientType("geth"), inlineSource.GetConfig().Participants[0].ELType)
			},
		},
		{
			name: "full observability with custom relay",
			options: []RunOption{
				WithFullObservability(),
				WithMEVBoostRelay("http://custom-relay.example.com"),
				WithVerbose(true),
			},
			validate: func(t *testing.T, cfg *RunConfig) {
				// Check all observability services
				assert.Len(t, cfg.AdditionalServices, 3)
				serviceNames := []string{}
				for _, svc := range cfg.AdditionalServices {
					serviceNames = append(serviceNames, svc.Name)
				}
				assert.Contains(t, serviceNames, "prometheus")
				assert.Contains(t, serviceNames, "grafana")
				assert.Contains(t, serviceNames, "blockscout")
				
				// Check MEV relay
				require.NotNil(t, cfg.MEV)
				assert.Equal(t, "http://custom-relay.example.com", cfg.MEV.RelayURL)
				
				// Check verbose
				assert.True(t, cfg.VerboseMode)
			},
		},
		{
			name: "complex configuration",
			options: []RunOption{
				WithParticipants([]types.ParticipantConfig{
					{ELType: "geth", CLType: "lighthouse", Count: 2},
					{ELType: "besu", CLType: "teku", Count: 1},
				}),
				WithCustomChain(77777, 6, 32),
				WithMEVBoost(),
				WithFullObservability(),
				WithTimeout(20 * time.Minute),
				WithParallelism(8),
				WithEnclaveName("test-complex-network"),
			},
			validate: func(t *testing.T, cfg *RunConfig) {
				// Check participants
				inlineSource, ok := cfg.ConfigSource.(*types.InlineConfigSource)
				require.True(t, ok)
				assert.Len(t, inlineSource.GetConfig().Participants, 2)
				
				// Check network params
				require.NotNil(t, cfg.NetworkParams)
				assert.Equal(t, uint64(77777), cfg.NetworkParams.ChainID)
				
				// Check MEV
				assert.NotNil(t, cfg.MEV)
				
				// Check services
				assert.Len(t, cfg.AdditionalServices, 3)
				
				// Check other options
				assert.Equal(t, 20*time.Minute, cfg.Timeout)
				assert.Equal(t, 8, cfg.Parallelism)
				assert.Equal(t, "test-complex-network", cfg.EnclaveName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultRunConfig()
			
			// Apply all options
			for _, opt := range tt.options {
				opt(cfg)
			}
			
			// Validate
			tt.validate(t, cfg)
		})
	}
}

func TestAdditionalServiceWithConfig(t *testing.T) {
	cfg := defaultRunConfig()
	
	// Add service with configuration
	service := types.AdditionalService{
		Name: "prometheus",
		Config: map[string]interface{}{
			"retention": "30d",
			"storage": map[string]interface{}{
				"tsdb": map[string]interface{}{
					"retention.time": "30d",
					"retention.size": "10GB",
				},
			},
		},
	}
	
	opt := WithAdditionalService(service)
	opt(cfg)

	require.Len(t, cfg.AdditionalServices, 1)
	assert.Equal(t, "prometheus", cfg.AdditionalServices[0].Name)
	assert.NotNil(t, cfg.AdditionalServices[0].Config)
	assert.Equal(t, "30d", cfg.AdditionalServices[0].Config["retention"])
	
	storage, ok := cfg.AdditionalServices[0].Config["storage"].(map[string]interface{})
	require.True(t, ok)
	tsdb, ok := storage["tsdb"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "30d", tsdb["retention.time"])
	assert.Equal(t, "10GB", tsdb["retention.size"])
}

func TestNetworkParamsValidation(t *testing.T) {
	tests := []struct {
		name   string
		params *types.NetworkParams
		valid  bool
	}{
		{
			name: "valid params",
			params: &types.NetworkParams{
				ChainID:        12345,
				NetworkID:      12345,
				SecondsPerSlot: 12,
				SlotsPerEpoch:  32,
			},
			valid: true,
		},
		{
			name: "zero chain ID",
			params: &types.NetworkParams{
				ChainID:        0,
				NetworkID:      12345,
				SecondsPerSlot: 12,
				SlotsPerEpoch:  32,
			},
			valid: false,
		},
		{
			name: "invalid seconds per slot",
			params: &types.NetworkParams{
				ChainID:        12345,
				NetworkID:      12345,
				SecondsPerSlot: 0,
				SlotsPerEpoch:  32,
			},
			valid: false,
		},
		{
			name: "invalid slots per epoch",
			params: &types.NetworkParams{
				ChainID:        12345,
				NetworkID:      12345,
				SecondsPerSlot: 12,
				SlotsPerEpoch:  0,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Implement NetworkParams.Validate() method
			t.Skip("NetworkParams.Validate() not implemented yet")
			// err := tt.params.Validate()
			// if tt.valid {
			// 	assert.NoError(t, err)
			// } else {
			// 	assert.Error(t, err)
			// }
		})
	}
}